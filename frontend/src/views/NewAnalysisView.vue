<template>
  <div class="new-analysis-view">
    <div class="mb-6">
      <button class="btn-ghost btn-sm mb-4" @click="$router.back()">
        <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
        </svg>
        返回
      </button>
      <h2 class="text-2xl font-bold text-gray-800 dark:text-gray-100">新建分析</h2>
      <p class="text-gray-500 dark:text-gray-400 text-sm mt-1">选择文档和工作流进行分析</p>
    </div>

    <div class="flex items-center gap-4 mb-8">
      <div v-for="(step, index) in steps" :key="index" class="flex items-center">
        <div
          class="w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium"
          :class="currentStep >= index ? 'bg-primary-600 text-white' : 'bg-gray-200 dark:bg-gray-700 text-gray-500 dark:text-gray-300'"
        >
          {{ index + 1 }}
        </div>
        <span class="ml-2 text-sm" :class="currentStep >= index ? 'text-gray-800 dark:text-gray-100' : 'text-gray-400 dark:text-gray-500'">
          {{ step }}
        </span>
        <div v-if="index < steps.length - 1" class="w-12 h-0.5 mx-4" :class="currentStep > index ? 'bg-primary-600' : 'bg-gray-200 dark:bg-gray-700'"></div>
      </div>
    </div>

    <div v-if="currentStep === 0" class="card">
      <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-100 mb-4">选择要分析的文档</h3>

      <div class="grid grid-cols-2 gap-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-2">旧版政策（基准）</label>
          <div
            class="border-2 border-dashed rounded-lg p-6 text-center cursor-pointer transition-colors"
            :class="selectedDocs.old ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20' : 'border-gray-300 dark:border-gray-600 hover:border-primary-400'"
            @click="openDocSelector('old')"
          >
            <div v-if="selectedDocs.old">
              <p class="font-medium text-gray-800 dark:text-gray-100">{{ selectedDocs.old.name }}</p>
              <p class="text-sm text-gray-500 dark:text-gray-400">{{ formatSize(selectedDocs.old.size) }}</p>
              <p class="text-xs text-primary-600 mt-2">点击重新选择已上传文档</p>
            </div>
            <div v-else>
              <svg class="w-8 h-8 mx-auto text-gray-400 dark:text-gray-500 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
              <p class="text-gray-600 dark:text-gray-300">点击选择已上传的旧版政策</p>
            </div>
          </div>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-2">新版政策（目标）</label>
          <div
            class="border-2 border-dashed rounded-lg p-6 text-center cursor-pointer transition-colors"
            :class="selectedDocs.new ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20' : 'border-gray-300 dark:border-gray-600 hover:border-primary-400'"
            @click="openDocSelector('new')"
          >
            <div v-if="selectedDocs.new">
              <p class="font-medium text-gray-800 dark:text-gray-100">{{ selectedDocs.new.name }}</p>
              <p class="text-sm text-gray-500 dark:text-gray-400">{{ formatSize(selectedDocs.new.size) }}</p>
              <p class="text-xs text-primary-600 mt-2">点击重新选择已上传文档</p>
            </div>
            <div v-else>
              <svg class="w-8 h-8 mx-auto text-gray-400 dark:text-gray-500 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
              <p class="text-gray-600 dark:text-gray-300">点击选择已上传的新版政策</p>
            </div>
          </div>
        </div>
      </div>

      <p v-if="documents.length === 0" class="text-sm text-amber-600 dark:text-amber-400 mt-4">默认工作空间暂无文档，请先导入至少两份文档。</p>

      <div class="flex justify-end mt-6">
        <button
          class="btn-primary"
          :disabled="!selectedDocs.old || !selectedDocs.new"
          @click="currentStep = 1"
        >
          下一步
        </button>
      </div>
    </div>

    <div v-if="currentStep === 1" class="card">
      <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-100 mb-4">选择分析工作流</h3>

      <div class="space-y-3">
        <div
          v-for="workflow in workflows"
          :key="workflow.id"
          class="border rounded-lg p-4 cursor-pointer transition-colors"
          :class="selectedWorkflow === workflow.id ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20' : 'border-gray-200 dark:border-gray-700 hover:border-primary-400'"
          @click="selectedWorkflow = workflow.id"
        >
          <div class="flex items-center gap-3">
            <div
              class="w-5 h-5 rounded-full border-2 flex items-center justify-center"
              :class="selectedWorkflow === workflow.id ? 'border-primary-500' : 'border-gray-300 dark:border-gray-600'"
            >
              <div v-if="selectedWorkflow === workflow.id" class="w-2.5 h-2.5 rounded-full bg-primary-500"></div>
            </div>
            <div>
              <h4 class="font-medium text-gray-800 dark:text-gray-100">{{ workflow.name }}</h4>
              <p class="text-sm text-gray-500 dark:text-gray-400">{{ workflow.description }}</p>
            </div>
          </div>
        </div>
      </div>

      <div class="flex justify-between mt-6">
        <button class="btn-secondary" @click="currentStep = 0">上一步</button>
        <button class="btn-primary" :disabled="!selectedWorkflow" @click="currentStep = 2">
          下一步
        </button>
      </div>
    </div>

    <div v-if="currentStep === 2" class="card">
      <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-100 mb-4">配置分析参数</h3>

      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-2">AI提供商</label>
          <select v-model="selectedProvider" class="select" @change="syncSelectedModel">
            <option v-for="provider in providers" :key="provider.name" :value="provider.name">
              {{ provider.name }}
            </option>
          </select>
          <p v-if="providers.length === 0" class="text-xs text-amber-600 mt-2">暂无可用 AI 提供商，请先在设置中启用并保存。</p>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-2">AI模型</label>
          <select v-model="selectedModel" class="select" :disabled="currentModels.length === 0">
            <option v-for="model in currentModels" :key="model.id" :value="model.id">
              {{ model.name }}
            </option>
          </select>
          <p v-if="selectedProvider && currentModels.length === 0" class="text-xs text-amber-600 mt-2">当前提供商暂无可用模型，请先到设置页添加自定义模型或获取模型列表。</p>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-2">重点关注领域（可选）</label>
          <div class="flex flex-wrap gap-2">
            <label v-for="area in focusAreas" :key="area" class="flex items-center">
              <input type="checkbox" :value="area" v-model="selectedAreas" class="rounded border-gray-300 text-primary-600 mr-2" />
              <span class="text-sm text-gray-700 dark:text-gray-300">{{ area }}</span>
            </label>
          </div>
        </div>
      </div>

      <div class="flex justify-between mt-6">
        <button class="btn-secondary" @click="currentStep = 1">上一步</button>
        <button class="btn-primary" :disabled="!selectedDocs.old || !selectedDocs.new || !selectedWorkflow || !selectedProvider || !selectedModel" @click="startAnalysis">
          开始分析
        </button>
      </div>
    </div>

    <div v-if="isDocSelectorOpen" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="closeDocSelector">
      <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl w-full max-w-2xl max-h-[80vh] overflow-hidden animate-slide-up border border-gray-200 dark:border-gray-700">
        <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between">
          <div>
            <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-100">选择已上传文档</h3>
            <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">当前选择：{{ docSelectorType === 'old' ? '旧版政策（基准）' : '新版政策（目标）' }}</p>
          </div>
          <button @click="closeDocSelector" class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div class="p-6 border-b border-gray-100 dark:border-gray-700">
          <input
            v-model="docSearchKeyword"
            type="text"
            class="input"
            placeholder="输入文档名称搜索"
          />
        </div>

        <div class="p-6 overflow-y-auto max-h-[50vh] space-y-3">
          <div v-if="availableDocuments.length === 0" class="text-sm text-gray-500 dark:text-gray-400 text-center py-8">
            没有可供选择的已上传文档
          </div>

          <button
            v-for="doc in availableDocuments"
            :key="doc.id"
            type="button"
            class="w-full text-left border border-gray-200 dark:border-gray-700 rounded-lg p-4 transition-colors hover:border-primary-400 hover:bg-primary-50 dark:hover:bg-primary-900/20"
            @click="chooseDocument(doc)"
          >
            <div class="flex items-start justify-between gap-4">
              <div class="min-w-0">
                <p class="font-medium text-gray-800 dark:text-gray-100 break-all">{{ doc.name }}</p>
                <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">{{ formatSize(doc.size) }}</p>
              </div>
              <span class="text-xs text-primary-600 dark:text-primary-300 whitespace-nowrap">选择此文档</span>
            </div>
          </button>
        </div>

        <div class="px-6 py-4 border-t border-gray-200 dark:border-gray-700 flex justify-end">
          <button class="btn-secondary btn-sm" @click="closeDocSelector">取消</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { appApi, documentApi, workflowApi, type Document, type Workflow } from '../api'
import type { config } from '../../wailsjs/go/models'

interface ProviderOption {
  name: string
  defaultModel: string
  models: Array<{ id: string; name: string; maxTokens: number; contextWindow: number }>
}

const route = useRoute()
const router = useRouter()

const steps = ['选择文档', '选择工作流', '配置参数']
const currentStep = ref(0)
const selectedDocs = ref<{ old: Document | null; new: Document | null }>({ old: null, new: null })
const selectedWorkflow = ref('standard-compare')
const selectedProvider = ref('')
const selectedModel = ref('')
const selectedAreas = ref<string[]>([])
const documents = ref<Document[]>([])
const workflows = ref<Workflow[]>([])
const providers = ref<ProviderOption[]>([])
const focusAreas = ['条款变化', '执行力度', '处罚标准', '适用范围', '执行时间']
const isDocSelectorOpen = ref(false)
const docSelectorType = ref<'old' | 'new'>('old')
const docSearchKeyword = ref('')

const currentModels = computed(() => {
  const provider = providers.value.find(item => item.name === selectedProvider.value)
  return provider?.models || []
})

const availableDocuments = computed(() => {
  const keyword = docSearchKeyword.value.trim().toLowerCase()
  const excludedId = docSelectorType.value === 'old' ? selectedDocs.value.new?.id : selectedDocs.value.old?.id

  return documents.value.filter(doc => {
    if (excludedId && doc.id === excludedId) {
      return false
    }
    if (!keyword) {
      return true
    }
    return doc.name.toLowerCase().includes(keyword) || doc.original_name.toLowerCase().includes(keyword)
  })
})

onMounted(async () => {
  const [workflowList, documentList, cfg] = await Promise.all([
    workflowApi.list(),
    documentApi.list('default'),
    appApi.getConfig()
  ])

  workflows.value = workflowList
  documents.value = documentList
  providers.value = mapEnabledProviders(cfg)

  if (route.query.workflow) {
    selectedWorkflow.value = route.query.workflow as string
  } else if (workflowList.length > 0) {
    selectedWorkflow.value = workflowList[0].id
  }

  initProviderAndModel()
})

function mapEnabledProviders(cfg: config.Config): ProviderOption[] {
  return (cfg.Providers || [])
    .filter(provider => provider.Enabled)
    .map(provider => ({
      name: provider.Name,
      defaultModel: provider.DefaultModel || '',
      models: (provider.Models || []).map(model => ({
        id: model.ID,
        name: model.Name,
        maxTokens: model.MaxTokens,
        contextWindow: model.ContextWindow || 32768
      }))
    }))
}

function initProviderAndModel() {
  if (providers.value.length === 0) {
    selectedProvider.value = ''
    selectedModel.value = ''
    return
  }

  selectedProvider.value = providers.value[0].name
  syncSelectedModel()
}

function syncSelectedModel() {
  const provider = providers.value.find(item => item.name === selectedProvider.value)
  if (!provider) {
    selectedModel.value = ''
    return
  }

  const hasCurrentModel = provider.models.some(model => model.id === selectedModel.value)
  if (hasCurrentModel) {
    return
  }

  if (provider.defaultModel && provider.models.some(model => model.id === provider.defaultModel)) {
    selectedModel.value = provider.defaultModel
    return
  }

  selectedModel.value = provider.models[0]?.id || ''
}

function openDocSelector(type: 'old' | 'new') {
  docSelectorType.value = type
  docSearchKeyword.value = ''
  isDocSelectorOpen.value = true
}

function closeDocSelector() {
  isDocSelectorOpen.value = false
}

function chooseDocument(doc: Document) {
  selectedDocs.value[docSelectorType.value] = doc
  closeDocSelector()
}

function formatSize(bytes: number): string {
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

async function startAnalysis() {
  if (!selectedDocs.value.old || !selectedDocs.value.new || !selectedProvider.value || !selectedModel.value) return

  // 获取当前选中模型的上下文窗口
  const currentModel = currentModels.value.find(m => m.id === selectedModel.value)
  const contextWindow = currentModel?.contextWindow || 32768

  const execID = await workflowApi.execute(selectedWorkflow.value, {
    workspace_id: 'default',
    old_policy: selectedDocs.value.old.id,
    new_policy: selectedDocs.value.new.id,
    focus_areas: selectedAreas.value,
    provider: selectedProvider.value,
    model: selectedModel.value,
    context_window: contextWindow
  })

  router.push({
    path: '/analysis/running',
    query: { execID, workflow: selectedWorkflow.value }
  })
}
</script>
