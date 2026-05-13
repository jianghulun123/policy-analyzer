package ai

import (
	"context"
	"fmt"
	"sync"
	"time"

	openai "github.com/sashabaranov/go-openai"

	"policy-analyzer/config"
	"policy-analyzer/logger"
)

var aiLog = logger.WithModule("ai")

// Provider AI提供商接口
type Provider interface {
	Name() string
	Configure(cfg config.ProviderConfig) error
	Validate() error
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
	StreamChat(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error)
	ListModels(ctx context.Context) ([]ModelInfo, error)
	GetModelInfo(model string) (ModelInfo, bool)
	Close()
}

// ProviderInfo 提供商信息（供前端展示）
type ProviderInfo struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	Enabled      bool        `json:"enabled"`
	DefaultModel string      `json:"default_model"`
	Models       []ModelInfo `json:"models"`
}

// ModelInfo 模型信息
type ModelInfo struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	MaxTokens     int    `json:"max_tokens"`
	ContextWindow int    `json:"context_window"`
}

// ChatRequest 聊天请求（OpenAI兼容格式）
type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []Message     `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Tools       []openai.Tool `json:"-"`
	ToolChoice  any           `json:"-"`
}

// Message 聊天消息
type Message struct {
	Role           string            `json:"role"` // system/user/assistant/tool
	Content        string            `json:"content"`
	DisplayContent string            `json:"display_content,omitempty"`
	Citations      []Citation        `json:"citations,omitempty"`
	ToolCalls      []openai.ToolCall `json:"-"`
	ToolCallID     string            `json:"-"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	ID        string            `json:"id"`
	Model     string            `json:"model"`
	Content   string            `json:"content"`
	Usage     Usage             `json:"usage"`
	ToolCalls []openai.ToolCall `json:"-"`
}

// Usage Token使用量
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// StreamChunk 流式响应块
type StreamChunk struct {
	SessionID string `json:"session_id,omitempty"`
	Index     int    `json:"index"`
	Content   string `json:"content"`
	Reasoning string `json:"reasoning,omitempty"` // DeepSeek R1等模型的推理过程
	Done      bool   `json:"done"`
	Error     string `json:"error,omitempty"`
}

// ========== OpenAI兼容提供商实现 ==========

// OpenAICompatibleProvider OpenAI兼容的提供商实现
// 支持OpenAI、Claude、DeepSeek等兼容OpenAI API的服务
type OpenAICompatibleProvider struct {
	name   string
	config config.ProviderConfig
	client *openai.Client
	mu     sync.Mutex
}

// NewOpenAICompatibleProvider 创建OpenAI兼容提供商
func NewOpenAICompatibleProvider(name string) *OpenAICompatibleProvider {
	return &OpenAICompatibleProvider{name: name}
}

// Name 获取提供商名称
func (p *OpenAICompatibleProvider) Name() string {
	return p.name
}

// Configure 配置提供商
func (p *OpenAICompatibleProvider) Configure(cfg config.ProviderConfig) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.config = cfg

	// 创建客户端配置
	clientConfig := openai.DefaultConfig(cfg.APIKey)
	if cfg.Endpoint != "" {
		clientConfig.BaseURL = cfg.Endpoint
	}

	p.client = openai.NewClientWithConfig(clientConfig)
	return nil
}

// Validate 验证配置
func (p *OpenAICompatibleProvider) Validate() error {
	if p.config.APIKey == "" && p.config.Endpoint != "http://localhost:11434/v1" {
		return fmt.Errorf("API Key未配置")
	}
	if p.config.Endpoint == "" {
		return fmt.Errorf("API Endpoint未配置")
	}
	return nil
}

// Chat 普通聊天（非流式）
func (p *OpenAICompatibleProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	aiLog.Info("发起AI非流式请求",
		logger.F("provider", p.name),
		logger.F("model", req.Model),
		logger.F("endpoint", p.config.Endpoint),
		logger.F("message_count", len(req.Messages)),
		logger.F("temperature", req.Temperature),
		logger.F("max_tokens", req.MaxTokens))

	// 转换消息格式
	messages := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, m := range req.Messages {
		messages[i] = openai.ChatCompletionMessage{
			Role:       m.Role,
			Content:    m.Content,
			ToolCalls:  m.ToolCalls,
			ToolCallID: m.ToolCallID,
		}
	}

	// 创建请求
	chatReq := openai.ChatCompletionRequest{
		Model:       req.Model,
		Messages:    messages,
		Temperature: float32(req.Temperature),
		MaxTokens:   req.MaxTokens,
		Tools:       req.Tools,
		ToolChoice:  req.ToolChoice,
	}

	// 调用API
	resp, err := p.client.CreateChatCompletion(ctx, chatReq)
	if err != nil {
		aiLog.ErrorErr("AI非流式请求失败", err,
			logger.F("provider", p.name),
			logger.F("model", req.Model),
			logger.F("endpoint", p.config.Endpoint),
			logger.F("message_count", len(req.Messages)))
		return nil, fmt.Errorf("API调用失败: %w", err)
	}

	aiLog.Info("AI非流式请求成功",
		logger.F("provider", p.name),
		logger.F("model", resp.Model),
		logger.F("response_id", resp.ID),
		logger.F("prompt_tokens", resp.Usage.PromptTokens),
		logger.F("completion_tokens", resp.Usage.CompletionTokens),
		logger.F("total_tokens", resp.Usage.TotalTokens))

	// 转换响应
	message := resp.Choices[0].Message
	return &ChatResponse{
		ID:        resp.ID,
		Model:     resp.Model,
		Content:   message.Content,
		ToolCalls: message.ToolCalls,
		Usage: Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}, nil
}

// StreamChat 流式聊天
func (p *OpenAICompatibleProvider) StreamChat(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error) {
	p.mu.Lock()
	client := p.client
	endpoint := p.config.Endpoint
	providerName := p.name
	_ = p.config // 保留配置供将来使用
	p.mu.Unlock()

	aiLog.Info("发起AI流式请求",
		logger.F("provider", providerName),
		logger.F("model", req.Model),
		logger.F("endpoint", endpoint),
		logger.F("message_count", len(req.Messages)),
		logger.F("temperature", req.Temperature),
		logger.F("max_tokens", req.MaxTokens))

	// 创建输出通道
	output := make(chan StreamChunk, 100)

	// 转换消息格式
	messages := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, m := range req.Messages {
		messages[i] = openai.ChatCompletionMessage{
			Role:       m.Role,
			Content:    m.Content,
			ToolCalls:  m.ToolCalls,
			ToolCallID: m.ToolCallID,
		}
	}

	// 创建流式请求
	chatReq := openai.ChatCompletionRequest{
		Model:       req.Model,
		Messages:    messages,
		Temperature: float32(req.Temperature),
		MaxTokens:   req.MaxTokens,
		Stream:      true,
	}

	stream, err := client.CreateChatCompletionStream(ctx, chatReq)
	if err != nil {
		aiLog.ErrorErr("创建AI流式请求失败", err,
			logger.F("provider", providerName),
			logger.F("model", req.Model),
			logger.F("endpoint", endpoint),
			logger.F("message_count", len(req.Messages)))
		return nil, fmt.Errorf("创建流式请求失败: %w", err)
	}

	aiLog.Info("AI流式请求已建立",
		logger.F("provider", providerName),
		logger.F("model", req.Model))

	// 启动goroutine处理流式响应
	go func() {
		defer close(output)
		defer stream.Close()

		index := 0
		for {
			response, err := stream.Recv()
			if err != nil {
				if err.Error() == "EOF" {
					aiLog.Info("AI流式响应结束",
						logger.F("provider", providerName),
						logger.F("model", req.Model),
						logger.F("chunk_count", index))
					output <- StreamChunk{Done: true}
					return
				}
				aiLog.ErrorErr("接收AI流式响应失败", err,
					logger.F("provider", providerName),
					logger.F("model", req.Model),
					logger.F("chunk_count", index))
				output <- StreamChunk{Error: err.Error()}
				return
			}

			if len(response.Choices) > 0 {
				chunk := StreamChunk{
					Index:   index,
					Content: response.Choices[0].Delta.Content,
				}

				// DeepSeek R1等模型可能返回推理过程
				if response.Choices[0].Delta.Content == "" && len(response.Choices[0].Delta.ToolCalls) > 0 {
					// 处理特殊响应（如推理过程）
					for _, tc := range response.Choices[0].Delta.ToolCalls {
						if tc.Function.Arguments != "" {
							chunk.Reasoning = tc.Function.Arguments
						}
					}
				}

				if index == 0 {
					aiLog.Info("收到首个AI流式分片",
						logger.F("provider", providerName),
						logger.F("model", req.Model),
						logger.F("has_content", chunk.Content != ""),
						logger.F("has_reasoning", chunk.Reasoning != ""))
				}

				output <- chunk
				index++
			}
		}
	}()

	return output, nil
}

// ListModels 获取模型列表
func (p *OpenAICompatibleProvider) ListModels(ctx context.Context) ([]ModelInfo, error) {
	p.mu.Lock()
	client := p.client
	p.mu.Unlock()

	modelsResp, err := client.ListModels(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取模型列表失败: %w", err)
	}

	models := make([]ModelInfo, 0, len(modelsResp.Models))
	for _, m := range modelsResp.Models {
		models = append(models, ModelInfo{
			ID:   m.ID,
			Name: m.ID,
			// MaxTokens需要根据模型类型设置
		})
	}

	return models, nil
}

// GetModelInfo 获取配置中的模型信息
func (p *OpenAICompatibleProvider) GetModelInfo(model string) (ModelInfo, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, m := range p.config.Models {
		if m.ID == model {
			return ModelInfo{
				ID:            m.ID,
				Name:          m.Name,
				MaxTokens:     m.MaxTokens,
				ContextWindow: m.ContextWindow,
			}, true
		}
	}
	return ModelInfo{}, false
}

// Close 关闭连接
func (p *OpenAICompatibleProvider) Close() {
	// OpenAI客户端无需特殊关闭
}

// ========== AI管理器 ==========

// Manager AI提供商管理器
type Manager struct {
	providers map[string]Provider
	configs   []config.ProviderConfig
	streamMgr *StreamManager
	mu        sync.RWMutex
}

// NewManager 创建AI管理器
func NewManager(configs []config.ProviderConfig) *Manager {
	mgr := &Manager{
		providers: make(map[string]Provider),
		configs:   configs,
		streamMgr: NewStreamManager(),
	}

	// 初始化所有提供商
	for _, cfg := range configs {
		if cfg.Enabled {
			p := NewOpenAICompatibleProvider(cfg.Name)
			p.Configure(cfg)
			mgr.providers[cfg.Name] = p
		}
	}

	return mgr
}

func (m *Manager) UpdateConfigs(configs []config.ProviderConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()

	providers := make(map[string]Provider)
	for _, cfg := range configs {
		if !cfg.Enabled {
			continue
		}

		p, ok := m.providers[cfg.Name]
		if !ok {
			p = NewOpenAICompatibleProvider(cfg.Name)
		}
		if err := p.Configure(cfg); err != nil {
			aiLog.Warn("AI提供商配置更新失败",
				logger.F("provider", cfg.Name),
				logger.F("error", err.Error()))
			continue
		}
		providers[cfg.Name] = p
	}

	for name, p := range m.providers {
		if _, ok := providers[name]; !ok {
			p.Close()
		}
	}

	m.configs = configs
	m.providers = providers
}

func (m *Manager) ListModelsWithConfig(ctx context.Context, cfg config.ProviderConfig) ([]ModelInfo, error) {
	p := NewOpenAICompatibleProvider(cfg.Name)
	if err := p.Configure(cfg); err != nil {
		return nil, err
	}
	return p.ListModels(ctx)
}

func (m *Manager) TestConnectionWithConfig(ctx context.Context, cfg config.ProviderConfig, model string) (*ConnectionTestResult, error) {
	p := NewOpenAICompatibleProvider(cfg.Name)
	if err := p.Configure(cfg); err != nil {
		return nil, err
	}

	req := ChatRequest{
		Model: model,
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
		MaxTokens: 10,
	}

	start := time.Now()
	resp, err := p.Chat(ctx, req)
	if err != nil {
		return &ConnectionTestResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &ConnectionTestResult{
		Success:      true,
		ResponseTime: time.Since(start).Milliseconds(),
		Model:        resp.Model,
		TokensUsed:   resp.Usage.TotalTokens,
		TestMessage:  "连接成功",
	}, nil
}

// ListProviders 获取提供商列表
func (m *Manager) ListProviders() []ProviderInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]ProviderInfo, 0, len(m.configs))
	for _, cfg := range m.configs {
		info := ProviderInfo{
			Name:         cfg.Name,
			Type:         cfg.Type,
			Enabled:      cfg.Enabled,
			DefaultModel: cfg.DefaultModel,
			Models:       make([]ModelInfo, len(cfg.Models)),
		}
		for i, model := range cfg.Models {
			info.Models[i] = ModelInfo{
				ID:            model.ID,
				Name:          model.Name,
				MaxTokens:     model.MaxTokens,
				ContextWindow: model.ContextWindow,
			}
		}
		result = append(result, info)
	}
	return result
}

// GetProvider 获取指定提供商
func (m *Manager) GetProvider(name string) (Provider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, ok := m.providers[name]
	if !ok {
		return nil, fmt.Errorf("提供商不存在: %s", name)
	}
	return p, nil
}

// ListModels 获取指定提供商的模型列表
func (m *Manager) ListModels(ctx context.Context, providerName string) ([]ModelInfo, error) {
	p, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}
	return p.ListModels(ctx)
}

// GetModelInfo 获取指定提供商的模型配置
func (m *Manager) GetModelInfo(providerName, model string) (ModelInfo, bool) {
	p, err := m.GetProvider(providerName)
	if err != nil {
		return ModelInfo{}, false
	}
	return p.GetModelInfo(model)
}

// Chat 聊天（自动选择提供商）
func (m *Manager) Chat(ctx context.Context, providerName string, req ChatRequest) (*ChatResponse, error) {
	p, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}
	return p.Chat(ctx, req)
}

// StreamChat 流式聊天
func (m *Manager) StreamChat(ctx context.Context, providerName string, req ChatRequest) (<-chan StreamChunk, error) {
	p, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}
	return p.StreamChat(ctx, req)
}

// Close 关闭所有连接
func (m *Manager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, p := range m.providers {
		p.Close()
	}
}

// TestConnection 测试模型连通性
func (m *Manager) TestConnection(ctx context.Context, providerName, model string) (*ConnectionTestResult, error) {
	p, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	// 发送一个简单的测试请求
	req := ChatRequest{
		Model: model,
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
		MaxTokens: 10,
	}

	start := time.Now()
	resp, err := p.Chat(ctx, req)
	if err != nil {
		return &ConnectionTestResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &ConnectionTestResult{
		Success:      true,
		ResponseTime: time.Since(start).Milliseconds(),
		Model:        resp.Model,
		TokensUsed:   resp.Usage.TotalTokens,
		TestMessage:  "连接成功",
	}, nil
}

// ConnectionTestResult 连接测试结果
type ConnectionTestResult struct {
	Success      bool   `json:"success"`
	Error        string `json:"error,omitempty"`
	ResponseTime int64  `json:"response_time,omitempty"` // 毫秒
	Model        string `json:"model,omitempty"`
	TokensUsed   int    `json:"tokens_used,omitempty"`
	TestMessage  string `json:"test_message,omitempty"`
}

// ========== 流式会话管理 ==========

// StreamManager 流式会话管理器
type StreamManager struct {
	sessions map[string]*StreamSession
	mu       sync.RWMutex
}

// StreamSession 流式会话
type StreamSession struct {
	ID       string
	Provider Provider
	Cancel   context.CancelFunc
	Active   bool
}

// NewStreamManager 创建流式管理器
func NewStreamManager() *StreamManager {
	return &StreamManager{
		sessions: make(map[string]*StreamSession),
	}
}

// CreateSession 创建流式会话
func (sm *StreamManager) CreateSession(id string, provider Provider) *StreamSession {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session := &StreamSession{
		ID:       id,
		Provider: provider,
		Active:   true,
	}
	sm.sessions[id] = session
	return session
}

// GetSession 获取会话
func (sm *StreamManager) GetSession(id string) (*StreamSession, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	session, ok := sm.sessions[id]
	return session, ok
}

// CancelSession 取消会话
func (sm *StreamManager) CancelSession(id string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, ok := sm.sessions[id]; ok {
		if session.Cancel != nil {
			session.Cancel()
		}
		session.Active = false
	}
}

// RemoveSession 移除会话
func (sm *StreamManager) RemoveSession(id string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, id)
}
