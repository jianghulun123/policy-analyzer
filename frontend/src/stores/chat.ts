import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export interface ChatCitation {
  document_id: string
  document_name: string
  chunk_index: number
  snippet: string
  score: number
}

// 聊天消息接口
export interface ChatMessage {
  role: 'system' | 'user' | 'assistant'
  content: string
  displayContent?: string
  citations?: ChatCitation[]
}

// 聊天会话接口
export interface ChatSession {
  id: string
  title: string
  provider: string
  model: string
  messages: ChatMessage[]
  created_at: string
  last_active_at: string
}

// 聊天事件
export interface ChatEvent {
  session_id: string
  type: 'start' | 'chunk' | 'done' | 'error'
  content?: string
  reasoning?: string
  index?: number
  total_tokens?: number
  error?: string
  citations?: ChatCitation[]
}

// AI聊天状态Store
export const useChatStore = defineStore('chat', () => {
  const sessions = ref<ChatSession[]>([])
  const currentSession = ref<ChatSession | null>(null)
  const isStreaming = ref(false)
  const streamingContent = ref('')
  const streamingReasoning = ref('')
  const streamingCitations = ref<ChatCitation[]>([])
  const error = ref<string | null>(null)

  const hasSession = computed(() => currentSession.value !== null)
  const messageCount = computed(() => currentSession.value?.messages.length || 0)

  function createSession(id: string, provider: string, model: string) {
    const session: ChatSession = {
      id,
      title: '新对话',
      provider,
      model,
      messages: [],
      created_at: new Date().toISOString(),
      last_active_at: new Date().toISOString()
    }
    sessions.value.push(session)
    currentSession.value = session
    return session
  }

  function setSessions(nextSessions: ChatSession[]) {
    sessions.value = nextSessions
  }

  function setCurrentSession(session: ChatSession | null) {
    currentSession.value = session
  }

  function addMessage(
    role: 'system' | 'user' | 'assistant',
    content: string,
    options: { displayContent?: string; citations?: ChatCitation[] } = {}
  ) {
    if (currentSession.value) {
      currentSession.value.messages.push({
        role,
        content,
        displayContent: options.displayContent ?? content,
        citations: options.citations || []
      })
      currentSession.value.last_active_at = new Date().toISOString()
    }
  }

  function setCurrentSessionMessages(messages: ChatMessage[]) {
    if (currentSession.value) {
      currentSession.value.messages = messages
      currentSession.value.last_active_at = new Date().toISOString()
    }
  }

  function startStreaming(citations: ChatCitation[] = []) {
    isStreaming.value = true
    streamingContent.value = ''
    streamingReasoning.value = ''
    streamingCitations.value = citations
    error.value = null
  }

  function appendStreamContent(content: string) {
    streamingContent.value += content
  }

  function appendStreamReasoning(reasoning: string) {
    streamingReasoning.value += reasoning
  }

  function setStreamingCitations(citations: ChatCitation[] = []) {
    streamingCitations.value = citations
  }

  function finishStreaming(content?: string, citations?: ChatCitation[]) {
    const finalContent = content ?? streamingContent.value
    const finalCitations = citations ?? streamingCitations.value
    if (currentSession.value && finalContent) {
      addMessage('assistant', finalContent, { citations: finalCitations })
    }
    isStreaming.value = false
    streamingContent.value = ''
    streamingReasoning.value = ''
    streamingCitations.value = []
  }

  function setError(err: string) {
    error.value = err
    isStreaming.value = false
    streamingCitations.value = []
  }

  function clearError() {
    error.value = null
  }

  function clearHistory() {
    if (currentSession.value) {
      currentSession.value.messages = currentSession.value.messages.filter(
        msg => msg.role === 'system'
      )
    }
    streamingCitations.value = []
  }

  function deleteSession(sessionId: string) {
    const index = sessions.value.findIndex(s => s.id === sessionId)
    if (index > -1) {
      sessions.value.splice(index, 1)
      if (currentSession.value?.id === sessionId) {
        currentSession.value = null
      }
    }
  }

  return {
    sessions,
    currentSession,
    isStreaming,
    streamingContent,
    streamingReasoning,
    streamingCitations,
    error,
    hasSession,
    messageCount,
    createSession,
    setSessions,
    setCurrentSession,
    setCurrentSessionMessages,
    addMessage,
    startStreaming,
    appendStreamContent,
    appendStreamReasoning,
    setStreamingCitations,
    finishStreaming,
    setError,
    clearError,
    clearHistory,
    deleteSession
  }
})

