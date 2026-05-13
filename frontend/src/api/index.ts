/**
 * Wails后端API封装
 * 提供前端调用Go后端方法的统一接口
 */

import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime.js'
import {
  GetVersion,
  GetConfig,
  UpdateConfig,
  GetLogDirectory,
  ExportLogs,
  DeleteHistoricalLogs,
  CleanupLogs,
  GetWorkspaces,
  CreateWorkspace,
  GetDocuments,
  GetDocument,
  SelectDocuments,
  ImportDocument,
  GetDocumentContent,
  UpdateDocumentMeta,
  DeleteDocument,
  GetProviders,
  GetModels,
  GetModelsWithConfig,
  GetWorkflows,
  GetWorkflowDefinition,
  NewWorkflowDefinition,
  SaveWorkflowDefinition,
  ExecuteWorkflow,
  CancelExecution,
  GetAnalysisHistory,
  GetAnalysisResult,
  ExportReport,
  DeleteAnalysisRecord,
  TestModelConnection,
  TestModelConnectionWithConfig,
  CreateChatSession,
  GetChatSession,
  DeleteChatSession,
  RenameChatSession,
  ListChatSessions,
  SendChatMessage,
  SendDocumentChatMessage,
  StopChatGeneration,
  GetChatHistory,
  ClearChatHistory,
  QuickChat
} from '../../wailsjs/go/main/App.js'

// 导出类型（从stores重新导出）
export type {
  Document,
  Workflow,
  WorkflowDefinitionDetail,
  Provider,
  Model,
  AnalysisRecord,
  AnalysisResult,
  StepResult
} from '../stores/app'
export type { ChatMessage, ChatSession, ChatEvent } from '../stores/chat'

// ========== 应用API ==========

export const appApi = {
  getVersion: GetVersion,
  getConfig: GetConfig,
  updateConfig: UpdateConfig,
  getLogDirectory: GetLogDirectory,
  exportLogs: ExportLogs,
  deleteHistoricalLogs: DeleteHistoricalLogs,
  cleanupLogs: CleanupLogs
}

// ========== 工作空间API ==========

export const workspaceApi = {
  list: GetWorkspaces,
  create: CreateWorkspace
}

// ========== 文档API ==========

export const documentApi = {
  list: GetDocuments,
  get: GetDocument,
  select: SelectDocuments,
  import: ImportDocument,
  getContent: GetDocumentContent,
  updateMeta: UpdateDocumentMeta,
  delete: DeleteDocument
}

// ========== AI提供商API ==========

export const providerApi = {
  list: GetProviders,
  getModels: GetModels,
  getModelsWithConfig: GetModelsWithConfig,
  testConnection: TestModelConnection,
  testConnectionWithConfig: TestModelConnectionWithConfig
}

// ========== 工作流API ==========

export const workflowApi = {
  list: GetWorkflows,
  getDefinition: GetWorkflowDefinition,
  newDefinition: NewWorkflowDefinition,
  saveDefinition: SaveWorkflowDefinition,
  execute: ExecuteWorkflow,
  cancel: CancelExecution,
  getHistory: GetAnalysisHistory,
  getResult: GetAnalysisResult,
  exportReport: ExportReport,
  deleteRecord: DeleteAnalysisRecord
}

// ========== 聊天API ==========

export const chatApi = {
  createSession: CreateChatSession,
  getSession: GetChatSession,
  deleteSession: DeleteChatSession,
  renameSession: RenameChatSession,
  listSessions: ListChatSessions,
  sendMessage: SendChatMessage,
  sendDocumentMessage: SendDocumentChatMessage,
  stopGeneration: StopChatGeneration,
  getHistory: GetChatHistory,
  clearHistory: ClearChatHistory,
  quickChat: QuickChat
}

// ========== 事件监听 ==========

export type ChatEventHandler = (event: any) => void
export type WorkflowEventHandler = (event: any) => void

function normalizeEventPayload<T extends Record<string, any>>(event: any): T {
  const payload = Array.isArray(event) ? event[0] : event
  return (payload || {}) as T
}

function normalizeChatEvent(event: any) {
  const payload = normalizeEventPayload<any>(event)
  return {
    session_id: payload.session_id ?? payload.SessionID ?? '',
    type: payload.type ?? payload.Type ?? '',
    content: payload.content ?? payload.Content ?? '',
    reasoning: payload.reasoning ?? payload.Reasoning ?? '',
    index: payload.index ?? payload.Index ?? 0,
    total_tokens: payload.total_tokens ?? payload.TotalTokens ?? 0,
    error: payload.error ?? payload.Error ?? '',
    citations: payload.citations ?? payload.Citations ?? []
  }
}

function normalizeWorkflowEvent(event: any) {
  return normalizeEventPayload<any>(event)
}

/**
 * 订阅聊天事件
 * @param handlers 事件处理器
 * @returns 取消订阅函数
 */
export function subscribeChatEvents(handlers: {
  onStart?: ChatEventHandler
  onChunk?: ChatEventHandler
  onDone?: ChatEventHandler
  onError?: ChatEventHandler
}) {
  EventsOn('chat:start', event => handlers.onStart?.(normalizeChatEvent(event)))
  EventsOn('chat:chunk', event => handlers.onChunk?.(normalizeChatEvent(event)))
  EventsOn('chat:done', event => handlers.onDone?.(normalizeChatEvent(event)))
  EventsOn('chat:error', event => handlers.onError?.(normalizeChatEvent(event)))

  return () => {
    EventsOff('chat:start')
    EventsOff('chat:chunk')
    EventsOff('chat:done')
    EventsOff('chat:error')
  }
}

/**
 * 订阅工作流事件
 * @param handlers 事件处理器
 * @returns 取消订阅函数
 */
export function subscribeWorkflowEvents(handlers: {
  onStart?: WorkflowEventHandler
  onStepStart?: WorkflowEventHandler
  onStepComplete?: WorkflowEventHandler
  onProgress?: WorkflowEventHandler
  onComplete?: WorkflowEventHandler
  onError?: WorkflowEventHandler
}) {
  EventsOn('workflow:start', event => handlers.onStart?.(normalizeWorkflowEvent(event)))
  EventsOn('workflow:step:start', event => (handlers.onStepStart || handlers.onProgress)?.(normalizeWorkflowEvent(event)))
  EventsOn('workflow:step:complete', event => (handlers.onStepComplete || handlers.onProgress)?.(normalizeWorkflowEvent(event)))
  EventsOn('workflow:progress', event => handlers.onProgress?.(normalizeWorkflowEvent(event)))
  EventsOn('workflow:complete', event => handlers.onComplete?.(normalizeWorkflowEvent(event)))
  EventsOn('workflow:error', event => handlers.onError?.(normalizeWorkflowEvent(event)))

  return () => {
    EventsOff('workflow:start')
    EventsOff('workflow:step:start')
    EventsOff('workflow:step:complete')
    EventsOff('workflow:progress')
    EventsOff('workflow:complete')
    EventsOff('workflow:error')
  }
}

// ========== 高级API封装 ==========

/**
 * 流式聊天封装
 * 提供更便捷的流式聊天接口
 */
export async function streamChat(
  provider: string,
  model: string,
  content: string,
  options: {
    sessionId?: string
    systemPrompt?: string
    temperature?: number
    maxTokens?: number
    onChunk?: (content: string, reasoning?: string) => void
    onDone?: (fullContent: string) => void
    onError?: (error: string) => void
  } = {}
): Promise<string> {
  const useChatStore = await import('../stores/chat').then(m => m.useChatStore)
  const store = useChatStore()

  // 获取或创建会话
  let sessionId = options.sessionId
  if (!sessionId) {
    const session = await CreateChatSession(provider, model, options.systemPrompt || '')
    sessionId = session.ID
    store.createSession(session.ID, provider, model)
  }

  // 订阅事件
  const unsubscribe = subscribeChatEvents({
    onChunk: (event: any) => {
      store.appendStreamContent(event.content || '')
      if (event.reasoning) {
        store.appendStreamReasoning(event.reasoning)
      }
      options.onChunk?.(event.content || '', event.reasoning)
    },
    onDone: (event: any) => {
      store.finishStreaming()
      options.onDone?.(event.content || '')
    },
    onError: (event: any) => {
      store.setError(event.error)
      options.onError?.(event.error)
    }
  })

  try {
    // 添加用户消息
    store.addMessage('user', content)
    store.startStreaming()

    // 发送消息
    await SendChatMessage(
      sessionId as string,
      content,
      options.temperature || 0.7,
      options.maxTokens || 4096,
      true // 流式
    )

    // 返回流式内容
    return store.streamingContent
  } finally {
    unsubscribe()
  }
}

/**
 * 非流式聊天封装
 */
export async function chat(
  provider: string,
  model: string,
  messages: Array<{ role: string; content: string }>,
  options: {
    temperature?: number
    maxTokens?: number
  } = {}
): Promise<string> {
  const response = await QuickChat(
    provider,
    model,
    messages as any,
    options.temperature || 0.7,
    options.maxTokens || 4096
  )
  return response.content
}
