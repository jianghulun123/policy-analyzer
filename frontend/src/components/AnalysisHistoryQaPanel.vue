<template>
  <section class="card mt-8">
    <div class="mb-4 flex items-start justify-between gap-4">
      <div>
        <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-100">AI 问答</h3>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
          选择一份或多份已导入文档后，可基于文档内容进行问答。
        </p>
      </div>
      <button class="btn-ghost btn-sm" :disabled="isLoading" @click="loadDocuments">
        {{ isLoading ? '刷新中...' : '刷新文档' }}
      </button>
    </div>

    <div v-if="errorMessage" class="mb-4 rounded-lg border border-error/20 bg-red-50 px-4 py-3 text-sm text-error dark:bg-red-900/20 dark:text-red-300">
      {{ errorMessage }}
    </div>

    <div class="mb-6 rounded-lg border border-gray-200 p-4 dark:border-gray-700">
      <div class="mb-3 flex items-center justify-between gap-3">
        <div>
          <h4 class="font-medium text-gray-800 dark:text-gray-100">选择文档</h4>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            已选择 {{ selectedCount }} 份{{ availableDocuments.length ? ` / 共 ${availableDocuments.length} 份` : '' }}
          </p>
        </div>
        <button
          v-if="selectedDocumentIds.length > 0"
          class="btn-ghost btn-sm"
          @click="clearSelection"
        >
          清空选择
        </button>
      </div>

      <div v-if="isLoading" class="py-6 text-center text-sm text-gray-500 dark:text-gray-400">
        正在加载文档...
      </div>
      <div v-else-if="availableDocuments.length === 0" class="py-6 text-center text-sm text-gray-500 dark:text-gray-400">
        暂无可选文档，请先到文档库导入文件。
      </div>
      <div v-else class="max-h-64 space-y-2 overflow-y-auto pr-1">
        <label
          v-for="doc in availableDocuments"
          :key="doc.id"
          class="flex cursor-pointer items-start gap-3 rounded-lg border border-gray-200 px-3 py-3 transition-colors hover:border-primary-300 hover:bg-gray-50 dark:border-gray-700 dark:hover:border-primary-700 dark:hover:bg-gray-800"
        >
          <input
            :checked="selectedDocumentIds.includes(doc.id)"
            class="mt-1 h-4 w-4"
            type="checkbox"
            @change="toggleDocument(doc.id)"
          />
          <div class="min-w-0 flex-1">
            <div class="truncate text-sm font-medium text-gray-800 dark:text-gray-100">{{ doc.name }}</div>
            <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">
              {{ normalizeType(doc.type) }}
              <span v-if="doc.issuer"> · {{ doc.issuer }}</span>
              <span v-if="doc.created_at"> · {{ formatDate(doc.created_at) }}</span>
            </div>
            <p class="mt-2 text-sm text-gray-600 dark:text-gray-300">
              {{ doc.text_content?.slice(0, 120) || '暂无预览内容' }}{{ doc.text_content ? '...' : '' }}
            </p>
          </div>
        </label>
      </div>
    </div>

    <div v-if="selectedDocuments.length > 0" class="mb-4 flex flex-wrap gap-2">
      <span
        v-for="doc in selectedDocuments"
        :key="doc.id"
        class="inline-flex items-center gap-2 rounded-full bg-primary-50 px-3 py-1 text-sm text-primary-700 dark:bg-primary-900/20 dark:text-primary-300"
      >
        <span class="max-w-56 truncate">{{ doc.name }}</span>
        <button class="text-primary-600 hover:text-primary-800 dark:text-primary-300 dark:hover:text-primary-100" @click="toggleDocument(doc.id)">
          ×
        </button>
      </span>
    </div>

    <ChatPanel
      :selected-documents="selectedDocumentIds"
      :system-prompt="systemPrompt"
    />
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import ChatPanel from '@/components/ChatPanel.vue'
import { documentApi } from '@/api'
import { useAppStore } from '@/stores/app'

const appStore = useAppStore()
const isLoading = ref(false)
const errorMessage = ref('')
const selectedDocumentIds = ref<string[]>([])

const availableDocuments = computed(() => appStore.documents)
const selectedDocuments = computed(() => {
  const idSet = new Set(selectedDocumentIds.value)
  return appStore.documents.filter(doc => idSet.has(doc.id))
})
const selectedCount = computed(() => selectedDocumentIds.value.length)
const systemPrompt = computed(() => {
  return '你是政策分析助手。回答时请优先依据当前问题附带的文档上下文；如果文档中没有足够信息，请明确说明，不要编造。'
})

onMounted(() => {
  loadDocuments()
})

async function loadDocuments() {
  isLoading.value = true
  errorMessage.value = ''

  try {
    const docs = await documentApi.list(appStore.currentWorkspace)
    appStore.setDocuments(docs)
    const availableIds = new Set(docs.map(doc => doc.id))
    selectedDocumentIds.value = selectedDocumentIds.value.filter(id => availableIds.has(id))
  } catch (error) {
    console.error('加载 AI 问答文档失败:', error)
    errorMessage.value = error instanceof Error ? error.message : '加载文档失败'
  } finally {
    isLoading.value = false
  }
}

function toggleDocument(id: string) {
  if (selectedDocumentIds.value.includes(id)) {
    selectedDocumentIds.value = selectedDocumentIds.value.filter(item => item !== id)
    return
  }

  selectedDocumentIds.value = [...selectedDocumentIds.value, id]
}

function clearSelection() {
  selectedDocumentIds.value = []
}

function normalizeType(type: string): string {
  return type.replace('.', '').toUpperCase()
}

function formatDate(date: string): string {
  return new Date(date).toLocaleDateString('zh-CN')
}
</script>
