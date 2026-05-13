<template>
  <div class="workflow-editor-view">
    <button class="btn-ghost btn-sm mb-4" @click="goBack">
      <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
      </svg>
      返回工作流列表
    </button>

    <div v-if="isLoading" class="card text-center text-gray-500 dark:text-gray-400">正在加载工作流定义...</div>

    <div v-else class="space-y-6">
      <div v-if="errorMessage" class="rounded-lg border border-error/20 bg-red-50 dark:bg-red-900/30 px-4 py-3 text-sm text-error dark:text-red-300">
        {{ errorMessage }}
      </div>

      <div v-if="saveMessage" class="rounded-lg border border-green-200 dark:border-green-800 bg-green-50 dark:bg-green-900/30 px-4 py-3 text-sm text-green-700 dark:text-green-300">
        {{ saveMessage }}
      </div>

      <div class="card">
        <div class="flex items-start justify-between gap-4 mb-6">
          <div>
            <h1 class="text-2xl font-bold text-gray-800 dark:text-gray-100 mb-2">
              {{ isCreateMode ? '新建工作流' : editorData?.info.name || '工作流编辑器' }}
            </h1>
            <p class="text-sm text-gray-500 dark:text-gray-400">
              直接编辑工作流 YAML。保存时会进行语法与基础字段校验。
            </p>
          </div>
          <div class="flex items-center gap-3">
            <button class="btn-secondary" :disabled="isSaving" @click="resetContent">重置</button>
            <button class="btn-primary" :disabled="isSaving" @click="saveWorkflow">
              {{ isSaving ? '保存中...' : '保存工作流' }}
            </button>
          </div>
        </div>

        <div v-if="editorData?.info" class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
          <div class="rounded-lg border border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-900 px-4 py-3">
            <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">工作流 ID</p>
            <p class="text-sm font-medium text-gray-800 dark:text-gray-100 break-all">{{ editorData.info.id }}</p>
          </div>
          <div class="rounded-lg border border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-900 px-4 py-3">
            <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">版本</p>
            <p class="text-sm font-medium text-gray-800 dark:text-gray-100">{{ editorData.info.version || '1.0.0' }}</p>
          </div>
          <div class="rounded-lg border border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-900 px-4 py-3">
            <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">步骤数</p>
            <p class="text-sm font-medium text-gray-800 dark:text-gray-100">{{ editorData.info.steps }}</p>
          </div>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-2">YAML 定义</label>
          <textarea
            v-model="content"
            class="w-full min-h-[560px] rounded-lg border border-gray-200 dark:border-gray-700 bg-gray-950 px-4 py-3 font-mono text-sm text-gray-100 focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-100 dark:focus:ring-primary-900/40"
            spellcheck="false"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { workflowApi, type WorkflowDefinitionDetail } from '@/api'

const route = useRoute()
const router = useRouter()

const editorData = ref<WorkflowDefinitionDetail | null>(null)
const initialContent = ref('')
const content = ref('')
const isLoading = ref(false)
const isSaving = ref(false)
const errorMessage = ref('')
const saveMessage = ref('')

const isCreateMode = computed(() => route.name === 'WorkflowCreate')
const originalId = computed(() => (isCreateMode.value ? '' : String(route.params.id || '')))

onMounted(() => {
  loadWorkflow()
})

async function loadWorkflow() {
  isLoading.value = true
  errorMessage.value = ''
  saveMessage.value = ''

  try {
    const detail = isCreateMode.value
      ? await workflowApi.newDefinition()
      : await workflowApi.getDefinition(originalId.value)

    editorData.value = detail
    initialContent.value = detail.content || ''
    content.value = detail.content || ''
  } catch (error) {
    console.error('加载工作流定义失败:', error)
    errorMessage.value = error instanceof Error ? error.message : '加载工作流定义失败'
  } finally {
    isLoading.value = false
  }
}

function resetContent() {
  content.value = initialContent.value
  saveMessage.value = ''
  errorMessage.value = ''
}

async function saveWorkflow() {
  if (isSaving.value) return

  isSaving.value = true
  errorMessage.value = ''
  saveMessage.value = ''

  try {
    const detail = await workflowApi.saveDefinition(originalId.value, content.value)
    editorData.value = detail
    initialContent.value = detail.content || ''
    content.value = detail.content || ''
    saveMessage.value = '工作流已保存'

    if (isCreateMode.value && detail.info?.id) {
      router.replace(`/workflows/${detail.info.id}/edit`)
    }
  } catch (error) {
    console.error('保存工作流失败:', error)
    errorMessage.value = error instanceof Error ? error.message : '保存工作流失败'
  } finally {
    isSaving.value = false
  }
}

function goBack() {
  router.push('/workflows')
}
</script>
