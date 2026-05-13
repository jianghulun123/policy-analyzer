package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/flosch/pongo2/v6"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gopkg.in/yaml.v3"

	"policy-analyzer/ai"
	"policy-analyzer/document"
	"policy-analyzer/logger"
	"policy-analyzer/report"
	"policy-analyzer/storage"
	"policy-analyzer/textutil"
)

// log 模块级日志记录器
var log = logger.WithModule("workflow")

// WorkflowInfo 工作流信息（供前端展示）
type WorkflowInfo struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Version     string     `json:"version"`
	Inputs      []InputDef `json:"inputs"`
	Steps       int        `json:"steps"`
}

// InputDef 输入参数定义
type InputDef struct {
	Name        string `json:"name" yaml:"name"`
	Type        string `json:"type" yaml:"type"`
	Required    bool   `json:"required" yaml:"required"`
	Description string `json:"description" yaml:"description"`
}

type WorkflowEditorData struct {
	Info    WorkflowInfo `json:"info"`
	Content string       `json:"content"`
	Source  string       `json:"source"`
}

// ExecutionRecord 执行记录（供前端展示）
type ExecutionRecord struct {
	ID           string     `json:"id"`
	WorkflowID   string     `json:"workflow_id"`
	WorkflowName string     `json:"workflow_name"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	StartedAt    *time.Time `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at"`
	Summary      string     `json:"summary"`
}

// ExecutionResult 执行结果
type ExecutionResult struct {
	ID          string                 `json:"id"`
	WorkflowID  string                 `json:"workflow_id"`
	Status      string                 `json:"status"`
	Steps       []StepResult           `json:"steps,omitempty"`
	Outputs     map[string]interface{} `json:"outputs"`
	StepResults map[string]StepResult  `json:"step_results"`
	Summary     string                 `json:"summary"`
	Metrics     ExecutionMetrics       `json:"metrics"`
	CreatedAt   time.Time              `json:"created_at"`
	CompletedAt time.Time              `json:"completed_at"`
}

// StepResult 步骤结果
type StepResult struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Status   string                 `json:"status"`
	Output   map[string]interface{} `json:"output"`
	Error    string                 `json:"error,omitempty"`
	Duration int                    `json:"duration"` // 毫秒
}

// ExecutionMetrics 执行指标
type ExecutionMetrics struct {
	TokensInput  int     `json:"tokens_input"`
	TokensOutput int     `json:"tokens_output"`
	APICalls     int     `json:"api_calls"`
	CostUSD      float64 `json:"cost_usd"`
	Duration     int     `json:"duration"` // 秒
}

// Skill 工作流技能接口
type Skill interface {
	Name() string
	Description() string
	ValidateConfig(config map[string]interface{}) error
	Execute(ctx context.Context, input SkillInput) (SkillOutput, error)
}

// SkillInput 技能输入
type SkillInput struct {
	Config  map[string]interface{} // 步骤配置
	Inputs  map[string]interface{} // 输入数据
	Context *ExecutionContext      // 执行上下文
	StepID  string                 // 步骤ID
}

// SkillOutput 技能输出
type SkillOutput struct {
	Data      map[string]interface{} // 输出数据
	Artifacts []Artifact             // 产物（文件等）
	Metrics   StepMetrics            // 执行指标
}

// Artifact 执行产物
type Artifact struct {
	Name    string
	Type    string // file/text/markdown
	Path    string
	Content string
}

// StepMetrics 步骤指标
type StepMetrics struct {
	TokensInput  int
	TokensOutput int
	APICalls     int
	CostUSD      float64
	Duration     time.Duration
}

// ExecutionContext 执行上下文
type ExecutionContext struct {
	WorkflowID   string
	ExecutionID  string
	Inputs       map[string]interface{} // 用户输入
	Variables    map[string]interface{}
	StepResults  map[string]SkillOutput
	Cancelled    bool
	EventEmitter func(event string, data interface{})
}

// Engine 工作流引擎
type Engine struct {
	aiMgr            *ai.Manager
	docSvc           *document.Service
	store            *storage.Manager
	skills           map[string]Skill
	workflows        map[string]*WorkflowDefinition
	workflowSources  map[string]string
	templateDirs     []string
	executions       map[string]*ExecutionSession
	completedResults map[string]*ExecutionResult
	mu               sync.RWMutex
	ctx              context.Context
}

// WorkflowDefinition 工作流定义（内部结构）
type WorkflowDefinition struct {
	ID          string                 `yaml:"id"`
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description"`
	Version     string                 `yaml:"version"`
	Inputs      []InputDef             `yaml:"inputs"`
	Variables   map[string]interface{} `yaml:"variables"`
	Steps       []StepDef              `yaml:"steps"`
	Outputs     []OutputDef            `yaml:"outputs"`
}

// StepDef 步骤定义
type StepDef struct {
	ID        string                 `json:"id" yaml:"id"`
	Name      string                 `json:"name" yaml:"name"`
	Type      string                 `json:"type" yaml:"type"`
	DependsOn []string               `json:"depends_on" yaml:"depends_on"`
	Config    map[string]interface{} `json:"config" yaml:"config"`
	Inputs    map[string]interface{} `json:"inputs" yaml:"inputs"`
	Outputs   []OutputMapping        `json:"outputs" yaml:"outputs"`
	OnError   string                 `json:"on_error" yaml:"on_error"`
}

// OutputMapping 输出映射
type OutputMapping struct {
	Name string `json:"name" yaml:"name"`
	From string `json:"from" yaml:"from"`
}

// OutputDef 输出定义
type OutputDef struct {
	Name string `json:"name" yaml:"name"`
	From string `json:"from" yaml:"from"`
}

// ExecutionSession 执行会话
type ExecutionSession struct {
	ID         string
	WorkflowID string
	Status     string
	Context    *ExecutionContext
	CancelFunc context.CancelFunc
	StartTime  time.Time
}

// NewEngine 创建工作流引擎
func NewEngine(aiMgr *ai.Manager, docSvc *document.Service, store *storage.Manager) *Engine {
	return &Engine{
		aiMgr:            aiMgr,
		docSvc:           docSvc,
		store:            store,
		skills:           make(map[string]Skill),
		workflows:        make(map[string]*WorkflowDefinition),
		workflowSources:  make(map[string]string),
		templateDirs:     make([]string, 0),
		executions:       make(map[string]*ExecutionSession),
		completedResults: make(map[string]*ExecutionResult),
	}
}

// RegisterBuiltInSkills 注册内置技能
func (e *Engine) RegisterBuiltInSkills() {
	e.RegisterSkill(&AIChatSkill{aiMgr: e.aiMgr})
	e.RegisterSkill(&TextCompareSkill{})
	e.RegisterSkill(&FileReadSkill{docSvc: e.docSvc})
	e.RegisterSkill(&TemplateRenderSkill{})
	e.RegisterSkill(&SummarizeSkill{aiMgr: e.aiMgr})
	e.RegisterSkill(&ExtractKeyPointsSkill{})
	e.RegisterSkill(&FormatReportSkill{})
}

// RegisterSkill 注册技能
func (e *Engine) RegisterSkill(skill Skill) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.skills[skill.Name()] = skill
}

// LoadWorkflows 加载工作流模板
func (e *Engine) LoadWorkflows(templateDirs []string) error {
	e.mu.Lock()
	e.templateDirs = append([]string(nil), templateDirs...)
	e.mu.Unlock()

	if err := e.reloadWorkflows(); err != nil {
		return err
	}
	return nil
}

func (e *Engine) reloadWorkflows() error {
	e.mu.Lock()
	e.workflows = make(map[string]*WorkflowDefinition)
	e.workflowSources = make(map[string]string)
	templateDirs := append([]string(nil), e.templateDirs...)
	e.mu.Unlock()

	for _, dir := range templateDirs {
		if err := e.loadWorkflowDir(dir); err != nil {
			log.Warn("加载工作流目录失败", logger.F("dir", dir), logger.F("error", err.Error()))
		}
	}

	e.loadBuiltinWorkflows()
	return nil
}

func (e *Engine) loadWorkflowDir(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		path := filepath.Join(dir, file.Name())
		if filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml" {
			content, wf, err := e.parseWorkflowFile(path)
			if err != nil {
				log.Warn("解析工作流失败", logger.F("path", path), logger.F("error", err.Error()))
				continue
			}
			e.mu.Lock()
			e.workflows[wf.ID] = wf
			e.workflowSources[wf.ID] = content
			e.mu.Unlock()
		}
	}

	return nil
}

func (e *Engine) loadBuiltinWorkflows() {
	wf := e.createStandardCompareWorkflow()
	e.mu.Lock()
	defer e.mu.Unlock()
	e.workflows[wf.ID] = wf
	data, err := yaml.Marshal(wf)
	if err != nil {
		log.Warn("序列化内建工作流失败，回退到默认骨架", logger.F("workflow_id", wf.ID), logger.F("error", err.Error()))
		e.workflowSources[wf.ID] = e.defaultWorkflowContent()
		return
	}
	e.workflowSources[wf.ID] = string(data)
}

func (e *Engine) createStandardCompareWorkflow() *WorkflowDefinition {
	return &WorkflowDefinition{
		ID:          "standard-compare",
		Name:        "标准政策对比",
		Description: "对比分析两份政策文件的差异，生成详细报告",
		Version:     "2.0.0",
		Inputs: []InputDef{
			{Name: "old_policy", Type: "document", Required: true, Description: "旧版政策文件"},
			{Name: "new_policy", Type: "document", Required: true, Description: "新版政策文件"},
			{Name: "focus_areas", Type: "array", Required: false, Description: "重点关注领域"},
			{Name: "provider", Type: "string", Required: false, Description: "AI提供商"},
			{Name: "model", Type: "string", Required: false, Description: "AI模型"},
		},
		Variables: map[string]interface{}{
			"system_prompt": "你是一位资深的政策分析专家，擅长识别政策文件的细微变化及其影响。",
		},
		Steps: []StepDef{
			{
				ID:   "read_old",
				Name: "读取旧版政策",
				Type: "file-read",
				Inputs: map[string]interface{}{
					"document_id": "{{ inputs.old_policy }}",
				},
			},
			{
				ID:   "read_new",
				Name: "读取新版政策",
				Type: "file-read",
				Inputs: map[string]interface{}{
					"document_id": "{{ inputs.new_policy }}",
				},
			},
			{
				ID:        "generate_full_report",
				Name:      "生成完整分析报告",
				Type:      "ai-chat",
				DependsOn: []string{"read_old", "read_new"},
				Config: map[string]interface{}{
					"model":                  "{{ inputs.model }}",
					"provider":               "{{ inputs.provider }}",
					"temperature":            0.3,
					"max_tokens":             16000,
					"continue_on_incomplete": true,
					"context_window":         "{{ inputs.context_window }}",
				},
				Inputs: map[string]interface{}{
					"system_prompt": "你是一位资深的政策分析专家，擅长识别政策文件的细微变化及其影响。你擅长使用 Markdown 格式撰写报告，能够在需要时使用 Mermaid 图表语法绘制数据可视化图表。",
					"prompt": `请对比以下两份政策文件，并输出一份“结论优先、信息密度高”的 Markdown 分析报告。

【旧版政策】
{{ steps.read_old.outputs.content }}

【新版政策】
{{ steps.read_new.outputs.content }}

【关注领域】
{{ inputs.focus_areas }}

---

# 输出要求

请直接输出 Markdown 报告，不要输出 JSON，不要解释，不要加代码块外包装。

请尽量按以下结构组织内容：

# 执行摘要

## 关键指标

## 关键变化对比

## 政策重心趋势

## 核心结论

## 行动建议

## 详细解读

## 补充观察

# 内容规范

1. 强调“变了什么、为什么重要、影响谁、如何应对”。
2. 不追求长篇幅，追求高信息密度，避免空话、套话、重复表述。
3. 执行摘要控制在 200-350 字。
4. 关键指标输出 4-6 项即可，优先使用 Markdown 表格，每项一句洞察。
5. 关键变化对比输出 6-10 行，只保留最关键变化，优先使用 Markdown 表格。
6. 政策重心趋势优先使用 Markdown 表格；如确有必要补充图表，只使用 Mermaid 代码块，不要使用其他图表格式。
7. 核心结论与行动建议优先使用 Markdown 列表表达，每条一个要点，便于前端直观展示。
8. 详细解读建议按“总体变化概览 / 关键变化拆解 / 影响评估”分节展开。
9. 补充观察可包含“措辞变化分析 / 排序变化 / 新提法 / 消失提法 / 重心迁移”等内容；信息不明显时可省略，不要为了凑数编造。
10. 所有内容必须基于给定文本，不要虚构政策事实。`,
				},
			},
			{
				ID:        "summarize",
				Name:      "摘要生成",
				Type:      "summarize",
				DependsOn: []string{"generate_full_report"},
				Inputs: map[string]interface{}{
					"content": "{{ steps.generate_full_report.outputs.content }}",
				},
			},
			{
				ID:        "format",
				Name:      "报告格式化",
				Type:      "format-report",
				DependsOn: []string{"generate_full_report", "summarize"},
				Config: map[string]interface{}{
					"template": "# 政策对比分析报告\n\n## 基本信息\n- 旧版政策: {{ data.old_policy }}\n- 新版政策: {{ data.new_policy }}\n- 分析日期: {{ data.generated_at }}\n\n## 执行摘要\n{{ steps.summarize.outputs.summary }}\n\n## 结构化报告\n{{ steps.generate_full_report.outputs.content }}",
				},
				Inputs: map[string]interface{}{
					"structured": "{{ steps.generate_full_report.outputs.content }}",
					"data": map[string]interface{}{
						"old_policy":   "{{ inputs.old_policy }}",
						"new_policy":   "{{ inputs.new_policy }}",
						"generated_at": "{{ inputs.generated_at }}",
					},
				},
			},
		},
		Outputs: []OutputDef{
			{Name: "comparison", From: "steps.generate_full_report.outputs.content"},
			{Name: "summary", From: "steps.summarize.outputs.summary"},
			{Name: "report", From: "steps.format.outputs.report"},
			{Name: "report_structured", From: "steps.format.outputs.report_structured"},
		},
	}
}

func (e *Engine) defaultWorkflowContent() string {
	return `id: custom-workflow
name: 自定义工作流
description: 在桌面端新建的工作流模板
version: 1.0.0
inputs:
  - name: document_id
    type: document
    required: true
    description: 待分析文档
variables: {}
steps:
  - id: read_document
    name: 读取文档
    type: file-read
    inputs:
      document_id: "{{ inputs.document_id }}"
outputs:
  - name: content
    from: steps.read_document.outputs.content
`
}

// parseWorkflowFile 解析工作流文件（简化实现）
func (e *Engine) parseWorkflowFile(path string) (string, *WorkflowDefinition, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", nil, fmt.Errorf("读取工作流文件失败: %w", err)
	}

	content := string(data)
	wf, err := e.parseWorkflowContent(content)
	if err != nil {
		return "", nil, err
	}
	return content, wf, nil
}

func (e *Engine) parseWorkflowContent(content string) (*WorkflowDefinition, error) {
	var wf WorkflowDefinition
	if err := yaml.Unmarshal([]byte(content), &wf); err != nil {
		return nil, fmt.Errorf("解析YAML失败: %w", err)
	}

	if wf.ID == "" {
		return nil, fmt.Errorf("工作流ID不能为空")
	}
	if wf.Name == "" {
		return nil, fmt.Errorf("工作流名称不能为空")
	}
	if len(wf.Steps) == 0 {
		return nil, fmt.Errorf("工作流至少需要一个步骤")
	}
	if wf.Version == "" {
		wf.Version = "1.0.0"
	}
	if wf.Variables == nil {
		wf.Variables = make(map[string]interface{})
	}
	if wf.Inputs == nil {
		wf.Inputs = make([]InputDef, 0)
	}
	if wf.Outputs == nil {
		wf.Outputs = make([]OutputDef, 0)
	}

	stepIDs := make(map[string]struct{}, len(wf.Steps))
	for i := range wf.Steps {
		step := &wf.Steps[i]
		if step.ID == "" {
			return nil, fmt.Errorf("存在未设置ID的步骤")
		}
		if step.Type == "" {
			return nil, fmt.Errorf("步骤 %s 未设置类型", step.ID)
		}
		if _, exists := stepIDs[step.ID]; exists {
			return nil, fmt.Errorf("步骤ID重复: %s", step.ID)
		}
		stepIDs[step.ID] = struct{}{}
		if step.Config == nil {
			step.Config = make(map[string]interface{})
		}
		if step.Inputs == nil {
			step.Inputs = make(map[string]interface{})
		}
	}

	for _, step := range wf.Steps {
		for _, dep := range step.DependsOn {
			if _, exists := stepIDs[dep]; !exists {
				return nil, fmt.Errorf("步骤 %s 依赖不存在的步骤: %s", step.ID, dep)
			}
		}
	}

	return &wf, nil
}

// ListWorkflows 获取工作流列表
func (e *Engine) ListWorkflows() []WorkflowInfo {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]WorkflowInfo, 0, len(e.workflows))
	for _, wf := range e.workflows {
		info := WorkflowInfo{
			ID:          wf.ID,
			Name:        wf.Name,
			Description: wf.Description,
			Version:     wf.Version,
			Inputs:      wf.Inputs,
			Steps:       len(wf.Steps),
		}
		result = append(result, info)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

func (e *Engine) GetWorkflowDefinition(id string) (*WorkflowEditorData, error) {
	e.mu.RLock()
	wf, ok := e.workflows[id]
	content, hasContent := e.workflowSources[id]
	e.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("工作流不存在: %s", id)
	}

	if !hasContent || strings.TrimSpace(content) == "" {
		data, err := yaml.Marshal(wf)
		if err != nil {
			return nil, fmt.Errorf("序列化工作流失败: %w", err)
		}
		content = string(data)
	}

	return &WorkflowEditorData{
		Info: WorkflowInfo{
			ID:          wf.ID,
			Name:        wf.Name,
			Description: wf.Description,
			Version:     wf.Version,
			Inputs:      wf.Inputs,
			Steps:       len(wf.Steps),
		},
		Content: content,
		Source:  id,
	}, nil
}

func (e *Engine) SaveWorkflowDefinition(originalID, content string) (*WorkflowEditorData, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, fmt.Errorf("工作流内容不能为空")
	}

	wf, err := e.parseWorkflowContent(content)
	if err != nil {
		return nil, err
	}

	templateDir, err := e.getPrimaryTemplateDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		return nil, fmt.Errorf("创建模板目录失败: %w", err)
	}

	targetPath := filepath.Join(templateDir, wf.ID+".yaml")
	if originalID != "" && originalID != wf.ID {
		oldPath := filepath.Join(templateDir, originalID+".yaml")
		if oldPath != targetPath {
			_ = os.Remove(oldPath)
		}
	}
	if err := os.WriteFile(targetPath, []byte(content+"\n"), 0644); err != nil {
		return nil, fmt.Errorf("保存工作流失败: %w", err)
	}
	if err := e.reloadWorkflows(); err != nil {
		return nil, err
	}
	return e.GetWorkflowDefinition(wf.ID)
}

func (e *Engine) NewWorkflowDefinition() *WorkflowEditorData {
	content := e.defaultWorkflowContent()
	wf, err := e.parseWorkflowContent(content)
	if err != nil {
		return &WorkflowEditorData{Content: content, Source: "new"}
	}
	return &WorkflowEditorData{
		Info: WorkflowInfo{
			ID:          wf.ID,
			Name:        wf.Name,
			Description: wf.Description,
			Version:     wf.Version,
			Inputs:      wf.Inputs,
			Steps:       len(wf.Steps),
		},
		Content: content,
		Source:  "new",
	}
}

// ExecuteAsync 异步执行工作流
func (e *Engine) ExecuteAsync(ctx context.Context, execID string, workflowID string, inputs map[string]interface{}) {
	e.mu.Lock()

	wf, ok := e.workflows[workflowID]
	if !ok {
		e.mu.Unlock()
		runtime.EventsEmit(ctx, "workflow:error", map[string]interface{}{
			"execution_id": execID,
			"error":        fmt.Sprintf("工作流不存在: %s", workflowID),
		})
		return
	}

	execCtx, cancel := context.WithCancel(ctx)
	workspaceID := e.extractWorkspaceID(inputs)
	preparedInputs := e.prepareExecutionInputs(inputs)
	session := &ExecutionSession{
		ID:         execID,
		WorkflowID: workflowID,
		Status:     "running",
		CancelFunc: cancel,
		StartTime:  time.Now(),
		Context: &ExecutionContext{
			WorkflowID:  workflowID,
			ExecutionID: execID,
			Inputs:      preparedInputs,
			Variables:   cloneMap(wf.Variables),
			StepResults: make(map[string]SkillOutput),
			Cancelled:   false,
			EventEmitter: func(event string, data interface{}) {
				runtime.EventsEmit(ctx, event, data)
			},
		},
	}
	e.executions[execID] = session
	e.mu.Unlock()

	record := &storage.AnalysisRecord{
		ID:          execID,
		Name:        wf.Name,
		WorkflowID:  workflowID,
		Workspace:   workspaceID,
		Status:      "running",
		Config:      e.mustMarshalJSON(preparedInputs),
		DocumentIDs: e.mustMarshalJSON(e.extractDocumentIDs(preparedInputs)),
		StartedAt:   timePtr(session.StartTime),
	}
	if err := e.store.SaveAnalysisRecord(record); err != nil {
		log.Warn("保存分析记录失败", logger.F("error", err.Error()))
	}

	runtime.EventsEmit(ctx, "workflow:start", map[string]interface{}{
		"execution_id":  execID,
		"workflow_id":   workflowID,
		"workflow_name": wf.Name,
	})

	go e.executeWorkflow(execCtx, session, wf, preparedInputs)
}

// executeWorkflow 执行工作流
func (e *Engine) executeWorkflow(ctx context.Context, session *ExecutionSession, wf *WorkflowDefinition, inputs map[string]interface{}) {
	result := &ExecutionResult{
		ID:          session.ID,
		WorkflowID:  wf.ID,
		Status:      "running",
		Steps:       buildStepSkeleton(wf),
		StepResults: make(map[string]StepResult),
		Outputs:     make(map[string]interface{}),
		CreatedAt:   session.StartTime,
	}

	for _, step := range wf.Steps {
		if session.Context.Cancelled {
			result.Status = "cancelled"
			break
		}

		missingDep := ""
		for _, dep := range step.DependsOn {
			if _, ok := session.Context.StepResults[dep]; !ok {
				missingDep = dep
				break
			}
		}
		if missingDep != "" {
			stepResult := StepResult{
				ID:     step.ID,
				Name:   step.Name,
				Status: "skipped",
				Error:  fmt.Sprintf("依赖步骤未完成: %s", missingDep),
			}
			result.StepResults[step.ID] = stepResult
			session.Context.EventEmitter("workflow:step:complete", map[string]interface{}{
				"execution_id": session.ID,
				"step_id":      step.ID,
				"step_result":  stepResult,
			})
			continue
		}

		session.Context.EventEmitter("workflow:step:start", map[string]interface{}{
			"execution_id": session.ID,
			"step_id":      step.ID,
			"step_name":    step.Name,
		})

		startTime := time.Now()
		output, err := e.executeStep(ctx, step, session.Context)

		stepResult := StepResult{
			ID:       step.ID,
			Name:     step.Name,
			Status:   "completed",
			Output:   output.Data,
			Duration: int(time.Since(startTime).Milliseconds()),
		}

		if err != nil {
			stepResult.Status = "failed"
			stepResult.Error = err.Error()
			if step.OnError == "abort" || step.OnError == "" {
				result.StepResults[step.ID] = stepResult
				result.Status = "failed"
				break
			}
		} else {
			session.Context.StepResults[step.ID] = output
		}

		result.StepResults[step.ID] = stepResult
		session.Context.EventEmitter("workflow:step:complete", map[string]interface{}{
			"execution_id": session.ID,
			"step_id":      step.ID,
			"step_result":  stepResult,
		})
	}

	if result.Status == "running" {
		result.Status = "completed"
		for _, outputDef := range wf.Outputs {
			value := e.resolveOutput(outputDef.From, session.Context)
			result.Outputs[outputDef.Name] = value
		}
	}
	if result.Status == "cancelled" {
		result.Summary = "执行已取消"
	}
	if result.Summary == "" {
		result.Summary = e.buildSummary(result)
	}
	result.Metrics = e.collectExecutionMetrics(session.Context, result)
	result.CompletedAt = time.Now()

	e.persistExecutionResult(session, wf, result)

	if result.Status == "failed" {
		session.Context.EventEmitter("workflow:error", map[string]interface{}{
			"execution_id": session.ID,
			"error":        result.Summary,
		})
	}

	session.Context.EventEmitter("workflow:complete", map[string]interface{}{
		"execution_id": session.ID,
		"result":       result,
	})

	e.mu.Lock()
	e.completedResults[session.ID] = result
	delete(e.executions, session.ID)
	e.mu.Unlock()
}

// executeStep 执行单个步骤
func (e *Engine) executeStep(ctx context.Context, step StepDef, execCtx *ExecutionContext) (SkillOutput, error) {
	skill, ok := e.skills[step.Type]
	if !ok {
		return SkillOutput{}, fmt.Errorf("未知技能类型: %s", step.Type)
	}

	if err := skill.ValidateConfig(step.Config); err != nil {
		return SkillOutput{}, err
	}

	resolvedInputs := e.resolveStepInputs(step.Inputs, execCtx)

	// 添加调试日志
	log.Info("executeStep解析后的输入",
		logger.F("step_id", step.ID),
		logger.F("step_type", step.Type),
		logger.F("inputs_keys", getMapKeys(resolvedInputs)))

	input := SkillInput{
		Config:  e.resolveStepInputs(step.Config, execCtx),
		Inputs:  resolvedInputs,
		Context: execCtx,
		StepID:  step.ID,
	}

	output, err := skill.Execute(ctx, input)
	if err != nil {
		log.Error("步骤执行失败",
			logger.F("step_id", step.ID),
			logger.F("error", err.Error()))
	} else {
		log.Info("步骤执行成功",
			logger.F("step_id", step.ID),
			logger.F("output_keys", getMapKeys(output.Data)))
	}

	return output, err
}

// resolveStepInputs 解析步骤输入
func (e *Engine) resolveStepInputs(inputs map[string]interface{}, ctx *ExecutionContext) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range inputs {
		resolved := e.resolveValue(value, ctx)
		result[key] = resolved
	}

	return result
}

// resolveValue 解析值（支持模板表达式）
func (e *Engine) resolveValue(value interface{}, ctx *ExecutionContext) interface{} {
	switch v := value.(type) {
	case string:
		if strings.Contains(v, "{{") && strings.Contains(v, "}}") {
			return e.evaluateTemplate(v, ctx)
		}
		return v
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, val := range v {
			result[k] = e.resolveValue(val, ctx)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, val := range v {
			result[i] = e.resolveValue(val, ctx)
		}
		return result
	default:
		return value
	}
}

// evaluateTemplate 评估模板表达式
func (e *Engine) evaluateTemplate(templateStr string, ctx *ExecutionContext) interface{} {
	tpl, err := pongo2.FromString(templateStr)
	if err != nil {
		log.Warn("模板解析失败", logger.F("template", templateStr), logger.F("error", err.Error()))
		return templateStr
	}

	stepsData := e.buildTemplateSteps(ctx)

	// 详细日志：记录模板上下文
	log.Info("evaluateTemplate上下文",
		logger.F("inputs_keys", getMapKeys(ctx.Inputs)),
		logger.F("steps_keys", getMapKeys(stepsData)),
		logger.F("template_preview", truncateString(templateStr, 100)))

	data := pongo2.Context{
		"inputs":    ctx.Inputs,
		"variables": ctx.Variables,
		"steps":     stepsData,
	}

	result, err := tpl.Execute(data)
	if err != nil {
		log.Warn("模板执行失败", logger.F("template", templateStr), logger.F("error", err.Error()))
		return templateStr
	}

	// 调试日志：记录模板解析结果
	log.Info("模板解析结果", logger.F("result_length", len(result)), logger.F("result_preview", truncateString(result, 200)))

	return result
}

// resolveOutput 解析输出表达式
func (e *Engine) resolveOutput(from string, ctx *ExecutionContext) interface{} {
	from = strings.TrimSpace(from)
	if from == "" {
		return nil
	}
	if strings.Contains(from, "{{") {
		return e.evaluateTemplate(from, ctx)
	}

	parts := strings.Split(from, ".")
	if len(parts) == 0 {
		return nil
	}

	switch parts[0] {
	case "inputs":
		return resolvePath(ctx.Inputs, parts[1:])
	case "variables":
		return resolvePath(ctx.Variables, parts[1:])
	case "steps":
		if len(parts) < 2 {
			return nil
		}
		stepID := parts[1]
		if output, ok := ctx.StepResults[stepID]; ok {
			stepData := map[string]interface{}{
				"outputs": output.Data,
			}
			return resolvePath(stepData, parts[2:])
		}
	}

	return nil
}

// CancelExecution 取消执行
func (e *Engine) CancelExecution(execID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	session, ok := e.executions[execID]
	if !ok {
		return fmt.Errorf("执行会话不存在: %s", execID)
	}

	session.Context.Cancelled = true
	if session.CancelFunc != nil {
		session.CancelFunc()
	}
	session.Status = "cancelled"

	return nil
}

// DeleteExecution 删除执行记录
func (e *Engine) DeleteExecution(execID string) error {
	e.mu.Lock()
	// 如果正在执行，先取消
	if session, ok := e.executions[execID]; ok {
		session.Context.Cancelled = true
		if session.CancelFunc != nil {
			session.CancelFunc()
		}
	}
	// 从内存中删除
	delete(e.executions, execID)
	delete(e.completedResults, execID)
	e.mu.Unlock()

	// 从数据库中删除
	return e.store.DeleteAnalysisRecord(execID)
}

// GetHistory 获取执行历史
func (e *Engine) GetHistory(ctx context.Context, workspaceID string) ([]ExecutionRecord, error) {
	records, err := e.store.ListAnalysisRecords(workspaceID)
	if err != nil {
		return nil, err
	}

	result := make([]ExecutionRecord, 0, len(records))
	for _, record := range records {
		workflowName := record.Name
		if workflowName == "" {
			if wf, ok := e.workflows[record.WorkflowID]; ok {
				workflowName = wf.Name
			}
		}
		result = append(result, ExecutionRecord{
			ID:           record.ID,
			WorkflowID:   record.WorkflowID,
			WorkflowName: workflowName,
			Status:       record.Status,
			CreatedAt:    record.CreatedAt,
			StartedAt:    record.StartedAt,
			CompletedAt:  record.CompletedAt,
			Summary:      record.Summary,
		})
	}
	return result, nil
}

// GetResult 获取执行结果
func (e *Engine) GetResult(execID string) (*ExecutionResult, error) {
	e.mu.RLock()
	if result, ok := e.completedResults[execID]; ok {
		e.mu.RUnlock()
		return sanitizeExecutionResult(result), nil
	}
	if session, ok := e.executions[execID]; ok {
		partial := &ExecutionResult{
			ID:          session.ID,
			WorkflowID:  session.WorkflowID,
			Status:      session.Status,
			Steps:       nil,
			Outputs:     map[string]interface{}{},
			StepResults: make(map[string]StepResult),
			CreatedAt:   session.StartTime,
		}
		if wf, exists := e.workflows[session.WorkflowID]; exists {
			partial.Steps = buildStepSkeleton(wf)
		}
		for stepID, output := range session.Context.StepResults {
			partial.StepResults[stepID] = StepResult{
				ID:       stepID,
				Name:     findStepName(partial.Steps, stepID),
				Status:   "completed",
				Output:   output.Data,
				Error:    "",
				Duration: 0,
			}
		}
		e.mu.RUnlock()
		return sanitizeExecutionResult(partial), nil
	}
	e.mu.RUnlock()

	record, err := e.store.GetAnalysisRecord(execID)
	if err != nil {
		return nil, fmt.Errorf("执行结果不存在: %s", execID)
	}
	if record.ResultPath == "" || !e.store.FileExists(record.ResultPath) {
		return nil, fmt.Errorf("执行结果不存在: %s", execID)
	}
	data, err := e.store.ReadFile(record.ResultPath)
	if err != nil {
		return nil, fmt.Errorf("读取执行结果失败: %w", err)
	}
	var result ExecutionResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("解析执行结果失败: %w", err)
	}
	if len(result.Steps) == 0 {
		e.mu.RLock()
		if wf, exists := e.workflows[result.WorkflowID]; exists {
			result.Steps = buildStepSkeleton(wf)
		}
		e.mu.RUnlock()
	}
	return sanitizeExecutionResult(&result), nil
}

func buildStepSkeleton(wf *WorkflowDefinition) []StepResult {
	if wf == nil || len(wf.Steps) == 0 {
		return nil
	}

	steps := make([]StepResult, 0, len(wf.Steps))
	for _, step := range wf.Steps {
		steps = append(steps, StepResult{
			ID:     step.ID,
			Name:   step.Name,
			Status: "pending",
			Output: map[string]interface{}{},
		})
	}
	return steps
}

func findStepName(steps []StepResult, stepID string) string {
	for _, step := range steps {
		if step.ID == stepID {
			return step.Name
		}
	}
	return ""
}

// ExportReport 导出报告
func (e *Engine) ExportReport(result *ExecutionResult, format string) ([]byte, error) {
	if result == nil {
		return nil, fmt.Errorf("执行结果不能为空")
	}

	structured, _ := report.ParseStructuredReport(result.Outputs["report_structured"])
	summary := textutil.NormalizeSummaryText(result.Summary)
	if structured != nil && structured.ExecutiveSummary != "" {
		summary = structured.ExecutiveSummary
	}
	content := firstNonEmptyString(
		textutil.SanitizeModelOutput(toString(result.Outputs["report"])),
		textutil.SanitizeModelOutput(toString(result.Outputs["comparison"])),
		summary,
	)
	if content == "" {
		content = "暂无可导出的报告内容"
	}

	format = strings.ToLower(format)

	// 创建报告数据
	reportData := report.ReportData{
		Title:      "政策对比分析报告",
		WorkflowID: result.WorkflowID,
		Status:     result.Status,
		Summary:    summary,
		Content:    content,
		Structured: structured,
		Metrics: report.Metrics{
			Duration:     result.Metrics.Duration,
			TokensInput:  result.Metrics.TokensInput,
			TokensOutput: result.Metrics.TokensOutput,
			APICalls:     result.Metrics.APICalls,
		},
		CreatedAt: result.CreatedAt,
	}

	generator := report.NewGenerator("Policy Analyzer")

	switch format {
	case "pdf":
		return generator.GeneratePDF(reportData)
	case "markdown", "md":
		return generator.GenerateMarkdown(reportData)
	case "txt", "text", "":
		return []byte(content), nil
	default:
		return nil, fmt.Errorf("暂不支持的导出格式: %s", format)
	}
}

func cloneMap(source map[string]interface{}) map[string]interface{} {
	if source == nil {
		return make(map[string]interface{})
	}
	cloned := make(map[string]interface{}, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}

func resolvePath(data interface{}, parts []string) interface{} {
	current := data
	for _, part := range parts {
		asMap, ok := current.(map[string]interface{})
		if !ok {
			return nil
		}
		current = asMap[part]
	}
	return current
}

func (e *Engine) buildTemplateSteps(ctx *ExecutionContext) map[string]interface{} {
	steps := make(map[string]interface{}, len(ctx.StepResults))
	for stepID, output := range ctx.StepResults {
		steps[stepID] = map[string]interface{}{
			"outputs": output.Data,
		}
		// 调试日志
		log.Info("buildTemplateSteps步骤数据",
			logger.F("step_id", stepID),
			logger.F("output_keys", getMapKeys(output.Data)))
	}
	return steps
}

// getMapKeys 获取 map 的所有键
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func (e *Engine) prepareExecutionInputs(inputs map[string]interface{}) map[string]interface{} {
	prepared := cloneMap(inputs)
	if _, ok := prepared["generated_at"]; !ok {
		prepared["generated_at"] = time.Now().Format(time.RFC3339)
	}
	if _, ok := prepared["workspace_id"]; !ok {
		prepared["workspace_id"] = e.extractWorkspaceID(inputs)
	}
	return prepared
}

func (e *Engine) extractWorkspaceID(inputs map[string]interface{}) string {
	if workspaceID, ok := inputs["workspace_id"].(string); ok && workspaceID != "" {
		return workspaceID
	}
	return "default"
}

func (e *Engine) extractDocumentIDs(inputs map[string]interface{}) []string {
	docIDs := make([]string, 0, 2)
	for _, key := range []string{"old_policy", "new_policy", "document_id"} {
		if value, ok := inputs[key].(string); ok && value != "" {
			docIDs = append(docIDs, value)
		}
	}
	return docIDs
}

func (e *Engine) mustMarshalJSON(value interface{}) string {
	data, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func (e *Engine) buildSummary(result *ExecutionResult) string {
	if result == nil {
		return ""
	}
	if structured, err := report.ParseStructuredReport(result.Outputs["report_structured"]); err == nil && structured != nil && structured.ExecutiveSummary != "" {
		return structured.ExecutiveSummary
	}
	if summary := textutil.NormalizeSummaryText(toString(result.Outputs["summary"])); summary != "" {
		return summary
	}
	if report := textutil.SanitizeModelOutput(toString(result.Outputs["report"])); report != "" {
		return truncateString(report, 200)
	}
	if comparison := textutil.SanitizeModelOutput(toString(result.Outputs["comparison"])); comparison != "" {
		return truncateString(comparison, 200)
	}
	return fmt.Sprintf("工作流执行状态：%s", result.Status)
}

func (e *Engine) collectExecutionMetrics(ctx *ExecutionContext, result *ExecutionResult) ExecutionMetrics {
	metrics := ExecutionMetrics{}
	for _, output := range ctx.StepResults {
		metrics.TokensInput += output.Metrics.TokensInput
		metrics.TokensOutput += output.Metrics.TokensOutput
		metrics.APICalls += output.Metrics.APICalls
		metrics.CostUSD += output.Metrics.CostUSD
	}
	if !result.CompletedAt.IsZero() {
		metrics.Duration = int(result.CompletedAt.Sub(result.CreatedAt).Seconds())
	}
	return metrics
}

func (e *Engine) persistExecutionResult(session *ExecutionSession, wf *WorkflowDefinition, result *ExecutionResult) {
	workspaceID := e.extractWorkspaceID(session.Context.Inputs)
	resultDir := e.store.GetAnalysisDir(workspaceID)
	resultPath := filepath.Join(resultDir, session.ID+".json")
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Warn("序列化执行结果失败", logger.F("error", err.Error()))
	} else if _, err := e.store.SaveFile(resultDir, session.ID+".json", data); err != nil {
		log.Warn("保存执行结果失败", logger.F("error", err.Error()))
	} else {
		resultPath = filepath.Join(resultDir, session.ID+".json")
	}

	record, err := e.store.GetAnalysisRecord(session.ID)
	if err != nil {
		log.Warn("读取分析记录失败", logger.F("error", err.Error()))
		return
	}
	record.Status = result.Status
	record.Summary = result.Summary
	record.ResultPath = resultPath
	record.CompletedAt = timePtr(result.CompletedAt)
	if record.StartedAt == nil {
		record.StartedAt = timePtr(session.StartTime)
	}
	if err := e.store.UpdateAnalysisRecord(record); err != nil {
		log.Warn("更新分析记录失败", logger.F("error", err.Error()))
	}
	_ = wf
}

func (e *Engine) getPrimaryTemplateDir() (string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if len(e.templateDirs) == 0 || strings.TrimSpace(e.templateDirs[0]) == "" {
		return "", fmt.Errorf("未配置工作流模板目录")
	}
	return e.templateDirs[0], nil
}

func toString(value interface{}) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	default:
		return fmt.Sprintf("%v", value)
	}
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func truncateString(value string, maxLen int) string {
	if len(value) <= maxLen {
		return value
	}
	return value[:maxLen] + "..."
}

func sanitizeExecutionResult(result *ExecutionResult) *ExecutionResult {
	if result == nil {
		return nil
	}

	structured, _ := report.ParseStructuredReport(result.Outputs["report_structured"])
	if structured != nil {
		result.Outputs["report_structured"] = structured
		if structured.ExecutiveSummary != "" {
			result.Summary = structured.ExecutiveSummary
		}
	}
	if summary, ok := result.Outputs["summary"]; ok {
		result.Summary = textutil.NormalizeSummaryText(toString(summary))
	} else {
		result.Summary = textutil.NormalizeSummaryText(result.Summary)
	}
	if result.Outputs == nil {
		return result
	}

	for _, key := range []string{"report", "comparison", "summary"} {
		if raw, ok := result.Outputs[key]; ok {
			if key == "summary" {
				result.Outputs[key] = textutil.NormalizeSummaryText(toString(raw))
				continue
			}
			result.Outputs[key] = textutil.SanitizeModelOutput(toString(raw))
		}
	}
	if structured != nil && structured.ExecutiveSummary != "" {
		result.Summary = structured.ExecutiveSummary
		result.Outputs["summary"] = structured.ExecutiveSummary
	}

	return result
}



