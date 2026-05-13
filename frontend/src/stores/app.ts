import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

// 文档接口
export interface Document {
  id: string
  name: string
  original_name: string
  type: string
  size: number
  text_content: string
  workspace: string
  tags: string[]
  source: string
  issuer: string
  publish_date?: string | null
  effective_date?: string | null
  created_at: string
  updated_at: string
}

// 工作流接口
export interface Workflow {
  id: string
  name: string
  description: string
  version: string
  inputs: InputDef[]
  steps: number
}

export interface InputDef {
  name: string
  type: string
  required: boolean
  description: string
}

export interface WorkflowDefinitionDetail {
  info: Workflow
  content: string
  source: string
}

export interface StepResult {
  id: string
  name: string
  status: 'pending' | 'running' | 'completed' | 'failed' | 'skipped'
  output?: Record<string, unknown>
  error?: string
  duration?: number
}

// AI提供商接口
export interface Provider {
  name: string
  type: string
  enabled: boolean
  default_model: string
  models: Model[]
}

export interface Model {
  id: string
  name: string
  max_tokens: number
  context_window: number
}

// 分析记录接口
export interface AnalysisRecord {
  id: string
  workflow_id: string
  workflow_name: string
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled'
  created_at: string
  started_at: string | null
  completed_at: string | null
  summary: string
}

export interface AnalysisResult {
  id: string
  workflow_id: string
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled'
  steps?: StepResult[]
  step_results: Record<string, StepResult>
  outputs: Record<string, unknown>
  summary: string
  metrics?: {
    duration?: number
    tokens_input?: number
    tokens_output?: number
    api_calls?: number
  }
  created_at: string
  completed_at?: string
}

// 应用状态Store
export const useAppStore = defineStore('app', () => {
  // 状态
  const isLoading = ref(false)
  const error = ref<string | null>(null)
  const currentWorkspace = ref('default')

  // 文档相关
  const documents = ref<Document[]>([])
  const selectedDocuments = ref<string[]>([])

  // 工作流相关
  const workflows = ref<Workflow[]>([])
  const currentWorkflow = ref<string | null>(null)

  // AI相关
  const providers = ref<Provider[]>([])
  const selectedProvider = ref<string>('DeepSeek')
  const selectedModel = ref<string>('deepseek-chat')

  // 分析相关
  const analysisHistory = ref<AnalysisRecord[]>([])
  const currentAnalysis = ref<string | null>(null)

  // 计算属性
  const documentCount = computed(() => documents.value.length)
  const hasSelection = computed(() => selectedDocuments.value.length > 0)

  // 操作方法
  function setLoading(value: boolean) {
    isLoading.value = value
  }

  function setError(value: string | null) {
    error.value = value
  }

  function setDocuments(docs: Document[]) {
    documents.value = docs
  }

  function selectDocument(id: string) {
    if (!selectedDocuments.value.includes(id)) {
      selectedDocuments.value.push(id)
    }
  }

  function deselectDocument(id: string) {
    const index = selectedDocuments.value.indexOf(id)
    if (index > -1) {
      selectedDocuments.value.splice(index, 1)
    }
  }

  function clearSelection() {
    selectedDocuments.value = []
  }

  function setWorkflows(wfs: Workflow[]) {
    workflows.value = wfs
  }

  function setProviders(provs: Provider[]) {
    providers.value = provs
  }

  function setAnalysisHistory(history: AnalysisRecord[]) {
    analysisHistory.value = history
  }

  return {
    // 状态
    isLoading,
    error,
    currentWorkspace,
    documents,
    selectedDocuments,
    workflows,
    currentWorkflow,
    providers,
    selectedProvider,
    selectedModel,
    analysisHistory,
    currentAnalysis,

    // 计算属性
    documentCount,
    hasSelection,

    // 方法
    setLoading,
    setError,
    setDocuments,
    selectDocument,
    deselectDocument,
    clearSelection,
    setWorkflows,
    setProviders,
    setAnalysisHistory
  }
})
