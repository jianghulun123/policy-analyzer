<template>
  <div class="analysis-detail-view">
    <!-- 加载中 -->
    <div v-if="isLoading" class="text-center py-16">
      <svg class="w-8 h-8 mx-auto text-primary-500 animate-spin mb-4" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
      </svg>
      <p class="text-gray-500 dark:text-gray-400">加载分析结果...</p>
    </div>

    <!-- 错误状态 -->
    <div v-else-if="error" class="text-center py-16">
      <svg class="w-16 h-16 mx-auto text-error mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
      </svg>
      <h3 class="text-lg font-medium text-gray-600 dark:text-gray-300 mb-2">加载失败</h3>
      <p class="text-gray-500 dark:text-gray-400 mb-4">{{ error }}</p>
      <div class="flex justify-center gap-3">
        <button class="btn-primary" @click="loadResult">重试</button>
        <button class="btn-secondary" @click="$router.back()">返回</button>
      </div>
    </div>

    <!-- 正常内容 -->
    <template v-else-if="result">
      <!-- 页面标题 -->
      <div class="mb-6">
        <button class="btn-ghost btn-sm mb-4" @click="$router.back()">
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
          </svg>
          返回
        </button>
        <div class="flex items-center justify-between">
          <div>
            <h2 class="text-2xl font-bold text-gray-800 dark:text-gray-100">分析结果</h2>
            <p class="text-gray-500 dark:text-gray-400 text-sm mt-1">
              {{ result.workflow_id }} ·
              <span :class="getStatusTextClass(result.status)">{{ getStatusText(result.status) }}</span>
              <span v-if="result.completed_at"> · 完成于 {{ formatDate(result.completed_at) }}</span>
            </p>
          </div>
          <div class="flex gap-2">
            <!-- 导出下拉菜单 -->
            <div class="relative">
              <button class="btn-primary" @click="toggleExportMenu">
                <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                </svg>
                导出报告
                <svg class="w-4 h-4 ml-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                </svg>
              </button>
              <div v-if="showExportMenu" class="absolute right-0 mt-2 w-40 bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-100 dark:border-gray-700 py-1 z-10">
                <button class="w-full px-4 py-2 text-left text-sm text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-700 flex items-center" @click="exportReportWithMenu('pdf')">
                  <svg class="w-4 h-4 mr-2 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
                  </svg>
                  PDF 文档
                </button>
                <button class="w-full px-4 py-2 text-left text-sm text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-700 flex items-center" @click="exportReportWithMenu('markdown')">
                  <svg class="w-4 h-4 mr-2 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                  Markdown
                </button>
                <button class="w-full px-4 py-2 text-left text-sm text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-700 flex items-center" @click="exportReportWithMenu('txt')">
                  <svg class="w-4 h-4 mr-2 text-gray-500 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                  纯文本
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 执行摘要 - 始终显示，增强容错 -->
      <div v-if="!hasStructuredData" class="card mb-6 bg-primary-50 dark:bg-primary-900/20 border-primary-100 dark:border-primary-800">
        <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-100 mb-3">执行摘要</h3>
        <div class="markdown-body prose max-w-none text-gray-700 dark:text-gray-300 dark:prose-invert" v-html="renderedSummary"></div>
      </div>

      <!-- 执行指标 -->
      <div class="card mb-6" v-if="result.metrics">
        <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-100 mb-4">执行指标</h3>
        <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div class="text-center p-3 bg-gray-50 dark:bg-gray-800 rounded">
            <p class="text-2xl font-bold text-primary-600">{{ result.metrics.duration || 0 }}</p>
            <p class="text-sm text-gray-500 dark:text-gray-400">执行时长(秒)</p>
          </div>
          <div class="text-center p-3 bg-gray-50 dark:bg-gray-800 rounded">
            <p class="text-2xl font-bold text-primary-600">{{ result.metrics.tokens_input || 0 }}</p>
            <p class="text-sm text-gray-500 dark:text-gray-400">输入Token</p>
          </div>
          <div class="text-center p-3 bg-gray-50 dark:bg-gray-800 rounded">
            <p class="text-2xl font-bold text-primary-600">{{ result.metrics.tokens_output || 0 }}</p>
            <p class="text-sm text-gray-500 dark:text-gray-400">输出Token</p>
          </div>
          <div class="text-center p-3 bg-gray-50 dark:bg-gray-800 rounded">
            <p class="text-2xl font-bold text-primary-600">{{ result.metrics.api_calls || 0 }}</p>
            <p class="text-sm text-gray-500 dark:text-gray-400">API调用</p>
          </div>
        </div>
      </div>

      <!-- 步骤结果 -->
      <div class="card mb-6" v-if="displaySteps.length > 0">
        <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-100 mb-4">步骤执行详情</h3>
        <div class="space-y-3">
          <div
            v-for="step in displaySteps"
            :key="step.id"
            class="border border-gray-100 dark:border-gray-700 rounded-lg p-4"
          >
            <div class="flex items-center justify-between mb-2">
              <span class="font-medium text-gray-800 dark:text-gray-100">{{ step.name || step.id }}</span>
              <div class="flex items-center gap-2">
                <span v-if="step.duration" class="text-sm text-gray-500 dark:text-gray-400">{{ step.duration }}ms</span>
                <span class="tag" :class="getStatusTagClass(step.status)">{{ getStatusText(step.status) }}</span>
              </div>
            </div>
            <p v-if="step.error" class="text-sm text-error">{{ step.error }}</p>
          </div>
        </div>
      </div>

      <!-- 结构化报告视图 - 如果有结构化数据则优先显示 -->
      <StructuredReportView
        v-if="hasStructuredData"
        :report="structuredReport"
        :fallback-summary="displaySummary"
      />

      <!-- Markdown 报告视图 -->
      <div v-else-if="hasReportContent" class="card">
        <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-100 mb-4">分析报告</h3>
        <div class="markdown-body prose max-w-none dark:prose-invert" v-html="renderedReport"></div>
      </div>

      <!-- 无数据提示 -->
      <div v-else class="card text-center py-8">
        <svg class="w-12 h-12 mx-auto text-gray-300 dark:text-gray-600 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
        </svg>
        <p class="text-gray-500 dark:text-gray-400">暂无分析报告内容</p>
        <p class="text-gray-400 dark:text-gray-500 text-sm mt-1">请检查分析任务是否正常完成</p>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { useRoute } from 'vue-router'
import { workflowApi, subscribeWorkflowEvents } from '@/api'
import StructuredReportView from '@/components/StructuredReportView.vue'
import type { StructuredReport } from '@/types/report'
import { renderMarkdown, renderMermaidInDOM } from '@/utils/markdown'
import type { AnalysisResult, StepResult } from '@/api'

const route = useRoute()

const result = ref<AnalysisResult | null>(null)
const isLoading = ref(false)
const error = ref<string | null>(null)
const reportContent = ref('')
const structuredReport = ref<StructuredReport | null>(null)
const showExportMenu = ref(false)
let unsubscribe: (() => void) | null = null

// 增强容错：判断是否有结构化数据（即使部分字段缺失）
const hasStructuredData = computed(() => {
  if (!structuredReport.value) return false
  const r = structuredReport.value
  // 只要有关键字段就认为有数据
  return !!(
    r.executive_summary ||
    (r.headline_metrics && r.headline_metrics.length > 0) ||
    (r.comparison_table && r.comparison_table.length > 0) ||
    (r.trend_points && r.trend_points.length > 0) ||
    (r.sections && (r.sections.overview || r.sections.key_changes || r.sections.impacts || r.sections.actions))
  )
})

// 判断是否有报告内容
const hasReportContent = computed(() => {
  return !!reportContent.value && reportContent.value.trim().length > 0
})

// 显示摘要 - 多级fallback
const displaySummary = computed(() => {
  if (structuredReport.value?.executive_summary) {
    return structuredReport.value.executive_summary
  }
  if (result.value?.summary) {
    return result.value.summary
  }
  return '暂无摘要'
})

const renderedReport = computed(() => {
  if (!reportContent.value) return ''
  return renderMarkdown(reportContent.value)
})

const renderedSummary = computed(() => renderMarkdown(displaySummary.value))

const displaySteps = computed<StepResult[]>(() => {
  if (!result.value) return []

  if (Array.isArray(result.value.steps) && result.value.steps.length > 0) {
    return result.value.steps.map(step => ({
      ...step,
      ...(result.value?.step_results?.[step.id] || {})
    }))
  }

  return Object.entries(result.value.step_results || {}).map(([stepId, step]) => ({
    id: step.id || stepId,
    name: step.name || stepId,
    status: step.status,
    output: step.output,
    error: step.error,
    duration: step.duration
  }))
})

onMounted(async () => {
  await loadResult()
  setupEventListeners()
  // 点击外部关闭下拉菜单
  document.addEventListener('click', handleClickOutside)
  // 渲染 Mermaid 图表
  await nextTick()
  renderMermaidInDOM()
})

// 监听报告内容变化，重新渲染 Mermaid
watch(renderedReport, async () => {
  await nextTick()
  renderMermaidInDOM()
})

onUnmounted(() => {
  if (unsubscribe) {
    unsubscribe()
  }
  document.removeEventListener('click', handleClickOutside)
})

function handleClickOutside(e: MouseEvent) {
  const target = e.target as HTMLElement
  if (!target.closest('.relative')) {
    showExportMenu.value = false
  }
}

async function loadResult() {
  const analysisId = route.params.id as string
  if (!analysisId) {
    error.value = '无效的分析ID'
    return
  }

  isLoading.value = true
  error.value = null

  try {
    const data = await workflowApi.getResult(analysisId) as AnalysisResult
    result.value = data

    const outputs = data.outputs || {}
    
    // 增强容错：尝试多种方式解析结构化报告
    let structured: StructuredReport | null = null
    for (const candidate of [outputs.report_structured, outputs.report, outputs.comparison]) {
      structured = tryParseStructuredReport(candidate)
      if (structured) {
        break
      }
    }
    structuredReport.value = structured

    // 多级fallback获取报告内容
    reportContent.value = toDisplayText(outputs.comparison)
      || toDisplayText(outputs.report)
      || toDisplayText(outputs.summary)
      || data.summary
      || ''

    // 如果结果正在运行，订阅更新
    if (data.status === 'running') {
      setupEventListeners()
    }
  } catch (e: any) {
    console.error('加载分析结果失败:', e)
    error.value = e.message || '加载失败'
  } finally {
    isLoading.value = false
  }
}

function setupEventListeners() {
  if (unsubscribe) return

  unsubscribe = subscribeWorkflowEvents({
    onComplete: (event: any) => {
      if (event.execution_id === route.params.id) {
        loadResult()
      }
    },
    onStepComplete: (event: any) => {
      if (event.execution_id === route.params.id && result.value) {
        // 更新步骤结果
        if (!result.value.step_results) {
          result.value.step_results = {}
        }
        result.value.step_results[event.step_id] = event.step_result
      }
    }
  })
}

function toggleExportMenu() {
  showExportMenu.value = !showExportMenu.value
}

function exportReportWithMenu(format: string) {
  showExportMenu.value = false
  exportReport(format)
}

function buildDownloadBlob(data: string, format: string): Blob {
  if (format === 'pdf') {
    const binary = window.atob(data)
    const bytes = new Uint8Array(binary.length)
    for (let i = 0; i < binary.length; i++) {
      bytes[i] = binary.charCodeAt(i)
    }
    return new Blob([bytes], { type: 'application/pdf' })
  }

  const mimeTypes: Record<string, string> = {
    markdown: 'text/markdown;charset=utf-8',
    md: 'text/markdown;charset=utf-8',
    txt: 'text/plain;charset=utf-8',
    text: 'text/plain;charset=utf-8'
  }
  return new Blob([data], { type: mimeTypes[format] || 'text/plain;charset=utf-8' })
}

async function exportReport(format: string) {
  if (!result.value) return

  try {
    const data = await workflowApi.exportReport(result.value.id, format)
    console.log('导出数据类型:', typeof data, '长度:', data?.length)

    if (!data || data.length === 0) {
      alert('导出失败: 返回数据为空')
      return
    }

    const blob = buildDownloadBlob(data as string, format)
    console.log('Blob 大小:', blob.size, '类型:', blob.type)

    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url

    // 根据格式设置文件扩展名
    const extensions: Record<string, string> = {
      markdown: 'md',
      md: 'md',
      pdf: 'pdf',
      txt: 'txt',
      text: 'txt'
    }
    const ext = extensions[format] || 'txt'
    a.download = `analysis-${result.value.id}.${ext}`

    console.log('下载文件:', a.download)
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  } catch (e: any) {
    console.error('导出报告失败:', e)
    alert('导出失败: ' + (e.message || '未知错误'))
  }
}

function formatDate(dateStr: string): string {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleString('zh-CN')
}

function tryParseStructuredReport(value: unknown): StructuredReport | null {
  if (!value) return null

  if (typeof value === 'object') {
    return value as StructuredReport
  }

  if (typeof value !== 'string') {
    return null
  }

  const trimmed = value.trim()
  if (!trimmed.startsWith('{') || !trimmed.endsWith('}')) {
    return null
  }

  try {
    return JSON.parse(trimmed) as StructuredReport
  } catch (e) {
    console.warn('解析结构化报告JSON失败，将回退为Markdown展示:', e)
    return null
  }
}

function toDisplayText(value: unknown): string {
  return typeof value === 'string' ? value : ''
}

function getStatusText(status: string): string {
  const texts: Record<string, string> = {
    completed: '已完成',
    running: '进行中',
    failed: '失败',
    cancelled: '已取消',
    pending: '等待中',
    skipped: '已跳过'
  }
  return texts[status] || status
}

function getStatusTextClass(status: string): string {
  const classes: Record<string, string> = {
    completed: 'text-success',
    running: 'text-primary-600',
    failed: 'text-error',
    cancelled: 'text-gray-500 dark:text-gray-400',
    pending: 'text-gray-500 dark:text-gray-400'
  }
  return classes[status] || 'text-gray-500 dark:text-gray-400'
}

function getStatusTagClass(status: string): string {
  const classes: Record<string, string> = {
    completed: 'tag-success',
    running: 'tag-primary',
    failed: 'tag-error',
    cancelled: 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300',
    pending: 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300',
    skipped: 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300'
  }
  return classes[status] || 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300'
}
</script>
