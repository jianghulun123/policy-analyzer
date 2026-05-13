package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"policy-analyzer/logger"
)

var streamLog = logger.WithModule("ai.stream")

// StreamService 流式聊天服务
// 提供前端友好的流式AI对话能力
type StreamService struct {
	manager   *Manager
	docSvc    DocumentProvider
	sessions  map[string]*ChatSession
	mu        sync.RWMutex
	appCtx    context.Context
	emitEvent func(string, interface{}) // Wails事件发射函数
}

// ChatSession 聊天会话
type ChatSession struct {
	ID           string
	Title        string
	Provider     string
	Model        string
	Messages     []Message
	CreatedAt    time.Time
	LastActiveAt time.Time
}

// ChatStartRequest 开始聊天请求
type ChatStartRequest struct {
	Provider     string    `json:"provider"`
	Model        string    `json:"model"`
	SystemPrompt string    `json:"system_prompt,omitempty"`
	Messages     []Message `json:"messages,omitempty"`
}

// ChatMessageRequest 聊天消息请求
type ChatMessageRequest struct {
	SessionID   string  `json:"session_id"`
	Content     string  `json:"content"`
	Temperature float64 `json:"temperature,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Stream      bool    `json:"stream"` // 是否流式输出
}

// ChatEvent 聊天事件（用于前端事件监听）
type ChatEvent struct {
	SessionID   string     `json:"session_id"`
	Type        string     `json:"type"` // start/chunk/done/error
	Content     string     `json:"content,omitempty"`
	Reasoning   string     `json:"reasoning,omitempty"`
	Index       int        `json:"index,omitempty"`
	TotalTokens int        `json:"total_tokens,omitempty"`
	Error       string     `json:"error,omitempty"`
	Citations   []Citation `json:"citations,omitempty"`
}

// NewStreamService 创建流式服务
func NewStreamService(manager *Manager, docSvc DocumentProvider) *StreamService {
	return &StreamService{
		manager:  manager,
		docSvc:   docSvc,
		sessions: make(map[string]*ChatSession),
	}
}

// SetAppContext 设置应用上下文（用于Wails事件）
func (s *StreamService) SetAppContext(ctx context.Context) {
	s.appCtx = ctx
}

// SetEventEmitter 设置事件发射器
func (s *StreamService) SetEventEmitter(emit func(string, interface{})) {
	s.emitEvent = emit
}

// emit 发送事件到前端
func (s *StreamService) emit(eventName string, data interface{}) {
	if s.appCtx != nil {
		// 使用Wails的运行时事件系统
		runtime.EventsEmit(s.appCtx, eventName, data)
	}
}

// CreateSession 创建聊天会话
func (s *StreamService) CreateSession(req ChatStartRequest) (*ChatSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session := &ChatSession{
		ID:           uuid.New().String(),
		Provider:     req.Provider,
		Model:        req.Model,
		Title:        "新对话",
		Messages:     make([]Message, 0),
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}

	// 添加系统提示
	if req.SystemPrompt != "" {
		session.Messages = append(session.Messages, Message{
			Role:    "system",
			Content: req.SystemPrompt,
		})
	}

	// 添加初始消息
	if len(req.Messages) > 0 {
		session.Messages = append(session.Messages, req.Messages...)
	}

	s.sessions[session.ID] = session
	streamLog.Info("创建聊天会话",
		logger.F("session_id", session.ID),
		logger.F("provider", session.Provider),
		logger.F("model", session.Model),
		logger.F("initial_message_count", len(session.Messages)))
	return session, nil
}

// GetSession 获取会话
func (s *StreamService) GetSession(sessionID string) (*ChatSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("会话不存在: %s", sessionID)
	}
	return session, nil
}

// DeleteSession 删除会话
func (s *StreamService) DeleteSession(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sessionID)
	streamLog.Info("删除聊天会话", logger.F("session_id", sessionID))
}

// ListSessions 列出所有会话
func (s *StreamService) ListSessions() []*ChatSession {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make([]*ChatSession, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// SendMessage 发送消息（支持流式和非流式）
func (s *StreamService) SendMessage(ctx context.Context, req ChatMessageRequest) (*ChatResponse, error) {
	session, err := s.GetSession(req.SessionID)
	if err != nil {
		streamLog.ErrorErr("发送聊天消息前获取会话失败", err,
			logger.F("session_id", req.SessionID))
		return nil, err
	}

	streamLog.Info("收到聊天消息发送请求",
		logger.F("session_id", req.SessionID),
		logger.F("provider", session.Provider),
		logger.F("model", session.Model),
		logger.F("stream", req.Stream),
		logger.F("content_length", len(req.Content)),
		logger.F("temperature", req.Temperature),
		logger.F("max_tokens", req.MaxTokens),
		logger.F("history_message_count", len(session.Messages)))

	// 添加用户消息
	userMsg := Message{
		Role:    "user",
		Content: req.Content,
	}
	session.Messages = append(session.Messages, userMsg)
	session.LastActiveAt = time.Now()

	// 构建聊天请求
	chatReq := ChatRequest{
		Model:       session.Model,
		Messages:    session.Messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
	}

	if req.Stream {
		// 流式模式：启动goroutine处理流式响应
		if err := s.streamMessage(ctx, session, chatReq); err != nil {
			streamLog.ErrorErr("启动流式聊天失败", err,
				logger.F("session_id", req.SessionID),
				logger.F("provider", session.Provider),
				logger.F("model", session.Model))
			return nil, err
		}
		return &ChatResponse{Model: session.Model, Content: ""}, nil
	}

	// 非流式模式：直接调用
	return s.syncMessage(ctx, session, chatReq)
}

// syncMessage 同步发送消息
func (s *StreamService) syncMessage(ctx context.Context, session *ChatSession, req ChatRequest) (*ChatResponse, error) {
	resp, err := s.manager.Chat(ctx, session.Provider, req)
	if err != nil {
		streamLog.ErrorErr("同步聊天请求失败", err,
			logger.F("session_id", session.ID),
			logger.F("provider", session.Provider),
			logger.F("model", session.Model))
		return nil, err
	}

	// 添加助手回复到会话
	session.Messages = append(session.Messages, Message{
		Role:    "assistant",
		Content: resp.Content,
	})
	session.LastActiveAt = time.Now()

	streamLog.Info("同步聊天请求成功",
		logger.F("session_id", session.ID),
		logger.F("provider", session.Provider),
		logger.F("model", resp.Model),
		logger.F("response_length", len(resp.Content)),
		logger.F("total_tokens", resp.Usage.TotalTokens))

	return resp, nil
}

// streamMessage 流式发送消息
func (s *StreamService) streamMessage(ctx context.Context, session *ChatSession, req ChatRequest) error {
	streamLog.Info("开始处理流式聊天",
		logger.F("session_id", session.ID),
		logger.F("provider", session.Provider),
		logger.F("model", session.Model),
		logger.F("message_count", len(req.Messages)))

	// 获取流式通道
	stream, err := s.manager.StreamChat(ctx, session.Provider, req)
	if err != nil {
		streamLog.ErrorErr("创建流式聊天通道失败", err,
			logger.F("session_id", session.ID),
			logger.F("provider", session.Provider),
			logger.F("model", session.Model))
		s.emit("chat:error", ChatEvent{
			SessionID: session.ID,
			Type:      "error",
			Error:     err.Error(),
		})
		return err
	}

	// 发送开始事件
	s.emit("chat:start", ChatEvent{
		SessionID: session.ID,
		Type:      "start",
	})
	streamLog.Info("已发送聊天开始事件", logger.F("session_id", session.ID))

	// 在goroutine中处理流式响应
	go func() {
		var fullContent strings.Builder
		var fullReasoning strings.Builder
		chunkCount := 0

		for chunk := range stream {
			if chunk.Error != "" {
				streamLog.Error("流式聊天收到错误分片",
					logger.F("session_id", session.ID),
					logger.F("provider", session.Provider),
					logger.F("model", session.Model),
					logger.F("chunk_count", chunkCount),
					logger.F("error", chunk.Error))
				s.emit("chat:error", ChatEvent{
					SessionID: session.ID,
					Type:      "error",
					Error:     chunk.Error,
				})
				return
			}

			// 累积内容
			if chunk.Content != "" {
				fullContent.WriteString(chunk.Content)
			}
			if chunk.Reasoning != "" {
				fullReasoning.WriteString(chunk.Reasoning)
			}

			chunkCount++
			if chunkCount == 1 {
				streamLog.Info("收到首个聊天事件分片",
					logger.F("session_id", session.ID),
					logger.F("provider", session.Provider),
					logger.F("model", session.Model),
					logger.F("index", chunk.Index),
					logger.F("content_length", len(chunk.Content)),
					logger.F("reasoning_length", len(chunk.Reasoning)))
			}

			// 发送块事件
			event := ChatEvent{
				SessionID: session.ID,
				Type:      "chunk",
				Index:     chunk.Index,
				Content:   chunk.Content,
				Reasoning: chunk.Reasoning,
			}
			s.emit("chat:chunk", event)

			// 如果完成
			if chunk.Done {
				// 添加完整回复到会话
				session.Messages = append(session.Messages, Message{
					Role:    "assistant",
					Content: fullContent.String(),
				})
				session.LastActiveAt = time.Now()

				streamLog.Info("流式聊天完成",
					logger.F("session_id", session.ID),
					logger.F("provider", session.Provider),
					logger.F("model", session.Model),
					logger.F("chunk_count", chunkCount),
					logger.F("response_length", fullContent.Len()),
					logger.F("reasoning_length", fullReasoning.Len()))

				// 发送完成事件
				s.emit("chat:done", ChatEvent{
					SessionID:   session.ID,
					Type:        "done",
					Content:     fullContent.String(),
					Reasoning:   fullReasoning.String(),
					TotalTokens: 0, // TODO: 从响应中提取
				})
			}
		}
	}()

	return nil
}

// StopGeneration 停止生成
func (s *StreamService) StopGeneration(sessionID string) {
	// 取消该会话的流式生成
	// 通过manager取消
	// TODO: 实现取消逻辑
	streamLog.Warn("收到停止生成请求，但尚未实现取消逻辑", logger.F("session_id", sessionID))
}

// RenameSession 重命名会话
func (s *StreamService) RenameSession(sessionID, title string) (*ChatSession, error) {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	title = strings.TrimSpace(title)
	if title == "" {
		return nil, fmt.Errorf("会话标题不能为空")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	session.Title = title
	session.LastActiveAt = time.Now()

	streamLog.Info("重命名聊天会话",
		logger.F("session_id", sessionID),
		logger.F("title", title))
	return session, nil
}

// GetSessionHistory 获取会话历史
func (s *StreamService) GetSessionHistory(sessionID string) ([]Message, error) {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return nil, err
	}
	return session.Messages, nil
}

// ClearSessionHistory 清除会话历史（保留系统提示）
func (s *StreamService) ClearSessionHistory(sessionID string) error {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return err
	}

	// 保留系统提示
	newMessages := make([]Message, 0)
	for _, msg := range session.Messages {
		if msg.Role == "system" {
			newMessages = append(newMessages, msg)
		}
	}
	session.Messages = newMessages
	session.LastActiveAt = time.Now()
	streamLog.Info("已清空聊天历史", logger.F("session_id", sessionID), logger.F("remaining_message_count", len(newMessages)))

	return nil
}

// QuickChat 快速对话（无状态，一次性对话）
func (s *StreamService) QuickChat(ctx context.Context, provider, model string, messages []Message, opts *ChatOptions) (*ChatResponse, error) {
	req := ChatRequest{
		Model:    model,
		Messages: messages,
	}

	if opts != nil {
		req.Temperature = opts.Temperature
		req.MaxTokens = opts.MaxTokens
	}

	streamLog.Info("收到快速对话请求",
		logger.F("provider", provider),
		logger.F("model", model),
		logger.F("message_count", len(messages)),
		logger.F("temperature", req.Temperature),
		logger.F("max_tokens", req.MaxTokens))
	return s.manager.Chat(ctx, provider, req)
}

// QuickStreamChat 快速流式对话
func (s *StreamService) QuickStreamChat(ctx context.Context, provider, model string, messages []Message, opts *ChatOptions, onChunk func(ChatEvent)) error {
	req := ChatRequest{
		Model:    model,
		Messages: messages,
	}

	if opts != nil {
		req.Temperature = opts.Temperature
		req.MaxTokens = opts.MaxTokens
	}

	streamLog.Info("收到快速流式对话请求",
		logger.F("provider", provider),
		logger.F("model", model),
		logger.F("message_count", len(messages)),
		logger.F("temperature", req.Temperature),
		logger.F("max_tokens", req.MaxTokens))

	stream, err := s.manager.StreamChat(ctx, provider, req)
	if err != nil {
		streamLog.ErrorErr("快速流式对话创建失败", err,
			logger.F("provider", provider),
			logger.F("model", model))
		return err
	}

	index := 0
	for chunk := range stream {
		if chunk.Error != "" {
			streamLog.Error("快速流式对话收到错误分片",
				logger.F("provider", provider),
				logger.F("model", model),
				logger.F("chunk_count", index),
				logger.F("error", chunk.Error))
			onChunk(ChatEvent{
				Type:  "error",
				Error: chunk.Error,
			})
			return fmt.Errorf("%s", chunk.Error)
		}

		onChunk(ChatEvent{
			Type:      "chunk",
			Index:     index,
			Content:   chunk.Content,
			Reasoning: chunk.Reasoning,
		})

		if chunk.Done {
			streamLog.Info("快速流式对话完成",
				logger.F("provider", provider),
				logger.F("model", model),
				logger.F("chunk_count", index+1))
			onChunk(ChatEvent{
				Type: "done",
			})
		}
		index++
	}

	return nil
}

// ChatOptions 聊天选项
type ChatOptions struct {
	Temperature float64
	MaxTokens   int
}

// ToJSON 转换为JSON字符串
func (e *ChatEvent) ToJSON() string {
	data, _ := json.Marshal(e)
	return string(data)
}

// ParseMessages 从JSON解析消息列表
func ParseMessages(jsonData string) ([]Message, error) {
	var messages []Message
	err := json.Unmarshal([]byte(jsonData), &messages)
	return messages, err
}
