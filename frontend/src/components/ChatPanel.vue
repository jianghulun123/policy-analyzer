<template>
  <div class="chat-panel h-full flex bg-white dark:bg-gray-800 rounded-lg shadow border border-gray-200 dark:border-gray-700 overflow-hidden">
    <aside class="chat-sidebar w-72 border-r border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-900/40 flex flex-col">
      <div class="p-4 border-b border-gray-200 dark:border-gray-700">
        <div class="flex items-center justify-between gap-3">
          <div>
            <div class="text-sm font-semibold text-gray-800 dark:text-gray-100">会话管理</div>
            <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">共 {{ sortedSessions.length }} 个会话</div>
          </div>
          <button
            class="btn-secondary btn-sm"
            :disabled="isStreaming"
            @click="startNewSession"
          >
            新建
          </button>
        </div>
      </div>

      <div class="flex-1 overflow-y-auto p-3 space-y-2">
        <button
          type="button"
          class="w-full rounded-lg border px-3 py-3 text-left transition-colors"
          :class="!currentSession
            ? 'border-primary-300 bg-primary-50 text-primary-700 dark:border-primary-700 dark:bg-primary-900/20 dark:text-primary-300'
            : 'border-gray-200 bg-white text-gray-600 hover:border-primary-300 hover:bg-gray-50 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300 dark:hover:border-primary-700 dark:hover:bg-gray-700'"
          @click="startNewSession"
        >
          <div class="text-sm font-medium">新对话</div>
          <div class="mt-1 text-xs opacity-80">使用当前提供商和模型开始新的问答会话</div>
        </button>

        <div
          v-for="session in sortedSessions"
          :key="session.id"
          class="rounded-lg border px-3 py-3 transition-colors cursor-pointer"
          :class="currentSession?.id === session.id
            ? 'border-primary-300 bg-primary-50 dark:border-primary-700 dark:bg-primary-900/20'
            : 'border-gray-200 bg-white hover:border-primary-300 hover:bg-gray-50 dark:border-gray-700 dark:bg-gray-800 dark:hover:border-primary-700 dark:hover:bg-gray-700'"
          @click="selectSession(session)"
        >
          <div class="flex items-start gap-2">
            <div class="min-w-0 flex-1">
              <div v-if="editingSessionId === session.id" class="space-y-2">
                <input
                  ref="renameInputRef"
                  v-model="renamingTitle"
                  type="text"
                  maxlength="60"
                  class="input text-sm w-full"
                  @click.stop
                  @keydown.enter.prevent="submitRename(session)"
                  @keydown.esc.prevent="cancelRename"
                />
                <div class="flex gap-2">
                  <button class="btn-secondary btn-sm" :disabled="isRenaming" @click.stop="submitRename(session)">
                    {{ isRenaming ? '保存中...' : '保存' }}
                  </button>
                  <button class="btn-secondary btn-sm" :disabled="isRenaming" @click.stop="cancelRename">
                    取消
                  </button>
                </div>
              </div>
              <template v-else>
                <div class="truncate text-sm font-medium text-gray-800 dark:text-gray-100">
                  {{ getSessionTitle(session) }}
                </div>
                <div class="mt-1 text-xs text-gray-500 dark:text-gray-400 truncate">
                  {{ session.provider }} / {{ getSessionModelName(session) }}
                </div>
                <div class="mt-1 text-xs text-gray-500 dark:text-gray-400 truncate">
                  {{ getSessionPreview(session) }}
                </div>
                <div class="mt-2 text-xs text-gray-400 dark:text-gray-500">
                  {{ formatSessionTime(session.last_active_at) }}
                </div>
              </template>
            </div>
            <div v-if="editingSessionId !== session.id" class="flex shrink-0 items-center gap-1">
              <button
                class="rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-gray-700 dark:hover:text-gray-200"
                :disabled="isStreaming"
                @click.stop="beginRename(session)"
              >
                ✎
              </button>
              <button
                class="rounded p-1 text-gray-400 hover:bg-red-50 hover:text-red-500 dark:hover:bg-red-900/20 dark:hover:text-red-300"
                :disabled="isStreaming"
                @click.stop="removeSession(session.id)"
              >
                ×
              </button>
            </div>
          </div>
        </div>

        <div v-if="sortedSessions.length === 0" class="rounded-lg border border-dashed border-gray-300 px-3 py-6 text-center text-sm text-gray-500 dark:border-gray-700 dark:text-gray-400">
          暂无历史会话
        </div>
      </div>
    </aside>

    <div class="flex-1 min-w-0 flex flex-col">
      <div class="chat-header flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700">
        <div class="flex items-center gap-2 min-w-0">
          <span class="text-lg font-semibold text-gray-800 dark:text-gray-100">AI 助手</span>
          <span v-if="currentProvider" class="text-sm text-gray-500 dark:text-gray-400 truncate">
            {{ currentProvider }} / {{ currentModel }}
          </span>
        </div>
        <div class="flex items-center gap-2">
          <button
            @click="clearHistory"
            class="px-3 py-1 text-sm text-gray-600 dark:text-gray-300 hover:text-gray-800 dark:hover:text-gray-100 hover:bg-gray-100 dark:hover:bg-gray-700 rounded"
            :disabled="!currentSession"
          >
            清除对话
          </button>
          <select
            v-model="selectedProvider"
            class="px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-900 text-gray-700 dark:text-gray-200 rounded"
            @change="onProviderChange"
          >
            <option v-for="p in providers" :key="p.name" :value="p.name">
              {{ p.name }}
            </option>
          </select>
          <select
            v-model="selectedModel"
            class="px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-900 text-gray-700 dark:text-gray-200 rounded"
          >
            <option v-for="m in currentModels" :key="m.id" :value="m.id">
              {{ m.name }}
            </option>
          </select>
        </div>
      </div>

      <div ref="messagesContainer" class="chat-messages flex-1 overflow-y-auto p-4 space-y-4">
        <div v-if="messages.length === 0" class="text-center text-gray-400 dark:text-gray-500 py-8">
          <p>{{ hasSelectedDocuments ? '开始基于所选文档提问' : '请先在上方选择至少一份文档' }}</p>
          <p class="text-sm mt-2">
            {{ hasSelectedDocuments ? `当前已选择 ${selectedDocumentCount} 份文档，可直接提问总结、对比、解读等问题` : '选择文档后，AI 将结合文档内容回答你的问题' }}
          </p>
        </div>

        <div
          v-for="(msg, index) in messages"
          :key="index"
          :class="[
            'message p-3 rounded-lg max-w-[85%]',
            msg.role === 'user' ? 'ml-auto bg-blue-500 text-white' : 'bg-gray-100 dark:bg-gray-700 text-gray-800 dark:text-gray-100'
          ]"
        >
          <div v-if="msg.role === 'user'" class="whitespace-pre-wrap">{{ msg.displayContent || msg.content }}</div>
          <div v-else>
            <div v-if="msg.citations?.length" class="mb-3 rounded bg-gray-50 px-3 py-2 text-xs text-gray-500 dark:bg-gray-800 dark:text-gray-400">
              <div class="mb-1 font-medium">引用文档</div>
              <div class="flex flex-wrap gap-2">
                <span
                  v-for="citation in msg.citations"
                  :key="`${citation.document_id}-${citation.chunk_index}`"
                  class="rounded bg-white px-2 py-1 dark:bg-gray-900"
                  :title="citation.snippet"
                >
                  {{ citation.document_name }} #{{ citation.chunk_index + 1 }}
                </span>
              </div>
            </div>
            <div class="prose prose-sm max-w-none dark:prose-invert" v-html="renderMarkdown(msg.content)"></div>
          </div>
        </div>

        <div v-if="isStreaming" class="message p-3 rounded-lg bg-gray-100 dark:bg-gray-700 text-gray-800 dark:text-gray-100 max-w-[85%]">
          <div v-if="streamingCitations.length" class="mb-3 rounded bg-gray-50 px-3 py-2 text-xs text-gray-500 dark:bg-gray-800 dark:text-gray-400">
            <div class="mb-1 font-medium">引用文档</div>
            <div class="flex flex-wrap gap-2">
              <span
                v-for="citation in streamingCitations"
                :key="`${citation.document_id}-${citation.chunk_index}`"
                class="rounded bg-white px-2 py-1 dark:bg-gray-900"
                :title="citation.snippet"
              >
                {{ citation.document_name }} #{{ citation.chunk_index + 1 }}
              </span>
            </div>
          </div>
          <div v-if="streamingReasoning" class="text-sm text-gray-500 dark:text-gray-400 mb-2 p-2 bg-gray-50 dark:bg-gray-800 rounded">
            <div class="font-medium mb-1">思考过程:</div>
            <div class="whitespace-pre-wrap">{{ streamingReasoning }}</div>
          </div>
          <div v-if="isWaitingAssistant" class="text-lg tracking-widest text-gray-500 dark:text-gray-400">...</div>
          <div v-else class="prose prose-sm max-w-none dark:prose-invert" v-html="renderMarkdown(streamingContent)"></div>
          <div class="flex items-center gap-1 mt-2">
            <span class="animate-pulse">▊</span>
          </div>
        </div>
      </div>

      <div class="chat-input p-4 border-t border-gray-200 dark:border-gray-700">
        <div class="flex gap-2">
          <textarea
            v-model="inputText"
            @keydown.enter.exact="handleEnter"
            @keydown.enter.shift.exact="() => {}"
            :placeholder="hasSelectedDocuments ? '输入问题... (Shift+Enter换行，Enter发送)' : '请先选择至少一份文档后再提问'"
            class="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-900 text-gray-800 dark:text-gray-100 rounded-lg resize-none focus:outline-none focus:border-blue-500"
            rows="3"
            :disabled="isStreaming || !hasSelectedDocuments"
          ></textarea>
          <div class="flex flex-col gap-2">
            <button
              v-if="!isStreaming"
              @click="sendMessage"
              :disabled="!inputText.trim() || !hasSelectedDocuments"
              class="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:bg-gray-300 dark:disabled:bg-gray-600 disabled:cursor-not-allowed"
            >
              发送
            </button>
            <button
              v-else
              @click="stopGeneration"
              class="px-4 py-2 bg-red-500 text-white rounded-lg hover:bg-red-600"
            >
              停止
            </button>
          </div>
        </div>
        <div class="flex justify-between mt-2 text-sm text-gray-500 dark:text-gray-400">
          <span>温度: {{ temperature }}</span>
          <input
            type="range"
            v-model.number="temperature"
            min="0"
            max="2"
            step="0.1"
            class="w-32"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick, watch, onUnmounted } from 'vue'
import { useChatStore } from '../stores/chat'
import type { ChatMessage, ChatSession as LocalChatSession } from '../stores/chat'
import { useAppStore } from '../stores/app'
import { chatApi, providerApi, subscribeChatEvents } from '../api'
import { renderMarkdown } from '../utils/markdown'
import type { ai } from '../../wailsjs/go/models'

const props = defineProps<{
  systemPrompt?: string
  selectedDocuments?: string[]
}>()

const chatStore = useChatStore()
const appStore = useAppStore()

const messagesContainer = ref<HTMLElement | null>(null)
const renameInputRef = ref<HTMLInputElement | null>(null)
const inputText = ref('')
const renamingTitle = ref('')
const editingSessionId = ref<string | null>(null)
const isRenaming = ref(false)
const temperature = ref(0.7)
const unsubscribe = ref<(() => void) | null>(null)

const providers = computed(() => appStore.providers)
const selectedProvider = computed({
  get: () => appStore.selectedProvider,
  set: (v) => { appStore.selectedProvider = v }
})
const selectedModel = computed({
  get: () => appStore.selectedModel,
  set: (v) => { appStore.selectedModel = v }
})

const currentProvider = computed(() => {
  return providers.value.find(p => p.name === selectedProvider.value)?.name || ''
})

const currentModel = computed(() => {
  const provider = providers.value.find(p => p.name === selectedProvider.value)
  return provider?.models.find(m => m.id === selectedModel.value)?.name || selectedModel.value
})

const currentModels = computed(() => {
  const provider = providers.value.find(p => p.name === selectedProvider.value)
  return provider?.models || []
})

const currentSession = computed(() => chatStore.currentSession)
const sortedSessions = computed(() => {
  return [...chatStore.sessions].sort((a, b) => {
    return new Date(b.last_active_at).getTime() - new Date(a.last_active_at).getTime()
  })
})
const messages = computed(() => chatStore.currentSession?.messages || [])
const isStreaming = computed(() => chatStore.isStreaming)
const streamingContent = computed(() => chatStore.streamingContent)
const streamingReasoning = computed(() => chatStore.streamingReasoning)
const streamingCitations = computed(() => chatStore.streamingCitations)
const selectedDocumentCount = computed(() => props.selectedDocuments?.length || 0)
const hasSelectedDocuments = computed(() => selectedDocumentCount.value > 0)
const isWaitingAssistant = computed(() => {
  return isStreaming.value && !streamingContent.value && !streamingReasoning.value
})

function mapChatMessages(messages: ai.Message[] = []): ChatMessage[] {
  return messages.map(message => ({
    role: message.role as ChatMessage['role'],
    content: message.content,
    displayContent: message.display_content ?? message.content,
    citations: message.citations || []
  }))
}

function normalizeDate(value: unknown): string {
  if (!value) {
    return new Date().toISOString()
  }
  const date = new Date(value as string)
  if (Number.isNaN(date.getTime())) {
    return new Date().toISOString()
  }
  return date.toISOString()
}

function mapChatSession(session: ai.ChatSession): LocalChatSession {
  return {
    id: session.ID,
    title: session.Title || '新对话',
    provider: session.Provider,
    model: session.Model,
    messages: mapChatMessages(session.Messages || []),
    created_at: normalizeDate(session.CreatedAt),
    last_active_at: normalizeDate(session.LastActiveAt)
  }
}

function scrollToBottom() {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

function getSessionTitle(session: LocalChatSession) {
  return session.title?.trim() || '新对话'
}

function getSessionPreview(session: LocalChatSession) {
  const lastMessage = [...session.messages].reverse().find(message => message.role !== 'system')
  if (!lastMessage) {
    return '暂无消息'
  }
  return (lastMessage.displayContent || lastMessage.content).slice(0, 36)
}

function getSessionModelName(session: LocalChatSession) {
  const provider = providers.value.find(item => item.name === session.provider)
  return provider?.models.find(item => item.id === session.model)?.name || session.model
}

function formatSessionTime(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return ''
  }
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

async function loadSessionHistory(sessionId: string) {
  const history = await chatApi.getHistory(sessionId) as ai.Message[]
  chatStore.setCurrentSessionMessages(mapChatMessages(history))
}

async function loadSessions() {
  try {
    const sessions = await chatApi.listSessions() as ai.ChatSession[]
    const mappedSessions = sessions.map(mapChatSession)
    chatStore.setSessions(mappedSessions)

    if (mappedSessions.length > 0) {
      await selectSession(mappedSessions.sort((a, b) => new Date(b.last_active_at).getTime() - new Date(a.last_active_at).getTime())[0], false)
    }
  } catch (error) {
    console.error('加载会话列表失败:', error)
  }
}

async function initSession(force = false) {
  const needsNewSession =
    force ||
    !chatStore.currentSession ||
    chatStore.currentSession.provider !== selectedProvider.value ||
    chatStore.currentSession.model !== selectedModel.value

  if (!needsNewSession) {
    return
  }

  try {
    const session = await chatApi.createSession(
      selectedProvider.value,
      selectedModel.value,
      props.systemPrompt || ''
    ) as ai.ChatSession
    chatStore.createSession(session.ID, selectedProvider.value, selectedModel.value)
  } catch (err: any) {
    console.error('创建会话失败:', err)
    throw err
  }
}

async function selectSession(session: LocalChatSession, scroll = true) {
  if (isStreaming.value || editingSessionId.value) {
    return
  }

  selectedProvider.value = session.provider
  selectedModel.value = session.model
  chatStore.setCurrentSession(session)

  try {
    await loadSessionHistory(session.id)
    if (scroll) {
      scrollToBottom()
    }
  } catch (error) {
    console.error('加载会话历史失败:', error)
    handleRequestError(error instanceof Error ? error.message : '加载会话失败')
  }
}

function startNewSession() {
  if (isStreaming.value) {
    return
  }
  cancelRename()
  chatStore.setCurrentSession(null)
  chatStore.clearError()
  inputText.value = ''
}

function beginRename(session: LocalChatSession) {
  if (isStreaming.value) {
    return
  }
  editingSessionId.value = session.id
  renamingTitle.value = session.title || ''
  nextTick(() => {
    renameInputRef.value?.focus()
    renameInputRef.value?.select()
  })
}

function cancelRename() {
  editingSessionId.value = null
  renamingTitle.value = ''
}

async function submitRename(session: LocalChatSession) {
  const title = renamingTitle.value.trim()
  if (!title || isRenaming.value) {
    return
  }

  isRenaming.value = true
  try {
    const renamed = await chatApi.renameSession(session.id, title) as ai.ChatSession
    const mapped = mapChatSession(renamed)
    const nextSessions = chatStore.sessions.map(item => item.id === session.id ? { ...item, ...mapped } : item)
    chatStore.setSessions(nextSessions)
    if (chatStore.currentSession?.id === session.id) {
      chatStore.setCurrentSession(nextSessions.find(item => item.id === session.id) || null)
    }
    cancelRename()
  } catch (error) {
    console.error('重命名会话失败:', error)
    handleRequestError(error instanceof Error ? error.message : '重命名会话失败')
  } finally {
    isRenaming.value = false
  }
}

async function removeSession(sessionId: string) {
  if (isStreaming.value) {
    return
  }

  const deletingCurrent = chatStore.currentSession?.id === sessionId
  try {
    await chatApi.deleteSession(sessionId)
    chatStore.deleteSession(sessionId)
    if (editingSessionId.value === sessionId) {
      cancelRename()
    }

    if (deletingCurrent && sortedSessions.value.length > 0) {
      await selectSession(sortedSessions.value[0])
      return
    }

    if (deletingCurrent) {
      startNewSession()
    }
  } catch (error) {
    console.error('删除会话失败:', error)
    handleRequestError(error instanceof Error ? error.message : '删除会话失败')
  }
}

function isCurrentSessionEvent(event: any) {
  return event?.session_id === chatStore.currentSession?.id
}

async function sendMessage() {
  const content = inputText.value.trim()
  if (!content || isStreaming.value || !hasSelectedDocuments.value) return

  await initSession()

  if (!unsubscribe.value) {
    unsubscribe.value = subscribeChatEvents({
      onStart: (event: any) => {
        if (!isCurrentSessionEvent(event)) return
        chatStore.startStreaming(event.citations || [])
        scrollToBottom()
      },
      onChunk: (event: any) => {
        if (!isCurrentSessionEvent(event)) return
        chatStore.appendStreamContent(event.content || '')
        if (event.reasoning) {
          chatStore.appendStreamReasoning(event.reasoning)
        }
        scrollToBottom()
      },
      onDone: (event: any) => {
        if (!isCurrentSessionEvent(event)) return
        chatStore.finishStreaming(event.content || '', event.citations || [])
        scrollToBottom()
      },
      onError: (event: any) => {
        if (!isCurrentSessionEvent(event)) return
        handleRequestError(event.error)
      }
    })
  }

  inputText.value = ''
  chatStore.addMessage('user', content, { displayContent: content })
  scrollToBottom()

  try {
    const sessionId = chatStore.currentSession?.id
    if (!sessionId) throw new Error('无活动会话')

    await chatApi.sendDocumentMessage(
      sessionId,
      content,
      props.selectedDocuments || [],
      temperature.value,
      4096,
      true
    )
  } catch (err: any) {
    console.error('[AI问答] 文档问答请求发送异常', err)
    handleRequestError(err.message || '发送失败')
  }
}

function handleRequestError(message: string) {
  const errorMessage = message || '发送失败'
  const lastMessage = messages.value[messages.value.length - 1]
  if (lastMessage?.role === 'assistant' && lastMessage.content === `请求失败：${errorMessage}`) {
    chatStore.setError(errorMessage)
    scrollToBottom()
    return
  }
  chatStore.setError(errorMessage)
  chatStore.addMessage('assistant', `请求失败：${errorMessage}`)
  scrollToBottom()
}

function stopGeneration() {
  const sessionId = chatStore.currentSession?.id
  if (sessionId) {
    chatApi.stopGeneration(sessionId)
    chatStore.finishStreaming()
  }
}

async function clearHistory() {
  const sessionId = chatStore.currentSession?.id
  if (sessionId) {
    await chatApi.clearHistory(sessionId)
    chatStore.clearHistory()
  }
}

function onProviderChange() {
  const provider = providers.value.find(p => p.name === selectedProvider.value)
  if (provider && provider.default_model) {
    selectedModel.value = provider.default_model
  } else if (provider && provider.models.length > 0) {
    selectedModel.value = provider.models[0].id
  }

  chatStore.setCurrentSession(null)
}

function handleEnter(e: KeyboardEvent) {
  if (!e.shiftKey) {
    e.preventDefault()
    sendMessage()
  }
}

onMounted(async () => {
  if (providers.value.length === 0) {
    const provs = await providerApi.list()
    appStore.setProviders(provs)
  }

  await loadSessions()
})

watch([selectedProvider, selectedModel], ([provider, model], [prevProvider, prevModel]) => {
  if (!chatStore.currentSession) {
    return
  }
  if (provider === prevProvider && model === prevModel) {
    return
  }
  if (chatStore.currentSession.provider !== provider || chatStore.currentSession.model !== model) {
    chatStore.setCurrentSession(null)
  }
})

watch(streamingContent, () => {
  scrollToBottom()
})

onUnmounted(() => {
  if (unsubscribe.value) {
    unsubscribe.value()
  }
})
</script>

<style scoped>
.chat-panel {
  min-height: 520px;
}

.message {
  animation: fadeIn 0.2s ease-in;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.prose :deep(pre) {
  background: #1e293b;
  color: #e2e8f0;
  padding: 1rem;
  border-radius: 0.5rem;
  overflow-x: auto;
}

.prose :deep(code) {
  background: #334155;
  padding: 0.125rem 0.25rem;
  border-radius: 0.25rem;
  font-size: 0.875em;
}

.prose :deep(pre code) {
  background: transparent;
  padding: 0;
}
</style>
