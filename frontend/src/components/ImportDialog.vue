<template>
  <div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="$emit('close')">
    <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl w-full max-w-md animate-slide-up border border-gray-200 dark:border-gray-700">
      <!-- 头部 -->
      <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
        <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-100">导入文档</h3>
      </div>

      <!-- 内容 -->
      <div class="px-6 py-4">
        <!-- 文件选择 -->
        <div
          class="border-2 border-dashed border-gray-300 dark:border-gray-600 rounded-lg p-8 text-center hover:border-primary-400 transition-colors cursor-pointer"
          :class="{ 'border-primary-500 bg-primary-50 dark:bg-primary-900/20': isDragging }"
          @dragover.prevent="isDragging = true"
          @dragleave.prevent="isDragging = false"
          @drop.prevent="handleDrop"
          @click="selectFiles"
        >
          <svg class="w-12 h-12 mx-auto text-gray-400 dark:text-gray-500 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
          </svg>
          <p class="text-gray-600 dark:text-gray-300 mb-1">拖拽文件到此处，或点击选择</p>
          <p class="text-xs text-gray-400 dark:text-gray-500">支持 PDF、Word、TXT、Markdown 格式</p>
        </div>

        <!-- 已选文件 -->
        <div v-if="selectedFiles.length > 0" class="mt-4">
          <h4 class="text-sm font-medium text-gray-700 dark:text-gray-200 mb-2">已选择文件</h4>
          <ul class="space-y-2">
            <li v-for="(file, index) in selectedFiles" :key="file.path" class="flex items-center justify-between p-2 bg-gray-50 dark:bg-gray-900 rounded">
              <div class="flex items-center gap-2 min-w-0">
                <svg class="w-4 h-4 text-gray-400 dark:text-gray-500 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                <span class="text-sm text-gray-700 dark:text-gray-300 truncate">{{ file.name }}</span>
                <span class="text-xs text-gray-400 dark:text-gray-500 shrink-0">{{ formatSize(file.size) }}</span>
              </div>
              <button @click="removeFile(index)" class="text-gray-400 hover:text-error">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </li>
          </ul>
        </div>

        <div v-if="errorMessage" class="mt-4 rounded-lg border border-error/20 bg-red-50 dark:bg-red-900/30 px-3 py-2 text-sm text-error dark:text-red-300">
          {{ errorMessage }}
        </div>

        <!-- 标签输入 -->
        <div class="mt-4">
          <label class="text-sm font-medium text-gray-700 dark:text-gray-200 mb-1 block">标签（可选）</label>
          <input
            type="text"
            v-model="tags"
            placeholder="输入标签，用逗号分隔"
            class="input text-sm"
          />
        </div>
      </div>

      <!-- 底部按钮 -->
      <div class="px-6 py-4 border-t border-gray-200 dark:border-gray-700 flex justify-end gap-3">
        <button class="btn-secondary btn-sm" @click="$emit('close')">取消</button>
        <button
          class="btn-primary btn-sm"
          :disabled="selectedFiles.length === 0 || isImporting"
          @click="importFiles"
        >
          {{ isImporting ? '导入中...' : '开始导入' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { documentApi } from '@/api'

// 日志工具
const log = {
  info: (message: string, data?: Record<string, unknown>) => {
    console.log(`[ImportDialog] ${message}`, data || '')
  },
  warn: (message: string, data?: Record<string, unknown>) => {
    console.warn(`[ImportDialog] ${message}`, data || '')
  },
  error: (message: string, data?: Record<string, unknown>) => {
    console.error(`[ImportDialog] ${message}`, data || '')
  }
}

const props = withDefaults(defineProps<{
  workspaceId?: string
}>(), {
  workspaceId: 'default'
})

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'imported'): void
}>()

interface SelectedFile {
  name: string
  size: number
  path: string
}

const isDragging = ref(false)
const selectedFiles = ref<SelectedFile[]>([])
const tags = ref('')
const isImporting = ref(false)
const errorMessage = ref('')

function getErrorMessage(error: unknown): string {
  if (error instanceof Error && error.message) {
    return error.message
  }

  if (typeof error === 'string' && error) {
    return error
  }

  if (error && typeof error === 'object') {
    const errorObj = error as Record<string, unknown>

    const directMessage = errorObj.message
    if (typeof directMessage === 'string' && directMessage) {
      return directMessage
    }

    const nestedMessage = errorObj.err
    if (typeof nestedMessage === 'string' && nestedMessage) {
      return nestedMessage
    }

    const cause = errorObj.cause
    if (cause && typeof cause === 'object' && typeof (cause as Record<string, unknown>).message === 'string') {
      return (cause as Record<string, unknown>).message as string
    }
  }

  return '导入失败，请检查文件格式'
}

async function selectFiles() {
  log.info('打开文件选择对话框')
  try {
    const paths = await documentApi.select()
    log.info('文件选择结果', { paths, count: paths?.length || 0 })
    appendFiles(paths || [])
  } catch (error) {
    log.error('选择文件失败', { error })
    errorMessage.value = '选择文件失败，请重试'
  }
}

function appendFiles(paths: string[]) {
  log.info('添加文件到列表', { paths, count: paths.length })
  const existing = new Set(selectedFiles.value.map(file => file.path))

  for (const path of paths) {
    if (existing.has(path)) {
      log.info('文件已存在，跳过', { path })
      continue
    }

    const segments = path.split(/[/\\]/)
    const name = segments[segments.length - 1] || path
    selectedFiles.value.push({ name, size: 0, path })
    existing.add(path)
    log.info('文件已添加', { name, path })
  }
}

function handleDrop(event: DragEvent) {
  isDragging.value = false
  console.log('[ImportDialog] 拖拽事件触发', event)

  const files = Array.from(event.dataTransfer?.files || [])
  console.log('[ImportDialog] 拖拽文件数量:', files.length)

  if (files.length === 0) {
    errorMessage.value = '拖拽未检测到文件，请使用点击选择方式'
    console.warn('[ImportDialog] 拖拽事件中没有检测到文件')
    return
  }

  // 浏览器环境下 File 对象没有 path 属性，需要使用文件选择对话框
  // Wails 桌面应用也不支持直接从拖拽获取文件路径
  errorMessage.value = '拖拽导入暂不支持，请点击选择文件'
  console.warn('[ImportDialog] 浏览器安全限制，无法获取拖拽文件的本地路径')
}

function removeFile(index: number) {
  selectedFiles.value.splice(index, 1)
}

function formatSize(bytes: number): string {
  if (!bytes) return '待解析'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

async function importFiles() {
  if (selectedFiles.value.length === 0) {
    log.warn('没有选择任何文件')
    return
  }

  log.info('开始导入文件', { files: selectedFiles.value.map(f => f.name) })
  isImporting.value = true
  errorMessage.value = ''

  try {
    const parsedTags = tags.value
      .split(',')
      .map(tag => tag.trim())
      .filter(Boolean)

    log.info('解析标签完成', { tags: parsedTags })

    for (const file of selectedFiles.value) {
      log.info('导入单个文件', { name: file.name, path: file.path, tags: parsedTags })
      try {
        await documentApi.import(file.path, props.workspaceId, parsedTags)
        log.info('文件导入成功', { name: file.name })
      } catch (fileError) {
        log.error('单个文件导入失败', { name: file.name, error: fileError })
        throw fileError
      }
    }

    log.info('所有文件导入成功', { count: selectedFiles.value.length })
    emit('imported')
    emit('close')
  } catch (error) {
    const message = getErrorMessage(error)
    log.error('导入失败', { error, errorMessage: message })
    errorMessage.value = message
  } finally {
    isImporting.value = false
  }
}
</script>
