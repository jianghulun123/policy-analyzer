<template>
  <div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="$emit('close')">
    <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl w-full max-w-2xl max-h-[80vh] overflow-hidden animate-slide-up border border-gray-200 dark:border-gray-700">
      <!-- 头部 -->
      <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between">
        <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-100">设置</h3>
        <button @click="$emit('close')" class="text-gray-400 dark:text-gray-300 hover:text-gray-600 dark:hover:text-gray-100">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- 错误/成功提示 -->
      <div v-if="errorMessage" class="mx-6 mt-4 rounded-lg border border-error/20 bg-red-50 dark:bg-red-900/30 px-4 py-3 text-sm text-error dark:text-red-300">
        {{ errorMessage }}
      </div>
      <div v-if="successMessage" class="mx-6 mt-4 rounded-lg border border-green-200 dark:border-green-800 bg-green-50 dark:bg-green-900/30 px-4 py-3 text-sm text-green-700 dark:text-green-300">
        {{ successMessage }}
      </div>

      <!-- 内容 -->
      <div class="flex h-[60vh]">
        <!-- 左侧导航 -->
        <nav class="w-48 border-r border-gray-200 dark:border-gray-700 p-4">
          <ul class="space-y-1">
            <li v-for="tab in tabs" :key="tab.id">
              <button
                class="w-full text-left px-3 py-2 rounded-md text-sm transition-colors"
                :class="activeTab === tab.id ? 'bg-primary-50 dark:bg-primary-900/20 text-primary-700 dark:text-primary-300' : 'text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'"
                @click="activeTab = tab.id"
              >
                {{ tab.name }}
              </button>
            </li>
          </ul>
        </nav>

        <!-- 右侧内容 -->
        <div class="flex-1 p-6 overflow-y-auto">
          <!-- AI模型配置 -->
          <div v-if="activeTab === 'ai'" class="space-y-6">
            <h4 class="font-medium text-gray-800 dark:text-gray-100">AI模型配置</h4>

            <div v-for="provider in providers" :key="provider.name" class="card">
              <div class="flex items-center justify-between mb-3">
                <div class="flex items-center gap-3">
                  <div class="w-8 h-8 bg-gray-100 dark:bg-gray-700 rounded flex items-center justify-center">
                    <span class="text-xs font-medium text-gray-600 dark:text-gray-300">{{ provider.name.charAt(0) }}</span>
                  </div>
                  <div>
                    <h5 class="font-medium text-gray-800 dark:text-gray-100">{{ provider.name }}</h5>
                    <p class="text-xs text-gray-500 dark:text-gray-400">{{ provider.type }}</p>
                  </div>
                </div>
                <label class="relative inline-flex items-center cursor-pointer">
                  <input type="checkbox" v-model="provider.enabled" class="sr-only peer">
                  <div class="w-9 h-5 bg-gray-200 dark:bg-gray-600 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-primary-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-primary-600"></div>
                </label>
              </div>

              <div v-if="provider.enabled" class="space-y-3 pt-3 border-t border-gray-100 dark:border-gray-700">
                <div>
                  <label class="text-xs font-medium text-gray-600 dark:text-gray-300 mb-1 block">API Endpoint</label>
                  <input
                    type="text"
                    v-model="provider.endpoint"
                    class="input text-sm"
                    placeholder="API 地址"
                  />
                </div>

                <div>
                  <label class="text-xs font-medium text-gray-600 dark:text-gray-300 mb-1 block">API Key</label>
                  <div class="relative">
                    <input
                      :type="showApiKey[provider.name] ? 'text' : 'password'"
                      v-model="provider.apiKey"
                      class="input text-sm pr-10"
                      :placeholder="`输入 ${provider.name} API Key`"
                    />
                    <button
                      type="button"
                      class="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 dark:text-gray-300 hover:text-gray-600 dark:hover:text-gray-100"
                      @click="showApiKey[provider.name] = !showApiKey[provider.name]"
                    >
                      <svg v-if="showApiKey[provider.name]" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
                      </svg>
                      <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                      </svg>
                    </button>
                  </div>
                </div>

                <div>
                  <label class="text-xs font-medium text-gray-600 dark:text-gray-300 mb-1 block">默认模型</label>
                  <div class="flex gap-2">
                    <select v-model="provider.defaultModel" class="select text-sm flex-1">
                      <option v-for="model in provider.models" :key="model.id" :value="model.id">
                        {{ model.name }}
                      </option>
                    </select>
                    <button
                      type="button"
                      class="btn-secondary btn-sm whitespace-nowrap"
                      @click="fetchModels(provider)"
                      :disabled="isLoadingModels[provider.name]"
                    >
                      {{ isLoadingModels[provider.name] ? '获取中...' : '获取模型' }}
                    </button>
                  </div>
                  <p class="text-xs text-gray-400 dark:text-gray-500 mt-1">也可直接在下方输入自定义模型名称</p>
                </div>

                <div>
                  <label class="text-xs font-medium text-gray-600 dark:text-gray-300 mb-1 block">自定义模型名称</label>
                  <div class="flex gap-2">
                    <input
                      type="text"
                      v-model="provider.customModel"
                      class="input text-sm flex-1"
                      placeholder="输入自定义模型名称"
                      @keyup.enter="addCustomModel(provider)"
                    />
                    <button
                      type="button"
                      class="btn-secondary btn-sm"
                      @click="addCustomModel(provider)"
                      :disabled="!provider.customModel?.trim()"
                    >
                      添加
                    </button>
                  </div>
                </div>

                <!-- 已添加的自定义模型 -->
                <div v-if="provider.customModels && provider.customModels.length > 0" class="flex flex-wrap gap-2">
                  <span class="text-xs text-gray-500 dark:text-gray-400">自定义模型:</span>
                  <span
                    v-for="model in provider.customModels"
                    :key="model"
                    class="inline-flex items-center gap-1 px-2 py-1 bg-primary-50 dark:bg-primary-900/20 text-primary-700 dark:text-primary-300 rounded text-xs"
                  >
                    {{ model }}
                    <button
                      type="button"
                      class="hover:text-primary-900 dark:hover:text-primary-200"
                      @click="removeCustomModel(provider, model)"
                    >
                      <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                      </svg>
                    </button>
                  </span>
                </div>
              </div>
            </div>
          </div>

          <!-- 外观设置 -->
          <div v-if="activeTab === 'appearance'" class="space-y-6">
            <h4 class="font-medium text-gray-800 dark:text-gray-100">外观设置</h4>

            <div class="space-y-4">
              <div>
                <label class="text-sm font-medium text-gray-700 dark:text-gray-200 mb-2 block">主题</label>
                <select v-model="settings.theme" class="select">
                  <option value="light">浅色</option>
                  <option value="dark">深色</option>
                  <option value="system">跟随系统</option>
                </select>
              </div>

              <div>
                <label class="text-sm font-medium text-gray-700 dark:text-gray-200 mb-2 block">字体大小</label>
                <select v-model.number="settings.fontSize" class="select">
                  <option :value="12">小</option>
                  <option :value="14">中</option>
                  <option :value="16">大</option>
                </select>
              </div>

              <div>
                <label class="text-sm font-medium text-gray-700 dark:text-gray-200 mb-2 block">代码高亮主题</label>
                <select v-model="settings.codeTheme" class="select">
                  <option value="github">GitHub</option>
                  <option value="monokai">Monokai</option>
                  <option value="dracula">Dracula</option>
                </select>
              </div>
            </div>
          </div>

          <!-- 存储设置 -->
          <div v-if="activeTab === 'storage'" class="space-y-6">
            <h4 class="font-medium text-gray-800 dark:text-gray-100">存储设置</h4>

            <div class="space-y-4">
              <div>
                <label class="text-sm font-medium text-gray-700 dark:text-gray-200 mb-2 block">数据存储位置</label>
                <div class="flex gap-2">
                  <input type="text" v-model="settings.storagePath" class="input" readonly />
                  <button class="btn-secondary btn-sm" disabled>浏览</button>
                </div>
                <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">默认存储在用户目录下的 .policy-analyzer 文件夹</p>
              </div>

              <div>
                <label class="text-sm font-medium text-gray-700 dark:text-gray-200 mb-2 block">最大文件大小</label>
                <select v-model.number="settings.maxFileSize" class="select">
                  <option :value="50">50 MB</option>
                  <option :value="100">100 MB</option>
                  <option :value="200">200 MB</option>
                </select>
              </div>
            </div>
          </div>

          <!-- 关于 -->
          <div v-if="activeTab === 'about'" class="space-y-6">
            <h4 class="font-medium text-gray-800 dark:text-gray-100">关于</h4>

            <div class="text-center py-8">
              <div class="w-16 h-16 bg-primary-600 rounded-xl mx-auto flex items-center justify-center mb-4">
                <span class="text-white font-bold text-2xl">PA</span>
              </div>
              <h3 class="text-xl font-bold text-gray-800 dark:text-gray-100 mb-1">Policy Analyzer</h3>
              <p class="text-gray-500 dark:text-gray-400 mb-4">版本 {{ appVersion }}</p>
              <p class="text-sm text-gray-600 dark:text-gray-300 max-w-md mx-auto">
                面向政策研究人员、企业合规团队、法律从业者的AI辅助分析工具。
                提供政策文件的智能导入、多维度对比分析、自动化报告生成能力。
              </p>
            </div>

            <div class="border-t border-gray-200 pt-4 text-center text-xs text-gray-500 dark:text-gray-400">
              <p>Copyright © 2026 Policy Analyzer Team</p>
              <p class="mt-1">The Emperor Protects</p>
            </div>
          </div>
        </div>
      </div>

      <!-- 底部按钮 -->
      <div class="px-6 py-4 border-t border-gray-200 dark:border-gray-700 flex justify-end gap-3">
        <button class="btn-secondary btn-sm" @click="$emit('close')" :disabled="isSaving">取消</button>
        <button class="btn-primary btn-sm" @click="saveSettings" :disabled="isSaving">
          {{ isSaving ? '保存中...' : '保存设置' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { appApi, providerApi } from '@/api'
import { config } from '../../wailsjs/go/models'

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'saved'): void
}>()

const tabs = [
  { id: 'ai', name: 'AI模型配置' },
  { id: 'appearance', name: '外观' },
  { id: 'storage', name: '存储' },
  { id: 'about', name: '关于' }
]

const activeTab = ref('ai')
const showApiKey = reactive<Record<string, boolean>>({})
const isLoadingModels = reactive<Record<string, boolean>>({})
const isSaving = ref(false)
const errorMessage = ref('')
const successMessage = ref('')
const appVersion = ref('1.0.0')
const rawConfig = ref<config.Config | null>(null)

interface ProviderForm {
  name: string
  type: string
  enabled: boolean
  endpoint: string
  apiKey: string
  defaultModel: string
  timeout: number
  models: Array<{ id: string; name: string; maxTokens?: number }>
  customModel?: string
  customModels: string[]
}

const providers = ref<ProviderForm[]>([])

// 外观设置
const settings = reactive({
  theme: 'light',
  fontSize: 14,
  codeTheme: 'github',
  storagePath: '~/.policy-analyzer',
  maxFileSize: 100
})

onMounted(async () => {
  await loadSettings()
})

async function loadSettings() {
  try {
    const cfg = await appApi.getConfig()
    rawConfig.value = cfg
    appVersion.value = cfg.App?.Version || '1.0.0'

    providers.value = (cfg.Providers || []).map(provider => ({
      name: provider.Name,
      type: provider.Type,
      enabled: provider.Enabled,
      endpoint: provider.Endpoint || '',
      apiKey: provider.APIKey || '',
      defaultModel: provider.DefaultModel || '',
      timeout: provider.Timeout || 60,
      models: (provider.Models || []).map(model => ({
        id: model.ID,
        name: model.Name,
        maxTokens: model.MaxTokens
      })),
      customModels: []
    }))

    settings.theme = cfg.App?.Theme === 'auto' ? 'system' : (cfg.App?.Theme || 'light')
    settings.fontSize = cfg.App?.UI?.FontSize || 14
    settings.codeTheme = cfg.App?.UI?.CodeTheme || 'github'
    settings.storagePath = cfg.Storage?.BaseDir || ''
    settings.maxFileSize = cfg.Storage?.MaxFileSize || 100
  } catch (error) {
    console.error('加载设置失败:', error)
    errorMessage.value = error instanceof Error ? error.message : '加载设置失败'
  }
}

function buildProviderConfig(provider: ProviderForm) {
  return config.ProviderConfig.createFrom({
    Name: provider.name,
    Type: provider.type,
    Endpoint: provider.endpoint.trim(),
    APIKey: provider.apiKey.trim(),
    DefaultModel: provider.defaultModel,
    Timeout: Number(provider.timeout) || 60,
    Enabled: provider.enabled,
    Models: provider.models.map(model => ({
      ID: model.id,
      Name: model.name,
      MaxTokens: model.maxTokens || 4096
    }))
  })
}

async function fetchModels(provider: ProviderForm) {
  if (!provider.apiKey) {
    errorMessage.value = '请先输入 API Key'
    return
  }

  isLoadingModels[provider.name] = true
  errorMessage.value = ''

  try {
    const models = await providerApi.getModelsWithConfig(buildProviderConfig(provider)) as Array<{ id: string; name: string; max_tokens?: number }>
    if (models && models.length > 0) {
      // 合并新获取的模型
      const existingIds = new Set(provider.models.map(m => m.id))
      for (const model of models) {
        if (!existingIds.has(model.id)) {
          provider.models.push({
            id: model.id,
            name: model.name,
            maxTokens: model.max_tokens
          })
        }
      }
      successMessage.value = `获取到 ${models.length} 个模型`
      setTimeout(() => { successMessage.value = '' }, 2000)
    }
  } catch (error) {
    console.error('获取模型列表失败:', error)
    errorMessage.value = `获取模型失败: ${error instanceof Error ? error.message : '未知错误'}`
  } finally {
    isLoadingModels[provider.name] = false
  }
}

function addCustomModel(provider: ProviderForm) {
  const modelName = provider.customModel?.trim()
  if (!modelName) return

  // 检查是否已存在
  const exists = provider.models.some(m => m.id === modelName) ||
                 provider.customModels.includes(modelName)

  if (!exists) {
    provider.customModels.push(modelName)
    // 同时添加到 models 列表
    provider.models.push({
      id: modelName,
      name: modelName
    })
    // 设为默认
    provider.defaultModel = modelName
  }

  provider.customModel = ''
}

function removeCustomModel(provider: ProviderForm, modelName: string) {
  // 从自定义列表移除
  const idx = provider.customModels.indexOf(modelName)
  if (idx > -1) {
    provider.customModels.splice(idx, 1)
  }
  // 从 models 列表移除
  const modelIdx = provider.models.findIndex(m => m.id === modelName)
  if (modelIdx > -1) {
    provider.models.splice(modelIdx, 1)
  }
  // 如果移除的是当前选中的模型，重置
  if (provider.defaultModel === modelName && provider.models.length > 0) {
    provider.defaultModel = provider.models[0].id
  }
}

function buildConfigFromForm() {
  if (!rawConfig.value) {
    throw new Error('配置未加载')
  }

  return {
    ...rawConfig.value,
    App: {
      ...rawConfig.value.App,
      Theme: settings.theme,
      UI: {
        ...rawConfig.value.App?.UI,
        FontSize: settings.fontSize,
        CodeTheme: settings.codeTheme
      }
    },
    Providers: providers.value.map(provider => ({
      Name: provider.name,
      Type: provider.type,
      Endpoint: provider.endpoint.trim(),
      APIKey: provider.apiKey.trim(),
      DefaultModel: provider.defaultModel,
      Timeout: Number(provider.timeout) || 60,
      Enabled: provider.enabled,
      Models: provider.models.map(model => ({
        ID: model.id,
        Name: model.name,
        MaxTokens: model.maxTokens || 4096
      }))
    })),
    Storage: {
      ...rawConfig.value.Storage,
      BaseDir: settings.storagePath,
      MaxFileSize: settings.maxFileSize
    }
  } as config.Config
}

async function saveSettings() {
  if (!rawConfig.value || isSaving.value) return

  isSaving.value = true
  errorMessage.value = ''
  successMessage.value = ''

  try {
    const nextConfig = buildConfigFromForm()

    await appApi.updateConfig(nextConfig)
    successMessage.value = '设置已保存'
    emit('saved')
    window.dispatchEvent(new CustomEvent('settings-changed', { detail: { theme: settings.theme } }))

    // 重新加载配置
    await loadSettings()

    // 2秒后清除成功提示
    setTimeout(() => { successMessage.value = '' }, 2000)
  } catch (error) {
    console.error('保存设置失败:', error)
    errorMessage.value = error instanceof Error ? error.message : '保存设置失败'
  } finally {
    isSaving.value = false
  }
}
</script>
