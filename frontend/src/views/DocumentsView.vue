<template>
  <div class="documents-view">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h2 class="text-2xl font-bold text-gray-800 dark:text-gray-100">文档库</h2>
        <p class="text-gray-500 dark:text-gray-400 text-sm mt-1">管理政策文件和文档</p>
      </div>
      <button class="btn-primary" @click="showImportDialog = true">
        <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        导入文档
      </button>
    </div>

    <div v-if="successMessage" class="mb-6 rounded-lg border border-green-200 dark:border-green-800 bg-green-50 dark:bg-green-900/30 px-4 py-3 text-sm text-green-700 dark:text-green-300">
      {{ successMessage }}
    </div>

    <div class="flex flex-col gap-4 mb-6 md:flex-row">
      <div class="flex-1 relative">
        <input
          v-model="searchQuery"
          type="text"
          placeholder="搜索文档标题、内容、标签..."
          class="input pl-10"
        />
        <svg class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400 dark:text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
      </div>
      <select v-model="filterType" class="select w-full md:w-40">
        <option value="">全部类型</option>
        <option value="pdf">PDF</option>
        <option value="docx">Word</option>
        <option value="txt">文本</option>
        <option value="md">Markdown</option>
      </select>
    </div>

    <div class="mb-4 flex items-center justify-between text-sm text-gray-500 dark:text-gray-400">
      <span>共 {{ appStore.documents.length }} 份文档，当前展示 {{ documents.length }} 份</span>
      <button class="btn-ghost btn-sm" :disabled="isLoading" @click="loadDocuments">
        {{ isLoading ? '刷新中...' : '刷新列表' }}
      </button>
    </div>

    <div v-if="isLoading" class="py-16 text-center text-gray-500 dark:text-gray-400">正在加载文档...</div>

    <div v-else-if="errorMessage" class="mb-6 rounded-lg border border-error/20 bg-red-50 dark:bg-red-900/30 px-4 py-3 text-sm text-error dark:text-red-300">
      {{ errorMessage }}
    </div>

    <div v-else-if="hasActiveFilter && documents.length === 0" class="text-center py-16">
      <svg class="w-16 h-16 mx-auto text-gray-300 dark:text-gray-600 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-4.35-4.35m1.85-5.65a7.5 7.5 0 11-15 0 7.5 7.5 0 0115 0z" />
      </svg>
      <h3 class="text-lg font-medium text-gray-600 dark:text-gray-300 mb-2">未找到匹配文档</h3>
      <p class="text-gray-500 dark:text-gray-400 mb-4">请调整搜索关键词或文件类型筛选条件</p>
      <button class="btn-secondary" @click="resetFilters">清除筛选</button>
    </div>

    <div v-else-if="documents.length > 0" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <div
        v-for="doc in documents"
        :key="doc.id"
        class="card-hover cursor-pointer group"
        @click="viewDocument(doc.id)"
      >
        <div class="flex items-start gap-3 mb-3">
          <div class="w-10 h-10 rounded-lg flex items-center justify-center" :class="getTypeColor(doc.type)">
            <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
          </div>
          <div class="flex-1 min-w-0">
            <h3 class="font-medium text-gray-800 dark:text-gray-100 truncate">{{ doc.name }}</h3>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ normalizeType(doc.type) }} · {{ formatSize(doc.size) }} · {{ formatDate(doc.created_at) }}</p>
          </div>
        </div>

        <div class="space-y-2 mb-3 text-sm text-gray-600 dark:text-gray-300">
          <p class="truncate-2">{{ doc.text_content?.slice(0, 150) || '暂无预览内容' }}{{ doc.text_content ? '...' : '' }}</p>
          <p v-if="doc.issuer" class="truncate"><span class="text-gray-400 dark:text-gray-500">机构：</span>{{ doc.issuer }}</p>
          <p v-if="doc.source" class="truncate"><span class="text-gray-400 dark:text-gray-500">来源：</span>{{ doc.source }}</p>
        </div>

        <div v-if="doc.tags?.length" class="flex flex-wrap gap-1">
          <span v-for="tag in doc.tags" :key="tag" class="tag-primary">
            {{ tag }}
          </span>
        </div>
        <div v-else class="text-xs text-gray-400 dark:text-gray-500">暂无标签</div>

        <div class="flex justify-end gap-2 mt-3 opacity-0 group-hover:opacity-100 transition-opacity">
          <button class="btn-ghost btn-sm" :disabled="deletingId === doc.id" @click.stop="deleteDocument(doc.id)">
            <svg class="w-4 h-4 text-error" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
          </button>
        </div>
      </div>
    </div>

    <div v-else class="text-center py-16">
      <svg class="w-16 h-16 mx-auto text-gray-300 dark:text-gray-600 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
      </svg>
      <h3 class="text-lg font-medium text-gray-600 dark:text-gray-300 mb-2">暂无文档</h3>
      <p class="text-gray-500 dark:text-gray-400 mb-4">导入您的第一个政策文件开始分析</p>
      <button class="btn-primary" @click="showImportDialog = true">
        <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        导入文档
      </button>
    </div>

    <ImportDialog
      v-if="showImportDialog"
      :workspace-id="appStore.currentWorkspace"
      @close="showImportDialog = false"
      @imported="handleImported"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import ImportDialog from '@/components/ImportDialog.vue'
import { documentApi } from '@/api'
import { useAppStore } from '@/stores/app'

const router = useRouter()
const appStore = useAppStore()
const showImportDialog = ref(false)
const searchQuery = ref('')
const filterType = ref('')
const isLoading = ref(false)
const errorMessage = ref('')
const successMessage = ref('')
const deletingId = ref('')

const hasActiveFilter = computed(() => Boolean(searchQuery.value.trim() || filterType.value))

const documents = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()

  return appStore.documents.filter(doc => {
    const matchesType = !filterType.value || doc.type.replace('.', '') === filterType.value
    const matchesQuery = !query ||
      doc.name.toLowerCase().includes(query) ||
      doc.text_content?.toLowerCase().includes(query) ||
      doc.source?.toLowerCase().includes(query) ||
      doc.issuer?.toLowerCase().includes(query) ||
      doc.tags.some(tag => tag.toLowerCase().includes(query))

    return matchesType && matchesQuery
  })
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
  } catch (error) {
    console.error('加载文档失败:', error)
    errorMessage.value = error instanceof Error ? error.message : '加载文档失败'
  } finally {
    isLoading.value = false
  }
}

async function handleImported() {
  showImportDialog.value = false
  successMessage.value = '文档导入成功，列表已刷新'
  await loadDocuments()
}

function resetFilters() {
  searchQuery.value = ''
  filterType.value = ''
}

function normalizeType(type: string): string {
  return type.replace('.', '').toUpperCase()
}

function getTypeColor(type: string): string {
  const normalizedType = type.replace('.', '')
  const colors: Record<string, string> = {
    pdf: 'bg-error',
    docx: 'bg-primary-600',
    txt: 'bg-gray-500',
    md: 'bg-gray-700'
  }
  return colors[normalizedType] || 'bg-gray-500'
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function formatDate(date: string): string {
  return new Date(date).toLocaleDateString('zh-CN')
}

function viewDocument(id: string) {
  router.push(`/documents/${id}`)
}

async function deleteDocument(id: string) {
  if (deletingId.value) return

  deletingId.value = id
  errorMessage.value = ''
  successMessage.value = ''

  try {
    await documentApi.delete(id)
    appStore.setDocuments(appStore.documents.filter(d => d.id !== id))
    successMessage.value = '文档已删除'
  } catch (error) {
    console.error('删除文档失败:', error)
    errorMessage.value = error instanceof Error ? error.message : '删除文档失败'
  } finally {
    deletingId.value = ''
  }
}
</script>
