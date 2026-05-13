package workflow

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/flosch/pongo2/v6"

	"policy-analyzer/ai"
	"policy-analyzer/document"
	"policy-analyzer/logger"
	"policy-analyzer/report"
	"policy-analyzer/textutil"
)

// log 模块级日志记录器
var skillLog = logger.WithModule("workflow.skills")

// ========== AI聊天技能 ==========

// AIChatSkill AI聊天技能
type AIChatSkill struct {
	aiMgr *ai.Manager
}

func (s *AIChatSkill) Name() string {
	return "ai-chat"
}

func (s *AIChatSkill) Description() string {
	return "使用AI模型进行对话，支持流式输出"
}

func (s *AIChatSkill) ValidateConfig(config map[string]interface{}) error {
	if config == nil {
		return fmt.Errorf("配置不能为空")
	}
	if _, ok := config["model"]; !ok {
		return fmt.Errorf("必须指定model参数")
	}
	return nil
}

func (s *AIChatSkill) Execute(ctx context.Context, input SkillInput) (SkillOutput, error) {
	startTime := time.Now()
	output := SkillOutput{
		Data:    make(map[string]interface{}),
		Metrics: StepMetrics{},
	}

	// 获取配置参数
	model := toString(input.Config["model"])
	providerName := toString(input.Config["provider"])
	if providerName == "" {
		providerName = "DeepSeek"
	}

	temperature := 0.7
	switch t := input.Config["temperature"].(type) {
	case float64:
		temperature = t
	case int:
		temperature = float64(t)
	}

	maxTokens := 4096
	switch mt := input.Config["max_tokens"].(type) {
	case int:
		maxTokens = mt
	case float64:
		maxTokens = int(mt)
	}

	// 获取上下文窗口大小
	contextWindow := 0
	switch cw := input.Config["context_window"].(type) {
	case int:
		contextWindow = cw
	case float64:
		contextWindow = int(cw)
	case string:
		// 从模板解析出来的可能是字符串
		if val, err := parseIntFromString(cw); err == nil {
			contextWindow = val
		}
	}

	messages := s.buildMessages(input, contextWindow)
	if len(messages) == 0 {
		return output, fmt.Errorf("未提供可发送给AI的消息内容")
	}

	// 记录请求信息
	skillLog.Info("AIChatSkill开始执行",
		logger.F("provider", providerName),
		logger.F("model", model),
		logger.F("message_count", len(messages)),
		logger.F("last_message_length", len(messages[len(messages)-1].Content)))

	req := ai.ChatRequest{
		Model:       model,
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}

	stream, err := s.aiMgr.StreamChat(ctx, providerName, req)
	if err != nil {
		skillLog.Error("AI请求失败", logger.F("error", err.Error()))
		return output, fmt.Errorf("AI请求失败: %w", err)
	}

	var fullContent string
	var reasoningContent string

	for chunk := range stream {
		if chunk.Error != "" {
			skillLog.Error("AI响应错误", logger.F("error", chunk.Error))
			return output, fmt.Errorf("AI响应错误: %s", chunk.Error)
		}
		if chunk.Done {
			break
		}
		if chunk.Content != "" {
			fullContent += chunk.Content
			if input.Context.EventEmitter != nil {
				input.Context.EventEmitter("ai:chunk", map[string]interface{}{
					"execution_id": input.Context.ExecutionID,
					"step_id":      input.StepID,
					"content":      chunk.Content,
					"index":        chunk.Index,
				})
			}
		}
		if chunk.Reasoning != "" {
			reasoningContent += chunk.Reasoning
		}
	}

	// 过滤思考标签（适配 DeepSeek R1 等思考模型）
	fullContent = filterThinkTags(fullContent)

	// 记录响应信息
	if len(fullContent) == 0 {
		skillLog.Warn("AI返回空内容",
			logger.F("reasoning_length", len(reasoningContent)),
			logger.F("duration", time.Since(startTime).String()))
		// 返回错误而不是空内容
		if len(reasoningContent) > 0 {
			// 某些模型（如 DeepSeek R1）可能只返回 reasoning
			// 过滤思考标签后再使用
			fullContent = filterThinkTags(reasoningContent)
		}
		if len(fullContent) == 0 {
			return output, fmt.Errorf("AI返回空内容，请检查模型配置或减少输入内容长度")
		}
	}

	skillLog.Info("AIChatSkill执行完成",
		logger.F("response_length", len(fullContent)),
		logger.F("duration", time.Since(startTime).String()))

	output.Data["content"] = fullContent
	// 注意：reasoning 内容不存入输出，避免思考链混入报告
	// output.Data["reasoning"] = reasoningContent
	output.Data["model"] = model
	output.Metrics.Duration = time.Since(startTime)
	output.Metrics.APICalls = 1

	if shouldContinueGeneration(input.Config, fullContent, maxTokens) {
		completed, rounds, err := s.completeResponse(ctx, providerName, req, fullContent, 2)
		if err != nil {
			skillLog.Warn("AI续写补全失败", logger.F("step_id", input.StepID), logger.F("error", err.Error()))
		} else {
			fullContent = completed
			output.Data["content"] = fullContent
			output.Metrics.APICalls += rounds
		}
	}

	return output, nil
}

func (s *AIChatSkill) buildMessages(input SkillInput, contextWindow int) []ai.Message {
	messages := make([]ai.Message, 0)

	if msgs, ok := input.Inputs["messages"].([]interface{}); ok {
		for _, msg := range msgs {
			if m, ok := msg.(map[string]interface{}); ok {
				role := toString(m["role"])
				content := toString(m["content"])
				if role != "" && content != "" {
					messages = append(messages, ai.Message{Role: role, Content: content})
				}
			}
		}
	}
	if len(messages) > 0 {
		return messages
	}

	systemPrompt := firstNonEmptyString(toString(input.Inputs["system_prompt"]), toString(input.Config["system_prompt"]))
	prompt := firstNonEmptyString(toString(input.Inputs["prompt"]), toString(input.Inputs["content"]))

	// 根据上下文窗口动态计算最大 prompt 长度
	// 默认假设 1 token ≈ 1.5 字符（中文），上下文窗口默认 32k
	maxPromptChars := 32000 * 1.5 // 默认 32k tokens
	if contextWindow > 0 {
		// 留出 20% 给 system prompt 和响应
		maxPromptChars = float64(contextWindow) * 0.8 * 1.5
	}

	if len(prompt) > int(maxPromptChars) {
		// 智能截断：尝试均衡保留两份文档的内容
		prompt = s.truncatePrompt(prompt, int(maxPromptChars), systemPrompt)
		skillLog.Warn("prompt内容过长，已智能截断",
			logger.F("original_length", len(toString(input.Inputs["prompt"]))),
			logger.F("truncated_length", len(prompt)),
			logger.F("context_window", contextWindow))
	}

	if systemPrompt != "" {
		messages = append(messages, ai.Message{Role: "system", Content: systemPrompt})
	}
	if prompt != "" {
		messages = append(messages, ai.Message{Role: "user", Content: prompt})
	}
	return messages
}

// truncatePrompt 智能截断 prompt，尝试均衡保留两份文档内容
func (s *AIChatSkill) truncatePrompt(prompt string, maxLen int, systemPrompt string) string {
	// 查找旧版政策和新版政策的标记
	oldMarker := "【旧版政策】"
	newMarker := "【新版政策】"
	focusMarker := "【关注领域】"

	oldIdx := strings.Index(prompt, oldMarker)
	newIdx := strings.Index(prompt, newMarker)
	focusIdx := strings.Index(prompt, focusMarker)

	// 如果找不到标记，使用简单截断
	if oldIdx == -1 || newIdx == -1 {
		availableLen := maxLen - len(systemPrompt) - 500
		if availableLen < 1000 {
			availableLen = 1000
		}
		if len(prompt) <= availableLen {
			return prompt
		}
		return prompt[:availableLen] + "\n\n...[内容已截断]..."
	}

	// 提取各部分内容
	prefix := prompt[:oldIdx]
	oldContent := prompt[oldIdx+len(oldMarker) : newIdx]

	var newContent, suffix string
	if focusIdx != -1 {
		newContent = prompt[newIdx+len(newMarker) : focusIdx]
		suffix = prompt[focusIdx:]
	} else {
		newContent = prompt[newIdx+len(newMarker):]
	}

	// 计算可用空间
	availableLen := maxLen - len(prefix) - len(suffix) - len(systemPrompt) - len(oldMarker) - len(newMarker) - 500
	if availableLen < 2000 {
		availableLen = 2000
	}

	// 均衡分配空间给两份文档
	halfLen := availableLen / 2
	oldLen := len(oldContent)
	newLen := len(newContent)

	var truncatedOld, truncatedNew string

	if oldLen > halfLen && newLen > halfLen {
		// 两份都超长，各截一半
		truncatedOld = oldContent[:halfLen] + "\n...[旧版政策内容已截断]..."
		truncatedNew = newContent[:halfLen] + "\n...[新版政策内容已截断]..."
	} else if oldLen > halfLen {
		// 旧版超长，新版不长
		extraSpace := halfLen - newLen
		oldAllowed := halfLen + extraSpace/2
		truncatedOld = oldContent[:oldAllowed] + "\n...[旧版政策内容已截断]..."
		truncatedNew = newContent
	} else if newLen > halfLen {
		// 新版超长，旧版不长
		extraSpace := halfLen - oldLen
		newAllowed := halfLen + extraSpace/2
		truncatedOld = oldContent
		truncatedNew = newContent[:newAllowed] + "\n...[新版政策内容已截断]..."
	} else {
		// 都不超长
		return prompt
	}

	return prefix + oldMarker + truncatedOld + newMarker + truncatedNew + suffix
}

// ========== 文本对比技能 ==========

// TextCompareSkill 文本对比技能
type TextCompareSkill struct{}

func (s *TextCompareSkill) Name() string {
	return "text-compare"
}

func (s *TextCompareSkill) Description() string {
	return "对比两段文本，识别差异"
}

func (s *TextCompareSkill) ValidateConfig(config map[string]interface{}) error {
	return nil
}

func (s *TextCompareSkill) Execute(ctx context.Context, input SkillInput) (SkillOutput, error) {
	output := SkillOutput{
		Data: make(map[string]interface{}),
	}

	text1, ok1 := input.Inputs["text1"].(string)
	text2, ok2 := input.Inputs["text2"].(string)

	if !ok1 || !ok2 {
		return output, fmt.Errorf("必须提供text1和text2参数")
	}

	// 简单的行级对比
	diff := s.compareLines(text1, text2)
	output.Data["diff"] = diff

	return output, nil
}

func (s *TextCompareSkill) compareLines(text1, text2 string) []map[string]interface{} {
	lines1 := strings.Split(text1, "\n")
	lines2 := strings.Split(text2, "\n")

	diff := make([]map[string]interface{}, 0)
	maxLen := max(len(lines1), len(lines2))

	for i := 0; i < maxLen; i++ {
		line1 := ""
		line2 := ""

		if i < len(lines1) {
			line1 = lines1[i]
		}
		if i < len(lines2) {
			line2 = lines2[i]
		}

		if line1 != line2 {
			changeType := "modified"
			if line1 == "" {
				changeType = "added"
			} else if line2 == "" {
				changeType = "removed"
			}

			diff = append(diff, map[string]interface{}{
				"line":        i + 1,
				"type":        changeType,
				"old_content": line1,
				"new_content": line2,
			})
		}
	}

	return diff
}

// ========== 文件读取技能 ==========

// FileReadSkill 文件读取技能
type FileReadSkill struct {
	docSvc *document.Service
}

func (s *FileReadSkill) Name() string {
	return "file-read"
}

func (s *FileReadSkill) Description() string {
	return "读取文档内容"
}

func (s *FileReadSkill) ValidateConfig(config map[string]interface{}) error {
	return nil
}

func (s *FileReadSkill) Execute(ctx context.Context, input SkillInput) (SkillOutput, error) {
	output := SkillOutput{
		Data: make(map[string]interface{}),
	}

	docID, ok := input.Inputs["document_id"].(string)
	skillLog.Info("FileReadSkill执行", logger.F("document_id", docID), logger.F("ok", ok))
	if !ok || docID == "" {
		return output, fmt.Errorf("必须提供document_id参数，当前值: %v", input.Inputs["document_id"])
	}

	// 获取文档内容
	content, err := s.docSvc.GetContent(ctx, docID)
	if err != nil {
		return output, fmt.Errorf("获取文档内容失败: %w", err)
	}

	// 记录日志以便调试
	skillLog.Info("FileReadSkill执行完成",
		logger.F("document_id", docID),
		logger.F("content_length", len(content)))

	// 简化：直接返回内容
	output.Data["content"] = content
	output.Data["document_id"] = docID

	return output, nil
}

// ========== 模板渲染技能 ==========

// TemplateRenderSkill 模板渲染技能
type TemplateRenderSkill struct{}

func (s *TemplateRenderSkill) Name() string {
	return "template-render"
}

func (s *TemplateRenderSkill) Description() string {
	return "使用模板渲染输出内容"
}

func (s *TemplateRenderSkill) ValidateConfig(config map[string]interface{}) error {
	return nil
}

func (s *TemplateRenderSkill) Execute(ctx context.Context, input SkillInput) (SkillOutput, error) {
	output := SkillOutput{
		Data: make(map[string]interface{}),
	}

	// 获取模板内容
	templateContent, ok := input.Config["template"].(string)
	if !ok {
		// 使用默认模板
		templateContent = s.getDefaultReportTemplate()
	}

	// 使用pongo2渲染
	tpl, err := pongo2.FromString(templateContent)
	if err != nil {
		return output, fmt.Errorf("模板解析失败: %w", err)
	}

	// 构建上下文 - 确保数据结构可正确访问
	stepsData := make(map[string]interface{})
	for stepID, result := range input.Context.StepResults {
		stepsData[stepID] = map[string]interface{}{
			"outputs": result.Data,
		}
	}

	data := pongo2.Context{
		"data":      input.Inputs["data"],
		"inputs":    input.Context.Inputs,
		"variables": input.Context.Variables,
		"steps":     stepsData,
	}

	// 渲染
	rendered, err := tpl.Execute(data)
	if err != nil {
		return output, fmt.Errorf("模板渲染失败: %w", err)
	}

	output.Data["rendered"] = rendered

	return output, nil
}

func (s *TemplateRenderSkill) getDefaultReportTemplate() string {
	return `# 政策对比分析报告

## 基本信息
- 旧版政策: {{ data.old_policy }}
- 新版政策: {{ data.new_policy }}
- 分析日期: {{ data.generated_at }}

## 分析结果
{{ steps.compare.outputs.content }}

## 执行摘要
{{ steps.summarize.outputs.summary }}
`
}

// ========== 摘要生成技能 ==========

// SummarizeSkill 摘要生成技能
type SummarizeSkill struct {
	aiMgr *ai.Manager
}

func (s *SummarizeSkill) Name() string {
	return "summarize"
}

func (s *SummarizeSkill) Description() string {
	return "生成内容摘要"
}

func (s *SummarizeSkill) ValidateConfig(config map[string]interface{}) error {
	return nil
}

func (s *SummarizeSkill) Execute(ctx context.Context, input SkillInput) (SkillOutput, error) {
	output := SkillOutput{
		Data: make(map[string]interface{}),
	}

	content, ok := input.Inputs["content"].(string)
	if !ok {
		return output, fmt.Errorf("必须提供content参数")
	}
	content = filterThinkTags(content)

	// 如果内容为空，返回空摘要
	if strings.TrimSpace(content) == "" {
		output.Data["summary"] = "无内容可生成摘要"
		return output, nil
	}

	// 尝试使用 AI 生成摘要
	if s.aiMgr != nil {
		summary, err := s.generateSummaryWithAI(ctx, content, input)
		if err == nil && summary != "" {
			output.Data["summary"] = textutil.NormalizeSummaryText(summary)
			output.Metrics.APICalls = 1
			return output, nil
		}
		// AI 调用失败，降级到简单截断
	}

	// 降级：直接返回前300字作为摘要
	maxLen := 300
	if len(content) > maxLen {
		output.Data["summary"] = textutil.NormalizeSummaryText(content[:maxLen] + "...")
	} else {
		output.Data["summary"] = textutil.NormalizeSummaryText(content)
	}

	return output, nil
}

func (s *SummarizeSkill) generateSummaryWithAI(ctx context.Context, content string, input SkillInput) (string, error) {
	// 从上下文获取 provider 和 model
	provider := "DeepSeek"
	model := "deepseek-chat"

	if p := toString(input.Context.Inputs["provider"]); p != "" {
		provider = p
	}
	if m := toString(input.Context.Inputs["model"]); m != "" {
		model = m
	}

	// 截取内容以避免超出 token 限制
	maxContentLen := 8000
	if len(content) > maxContentLen {
		content = content[:maxContentLen] + "..."
	}

	req := ai.ChatRequest{
		Model: model,
		Messages: []ai.Message{
			{
				Role:    "system",
				Content: "你是一位专业的政策分析师。请为给定的分析内容生成一个简洁的执行摘要（200字以内），突出主要发现和结论。",
			},
			{
				Role:    "user",
				Content: fmt.Sprintf("请为以下内容生成执行摘要：\n\n%s", content),
			},
		},
		Temperature: 0.3,
		MaxTokens:   500,
	}

	resp, err := s.aiMgr.Chat(ctx, provider, req)
	if err != nil {
		return "", err
	}

	summary := textutil.NormalizeSummaryText(resp.Content)
	if shouldContinueGeneration(input.Config, summary, req.MaxTokens) {
		completed, _, err := completeChatResponse(ctx, s.aiMgr, provider, req, summary, 2)
		if err != nil {
			skillLog.Warn("摘要续写补全失败", logger.F("error", err.Error()))
			return summary, nil
		}
		summary = textutil.NormalizeSummaryText(completed)
	}

	return summary, nil
}

// ========== 关键点提取技能 ==========

// ExtractKeyPointsSkill 关键点提取技能
type ExtractKeyPointsSkill struct{}

func (s *ExtractKeyPointsSkill) Name() string {
	return "extract-key"
}

func (s *ExtractKeyPointsSkill) Description() string {
	return "提取文档关键点"
}

func (s *ExtractKeyPointsSkill) ValidateConfig(config map[string]interface{}) error {
	return nil
}

func (s *ExtractKeyPointsSkill) Execute(ctx context.Context, input SkillInput) (SkillOutput, error) {
	output := SkillOutput{
		Data: make(map[string]interface{}),
	}

	// 实际实现会使用AI
	// 这里返回占位数据
	output.Data["key_points"] = []string{
		"关键点1（待AI提取）",
		"关键点2（待AI提取）",
	}

	return output, nil
}

// ========== 报告格式化技能 ==========

// FormatReportSkill 报告格式化技能
type FormatReportSkill struct{}

func (s *FormatReportSkill) Name() string {
	return "format-report"
}

func (s *FormatReportSkill) Description() string {
	return "格式化分析报告"
}

func (s *FormatReportSkill) ValidateConfig(config map[string]interface{}) error {
	return nil
}

func (s *FormatReportSkill) Execute(ctx context.Context, input SkillInput) (SkillOutput, error) {
	output := SkillOutput{
		Data: make(map[string]interface{}),
	}

	var structuredValue interface{}
	if structured := input.Inputs["structured"]; structured != nil {
		structuredValue = structured
		output.Data["report_structured"] = structured
	}

	if structuredReport, err := report.ParseStructuredReport(structuredValue); err == nil && structuredReport != nil {
		output.Data["report_structured"] = structuredReport
		output.Data["report"] = report.RenderStructuredMarkdown(structuredReport, "")
		return output, nil
	}

	// 使用模板渲染
	renderSkill := &TemplateRenderSkill{}
	renderOutput, err := renderSkill.Execute(ctx, input)
	if err != nil {
		return output, err
	}

	output.Data["report"] = filterThinkTags(toString(renderOutput.Data["rendered"]))

	return output, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// parseIntFromString 从字符串解析整数
func parseIntFromString(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("空字符串")
	}
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

// filterThinkTags 过滤思考标签内容
// 适配 DeepSeek R1、QwQ、MiniMax 等思考模型输出的思考链
// 支持多种标签格式：<think＞</think＞、&lt;think&gt;&lt;/think&gt; 等
func filterThinkTags(content string) string {
	originalLen := len(content)
	content = textutil.SanitizeModelOutput(content)

	// 记录过滤效果
	if originalLen != len(content) {
		skillLog.Info("思考链过滤完成",
			logger.F("original_length", originalLen),
			logger.F("filtered_length", len(content)),
			logger.F("removed_bytes", originalLen-len(content)))
	}

	return content
}

func (s *AIChatSkill) completeResponse(ctx context.Context, providerName string, req ai.ChatRequest, content string, maxRounds int) (string, int, error) {
	completed, rounds, err := completeChatResponse(ctx, s.aiMgr, providerName, req, content, maxRounds)
	if err != nil {
		return content, rounds, err
	}
	return completed, rounds, nil
}

func completeChatResponse(ctx context.Context, aiMgr *ai.Manager, providerName string, req ai.ChatRequest, content string, maxRounds int) (string, int, error) {
	if aiMgr == nil {
		return content, 0, nil
	}

	completed := filterThinkTags(content)
	rounds := 0
	for attempt := 0; attempt < maxRounds; attempt++ {
		if !shouldContinueGeneration(nil, completed, req.MaxTokens) {
			break
		}

		messages := append([]ai.Message{}, req.Messages...)
		messages = append(messages,
			ai.Message{Role: "assistant", Content: completed},
			ai.Message{
				Role:    "user",
				Content: "你上一段输出在中途停止了。请从刚才中断的位置继续补完剩余内容，直接续写，不要重复前文，不要另起开头，不要添加“继续如下”之类说明。如果内容其实已经完整，只回复 [DONE]。",
			},
		)

		resp, err := aiMgr.Chat(ctx, providerName, ai.ChatRequest{
			Model:       req.Model,
			Messages:    messages,
			Temperature: req.Temperature,
			MaxTokens:   req.MaxTokens,
		})
		if err != nil {
			return completed, rounds, err
		}

		rounds++
		extra := strings.TrimSpace(filterThinkTags(resp.Content))
		if extra == "" || extra == "[DONE]" {
			break
		}

		completed = mergeContinuation(completed, extra)
	}

	return completed, rounds, nil
}

func shouldContinueGeneration(config map[string]interface{}, content string, maxTokens int) bool {
	content = strings.TrimSpace(content)
	if content == "" {
		return false
	}

	if enabled, ok := config["continue_on_incomplete"].(bool); ok && !enabled {
		return false
	}

	return isLikelyIncompleteContent(content)
}

func isLikelyIncompleteContent(content string) bool {
	content = strings.TrimSpace(content)
	if content == "" {
		return false
	}

	if strings.HasSuffix(content, "[DONE]") {
		return false
	}

	last, _ := utf8.DecodeLastRuneInString(content)
	switch last {
	case '。', '！', '？', '.', '!', '?', ')', '）', ']', '】', '}', '」', '』', '"', '\'', '`':
		return false
	case '：', ':', '；', ';', '，', ',', '、', '（', '(', '-', '*', '#', '《':
		return true
	}

	if strings.HasSuffix(content, "...") || strings.HasSuffix(content, "…") {
		return true
	}

	// 中文与 Markdown 正文若停在普通字词末尾，通常意味着模型被截断。
	return true
}

func mergeContinuation(base, extra string) string {
	base = strings.TrimRight(base, "\r\n\t ")
	extra = strings.TrimLeft(extra, "\r\n\t ")
	if base == "" {
		return extra
	}
	if extra == "" {
		return base
	}

	maxOverlap := min(len(base), len(extra))
	for overlap := maxOverlap; overlap >= 12; overlap-- {
		if strings.HasSuffix(base, extra[:overlap]) {
			return base + extra[overlap:]
		}
	}

	if needContinuationNewline(base, extra) {
		return base + "\n" + extra
	}
	return base + extra
}

func needContinuationNewline(base, extra string) bool {
	last, _ := utf8.DecodeLastRuneInString(base)
	first, _ := utf8.DecodeRuneInString(extra)
	if last == '\n' || first == '\n' {
		return false
	}
	if strings.HasPrefix(extra, "#") || strings.HasPrefix(extra, "-") || strings.HasPrefix(extra, "*") {
		return true
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
