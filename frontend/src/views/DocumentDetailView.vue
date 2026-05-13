<template>
  <div class="document-detail-view">
    <button class="btn-ghost btn-sm mb-4" @click="$router.back()">
      <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
      </svg>
      返回文档列表
    </button>

    <div v-if="isLoading" class="card text-center text-gray-500 dark:text-gray-400">正在加载文档详情...</div>

    <div v-else-if="errorMessage" class="rounded-lg border border-error/20 bg-red-50 dark:bg-red-900/30 px-4 py-3 text-sm text-error dark:text-red-300">
      {{ errorMessage }}
    </div>

    <template v-else-if="document">
      <div class="card mb-6">
        <div class="flex items-start justify-between gap-4">
          <div>
            <h1 class="text-2xl font-bold text-gray-800 dark:text-gray-100 mb-2">{{ document.name }}</h1>
            <div class="flex flex-wrap items-center gap-4 text-sm text-gray-500 dark:text-gray-400">
              <span>{{ document.type.replace('.', '').toUpperCase() }}</span>
              <span>{{ formatSize(document.size || 0) }}</span>
              <span>创建于 {{ formatDate(document.created_at || '') }}</span>
              <span v-if="document.publish_date">发布日期：{{ formatDate(document.publish_date) }}</span>
              <span v-if="document.effective_date">生效日期：{{ formatDate(document.effective_date) }}</span>
            </div>
          </div>
          <div class="flex gap-2">
            <button
              v-if="!isEditing"
              class="btn-secondary btn-sm"
              @click="startEditing"
            >
              编辑元数据
            </button>
            <button class="btn-primary btn-sm" @click="startAnalysis">
              <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.95-.083-1.867-.548-2.659a5 5 0 01-7.072 0l.547-.548A3.374 3.374 0 015.469 12H6a2 2 0 114 0v.531c0 .95.083 1.867.548 2.659z" />
              </svg>
              开始分析
            </button>
          </div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mt-6">
          <div>
            <label class="text-sm font-medium text-gray-700 dark:text-gray-200 block mb-1">来源</label>
            <input
              v-if="isEditing"
              v-model="editForm.source"
              type="text"
              class="input text-sm"
              placeholder="输入文档来源"
            />
            <p v-else class="text-sm text-gray-600 dark:text-gray-200 min-h-[2.5rem] rounded-lg border border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-900 px-3 py-2">
              {{ document.source || '未填写' }}
            </p>
          </div>
          <div>
            <label class="text-sm font-medium text-gray-700 dark:text-gray-200 block mb-1">发布机构</label>
            <input
              v-if="isEditing"
              v-model="editForm.issuer"
              type="text"
              class="input text-sm"
              placeholder="输入发布机构"
            />
            <p v-else class="text-sm text-gray-600 dark:text-gray-200 min-h-[2.5rem] rounded-lg border border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-900 px-3 py-2">
              {{ document.issuer || '未填写' }}
            </p>
          </div>
        </div>

        <div v-if="isEditing" class="flex items-center justify-end gap-3 mt-4">
          <button class="btn-secondary btn-sm" :disabled="isSaving" @click="cancelEditing">取消</button>
          <button class="btn-primary btn-sm" :disabled="isSaving" @click="saveMetadata">
            {{ isSaving ? '保存中...' : '保存元数据' }}
          </button>
        </div>

        <div v-if="saveMessage" class="mt-4 rounded-lg border border-green-200 dark:border-green-800 bg-green-50 dark:bg-green-900/30 px-4 py-3 text-sm text-green-700 dark:text-green-300">
          {{ saveMessage }}
        </div>

        <div v-if="document.tags?.length" class="flex gap-2 mt-4 flex-wrap">
          <span v-for="tag in document.tags" :key="tag" class="tag-primary">
            {{ tag }}
          </span>
        </div>
      </div>

      <div class="card">
        <h2 class="text-lg font-semibold text-gray-800 dark:text-gray-100 mb-4">文档内容</h2>
        <pre class="document-content whitespace-pre-wrap break-words text-sm text-gray-700 dark:text-gray-300">{{ content || '暂无内容' }}</pre>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { documentApi } from '@/api'
import { useAppStore, type Document } from '@/stores/app'

const route = useRoute()
const router = useRouter()
const appStore = useAppStore()
const document = ref<Document | null>(null)
const content = ref('')
const isLoading = ref(false)
const isSaving = ref(false)
const isEditing = ref(false)
const errorMessage = ref('')
const saveMessage = ref('')
const editForm = reactive({
  source: '',
  issuer: ''
})

onMounted(() => {
  loadDocumentDetail()
})

async function loadDocumentDetail() {
  isLoading.value = true
  errorMessage.value = ''
  saveMessage.value = ''

  try {
    const docId = String(route.params.id)
    const detail = await documentApi.get(docId)
    document.value = detail
    syncDocumentToStore(detail)
    resetEditForm(detail)
    content.value = await documentApi.getContent(detail.id)
  } catch (error) {
    console.error('加载文档详情失败:', error)
    errorMessage.value = error instanceof Error ? error.message : '加载文档详情失败'
  } finally {
    isLoading.value = false
  }
}

function resetEditForm(doc: Document) {
  editForm.source = doc.source || ''
  editForm.issuer = doc.issuer || ''
}

function syncDocumentToStore(doc: Document) {
  const nextDocuments = [...appStore.documents]
  const index = nextDocuments.findIndex(item => item.id === doc.id)

  if (index >= 0) {
    nextDocuments[index] = doc
  } else {
    nextDocuments.unshift(doc)
  }

  appStore.setDocuments(nextDocuments)
}

function startEditing() {
  if (!document.value) return
  resetEditForm(document.value)
  saveMessage.value = ''
  isEditing.value = true
}

function cancelEditing() {
  if (document.value) {
    resetEditForm(document.value)
  }
  isEditing.value = false
}

async function saveMetadata() {
  if (!document.value || isSaving.value) return

  isSaving.value = true
  errorMessage.value = ''
  saveMessage.value = ''

  try {
    await documentApi.updateMeta(document.value.id, {
      source: editForm.source.trim(),
      issuer: editForm.issuer.trim()
    })

    const detail = await documentApi.get(document.value.id)
    document.value = detail
    syncDocumentToStore(detail)
    resetEditForm(detail)
    isEditing.value = false
    saveMessage.value = '元数据已保存'
  } catch (error) {
    console.error('保存文档元数据失败:', error)
    errorMessage.value = error instanceof Error ? error.message : '保存文档元数据失败'
  } finally {
    isSaving.value = false
  }
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function formatDate(date: string | null | undefined): string {
  if (!date) return '—'
  return new Date(date).toLocaleDateString('zh-CN')
}

function startAnalysis() {
  router.push('/analysis/new')
}
</script>
