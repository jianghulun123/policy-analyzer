package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	openai "github.com/sashabaranov/go-openai"

	"policy-analyzer/document"
	"policy-analyzer/logger"
)

var documentQALog = logger.WithModule("ai.document-qa")

// Citation 引用片段
// 用于前端展示回答依据的文档来源
type Citation struct {
	DocumentID   string  `json:"document_id"`
	DocumentName string  `json:"document_name"`
	ChunkIndex   int     `json:"chunk_index"`
	Snippet      string  `json:"snippet"`
	Score        float64 `json:"score"`
}

// DocumentProvider 文档读取能力
type DocumentProvider interface {
	Get(context.Context, string) (*document.Document, error)
	GetContent(context.Context, string) (string, error)
}

// DocumentChatRequest 基于文档的问答请求
type DocumentChatRequest struct {
	SessionID      string   `json:"session_id"`
	DisplayContent string   `json:"display_content"`
	SelectedDocIDs []string `json:"selected_doc_ids"`
	Temperature    float64  `json:"temperature,omitempty"`
	MaxTokens      int      `json:"max_tokens,omitempty"`
	Stream         bool     `json:"stream"`
}

type retrievedChunk struct {
	citation Citation
	content  string
}

type documentSearchArgs struct {
	Queries []string `json:"queries"`
	Limit   int      `json:"limit"`
}

type documentPromptMode string

const (
	documentPromptModeFull      documentPromptMode = "full"
	documentPromptModeRetrieved documentPromptMode = "retrieved"
)

type documentPrompt struct {
	content   string
	citations []Citation
	mode      documentPromptMode
}

type documentContext struct {
	doc      *document.Document
	content  string
	citation Citation
}

func (s *StreamService) SendDocumentMessage(ctx context.Context, req DocumentChatRequest) (*ChatResponse, error) {
	if s.docSvc == nil {
		documentQALog.Error("文档问答服务未初始化", logger.F("session_id", req.SessionID))
		return nil, fmt.Errorf("文档问答服务未初始化")
	}
	if len(req.SelectedDocIDs) == 0 {
		documentQALog.Warn("发送文档问答时未选择文档", logger.F("session_id", req.SessionID))
		return nil, fmt.Errorf("请先选择至少一份文档")
	}

	session, err := s.GetSession(req.SessionID)
	if err != nil {
		documentQALog.ErrorErr("发送文档问答前获取会话失败", err,
			logger.F("session_id", req.SessionID))
		return nil, err
	}

	question := strings.TrimSpace(req.DisplayContent)
	if question == "" {
		documentQALog.Warn("发送文档问答时问题为空", logger.F("session_id", req.SessionID))
		return nil, fmt.Errorf("问题不能为空")
	}

	documentQALog.Info("收到文档问答发送请求",
		logger.F("session_id", req.SessionID),
		logger.F("provider", session.Provider),
		logger.F("model", session.Model),
		logger.F("stream", req.Stream),
		logger.F("selected_doc_count", len(req.SelectedDocIDs)),
		logger.F("question_length", len(question)),
		logger.F("temperature", req.Temperature),
		logger.F("max_tokens", req.MaxTokens))

	userMsg := Message{
		Role:           "user",
		Content:        question,
		DisplayContent: question,
	}
	session.Messages = append(session.Messages, userMsg)
	session.LastActiveAt = time.Now()

	docPrompt, err := s.buildDocumentPrompt(ctx, session.Provider, session.Model, question, req.SelectedDocIDs)
	if err != nil {
		documentQALog.ErrorErr("构建文档问答提示词失败", err,
			logger.F("session_id", req.SessionID),
			logger.F("selected_doc_count", len(req.SelectedDocIDs)),
			logger.F("question_length", len(question)))
		return nil, err
	}

	documentQALog.Info("文档问答提示词构建完成",
		logger.F("session_id", req.SessionID),
		logger.F("citation_count", len(docPrompt.citations)),
		logger.F("prompt_length", len(docPrompt.content)),
		logger.F("prompt_mode", string(docPrompt.mode)))

	modelMessages := cloneMessages(session.Messages)
	modelMessages[len(modelMessages)-1].Content = docPrompt.content
	citations := docPrompt.citations

	if docPrompt.mode != documentPromptModeFull {
		modelMessages, citations = s.runDocumentSearchTools(ctx, session.Provider, session.Model, question, req.SelectedDocIDs, modelMessages, citations)
	}

	chatReq := ChatRequest{
		Model:       session.Model,
		Messages:    modelMessages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
	}

	if req.Stream {
		if err := s.streamDocumentMessage(ctx, session, chatReq, citations); err != nil {
			documentQALog.ErrorErr("启动文档问答流式请求失败", err,
				logger.F("session_id", req.SessionID),
				logger.F("provider", session.Provider),
				logger.F("model", session.Model))
			return nil, err
		}
		return &ChatResponse{Model: session.Model, Content: ""}, nil
	}
	return s.syncDocumentMessage(ctx, session, chatReq, citations)
}

func (s *StreamService) syncDocumentMessage(ctx context.Context, session *ChatSession, req ChatRequest, citations []Citation) (*ChatResponse, error) {
	resp, err := s.manager.Chat(ctx, session.Provider, req)
	if err != nil {
		documentQALog.ErrorErr("文档问答非流式请求失败", err,
			logger.F("session_id", session.ID),
			logger.F("provider", session.Provider),
			logger.F("model", session.Model),
			logger.F("citation_count", len(citations)))
		return nil, err
	}

	session.Messages = append(session.Messages, Message{
		Role:      "assistant",
		Content:   resp.Content,
		Citations: citations,
	})
	session.LastActiveAt = time.Now()
	documentQALog.Info("文档问答非流式请求成功",
		logger.F("session_id", session.ID),
		logger.F("provider", session.Provider),
		logger.F("model", resp.Model),
		logger.F("citation_count", len(citations)),
		logger.F("response_length", len(resp.Content)),
		logger.F("total_tokens", resp.Usage.TotalTokens))
	return resp, nil
}

func (s *StreamService) streamDocumentMessage(ctx context.Context, session *ChatSession, req ChatRequest, citations []Citation) error {
	documentQALog.Info("开始处理文档问答流式请求",
		logger.F("session_id", session.ID),
		logger.F("provider", session.Provider),
		logger.F("model", session.Model),
		logger.F("citation_count", len(citations)),
		logger.F("message_count", len(req.Messages)))

	stream, err := s.manager.StreamChat(ctx, session.Provider, req)
	if err != nil {
		if isStreamUnsupportedError(err) {
			documentQALog.Warn("文档问答流式请求不支持，回退非流式",
				logger.F("session_id", session.ID),
				logger.F("provider", session.Provider),
				logger.F("model", session.Model),
				logger.F("error", err.Error()))
			resp, syncErr := s.syncDocumentMessage(ctx, session, req, citations)
			if syncErr != nil {
				documentQALog.ErrorErr("文档问答回退非流式失败", syncErr,
					logger.F("session_id", session.ID),
					logger.F("provider", session.Provider),
					logger.F("model", session.Model))
				s.emit("chat:error", ChatEvent{SessionID: session.ID, Type: "error", Error: syncErr.Error()})
				return syncErr
			}
			s.emit("chat:done", ChatEvent{
				SessionID:   session.ID,
				Type:        "done",
				Content:     resp.Content,
				TotalTokens: resp.Usage.TotalTokens,
				Citations:   citations,
			})
			documentQALog.Info("文档问答已回退非流式并完成",
				logger.F("session_id", session.ID),
				logger.F("provider", session.Provider),
				logger.F("model", resp.Model),
				logger.F("response_length", len(resp.Content)))
			return nil
		}
		documentQALog.ErrorErr("创建文档问答流式请求失败", err,
			logger.F("session_id", session.ID),
			logger.F("provider", session.Provider),
			logger.F("model", session.Model))
		return err
	}

	s.emit("chat:start", ChatEvent{
		SessionID: session.ID,
		Type:      "start",
		Citations: citations,
	})
	documentQALog.Info("已发送文档问答开始事件",
		logger.F("session_id", session.ID),
		logger.F("citation_count", len(citations)))

	go func() {
		var fullContent strings.Builder
		var fullReasoning strings.Builder
		chunkCount := 0

		for chunk := range stream {
			if chunk.Error != "" {
				documentQALog.Error("文档问答流式收到错误分片",
					logger.F("session_id", session.ID),
					logger.F("provider", session.Provider),
					logger.F("model", session.Model),
					logger.F("chunk_count", chunkCount),
					logger.F("error", chunk.Error))
				s.emit("chat:error", ChatEvent{SessionID: session.ID, Type: "error", Error: chunk.Error})
				return
			}

			if chunk.Content != "" {
				fullContent.WriteString(chunk.Content)
			}
			if chunk.Reasoning != "" {
				fullReasoning.WriteString(chunk.Reasoning)
			}

			chunkCount++
			if chunkCount == 1 {
				documentQALog.Info("收到首个文档问答流式分片",
					logger.F("session_id", session.ID),
					logger.F("provider", session.Provider),
					logger.F("model", session.Model),
					logger.F("index", chunk.Index),
					logger.F("content_length", len(chunk.Content)),
					logger.F("reasoning_length", len(chunk.Reasoning)))
			}

			s.emit("chat:chunk", ChatEvent{
				SessionID: session.ID,
				Type:      "chunk",
				Index:     chunk.Index,
				Content:   chunk.Content,
				Reasoning: chunk.Reasoning,
			})

			if chunk.Done {
				session.Messages = append(session.Messages, Message{
					Role:      "assistant",
					Content:   fullContent.String(),
					Citations: citations,
				})
				session.LastActiveAt = time.Now()

				documentQALog.Info("文档问答流式请求完成",
					logger.F("session_id", session.ID),
					logger.F("provider", session.Provider),
					logger.F("model", session.Model),
					logger.F("chunk_count", chunkCount),
					logger.F("citation_count", len(citations)),
					logger.F("response_length", fullContent.Len()),
					logger.F("reasoning_length", fullReasoning.Len()))

				s.emit("chat:done", ChatEvent{
					SessionID:   session.ID,
					Type:        "done",
					Content:     fullContent.String(),
					Reasoning:   fullReasoning.String(),
					TotalTokens: 0,
					Citations:   citations,
				})
			}
		}
	}()

	return nil
}

func isStreamUnsupportedError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	keywords := []string{
		"stream",
		"streaming",
		"unsupported",
		"not support",
		"does not support",
		"不支持",
	}
	for _, keyword := range keywords {
		if strings.Contains(message, keyword) {
			return true
		}
	}
	return false
}

func (s *StreamService) runDocumentSearchTools(ctx context.Context, provider, model, question string, selectedDocIDs []string, messages []Message, fallbackCitations []Citation) ([]Message, []Citation) {
	toolReq := ChatRequest{
		Model: model,
		Messages: []Message{
			{
				Role: "system",
				Content: strings.Join([]string{
					"你是文档检索规划助手。回答前必须调用 search_documents 工具。",
					"请先把用户问题泛化为 3 到 6 条覆盖不同表述、同义词和关键实体的检索 query。",
					"不要直接回答用户问题，只调用工具。",
				}, "\n"),
			},
			{Role: "user", Content: question},
		},
		Temperature: 0.2,
		MaxTokens:   300,
		Tools:       []openai.Tool{documentSearchTool()},
		ToolChoice: openai.ToolChoice{
			Type: openai.ToolTypeFunction,
			Function: openai.ToolFunction{
				Name: "search_documents",
			},
		},
	}

	resp, err := s.manager.Chat(ctx, provider, toolReq)
	if err != nil || resp == nil || len(resp.ToolCalls) == 0 {
		if err != nil {
			documentQALog.Warn("文档搜索工具调用失败，回退原提示词",
				logger.F("provider", provider),
				logger.F("model", model),
				logger.F("question_length", len(question)),
				logger.F("error", err.Error()))
		}
		return messages, fallbackCitations
	}

	toolMessages := append([]Message(nil), messages[:len(messages)-1]...)
	toolMessages = append(toolMessages, Message{
		Role:    "user",
		Content: question,
	})
	toolMessages = append(toolMessages, Message{
		Role:      "assistant",
		ToolCalls: resp.ToolCalls,
	})

	citations := make([]Citation, 0, len(fallbackCitations))
	for _, toolCall := range resp.ToolCalls {
		if toolCall.Function.Name != "search_documents" {
			continue
		}

		chunks, err := s.executeDocumentSearchTool(ctx, question, selectedDocIDs, toolCall.Function.Arguments)
		if err != nil {
			documentQALog.Warn("执行文档搜索工具失败",
				logger.F("provider", provider),
				logger.F("model", model),
				logger.F("error", err.Error()))
			continue
		}

		toolContent, chunkCitations := formatDocumentSearchToolResult(chunks)
		citations = append(citations, chunkCitations...)
		toolMessages = append(toolMessages, Message{
			Role:       "tool",
			Content:    toolContent,
			ToolCallID: toolCall.ID,
		})
	}

	if len(citations) == 0 {
		return messages, fallbackCitations
	}

	toolMessages = append(toolMessages, Message{
		Role: "user",
		Content: strings.Join([]string{
			"请基于 search_documents 工具返回的文档片段回答用户原问题。",
			"如果片段不足以支持结论，请明确说明，不要编造。",
			"回答时优先总结结论，再给出必要依据。",
			"",
			"## 用户原问题",
			question,
		}, "\n"),
	})

	documentQALog.Info("文档搜索工具调用完成",
		logger.F("provider", provider),
		logger.F("model", model),
		logger.F("tool_call_count", len(resp.ToolCalls)),
		logger.F("citation_count", len(citations)))
	return toolMessages, citations
}

func documentSearchTool() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "search_documents",
			Description: "在用户选中的政策文档中检索相关片段。必须先泛化用户问题，再用多条 queries 搜索。",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"queries": map[string]any{
						"type":        "array",
						"description": "3 到 6 条泛化检索短语，覆盖同义词、实体、政策名称、章节主题等不同角度。",
						"items": map[string]any{
							"type": "string",
						},
					},
					"limit": map[string]any{
						"type":        "integer",
						"description": "最多返回的片段数量，建议 20。",
						"minimum":     1,
						"maximum":     40,
					},
				},
				"required": []string{"queries"},
			},
		},
	}
}

func (s *StreamService) executeDocumentSearchTool(ctx context.Context, question string, selectedDocIDs []string, arguments string) ([]retrievedChunk, error) {
	args := documentSearchArgs{Limit: 20}
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return nil, fmt.Errorf("解析搜索参数失败: %w", err)
	}
	queries := append([]string{question}, args.Queries...)
	queries = append(queries, buildHeuristicQueries(question)...)
	limit := args.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 40 {
		limit = 40
	}
	return s.retrieveRelevantChunksWithLimit(ctx, question, selectedDocIDs, dedupeQueries(queries), limit)
}

func formatDocumentSearchToolResult(chunks []retrievedChunk) (string, []Citation) {
	citations := make([]Citation, 0, len(chunks))
	parts := make([]string, 0, len(chunks))
	for i, chunk := range chunks {
		citations = append(citations, chunk.citation)
		parts = append(parts, strings.Join([]string{
			fmt.Sprintf("## 片段%d", i+1),
			fmt.Sprintf("文档：%s", chunk.citation.DocumentName),
			fmt.Sprintf("片段序号：%d", chunk.citation.ChunkIndex+1),
			fmt.Sprintf("相关性分数：%.2f", chunk.citation.Score),
			"内容：",
			chunk.content,
		}, "\n"))
	}
	return strings.Join(parts, "\n\n"), citations
}
func (s *StreamService) buildDocumentPrompt(ctx context.Context, provider, model, question string, selectedDocIDs []string) (*documentPrompt, error) {
	fullContexts, totalRunes, err := s.loadSelectedDocumentContexts(ctx, selectedDocIDs)
	if err != nil {
		return nil, err
	}
	if len(fullContexts) == 0 {
		return nil, fmt.Errorf("未找到可用的文档内容")
	}

	contextBudget := s.documentContextBudget(provider, model)
	if totalRunes <= contextBudget {
		citations := make([]Citation, 0, len(fullContexts))
		contexts := make([]string, 0, len(fullContexts))
		for i, item := range fullContexts {
			citations = append(citations, item.citation)
			contexts = append(contexts, strings.Join([]string{
				fmt.Sprintf("## 文档%d：%s", i+1, item.doc.Name),
				item.content,
			}, "\n"))
		}

		prompt := strings.Join([]string{
			"请基于下面提供的完整文档内容回答问题。",
			"如果文档内容不足以支持结论，请明确说明，不要编造。",
			"回答时优先总结结论，再给出必要依据。",
			"",
			strings.Join(contexts, "\n\n"),
			"",
			"## 用户问题",
			question,
		}, "\n")

		documentQALog.Info("文档问答使用全文上下文",
			logger.F("provider", provider),
			logger.F("model", model),
			logger.F("doc_count", len(fullContexts)),
			logger.F("total_runes", totalRunes),
			logger.F("context_budget", contextBudget))
		return &documentPrompt{content: prompt, citations: citations, mode: documentPromptModeFull}, nil
	}

	queries := s.buildSearchQueries(ctx, provider, model, question)
	chunks, err := s.retrieveRelevantChunks(ctx, question, selectedDocIDs, queries)
	if err != nil {
		return nil, err
	}
	if len(chunks) == 0 {
		return nil, fmt.Errorf("未找到可用的文档内容")
	}

	citations := make([]Citation, 0, len(chunks))
	contexts := make([]string, 0, len(chunks))
	for i, chunk := range chunks {
		citations = append(citations, chunk.citation)
		contexts = append(contexts, strings.Join([]string{
			fmt.Sprintf("## 引用片段%d", i+1),
			fmt.Sprintf("文档：%s", chunk.citation.DocumentName),
			fmt.Sprintf("片段序号：%d", chunk.citation.ChunkIndex+1),
			"内容：",
			chunk.content,
		}, "\n"))
	}

	prompt := strings.Join([]string{
		"请基于下面提供的文档片段回答问题。",
		"如果文档片段不足以支持结论，请明确说明，不要编造。",
		"回答时优先总结结论，再给出必要依据。",
		"",
		strings.Join(contexts, "\n\n"),
		"",
		"## 用户问题",
		question,
	}, "\n")

	return &documentPrompt{content: prompt, citations: citations, mode: documentPromptModeRetrieved}, nil
}

func (s *StreamService) loadSelectedDocumentContexts(ctx context.Context, selectedDocIDs []string) ([]documentContext, int, error) {
	contexts := make([]documentContext, 0, len(selectedDocIDs))
	totalRunes := 0
	for _, docID := range selectedDocIDs {
		doc, err := s.docSvc.Get(ctx, docID)
		if err != nil {
			return nil, 0, fmt.Errorf("读取文档失败: %w", err)
		}

		content, err := s.docSvc.GetContent(ctx, docID)
		if err != nil {
			return nil, 0, fmt.Errorf("读取文档内容失败: %w", err)
		}
		content = strings.TrimSpace(content)
		if content == "" {
			continue
		}

		contexts = append(contexts, documentContext{
			doc:     doc,
			content: content,
			citation: Citation{
				DocumentID:   doc.ID,
				DocumentName: doc.Name,
				ChunkIndex:   0,
				Snippet:      clipText(content, 180),
				Score:        0,
			},
		})
		totalRunes += utf8.RuneCountInString(content)
	}
	return contexts, totalRunes, nil
}

func (s *StreamService) documentContextBudget(provider, model string) int {
	const defaultBudget = 200000
	info, ok := s.manager.GetModelInfo(provider, model)
	if !ok || info.ContextWindow <= 0 {
		return defaultBudget
	}

	budget := int(float64(info.ContextWindow) * 0.7)
	if budget < 8000 {
		return 8000
	}
	return budget
}

func (s *StreamService) retrieveRelevantChunks(ctx context.Context, question string, selectedDocIDs, queries []string) ([]retrievedChunk, error) {
	return s.retrieveRelevantChunksWithLimit(ctx, question, selectedDocIDs, queries, 6)
}

func (s *StreamService) retrieveRelevantChunksWithLimit(ctx context.Context, question string, selectedDocIDs, queries []string, limit int) ([]retrievedChunk, error) {
	type scoredChunk struct {
		retrievedChunk
		score float64
	}

	scored := make([]scoredChunk, 0, 32)
	for _, docID := range selectedDocIDs {
		doc, err := s.docSvc.Get(ctx, docID)
		if err != nil {
			return nil, fmt.Errorf("读取文档失败: %w", err)
		}

		content, err := s.docSvc.GetContent(ctx, docID)
		if err != nil {
			return nil, fmt.Errorf("读取文档内容失败: %w", err)
		}

		chunks := splitTextIntoChunks(content, 800, 120)
		for idx, chunk := range chunks {
			score := scoreChunk(question, doc.Name, chunk)
			for _, query := range queries {
				score += scoreChunk(query, doc.Name, chunk)
			}
			if score <= 0 {
				continue
			}

			scored = append(scored, scoredChunk{
				retrievedChunk: retrievedChunk{
					citation: Citation{
						DocumentID:   doc.ID,
						DocumentName: doc.Name,
						ChunkIndex:   idx,
						Snippet:      clipText(chunk, 180),
						Score:        score,
					},
					content: chunk,
				},
				score: score,
			})
		}
	}

	if len(scored) == 0 {
		return s.fallbackChunks(ctx, selectedDocIDs, limit)
	}

	sort.SliceStable(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	if limit <= 0 {
		limit = 6
	}
	if len(scored) < limit {
		limit = len(scored)
	}

	results := make([]retrievedChunk, 0, limit)
	seen := make(map[string]struct{}, limit)
	for _, item := range scored {
		key := fmt.Sprintf("%s:%d", item.citation.DocumentID, item.citation.ChunkIndex)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		results = append(results, item.retrievedChunk)
		if len(results) >= limit {
			break
		}
	}

	return results, nil
}

func (s *StreamService) fallbackChunks(ctx context.Context, selectedDocIDs []string, limit int) ([]retrievedChunk, error) {
	if limit <= 0 {
		limit = 3
	}
	results := make([]retrievedChunk, 0, minInt(limit, len(selectedDocIDs)*2))
	for _, docID := range selectedDocIDs {
		doc, err := s.docSvc.Get(ctx, docID)
		if err != nil {
			return nil, err
		}
		content, err := s.docSvc.GetContent(ctx, docID)
		if err != nil {
			return nil, err
		}
		chunks := splitTextIntoChunks(content, 800, 120)
		for idx, chunk := range chunks {
			results = append(results, retrievedChunk{
				citation: Citation{
					DocumentID:   doc.ID,
					DocumentName: doc.Name,
					ChunkIndex:   idx,
					Snippet:      clipText(chunk, 180),
					Score:        0,
				},
				content: chunk,
			})
			if len(results) >= limit {
				return results, nil
			}
		}
	}
	return results, nil
}

func (s *StreamService) buildSearchQueries(ctx context.Context, provider, model, question string) []string {
	queries := []string{question}
	fallback := buildHeuristicQueries(question)

	resp, err := s.manager.Chat(ctx, provider, ChatRequest{
		Model: model,
		Messages: []Message{
			{
				Role:    "system",
				Content: "你是检索查询改写助手。请根据用户问题给出3条简洁检索短语，每行一条，不要编号，不要解释。",
			},
			{
				Role:    "user",
				Content: question,
			},
		},
		Temperature: 0.2,
		MaxTokens:   120,
	})
	if err != nil || resp == nil {
		if err != nil {
			documentQALog.Warn("生成检索查询改写失败，使用启发式查询",
				logger.F("provider", provider),
				logger.F("model", model),
				logger.F("question_length", len(question)),
				logger.F("error", err.Error()))
		}
		return dedupeQueries(append(queries, fallback...))
	}

	for _, line := range strings.Split(resp.Content, "\n") {
		line = strings.TrimSpace(strings.TrimLeft(line, "-1234567890.、)） "))
		if line != "" {
			queries = append(queries, line)
		}
	}
	queries = append(queries, fallback...)
	result := dedupeQueries(queries)
	documentQALog.Info("生成检索查询改写完成",
		logger.F("provider", provider),
		logger.F("model", model),
		logger.F("query_count", len(result)))
	return result
}

func cloneMessages(messages []Message) []Message {
	cloned := make([]Message, len(messages))
	for i, msg := range messages {
		cloned[i] = msg
		if len(msg.Citations) > 0 {
			cloned[i].Citations = append([]Citation(nil), msg.Citations...)
		}
	}
	return cloned
}

func splitTextIntoChunks(text string, size, overlap int) []string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil
	}

	runes := []rune(trimmed)
	if len(runes) <= size {
		return []string{trimmed}
	}

	chunks := make([]string, 0, len(runes)/size+1)
	step := size - overlap
	if step <= 0 {
		step = size
	}
	for start := 0; start < len(runes); start += step {
		end := start + size
		if end > len(runes) {
			end = len(runes)
		}
		chunk := strings.TrimSpace(string(runes[start:end]))
		if chunk != "" {
			chunks = append(chunks, chunk)
		}
		if end == len(runes) {
			break
		}
	}
	return chunks
}

func scoreChunk(query, title, chunk string) float64 {
	terms := extractTerms(query)
	if len(terms) == 0 {
		return 0
	}

	normalizedTitle := normalizeForSearch(title)
	normalizedChunk := normalizeForSearch(chunk)
	score := 0.0
	for _, term := range terms {
		if term == "" {
			continue
		}
		score += float64(strings.Count(normalizedChunk, term)) * 2
		score += float64(strings.Count(normalizedTitle, term)) * 4
	}
	if strings.Contains(normalizedChunk, normalizeForSearch(query)) {
		score += 6
	}
	return score
}

func extractTerms(text string) []string {
	normalized := normalizeForSearch(text)
	parts := strings.FieldsFunc(normalized, func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r))
	})

	terms := make([]string, 0, len(parts)+8)
	for _, part := range parts {
		if utf8.RuneCountInString(part) >= 2 {
			terms = append(terms, part)
		}
	}

	if len(terms) == 0 && utf8.RuneCountInString(normalized) >= 2 {
		runes := []rune(normalized)
		for i := 0; i < len(runes)-1 && i < 10; i++ {
			terms = append(terms, string(runes[i:i+2]))
		}
	}
	return dedupeQueries(terms)
}

func buildHeuristicQueries(question string) []string {
	terms := extractTerms(question)
	queries := make([]string, 0, 4)
	if len(terms) >= 2 {
		queries = append(queries, strings.Join(terms[:minInt(4, len(terms))], " "))
	}
	if len(terms) >= 4 {
		queries = append(queries, strings.Join(terms[:2], " "), strings.Join(terms[2:4], " "))
	}
	return dedupeQueries(queries)
}

func dedupeQueries(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

func normalizeForSearch(text string) string {
	return strings.ToLower(strings.TrimSpace(text))
}

func clipText(text string, limit int) string {
	runes := []rune(strings.TrimSpace(text))
	if len(runes) <= limit {
		return string(runes)
	}
	return string(runes[:limit]) + "..."
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
