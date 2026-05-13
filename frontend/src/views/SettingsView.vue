<template>
  <div class="settings-view">
    <div class="mb-6">
      <h2 class="text-2xl font-bold text-gray-800 dark:text-gray-100">设置</h2>
      <p class="text-gray-500 dark:text-gray-400 text-sm mt-1">管理应用配置和AI模型</p>
    </div>

    <div v-if="loadError" class="mb-6 rounded-lg border border-error/20 bg-red-50 dark:bg-red-900/30 px-4 py-3 text-sm text-error dark:text-red-300">
      {{ loadError }}
    </div>

    <div v-if="saveMessage" class="mb-6 rounded-lg border border-green-200 dark:border-green-800 bg-green-50 dark:bg-green-900/30 px-4 py-3 text-sm text-green-700 dark:text-green-300">
      {{ saveMessage }}
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <div class="card">
        <nav class="space-y-1">
          <button
            v-for="item in navItems"
            :key="item.id"
            class="w-full text-left px-4 py-3 rounded-lg transition-colors"
            :class="activeSection === item.id ? 'bg-primary-50 dark:bg-primary-900/20 text-primary-700 dark:text-primary-300' : 'text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700'"
            @click="activeSection = item.id"
          >
            <div class="flex items-center gap-3">
              <component :is="item.icon" class="w-5 h-5" />
              <span>{{ item.name }}</span>
            </div>
          </button>
        </nav>
      </div>

      <div class="lg:col-span-2">
        <div v-if="isLoading" class="card text-center text-gray-500 dark:text-gray-400">正在加载设置...</div>

        <template v-else>
          <div v-if="activeSection === 'ai'" class="card">
            <div class="flex items-center justify-between gap-4 mb-6">
              <div>
                <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-100">AI模型配置</h3>
                <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">编辑已存在的 AI 提供商连接信息</p>
              </div>
              <button class="btn-primary" :disabled="isSaving" @click="saveSettings">
                {{ isSaving ? '保存中...' : '保存配置' }}
              </button>
            </div>

            <div class="space-y-6">
              <div v-for="provider in providers" :key="provider.name" class="border border-gray-200 dark:border-gray-700 rounded-lg p-4">
                <div class="flex items-center justify-between mb-4">
                  <div class="flex items-center gap-3">
                    <div class="w-10 h-10 bg-gray-100 dark:bg-gray-700 rounded-lg flex items-center justify-center">
                      <span class="text-sm font-bold text-gray-600 dark:text-gray-300">{{ provider.name.charAt(0) }}</span>
                    </div>
                    <div>
                      <h4 class="font-medium text-gray-800 dark:text-gray-100">{{ provider.name }}</h4>
                      <p class="text-xs text-gray-500 dark:text-gray-400">{{ provider.type }}</p>
                    </div>
                  </div>
                  <label class="relative inline-flex items-center cursor-pointer">
                    <input v-model="provider.enabled" type="checkbox" class="sr-only peer">
                    <div class="w-11 h-6 bg-gray-200 dark:bg-gray-600 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-primary-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary-600"></div>
                  </label>
                </div>

                <div class="grid grid-cols-1 md:grid-cols-2 gap-4 pt-4 border-t border-gray-100 dark:border-gray-700">
                  <div class="md:col-span-2">
                    <label class="text-sm font-medium text-gray-700 dark:text-gray-200 block mb-1">API Endpoint</label>
                    <input v-model="provider.endpoint" type="text" class="input text-sm" />
                  </div>
                  <div>
                    <label class="text-sm font-medium text-gray-700 dark:text-gray-200 block mb-1">API Key</label>
                    <input v-model="provider.apiKey" type="password" class="input text-sm" placeholder="输入API密钥" />
                  </div>
                  <div>
                    <label class="text-sm font-medium text-gray-700 dark:text-gray-200 block mb-1">超时时间（秒）</label>
                    <input v-model.number="provider.timeout" type="number" min="1" class="input text-sm" />
                  </div>
                  <div class="md:col-span-2">
                    <label class="text-sm font-medium text-gray-700 dark:text-gray-200 block mb-1">默认模型</label>
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
                  </div>

                  <div class="md:col-span-2">
                    <div class="flex gap-2 items-end">
                      <button
                        type="button"
                        class="btn-secondary btn-sm"
                        @click="testConnection(provider)"
                        :disabled="isTestingConnection[provider.name] || !provider.apiKey || !provider.defaultModel"
                      >
                        <svg v-if="isTestingConnection[provider.name]" class="w-4 h-4 mr-1 animate-spin" fill="none" viewBox="0 0 24 24">
                          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                        </svg>
                        {{ isTestingConnection[provider.name] ? '测试中...' : '测试连接' }}
                      </button>
                      <div v-if="connectionResults[provider.name]" class="flex items-center gap-2 text-sm">
                        <span v-if="connectionResults[provider.name].success" class="flex items-center gap-1 text-green-600">
                          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                          </svg>
                          连接成功 ({{ connectionResults[provider.name].response_time }}ms)
                        </span>
                        <span v-else class="flex items-center gap-1 text-red-600">
                          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                          </svg>
                          {{ connectionResults[provider.name].error || '连接失败' }}
                        </span>
                      </div>
                    </div>
                  </div>

                  <div class="md:col-span-2">
                    <label class="text-sm font-medium text-gray-700 dark:text-gray-200 block mb-2">模型上下文配置</label>
                    <div class="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden">
                      <table class="w-full text-sm">
                        <thead class="bg-gray-50 dark:bg-gray-800">
                          <tr>
                            <th class="px-3 py-2 text-left font-medium text-gray-600 dark:text-gray-300">模型名称</th>
                            <th class="px-3 py-2 text-left font-medium text-gray-600 dark:text-gray-300">最大输出Token</th>
                            <th class="px-3 py-2 text-left font-medium text-gray-600 dark:text-gray-300">上下文窗口</th>
                          </tr>
                        </thead>
                        <tbody class="divide-y divide-gray-100 dark:divide-gray-700">
                          <tr v-for="model in provider.models" :key="model.id" class="hover:bg-gray-50 dark:hover:bg-gray-800">
                            <td class="px-3 py-2">
                              <span :class="model.id === provider.defaultModel ? 'font-medium text-primary-600' : 'text-gray-700 dark:text-gray-200'">
                                {{ model.name }}
                              </span>
                            </td>
                            <td class="px-3 py-2">
                              <input
                                type="number"
                                v-model.number="model.maxTokens"
                                class="input text-sm w-24"
                                min="1"
                              />
                            </td>
                            <td class="px-3 py-2">
                              <input
                                type="number"
                                v-model.number="model.contextWindow"
                                class="input text-sm w-28"
                                min="1"
                              />
                            </td>
                          </tr>
                        </tbody>
                      </table>
                    </div>
                    <p class="text-xs text-gray-400 dark:text-gray-500 mt-1">上下文窗口 = 输入 + 输出的总体限制，请根据模型实际能力设置</p>
                  </div>

                  <div class="md:col-span-2">
                    <label class="text-sm font-medium text-gray-700 dark:text-gray-200 block mb-1">自定义模型名称</label>
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
                    <p class="text-xs text-gray-400 dark:text-gray-500 mt-1">添加后可在下拉列表中选择使用</p>
                  </div>

                  <div v-if="provider.customModels && provider.customModels.length > 0" class="md:col-span-2">
                    <div class="flex flex-wrap gap-2">
                      <span class="text-xs text-gray-500 dark:text-gray-400">自定义模型:</span>
                      <span
                        v-for="model in provider.customModels"
                        :key="model"
                        class="inline-flex items-center gap-1 px-2 py-1 bg-primary-50 text-primary-700 rounded text-xs"
                      >
                        {{ model }}
                        <button
                          type="button"
                          class="hover:text-primary-900"
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
            </div>
          </div>

          <div v-if="activeSection === 'appearance'" class="card">
            <div class="flex items-center justify-between gap-4 mb-6">
              <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-100">外观设置</h3>
              <button class="btn-primary" :disabled="isSaving" @click="saveSettings">
                {{ isSaving ? '保存中...' : '保存配置' }}
              </button>
            </div>

            <div class="space-y-6">
              <div>
                <label class="text-sm font-medium text-gray-700 dark:text-gray-200 block mb-2">主题</label>
                <div class="flex gap-4">
                  <label class="flex items-center gap-2 cursor-pointer">
                    <input v-model="appearance.theme" type="radio" value="light" class="text-primary-600" />
                    <span class="text-gray-700 dark:text-gray-300">浅色</span>
                  </label>
                  <label class="flex items-center gap-2 cursor-pointer">
                    <input v-model="appearance.theme" type="radio" value="dark" class="text-primary-600" />
                    <span class="text-gray-700 dark:text-gray-300">深色</span>
                  </label>
                  <label class="flex items-center gap-2 cursor-pointer">
                    <input v-model="appearance.theme" type="radio" value="system" class="text-primary-600" />
                    <span class="text-gray-700 dark:text-gray-300">跟随系统</span>
                  </label>
                </div>
              </div>

              <div>
                <label class="text-sm font-medium text-gray-700 dark:text-gray-200 block mb-2">界面字体大小</label>
                <select v-model.number="appearance.fontSize" class="select">
                  <option :value="12">小 (12px)</option>
                  <option :value="14">中 (14px)</option>
                  <option :value="16">大 (16px)</option>
                </select>
              </div>

              <div>
                <label class="text-sm font-medium text-gray-700 dark:text-gray-200 block mb-2">Markdown代码高亮主题</label>
                <select v-model="appearance.codeTheme" class="select">
                  <option value="github">GitHub</option>
                  <option value="monokai">Monokai</option>
                  <option value="dracula">Dracula</option>
                </select>
              </div>
            </div>
          </div>

          <div v-if="activeSection === 'data'" class="card">
            <div class="flex items-center justify-between gap-4 mb-6">
              <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-100">数据管理</h3>
              <button class="btn-primary" :disabled="isSaving" @click="saveSettings">
                {{ isSaving ? '保存中...' : '保存配置' }}
              </button>
            </div>

            <div class="space-y-6">
              <div>
                <label class="text-sm font-medium text-gray-700 dark:text-gray-200 block mb-2">数据存储位置</label>
                <input v-model="dataConfig.storagePath" type="text" class="input" readonly />
                <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">当前版本仅展示配置中的存储目录，不在此页切换目录</p>
              </div>

              <div>
                <label class="text-sm font-medium text-gray-700 dark:text-gray-200 block mb-2">最大文件大小限制</label>
                <select v-model.number="dataConfig.maxFileSize" class="select">
                  <option :value="50">50 MB</option>
                  <option :value="100">100 MB</option>
                  <option :value="200">200 MB</option>
                </select>
              </div>

              <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-4 space-y-4">
                <div>
                  <h4 class="text-base font-medium text-gray-800 dark:text-gray-100">日志管理</h4>
                  <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">支持导出日志、删除历史日志，并按保留天数自动清理。</p>
                </div>

                <div>
                  <label class="text-sm font-medium text-gray-700 dark:text-gray-200 block mb-2">日志目录</label>
                  <input :value="dataConfig.logDirectory" type="text" class="input text-sm" readonly />
                </div>

                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <label class="flex items-start gap-3 cursor-pointer rounded-lg border border-gray-200 dark:border-gray-700 p-3">
                    <input v-model="dataConfig.logAutoCleanup" type="checkbox" class="mt-1 h-4 w-4" />
                    <div>
                      <div class="text-sm font-medium text-gray-800 dark:text-gray-100">启用自动清理</div>
                      <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">应用启动和保存设置时，会自动清理超出保留天数的旧日志。</p>
                    </div>
                  </label>

                  <div>
                    <label class="text-sm font-medium text-gray-700 dark:text-gray-200 block mb-2">日志保留天数</label>
                    <input
                      v-model.number="dataConfig.logRetentionDays"
                      type="number"
                      min="1"
                      class="input text-sm"
                    />
                    <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">保留最近 N 天日志，始终保留当天正在写入的日志文件。</p>
                  </div>
                </div>

                <div class="flex flex-wrap gap-3">
                  <button
                    type="button"
                    class="btn-secondary btn-sm"
                    :disabled="isExportingLogs"
                    @click="exportLogs"
                  >
                    {{ isExportingLogs ? '导出中...' : '导出日志文件' }}
                  </button>
                  <button
                    type="button"
                    class="btn-secondary btn-sm"
                    :disabled="isCleaningLogs"
                    @click="cleanupLogsNow"
                  >
                    {{ isCleaningLogs ? '清理中...' : '立即清理过期日志' }}
                  </button>
                  <button
                    type="button"
                    class="btn-secondary btn-sm text-red-600 border-red-200 hover:bg-red-50 dark:text-red-300 dark:border-red-800 dark:hover:bg-red-900/20"
                    :disabled="isDeletingLogs"
                    @click="deleteHistoryLogs"
                  >
                    {{ isDeletingLogs ? '删除中...' : '删除历史日志' }}
                  </button>
                </div>
              </div>
            </div>
          </div>

          <div v-if="activeSection === 'about'" class="card">
            <div class="text-center py-8">
              <div class="w-20 h-20 bg-primary-600 rounded-2xl mx-auto flex items-center justify-center mb-4">
                <span class="text-white font-bold text-3xl">PA</span>
              </div>
              <h3 class="text-2xl font-bold text-gray-800 dark:text-gray-100 mb-2">Policy Analyzer</h3>
              <p class="text-gray-500 dark:text-gray-400 mb-4">版本 {{ appVersion }}</p>
              <p class="text-gray-600 dark:text-gray-300 max-w-md mx-auto mb-8">
                面向政策研究人员、企业合规团队、法律从业者的AI辅助分析工具。
                提供政策文件的智能导入、多维度对比分析、自动化报告生成能力。
              </p>

              <div class="border-t border-gray-200 pt-6 text-sm text-gray-500 dark:text-gray-400">
                <p>Copyright © 2026 Policy Analyzer Team</p>
                <p class="mt-1 font-medium">The Emperor Protects</p>
              </div>
            </div>
          </div>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { appApi, providerApi } from '@/api'
import { config } from '../../wailsjs/go/models'

interface ProviderForm {
  name: string
  type: string
  enabled: boolean
  endpoint: string
  apiKey: string
  defaultModel: string
  timeout: number
  models: Array<{ id: string; name: string; maxTokens: number; contextWindow: number }>
  customModel?: string
  customModels: string[]
}

const activeSection = ref('ai')
const isLoading = ref(false)
const isSaving = ref(false)
const loadError = ref('')
const saveMessage = ref('')
const appVersion = ref('1.0.0')
const rawConfig = ref<config.Config | null>(null)
const isLoadingModels = reactive<Record<string, boolean>>({})
const isTestingConnection = reactive<Record<string, boolean>>({})
const isExportingLogs = ref(false)
const isDeletingLogs = ref(false)
const isCleaningLogs = ref(false)
const connectionResults = reactive<Record<string, { success: boolean; response_time?: number; error?: string }>>({})

const navItems = [
  { id: 'ai', name: 'AI模型配置', icon: 'span' },
  { id: 'appearance', name: '外观', icon: 'span' },
  { id: 'data', name: '数据管理', icon: 'span' },
  { id: 'about', name: '关于', icon: 'span' }
]

const providers = ref<ProviderForm[]>([])

const appearance = reactive({
  theme: 'light',
  fontSize: 14,
  codeTheme: 'github'
})

const dataConfig = reactive({
  storagePath: '',
  maxFileSize: 100,
  logDirectory: '',
  logAutoCleanup: true,
  logRetentionDays: 7
})

onMounted(() => {
  loadSettings()
})

async function loadSettings() {
  isLoading.value = true
  loadError.value = ''
  saveMessage.value = ''

  try {
    const [cfg, logDirectory] = await Promise.all([
      appApi.getConfig(),
      appApi.getLogDirectory()
    ])
    rawConfig.value = cfg
    appVersion.value = cfg.App?.Version || '1.0.0'

    providers.value = (cfg.Providers || []).map((provider: config.ProviderConfig) => ({
      name: provider.Name,
      type: provider.Type,
      enabled: provider.Enabled,
      endpoint: provider.Endpoint || '',
      apiKey: provider.APIKey || '',
      defaultModel: provider.DefaultModel || '',
      timeout: provider.Timeout || 60,
      models: (provider.Models || []).map((model: config.ModelConfig) => ({
        id: model.ID,
        name: model.Name,
        maxTokens: model.MaxTokens,
        contextWindow: model.ContextWindow || 8192
      })),
      customModels: []
    }))

    appearance.theme = cfg.App?.Theme === 'auto' ? 'system' : (cfg.App?.Theme || 'light')
    appearance.fontSize = cfg.App?.UI?.FontSize || 14
    appearance.codeTheme = cfg.App?.UI?.CodeTheme || 'github'
    dataConfig.storagePath = cfg.Storage?.BaseDir || ''
    dataConfig.maxFileSize = cfg.Storage?.MaxFileSize || 100
    dataConfig.logDirectory = logDirectory || ''
    dataConfig.logAutoCleanup = cfg.Storage?.LogAutoCleanup ?? true
    dataConfig.logRetentionDays = cfg.Storage?.LogRetentionDays || 7
  } catch (error) {
    console.error('加载设置失败:', error)
    loadError.value = error instanceof Error ? error.message : '加载设置失败'
  } finally {
    isLoading.value = false
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
      MaxTokens: model.maxTokens,
      ContextWindow: model.contextWindow
    }))
  })
}

async function fetchModels(provider: ProviderForm) {
  if (!provider.apiKey) {
    loadError.value = '请先输入 API Key'
    return
  }

  isLoadingModels[provider.name] = true
  loadError.value = ''

  try {
    const models = await providerApi.getModelsWithConfig(buildProviderConfig(provider)) as Array<{ id: string; name: string; max_tokens?: number }>
    if (models && models.length > 0) {
      const existingIds = new Set(provider.models.map(m => m.id))
      for (const model of models) {
        if (!existingIds.has(model.id)) {
          provider.models.push({
            id: model.id,
            name: model.name,
            maxTokens: model.max_tokens || 4096,
            contextWindow: 8192
          })
        }
      }
      showSuccess(`获取到 ${models.length} 个模型`)
    }
  } catch (error) {
    console.error('获取模型列表失败:', error)
    loadError.value = `获取模型失败: ${error instanceof Error ? error.message : '未知错误'}`
  } finally {
    isLoadingModels[provider.name] = false
  }
}

async function testConnection(provider: ProviderForm) {
  if (!provider.apiKey || !provider.defaultModel) {
    loadError.value = '请先输入 API Key 并选择模型'
    return
  }

  isTestingConnection[provider.name] = true
  loadError.value = ''
  delete connectionResults[provider.name]

  try {
    const result = await providerApi.testConnectionWithConfig(buildProviderConfig(provider), provider.defaultModel) as {
      success: boolean
      response_time?: number
      error?: string
    }
    connectionResults[provider.name] = result
  } catch (error) {
    console.error('测试连接失败:', error)
    connectionResults[provider.name] = {
      success: false,
      error: error instanceof Error ? error.message : '未知错误'
    }
  } finally {
    isTestingConnection[provider.name] = false
  }
}

function addCustomModel(provider: ProviderForm) {
  const modelName = provider.customModel?.trim()
  if (!modelName) return

  const exists = provider.models.some(m => m.id === modelName) ||
                 provider.customModels.includes(modelName)

  if (!exists) {
    provider.customModels.push(modelName)
    provider.models.push({
      id: modelName,
      name: modelName,
      maxTokens: 4096,
      contextWindow: 8192
    })
    provider.defaultModel = modelName
  }

  provider.customModel = ''
}

function removeCustomModel(provider: ProviderForm, modelName: string) {
  const idx = provider.customModels.indexOf(modelName)
  if (idx > -1) {
    provider.customModels.splice(idx, 1)
  }
  const modelIdx = provider.models.findIndex(m => m.id === modelName)
  if (modelIdx > -1) {
    provider.models.splice(modelIdx, 1)
  }
  if (provider.defaultModel === modelName && provider.models.length > 0) {
    provider.defaultModel = provider.models[0].id
  }
}

async function exportLogs() {
  if (isExportingLogs.value) return

  isExportingLogs.value = true
  loadError.value = ''

  try {
    const exportPath = await appApi.exportLogs()
    if (!exportPath) return
    showSuccess(`日志已导出到：${exportPath}`)
  } catch (error) {
    console.error('导出日志失败:', error)
    loadError.value = error instanceof Error ? error.message : '导出日志失败'
  } finally {
    isExportingLogs.value = false
  }
}

async function deleteHistoryLogs() {
  if (isDeletingLogs.value) return

  isDeletingLogs.value = true
  loadError.value = ''

  try {
    const deletedCount = await appApi.deleteHistoricalLogs()
    showSuccess(`已删除 ${deletedCount} 个历史日志文件`)
  } catch (error) {
    console.error('删除历史日志失败:', error)
    loadError.value = error instanceof Error ? error.message : '删除历史日志失败'
  } finally {
    isDeletingLogs.value = false
  }
}

async function cleanupLogsNow() {
  if (isCleaningLogs.value) return

  isCleaningLogs.value = true
  loadError.value = ''

  try {
    const deletedCount = await appApi.cleanupLogs(safeRetentionDays())
    showSuccess(`已清理 ${deletedCount} 个过期日志文件`)
  } catch (error) {
    console.error('清理日志失败:', error)
    loadError.value = error instanceof Error ? error.message : '清理日志失败'
  } finally {
    isCleaningLogs.value = false
  }
}

function safeRetentionDays() {
  return Math.max(1, Number(dataConfig.logRetentionDays) || 7)
}

function showSuccess(message: string) {
  saveMessage.value = message
  window.setTimeout(() => {
    if (saveMessage.value === message) {
      saveMessage.value = ''
    }
  }, 2500)
}

function buildConfigFromForm() {
  if (!rawConfig.value) {
    throw new Error('配置未加载')
  }

  return config.Config.createFrom({
    ...rawConfig.value,
    App: {
      ...rawConfig.value.App,
      Theme: appearance.theme,
      UI: {
        ...rawConfig.value.App?.UI,
        FontSize: appearance.fontSize,
        CodeTheme: appearance.codeTheme
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
        MaxTokens: model.maxTokens,
        ContextWindow: model.contextWindow
      }))
    })),
    Storage: {
      ...rawConfig.value.Storage,
      BaseDir: dataConfig.storagePath,
      MaxFileSize: dataConfig.maxFileSize,
      LogAutoCleanup: dataConfig.logAutoCleanup,
      LogRetentionDays: safeRetentionDays()
    }
  })
}

async function saveSettings() {
  if (!rawConfig.value || isSaving.value) return

  isSaving.value = true
  loadError.value = ''
  saveMessage.value = ''

  try {
    const nextConfig = buildConfigFromForm()

    await appApi.updateConfig(nextConfig)
    showSuccess('设置已保存')

    window.dispatchEvent(new CustomEvent('settings-changed', { detail: { theme: appearance.theme } }))

    await loadSettings()
  } catch (error) {
    console.error('保存设置失败:', error)
    loadError.value = error instanceof Error ? error.message : '保存设置失败'
  } finally {
    isSaving.value = false
  }
}
</script>
