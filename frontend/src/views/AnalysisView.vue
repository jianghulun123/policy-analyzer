<template>
  <div class="analysis-view">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between mb-6">
      <div>
        <h2 class="text-2xl font-bold text-gray-800 dark:text-gray-100">分析历史</h2>
        <p class="text-gray-500 dark:text-gray-400 text-sm mt-1">查看历史分析记录</p>
      </div>
      <div class="flex items-center gap-3">
        <button
          class="btn-ghost btn-sm"
          @click="loadHistory"
          :disabled="isLoading"
        >
          <svg class="w-4 h-4 mr-1" :class="{ 'animate-spin': isLoading }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          刷新
        </button>
        <router-link to="/analysis/new" class="btn-primary">
          <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          新建分析
        </router-link>
      </div>
    </div>

    <!-- 删除确认对话框 -->
    <div v-if="showDeleteConfirm" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div class="bg-white dark:bg-gray-800 rounded-lg p-6 max-w-md w-full mx-4 shadow-xl border border-gray-200 dark:border-gray-700">
        <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-100 mb-2">确认删除</h3>
        <p class="text-gray-600 dark:text-gray-300 mb-4">确定要删除此分析记录吗？此操作不可撤销。</p>
        <div class="flex justify-end gap-3">
          <button class="btn-ghost" @click="showDeleteConfirm = false">取消</button>
          <button class="btn-danger" @click="confirmDelete" :disabled="isDeleting">
            {{ isDeleting ? '删除中...' : '确认删除' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 分析记录列表 -->
    <div v-if="isLoading" class="text-center py-16">
      <svg class="w-8 h-8 mx-auto text-primary-500 animate-spin mb-4" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
      </svg>
      <p class="text-gray-500 dark:text-gray-400">加载中...</p>
    </div>

    <div v-else-if="error" class="text-center py-16">
      <svg class="w-16 h-16 mx-auto text-error mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
      </svg>
      <h3 class="text-lg font-medium text-gray-600 dark:text-gray-300 mb-2">加载失败</h3>
      <p class="text-gray-500 dark:text-gray-400 mb-4">{{ error }}</p>
      <button class="btn-primary" @click="loadHistory">重试</button>
    </div>

    <div v-else-if="records.length > 0" class="space-y-4">
      <div
        v-for="record in records"
        :key="record.id"
        class="card hover:shadow-md transition-shadow"
      >
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-4 cursor-pointer flex-1" @click="viewAnalysis(record.id)">
            <div class="w-10 h-10 rounded-lg flex items-center justify-center" :class="getStatusColor(record.status)">
              <span class="status-dot" :class="getStatusDotClass(record.status)"></span>
            </div>
            <div>
              <h3 class="font-medium text-gray-800 dark:text-gray-100">{{ record.workflow_name }}</h3>
              <p class="text-sm text-gray-500 dark:text-gray-400">{{ formatDate(record.created_at) }}</p>
            </div>
          </div>
          <div class="flex items-center gap-3">
            <span class="tag" :class="getStatusTagClass(record.status)">
              {{ getStatusText(record.status) }}
            </span>
            <button
              class="p-2 text-gray-400 hover:text-error hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors"
              title="删除"
              @click.stop="promptDelete(record)"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
            </button>
            <svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
            </svg>
          </div>
        </div>

        <p v-if="record.summary" class="text-sm text-gray-600 dark:text-gray-300 mt-3 line-clamp-2 cursor-pointer" @click="viewAnalysis(record.id)">
          {{ record.summary }}
        </p>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-else class="text-center py-16">
      <svg class="w-16 h-16 mx-auto text-gray-300 dark:text-gray-600 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.95-.083-1.867-.548-2.659a5 5 0 01-7.072 0l.547-.548A3.374 3.374 0 015.469 12H6a2 2 0 114 0v.531c0 .95.083 1.867.548 2.659z" />
      </svg>
      <h3 class="text-lg font-medium text-gray-600 dark:text-gray-300 mb-2">暂无分析记录</h3>
      <p class="text-gray-500 dark:text-gray-400 mb-4">开始您的第一次政策分析</p>
      <router-link to="/analysis/new" class="btn-primary">
        <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        新建分析
      </router-link>
    </div>

  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { workflowApi } from '@/api'

interface ExecutionRecord {
  id: string
  workflow_id: string
  workflow_name: string
  status: string
  created_at: string
  started_at: string | null
  completed_at: string | null
  summary: string
}

const router = useRouter()

const records = ref<ExecutionRecord[]>([])
const isLoading = ref(false)
const error = ref<string | null>(null)
const showDeleteConfirm = ref(false)
const recordToDelete = ref<ExecutionRecord | null>(null)
const isDeleting = ref(false)

onMounted(async () => {
  await loadHistory()
})

async function loadHistory() {
  isLoading.value = true
  error.value = null
  try {
    const result = await workflowApi.getHistory('default')
    records.value = (result || []) as ExecutionRecord[]
  } catch (e: any) {
    console.error('加载分析历史失败:', e)
    error.value = e.message || '加载失败'
  } finally {
    isLoading.value = false
  }
}

function formatDate(date: string): string {
  return new Date(date).toLocaleString('zh-CN')
}

function getStatusColor(status: string): string {
  const colors: Record<string, string> = {
    completed: 'bg-success/10',
    running: 'bg-primary-500/10',
    failed: 'bg-error/10',
    pending: 'bg-gray-100'
  }
  return colors[status] || 'bg-gray-100'
}

function getStatusDotClass(status: string): string {
  const classes: Record<string, string> = {
    completed: 'status-success',
    running: 'status-running',
    failed: 'status-error',
    pending: 'status-pending'
  }
  return classes[status] || 'status-pending'
}

function getStatusText(status: string): string {
  const texts: Record<string, string> = {
    completed: '已完成',
    running: '进行中',
    failed: '失败',
    pending: '等待中'
  }
  return texts[status] || status
}

function getStatusTagClass(status: string): string {
  const classes: Record<string, string> = {
    completed: 'tag-success',
    running: 'tag-primary',
    failed: 'tag-error',
    pending: 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300'
  }
  return classes[status] || 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300'
}

function viewAnalysis(id: string) {
  router.push(`/analysis/${id}`)
}

function promptDelete(record: ExecutionRecord) {
  recordToDelete.value = record
  showDeleteConfirm.value = true
}

async function confirmDelete() {
  if (!recordToDelete.value) return

  isDeleting.value = true
  try {
    await workflowApi.deleteRecord(recordToDelete.value.id)
    // 从列表中移除
    records.value = records.value.filter(r => r.id !== recordToDelete.value!.id)
    showDeleteConfirm.value = false
    recordToDelete.value = null
  } catch (e: any) {
    console.error('删除分析记录失败:', e)
    error.value = `删除失败: ${e.message || '未知错误'}`
  } finally {
    isDeleting.value = false
  }
}
</script>