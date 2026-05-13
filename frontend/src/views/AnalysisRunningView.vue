<template>
  <div class="analysis-running-view">
    <!-- 页面标题 -->
    <div class="mb-6">
      <button class="btn-ghost btn-sm mb-4" @click="confirmCancel">
        <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
        </svg>
        返回
      </button>
      <div class="flex items-center justify-between">
        <div>
          <h2 class="text-2xl font-bold text-gray-800 dark:text-gray-100">分析执行中</h2>
          <p class="text-gray-500 dark:text-gray-400 text-sm mt-1">{{ workflowName }}</p>
        </div>
        <div class="flex gap-2">
          <button
            v-if="status === 'running'"
            class="btn-secondary text-error border-error hover:bg-error hover:text-white"
            @click="cancelExecution"
          >
            <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
            取消执行
          </button>
        </div>
      </div>
    </div>

    <!-- 执行进度 -->
    <div class="card mb-6">
      <div class="flex items-center justify-between mb-4">
        <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-100">执行进度</h3>
        <div class="flex items-center gap-2">
          <span class="status-dot" :class="statusDotClass"></span>
          <span class="text-sm" :class="statusTextClass">{{ statusText }}</span>
        </div>
      </div>

      <!-- 进度条 -->
      <div class="mb-4">
        <div class="h-2 bg-gray-100 dark:bg-gray-700 rounded-full overflow-hidden">
          <div
            class="h-full transition-all duration-500"
            :class="progressBarClass"
            :style="{ width: progressPercent + '%' }"
          ></div>
        </div>
        <div class="flex justify-between text-sm text-gray-500 dark:text-gray-400 mt-1">
          <span>{{ completedSteps }} / {{ totalSteps }} 步骤</span>
          <span v-if="duration > 0">{{ duration }}秒</span>
        </div>
      </div>

      <!-- 步骤列表 -->
      <div class="space-y-3">
        <div
          v-for="step in steps"
          :key="step.id"
          class="flex items-center gap-3 p-3 rounded-lg"
          :class="getStepBgClass(step.status)"
        >
          <div class="w-6 h-6 rounded-full flex items-center justify-center" :class="getStepIconClass(step.status)">
            <svg v-if="step.status === 'completed'" class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
            </svg>
            <svg v-else-if="step.status === 'failed'" class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
            <svg v-else-if="step.status === 'running'" class="w-4 h-4 text-white animate-spin" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            <span v-else class="text-sm font-medium text-gray-400 dark:text-gray-500">{{ step.index }}</span>
          </div>
          <div class="flex-1">
            <p class="font-medium" :class="getStepTextClass(step.status)">{{ step.name }}</p>
            <p v-if="step.error" class="text-sm text-error mt-1">{{ step.error }}</p>
            <p v-if="step.duration && step.status === 'completed'" class="text-xs text-gray-500 dark:text-gray-400 mt-1">
              耗时 {{ step.duration }}ms
            </p>
          </div>
        </div>
      </div>
    </div>

    <!-- 完成后操作 -->
    <div v-if="status === 'completed'" class="card bg-success/5 border-success/20 dark:bg-green-900/20 dark:border-green-800">
      <div class="flex items-center gap-4">
        <div class="w-12 h-12 rounded-full bg-success flex items-center justify-center">
          <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
          </svg>
        </div>
        <div class="flex-1">
          <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-100">分析完成</h3>
          <p class="text-sm text-gray-500 dark:text-gray-400">共耗时 {{ duration }} 秒</p>
        </div>
        <div class="flex gap-2">
          <button class="btn-secondary" @click="exportReport">
            <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
            </svg>
            导出报告
          </button>
          <button class="btn-primary" @click="viewResult">
            查看结果
          </button>
        </div>
      </div>
    </div>

    <!-- 失败提示 -->
    <div v-else-if="status === 'failed'" class="card bg-error/5 border-error/20 dark:bg-red-900/20 dark:border-red-800">
      <div class="flex items-center gap-4">
        <div class="w-12 h-12 rounded-full bg-error flex items-center justify-center">
          <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </div>
        <div class="flex-1">
          <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-100">分析失败</h3>
          <p class="text-sm text-error">{{ errorMessage || '执行过程中发生错误' }}</p>
        </div>
        <button class="btn-secondary" @click="$router.push('/analysis/new')">
          重新分析
        </button>
      </div>
    </div>

    <!-- 取消提示 -->
    <div v-else-if="status === 'cancelled'" class="card bg-gray-50 dark:bg-gray-800 border-gray-200 dark:border-gray-700">
      <div class="flex items-center gap-4">
        <div class="w-12 h-12 rounded-full bg-gray-400 flex items-center justify-center">
          <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636" />
          </svg>
        </div>
        <div class="flex-1">
          <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-100">已取消</h3>
          <p class="text-sm text-gray-500 dark:text-gray-400">分析已取消执行</p>
        </div>
        <button class="btn-secondary" @click="$router.push('/analysis/new')">
          重新分析
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { workflowApi, subscribeWorkflowEvents } from '@/api'
import type { AnalysisResult, StepResult, WorkflowDefinitionDetail } from '@/api'

interface StepInfo {
  id: string
  name: string
  status: 'pending' | 'running' | 'completed' | 'failed' | 'skipped'
  index: number
  error?: string
  duration?: number
}

const route = useRoute()
const router = useRouter()

const execID = ref('')
const workflowID = ref('')
const workflowName = ref('分析任务')
const status = ref<'pending' | 'running' | 'completed' | 'failed' | 'cancelled'>('running')
const steps = ref<StepInfo[]>([])
const duration = ref(0)
const errorMessage = ref('')
const startTime = ref<Date | null>(null)

let unsubscribe: (() => void) | null = null
let durationTimer: ReturnType<typeof setInterval> | null = null

const totalSteps = computed(() => steps.value.length)
const completedSteps = computed(() => steps.value.filter(step => step.status === 'completed').length)

const progressPercent = computed(() => {
  if (totalSteps.value === 0) return 0
  return Math.round((completedSteps.value / totalSteps.value) * 100)
})

const statusDotClass = computed(() => {
  const classes: Record<string, string> = {
    pending: 'status-pending',
    running: 'status-running',
    completed: 'status-success',
    failed: 'status-error',
    cancelled: 'status-pending'
  }
  return classes[status.value] || 'status-pending'
})

const statusTextClass = computed(() => {
  const classes: Record<string, string> = {
    pending: 'text-gray-500 dark:text-gray-400',
    running: 'text-primary-600 dark:text-primary-300',
    completed: 'text-success',
    failed: 'text-error',
    cancelled: 'text-gray-500 dark:text-gray-400'
  }
  return classes[status.value] || 'text-gray-500 dark:text-gray-400'
})

const statusText = computed(() => {
  const texts: Record<string, string> = {
    pending: '等待中',
    running: '执行中',
    completed: '已完成',
    failed: '失败',
    cancelled: '已取消'
  }
  return texts[status.value] || status.value
})

const progressBarClass = computed(() => {
  if (status.value === 'failed') return 'bg-error'
  if (status.value === 'completed') return 'bg-success'
  return 'bg-primary-500'
})

onMounted(async () => {
  execID.value = (route.query.execID as string) || ''
  workflowID.value = (route.query.workflow as string) || ''
  workflowName.value = workflowID.value || '分析任务'

  if (!execID.value) {
    router.push('/analysis')
    return
  }

  await initializeSteps()

  startTime.value = new Date()
  durationTimer = setInterval(updateDuration, 1000)

  setupEventListeners()
  await checkCurrentStatus()
})

onUnmounted(() => {
  if (unsubscribe) {
    unsubscribe()
  }
  if (durationTimer) {
    clearInterval(durationTimer)
  }
})

function getDefaultSteps(): StepInfo[] {
  // 与后端 standard-compare 工作流定义保持一致 (engine.go createStandardCompareWorkflow)
  return [
    { id: 'read_old', name: '读取旧版政策', status: 'pending', index: 1 },
    { id: 'read_new', name: '读取新版政策', status: 'pending', index: 2 },
    { id: 'generate_full_report', name: '生成完整分析报告', status: 'pending', index: 3 },
    { id: 'summarize', name: '摘要生成', status: 'pending', index: 4 },
    { id: 'format', name: '报告格式化', status: 'pending', index: 5 }
  ]
}

async function initializeSteps() {
  if (!workflowID.value) {
    steps.value = getDefaultSteps()
    return
  }

  try {
    const definition = await workflowApi.getDefinition(workflowID.value) as WorkflowDefinitionDetail
    const workflowSteps = extractStepsFromDefinition(definition)
    steps.value = workflowSteps.length > 0 ? workflowSteps : getDefaultSteps()
    if (definition?.info?.name) {
      workflowName.value = definition.info.name
    }
  } catch (e) {
    console.warn('加载工作流定义失败，使用默认步骤', e)
    steps.value = getDefaultSteps()
  }
}

function extractStepsFromDefinition(definition: WorkflowDefinitionDetail | null | undefined): StepInfo[] {
  const declaredSteps = Number(definition?.info?.steps || 0)
  if (declaredSteps > 0 && declaredSteps <= 1 && (definition?.source === 'standard-compare' || workflowID.value === 'standard-compare')) {
    return getDefaultSteps()
  }

  const content = definition?.content || ''
  const lines = content.split(/\r?\n/)
  const result: StepInfo[] = []
  let currentId = ''
  let currentName = ''
  let inStepsSection = false

  for (const rawLine of lines) {
    const line = rawLine.trim()
    if (!line) continue

    if (!inStepsSection) {
      if (line === 'steps:') {
        inStepsSection = true
      }
      continue
    }

    if (!rawLine.startsWith(' ') && !rawLine.startsWith('-')) {
      break
    }

    if (line.startsWith('- id:')) {
      if (currentId) {
        result.push({
          id: currentId,
          name: currentName || currentId,
          status: 'pending',
          index: result.length + 1
        })
      }
      currentId = line.slice('- id:'.length).trim()
      currentName = ''
      continue
    }

    if (line.startsWith('id:')) {
      if (currentId) {
        result.push({
          id: currentId,
          name: currentName || currentId,
          status: 'pending',
          index: result.length + 1
        })
      }
      currentId = line.slice('id:'.length).trim()
      currentName = ''
      continue
    }

    if (line.startsWith('name:')) {
      currentName = line.slice('name:'.length).trim().replace(/^['"]|['"]$/g, '')
    }
  }

  if (currentId) {
    result.push({
      id: currentId,
      name: currentName || currentId,
      status: 'pending',
      index: result.length + 1
    })
  }

  return result
}

function updateDuration() {
  if (startTime.value && status.value === 'running') {
    duration.value = Math.floor((Date.now() - startTime.value.getTime()) / 1000)
  }
}

function setupEventListeners() {
  unsubscribe = subscribeWorkflowEvents({
    onStart: (event: any) => {
      if (event.execution_id === execID.value) {
        workflowName.value = event.workflow_name || workflowName.value
        status.value = 'running'
      }
    },
    onStepStart: (event: any) => {
      if (event.execution_id === execID.value) {
        updateStepStatus(event.step_id, 'running', event.step_name)
      }
    },
    onStepComplete: (event: any) => {
      if (event.execution_id === execID.value) {
        const stepResult = event.step_result
        updateStepStatus(
          event.step_id,
          stepResult?.status || 'completed',
          stepResult?.name,
          stepResult?.error,
          stepResult?.duration
        )
      }
    },
    onComplete: (event: any) => {
      if (event.execution_id === execID.value) {
        const result = event.result as AnalysisResult | undefined
        mergeStepsFromResult(result)
        status.value = result?.status || 'completed'
        if (durationTimer) {
          clearInterval(durationTimer)
        }
        if (result?.metrics?.duration) {
          duration.value = result.metrics.duration
        }
      }
    },
    onError: (event: any) => {
      if (event.execution_id === execID.value && !event.step_id) {
        status.value = 'failed'
        errorMessage.value = event.error || '执行失败'
        if (durationTimer) {
          clearInterval(durationTimer)
        }
      }
    }
  })
}

async function checkCurrentStatus() {
  try {
    const result = await workflowApi.getResult(execID.value) as AnalysisResult
    if (result) {
      mergeStepsFromResult(result)
      status.value = result.status as any
      if (result.status === 'completed' || result.status === 'failed' || result.status === 'cancelled') {
        if (durationTimer) {
          clearInterval(durationTimer)
        }
      }
      if (result.step_results) {
        for (const [stepId, stepResult] of Object.entries(result.step_results)) {
          updateStepStatus(
            stepId,
            (stepResult as any).status,
            (stepResult as any).name,
            (stepResult as any).error,
            (stepResult as any).duration
          )
        }
      }
    }
  } catch (e) {
    console.error('检查执行状态失败:', e)
  }
}

function mergeStepsFromResult(result?: AnalysisResult | null) {
  if (!result) return

  if (Array.isArray(result.steps) && result.steps.length > 0) {
    steps.value = result.steps.map((step, index) => ({
      id: step.id,
      name: step.name || step.id,
      status: normalizeStepStatus(step.status),
      index: index + 1,
      error: step.error || undefined,
      duration: step.duration || undefined
    }))
  }

  if (result.step_results) {
    for (const [stepId, stepResult] of Object.entries(result.step_results)) {
      updateStepStatus(
        stepId,
        (stepResult as StepResult).status,
        (stepResult as StepResult).name,
        (stepResult as StepResult).error,
        (stepResult as StepResult).duration
      )
    }
  }
}

function normalizeStepStatus(statusValue?: string): StepInfo['status'] {
  switch (statusValue) {
    case 'running':
    case 'completed':
    case 'failed':
    case 'skipped':
      return statusValue
    default:
      return 'pending'
  }
}

function updateStepStatus(
  stepId: string,
  stepStatus: string,
  stepName?: string,
  error?: string,
  duration?: number
) {
  const stepIndex = steps.value.findIndex(s => s.id === stepId)
  if (stepIndex >= 0) {
    steps.value[stepIndex].status = normalizeStepStatus(stepStatus)
    if (stepName) steps.value[stepIndex].name = stepName
    steps.value[stepIndex].error = error || undefined
    steps.value[stepIndex].duration = duration || undefined
  } else {
    // 动态添加未知步骤，确保进度可见
    // 先检查是否已存在同名步骤，避免重复
    const existingByName = steps.value.findIndex(s => s.name === stepName && s.id !== stepId)
    if (existingByName >= 0 && stepName) {
      // 更新现有步骤的 ID
      steps.value[existingByName].id = stepId
      steps.value[existingByName].status = normalizeStepStatus(stepStatus)
      steps.value[existingByName].error = error || undefined
      steps.value[existingByName].duration = duration || undefined
    } else {
      steps.value.push({
        id: stepId,
        name: stepName || stepId,
        status: normalizeStepStatus(stepStatus),
        index: steps.value.length + 1,
        error,
        duration
      })
    }
  }
}

async function cancelExecution() {
  if (!execID.value) return
  try {
    await workflowApi.cancel(execID.value)
    status.value = 'cancelled'
    if (durationTimer) {
      clearInterval(durationTimer)
    }
  } catch (e: any) {
    console.error('取消执行失败:', e)
  }
}

function confirmCancel() {
  if (status.value === 'running') {
    if (confirm('分析正在执行中，确定要返回吗？')) {
      router.push('/analysis')
    }
  } else {
    router.push('/analysis')
  }
}

function buildMarkdownDownloadBlob(data: string): Blob {
  return new Blob([data], { type: 'text/markdown;charset=utf-8' })
}

async function exportReport() {
  if (!execID.value) return
  try {
    const data = await workflowApi.exportReport(execID.value, 'markdown')
    console.log('导出数据类型:', typeof data, '长度:', data?.length)

    if (!data || data.length === 0) {
      alert('导出失败: 返回数据为空')
      return
    }

    const blob = buildMarkdownDownloadBlob(data as string)
    console.log('Blob 大小:', blob.size)

    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `analysis-${execID.value}.md`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  } catch (e: any) {
    console.error('导出报告失败:', e)
    alert('导出失败: ' + (e.message || '未知错误'))
  }
}

function viewResult() {
  router.push(`/analysis/${execID.value}`)
}

function getStepBgClass(stepStatus: string): string {
  const classes: Record<string, string> = {
    pending: 'bg-gray-50 dark:bg-gray-800',
    running: 'bg-primary-50 dark:bg-primary-900/20',
    completed: 'bg-success/5 dark:bg-green-900/20',
    failed: 'bg-error/5 dark:bg-red-900/20',
    skipped: 'bg-gray-50 dark:bg-gray-800'
  }
  return classes[stepStatus] || 'bg-gray-50 dark:bg-gray-800'
}

function getStepIconClass(stepStatus: string): string {
  const classes: Record<string, string> = {
    pending: 'bg-gray-200',
    running: 'bg-primary-500',
    completed: 'bg-success',
    failed: 'bg-error',
    skipped: 'bg-gray-300'
  }
  return classes[stepStatus] || 'bg-gray-200'
}

function getStepTextClass(stepStatus: string): string {
  const classes: Record<string, string> = {
    pending: 'text-gray-500 dark:text-gray-300',
    running: 'text-primary-700 dark:text-primary-300',
    completed: 'text-gray-800 dark:text-gray-100',
    failed: 'text-error',
    skipped: 'text-gray-400 dark:text-gray-500'
  }
  return classes[stepStatus] || 'text-gray-500 dark:text-gray-300'
}
</script>
