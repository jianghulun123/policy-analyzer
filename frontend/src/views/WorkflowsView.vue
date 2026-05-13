<template>
  <div class="workflows-view">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h2 class="text-2xl font-bold text-gray-800 dark:text-gray-100">工作流</h2>
        <p class="text-gray-500 dark:text-gray-400 text-sm mt-1">选择分析工作流模板，或进入 YAML 编辑器维护模板</p>
      </div>
      <button class="btn-primary" @click="createWorkflow">
        <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        新建工作流
      </button>
    </div>

    <div v-if="errorMessage" class="mb-6 rounded-lg border border-error/20 bg-red-50 dark:bg-red-900/30 px-4 py-3 text-sm text-error dark:text-red-300">
      {{ errorMessage }}
    </div>

    <div v-if="isLoading" class="card text-center text-gray-500 dark:text-gray-400">正在加载工作流...</div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <div
        v-for="workflow in workflows"
        :key="workflow.id"
        class="card-hover cursor-pointer"
        @click="selectWorkflow(workflow)"
      >
        <div class="flex items-start justify-between gap-3 mb-3">
          <div class="flex items-start gap-3 min-w-0">
            <div class="w-10 h-10 rounded-lg bg-primary-100 dark:bg-primary-900/30 flex items-center justify-center shrink-0">
              <svg class="w-5 h-5 text-primary-600 dark:text-primary-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 10h16M4 14h16M4 18h16" />
              </svg>
            </div>
            <div class="min-w-0">
              <h3 class="font-medium text-gray-800 dark:text-gray-100 truncate">{{ workflow.name }}</h3>
              <p class="text-xs text-gray-500 dark:text-gray-400">v{{ workflow.version }} · {{ workflow.steps }} 步骤</p>
            </div>
          </div>
          <button
            class="btn-secondary btn-sm shrink-0"
            @click.stop="editWorkflow(workflow.id)"
          >
            编辑
          </button>
        </div>

        <p class="text-sm text-gray-600 dark:text-gray-300 mb-3 line-clamp-3">{{ workflow.description }}</p>

        <div class="flex flex-wrap gap-1">
          <span
            v-for="input in workflow.inputs"
            :key="input.name"
            class="tag"
            :class="input.required ? 'bg-primary-100 dark:bg-primary-900/30 text-primary-700 dark:text-primary-300' : 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300'"
          >
            {{ input.name }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { workflowApi, type Workflow } from '../api'

const router = useRouter()
const workflows = ref<Workflow[]>([])
const isLoading = ref(false)
const errorMessage = ref('')

onMounted(() => {
  loadWorkflows()
})

async function loadWorkflows() {
  isLoading.value = true
  errorMessage.value = ''

  try {
    workflows.value = await workflowApi.list()
  } catch (error) {
    console.error('加载工作流失败:', error)
    errorMessage.value = error instanceof Error ? error.message : '加载工作流失败'
  } finally {
    isLoading.value = false
  }
}

function selectWorkflow(workflow: Workflow) {
  router.push({
    path: '/analysis/new',
    query: { workflow: workflow.id }
  })
}

function editWorkflow(workflowId: string) {
  router.push(`/workflows/${workflowId}/edit`)
}

function createWorkflow() {
  router.push('/workflows/new')
}
</script>
