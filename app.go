package main

import (
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gopkg.in/yaml.v3"

	"policy-analyzer/ai"
	"policy-analyzer/config"
	"policy-analyzer/document"
	apperrors "policy-analyzer/errors"
	"policy-analyzer/logger"
	"policy-analyzer/storage"
	"policy-analyzer/workflow"
)

// log 模块级日志记录器
var log = logger.WithModule("app")

// App 主应用结构体
// 所有暴露给前端的方法都绑定在此结构体上
type App struct {
	ctx      context.Context
	config   *config.Config
	aiMgr    *ai.Manager
	docSvc   *document.Service
	wfEngine *workflow.Engine
	store    *storage.Manager
}

// NewApp 创建新的应用实例
func NewApp() *App {
	return &App{}
}

// startup 应用启动时调用
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 初始化日志系统
	logDir := a.getLogDir()
	logger.Init(logger.Config{
		Level:    logger.LevelInfo,
		LogDir:   logDir,
		UseColor: true,
	})

	log.Info("应用启动中...")

	// 初始化配置
	if err := a.initConfig(); err != nil {
		log.ErrorErr("配置初始化失败", err)
	}

	// 初始化存储层
	if err := a.initStorage(); err != nil {
		log.ErrorErr("存储层初始化失败", err)
	}

	// 初始化AI管理器
	a.initAIManager()

	// 初始化文档服务
	a.initDocumentService()

	// 初始化工作流引擎
	a.initWorkflowEngine()

	a.applyConfiguredLogCleanup()

	log.Info("应用启动完成")
}

// shutdown 应用关闭时调用
func (a *App) shutdown(ctx context.Context) {
	log.Info("应用关闭中...")
	// 清理资源
	if a.aiMgr != nil {
		a.aiMgr.Close()
	}
	if a.store != nil {
		a.store.Close()
	}
	logger.Close()
	log.Info("应用已关闭")
}

// initConfig 初始化配置
func (a *App) initConfig() error {
	configPath := a.getConfigPath()

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 创建默认配置
		a.config = config.DefaultConfig()
		// 保存默认配置
		if err := a.saveConfig(); err != nil {
			return apperrors.Wrap(apperrors.ErrConfigInvalid, "保存默认配置失败", err)
		}
		log.Info("已创建默认配置文件", logger.F("path", configPath))
		return nil
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return apperrors.Wrap(apperrors.ErrConfigInvalid, "读取配置文件失败", err)
	}

	a.config = &config.Config{}
	if err := yaml.Unmarshal(data, a.config); err != nil {
		return apperrors.Wrap(apperrors.ErrConfigInvalid, "解析配置文件失败", err)
	}

	log.Info("配置加载完成", logger.F("path", configPath))
	return nil
}

// getConfigPath 获取配置文件路径
func (a *App) getConfigPath() string {
	// 使用用户目录下的应用数据目录
	homeDir, _ := os.UserHomeDir()
	appDir := filepath.Join(homeDir, ".policy-analyzer")
	configDir := filepath.Join(appDir, "config")

	// 确保目录存在
	os.MkdirAll(configDir, 0755)

	return filepath.Join(configDir, "settings.yaml")
}

// saveConfig 保存配置
func (a *App) saveConfig() error {
	data, err := yaml.Marshal(a.config)
	if err != nil {
		return err
	}
	return os.WriteFile(a.getConfigPath(), data, 0644)
}

// initStorage 初始化存储层
func (a *App) initStorage() error {
	homeDir, _ := os.UserHomeDir()
	baseDir := filepath.Join(homeDir, ".policy-analyzer")

	a.store = storage.NewManager(baseDir)
	if err := a.store.Init(); err != nil {
		return apperrors.Wrap(apperrors.ErrStorageInit, "存储初始化失败", err)
	}

	log.Info("存储层初始化完成", logger.F("path", baseDir))
	return nil
}

// initAIManager 初始化AI管理器
func (a *App) initAIManager() {
	a.aiMgr = ai.NewManager(a.config.Providers)
	log.Info("AI管理器初始化完成", logger.F("provider_count", len(a.config.Providers)))
}

// initDocumentService 初始化文档服务
func (a *App) initDocumentService() {
	a.docSvc = document.NewService(a.store)
	log.Info("文档服务初始化完成")
}

// initWorkflowEngine 初始化工作流引擎
func (a *App) initWorkflowEngine() {
	a.wfEngine = workflow.NewEngine(a.aiMgr, a.docSvc, a.store)
	// 注册内置Skills
	a.wfEngine.RegisterBuiltInSkills()

	homeDir, _ := os.UserHomeDir()
	templateDir := filepath.Join(homeDir, ".policy-analyzer", "templates")
	if err := a.wfEngine.LoadWorkflows([]string{templateDir}); err != nil {
		log.Warn("工作流模板加载失败", logger.F("error", err.Error()))
	}
	log.Info("工作流引擎初始化完成")
}

// ========== 暴露给前端的方法 ==========

// GetVersion 获取应用版本
func (a *App) GetVersion() string {
	return "1.0.1"
}

// GetConfig 获取当前配置
func (a *App) GetConfig() *config.Config {
	return a.config
}

// UpdateConfig 更新配置
func (a *App) UpdateConfig(newConfig *config.Config) error {
	oldConfig := a.config
	a.config = newConfig
	if err := a.saveConfig(); err != nil {
		a.config = oldConfig
		return err
	}
	if a.aiMgr != nil {
		a.aiMgr.UpdateConfigs(a.config.Providers)
	}
	if a.config.Storage.LogAutoCleanup {
		a.applyConfiguredLogCleanup()
	}
	return nil
}

// GetWorkspaces 获取工作空间列表
func (a *App) GetWorkspaces() ([]storage.Workspace, error) {
	return a.store.ListWorkspaces()
}

// CreateWorkspace 创建新工作空间
func (a *App) CreateWorkspace(name, description string) (*storage.Workspace, error) {
	ws := &storage.Workspace{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
	}
	if err := a.store.CreateWorkspace(ws); err != nil {
		return nil, err
	}
	return ws, nil
}

// GetDocuments 获取文档列表
func (a *App) GetDocuments(workspaceID string) ([]document.Document, error) {
	return a.docSvc.List(a.ctx, workspaceID)
}

// SelectDocuments 打开文件选择对话框
func (a *App) SelectDocuments() ([]string, error) {
	return runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择要导入的文档",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "支持的文档",
				Pattern:     "*.pdf;*.docx;*.txt;*.md",
			},
		},
	})
}

// ImportDocument 导入文档
func (a *App) ImportDocument(filePath, workspaceID string, tags []string) (*document.Document, error) {
	log.Info("收到文档导入请求",
		logger.F("file_path", filePath),
		logger.F("workspace_id", workspaceID),
		logger.F("tags", tags))

	req := document.ImportRequest{
		SourcePath: filePath,
		Workspace:  workspaceID,
		Tags:       tags,
	}

	doc, err := a.docSvc.Import(a.ctx, req)
	if err != nil {
		log.ErrorErr("文档导入失败", err,
			logger.F("file_path", filePath),
			logger.F("workspace_id", workspaceID))
		return nil, err
	}

	log.Info("文档导入成功",
		logger.F("doc_id", doc.ID),
		logger.F("name", doc.Name))
	return doc, nil
}

// GetDocument 获取单个文档详情
func (a *App) GetDocument(docID string) (*document.Document, error) {
	return a.docSvc.Get(a.ctx, docID)
}

// GetDocumentContent 获取文档内容
func (a *App) GetDocumentContent(docID string) (string, error) {
	return a.docSvc.GetContent(a.ctx, docID)
}

// UpdateDocumentMeta 更新文档元数据
func (a *App) UpdateDocumentMeta(docID string, meta map[string]interface{}) error {
	return a.docSvc.UpdateMeta(a.ctx, docID, meta)
}

// DeleteDocument 删除文档
func (a *App) DeleteDocument(docID string) error {
	return a.docSvc.Delete(a.ctx, docID)
}

// GetProviders 获取AI提供商列表
func (a *App) GetProviders() []ai.ProviderInfo {
	return a.aiMgr.ListProviders()
}

// GetModels 获取指定提供商的模型列表
func (a *App) GetModels(providerName string) ([]ai.ModelInfo, error) {
	return a.aiMgr.ListModels(a.ctx, providerName)
}

// TestModelConnection 测试模型连通性
func (a *App) TestModelConnection(providerName, model string) (*ai.ConnectionTestResult, error) {
	return a.aiMgr.TestConnection(a.ctx, providerName, model)
}

// GetModelsWithConfig 使用临时提供商配置获取模型列表，不持久化配置
func (a *App) GetModelsWithConfig(provider config.ProviderConfig) ([]ai.ModelInfo, error) {
	return a.aiMgr.ListModelsWithConfig(a.ctx, provider)
}

// TestModelConnectionWithConfig 使用临时提供商配置测试模型连通性，不持久化配置
func (a *App) TestModelConnectionWithConfig(provider config.ProviderConfig, model string) (*ai.ConnectionTestResult, error) {
	return a.aiMgr.TestConnectionWithConfig(a.ctx, provider, model)
}

// GetWorkflows 获取工作流模板列表
func (a *App) GetWorkflows() []workflow.WorkflowInfo {
	return a.wfEngine.ListWorkflows()
}

// GetWorkflowDefinition 获取工作流完整定义
func (a *App) GetWorkflowDefinition(workflowID string) (*workflow.WorkflowEditorData, error) {
	return a.wfEngine.GetWorkflowDefinition(workflowID)
}

// NewWorkflowDefinition 获取新建工作流模板骨架
func (a *App) NewWorkflowDefinition() *workflow.WorkflowEditorData {
	return a.wfEngine.NewWorkflowDefinition()
}

// SaveWorkflowDefinition 保存工作流定义
func (a *App) SaveWorkflowDefinition(originalID, content string) (*workflow.WorkflowEditorData, error) {
	return a.wfEngine.SaveWorkflowDefinition(originalID, content)
}

// ExecuteWorkflow 执行工作流
func (a *App) ExecuteWorkflow(workflowID string, inputs map[string]interface{}) (string, error) {
	// 返回执行ID，前端通过事件监听进度
	execID := uuid.New().String()
	go a.wfEngine.ExecuteAsync(a.ctx, execID, workflowID, inputs)
	return execID, nil
}

// CancelExecution 取消工作流执行
func (a *App) CancelExecution(execID string) error {
	return a.wfEngine.CancelExecution(execID)
}

// GetAnalysisHistory 获取分析历史
func (a *App) GetAnalysisHistory(workspaceID string) ([]workflow.ExecutionRecord, error) {
	return a.wfEngine.GetHistory(a.ctx, workspaceID)
}

// DeleteAnalysisRecord 删除分析记录
func (a *App) DeleteAnalysisRecord(execID string) error {
	return a.wfEngine.DeleteExecution(execID)
}

// GetAnalysisResult 获取分析结果
func (a *App) GetAnalysisResult(execID string) (*workflow.ExecutionResult, error) {
	return a.wfEngine.GetResult(execID)
}

// ExportReport 导出报告
// 文本格式直接返回内容；PDF 返回 Base64 编码字符串，避免 Wails 传输二进制时损坏。
func (a *App) ExportReport(execID string, format string) (string, error) {
	log.Info("ExportReport调用", logger.F("execID", execID), logger.F("format", format))
	result, err := a.wfEngine.GetResult(execID)
	if err != nil {
		log.Error("获取执行结果失败", logger.F("error", err.Error()))
		return "", err
	}
	log.Info("获取结果成功",
		logger.F("outputs_keys", getKeysFromMap(result.Outputs)),
		logger.F("status", result.Status))
	data, err := a.wfEngine.ExportReport(result, format)
	if err != nil {
		log.Error("导出报告失败", logger.F("error", err.Error()))
		return "", err
	}
	log.Info("导出报告成功", logger.F("data_length", len(data)))

	if strings.EqualFold(format, "pdf") {
		return base64.StdEncoding.EncodeToString(data), nil
	}

	return string(data), nil
}

// getKeysFromMap 获取 map 的键列表
func getKeysFromMap(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// ========== AI流式聊天接口 ==========

// streamService 流式聊天服务实例（懒加载）
var streamService *ai.StreamService

// getStreamService 获取流式服务实例
func (a *App) getStreamService() *ai.StreamService {
	if streamService == nil {
		streamService = ai.NewStreamService(a.aiMgr, a.docSvc)
		streamService.SetAppContext(a.ctx)
	}
	return streamService
}

// CreateChatSession 创建聊天会话
func (a *App) CreateChatSession(provider, model, systemPrompt string) (*ai.ChatSession, error) {
	req := ai.ChatStartRequest{
		Provider:     provider,
		Model:        model,
		SystemPrompt: systemPrompt,
	}
	return a.getStreamService().CreateSession(req)
}

// GetChatSession 获取聊天会话
func (a *App) GetChatSession(sessionID string) (*ai.ChatSession, error) {
	return a.getStreamService().GetSession(sessionID)
}

// DeleteChatSession 删除聊天会话
func (a *App) DeleteChatSession(sessionID string) {
	a.getStreamService().DeleteSession(sessionID)
}

// RenameChatSession 重命名聊天会话
func (a *App) RenameChatSession(sessionID, title string) (*ai.ChatSession, error) {
	return a.getStreamService().RenameSession(sessionID, title)
}

// ListChatSessions 列出所有聊天会话
func (a *App) ListChatSessions() []*ai.ChatSession {
	return a.getStreamService().ListSessions()
}

// SendChatMessage 发送聊天消息
// stream参数控制是否流式输出，流式输出通过事件"chat:chunk"和"chat:done"通知前端
func (a *App) SendChatMessage(sessionID, content string, temperature float64, maxTokens int, stream bool) (*ai.ChatResponse, error) {
	log.Info("收到AI聊天请求",
		logger.F("session_id", sessionID),
		logger.F("stream", stream),
		logger.F("content_length", len(content)),
		logger.F("temperature", temperature),
		logger.F("max_tokens", maxTokens))

	req := ai.ChatMessageRequest{
		SessionID:   sessionID,
		Content:     content,
		Temperature: temperature,
		MaxTokens:   maxTokens,
		Stream:      stream,
	}
	resp, err := a.getStreamService().SendMessage(a.ctx, req)
	if err != nil {
		log.ErrorErr("AI聊天请求失败", err,
			logger.F("session_id", sessionID),
			logger.F("stream", stream))
		return nil, err
	}

	log.Info("AI聊天请求完成",
		logger.F("session_id", sessionID),
		logger.F("stream", stream),
		logger.F("response_model", resp.Model),
		logger.F("response_length", len(resp.Content)))
	return resp, nil
}

// SendDocumentChatMessage 发送基于文档检索的聊天消息
func (a *App) SendDocumentChatMessage(sessionID, displayContent string, selectedDocIDs []string, temperature float64, maxTokens int, stream bool) (*ai.ChatResponse, error) {
	log.Info("收到AI文档问答请求",
		logger.F("session_id", sessionID),
		logger.F("stream", stream),
		logger.F("selected_doc_count", len(selectedDocIDs)),
		logger.F("content_length", len(displayContent)),
		logger.F("temperature", temperature),
		logger.F("max_tokens", maxTokens))

	req := ai.DocumentChatRequest{
		SessionID:      sessionID,
		DisplayContent: displayContent,
		SelectedDocIDs: selectedDocIDs,
		Temperature:    temperature,
		MaxTokens:      maxTokens,
		Stream:         stream,
	}
	resp, err := a.getStreamService().SendDocumentMessage(a.ctx, req)
	if err != nil {
		log.ErrorErr("AI文档问答请求失败", err,
			logger.F("session_id", sessionID),
			logger.F("stream", stream),
			logger.F("selected_doc_count", len(selectedDocIDs)))
		return nil, err
	}

	log.Info("AI文档问答请求完成",
		logger.F("session_id", sessionID),
		logger.F("stream", stream),
		logger.F("selected_doc_count", len(selectedDocIDs)),
		logger.F("response_model", resp.Model),
		logger.F("response_length", len(resp.Content)))
	return resp, nil
}

// StopChatGeneration 停止生成
func (a *App) StopChatGeneration(sessionID string) {
	a.getStreamService().StopGeneration(sessionID)
}

// GetChatHistory 获取聊天历史
func (a *App) GetChatHistory(sessionID string) ([]ai.Message, error) {
	return a.getStreamService().GetSessionHistory(sessionID)
}

// ClearChatHistory 清除聊天历史
func (a *App) ClearChatHistory(sessionID string) error {
	return a.getStreamService().ClearSessionHistory(sessionID)
}

// QuickChat 快速对话（无状态）
func (a *App) QuickChat(provider, model string, messages []ai.Message, temperature float64, maxTokens int) (*ai.ChatResponse, error) {
	opts := &ai.ChatOptions{
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}
	return a.getStreamService().QuickChat(a.ctx, provider, model, messages, opts)
}
