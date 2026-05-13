<template>
  <div class="app-container h-screen flex flex-col" :class="themeClass">
    <!-- 顶部导航栏 -->
    <header class="app-header h-14 bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 flex items-center px-4 flex-shrink-0">
      <div class="flex items-center gap-3">
        <div class="w-8 h-8 bg-primary-600 rounded-lg flex items-center justify-center">
          <span class="text-white font-bold text-sm">PA</span>
        </div>
        <h1 class="text-lg font-semibold text-gray-800 dark:text-gray-100">Policy Analyzer</h1>
      </div>

      <!-- 全局搜索 -->
      <div class="flex-1 max-w-md mx-4">
        <div class="relative">
          <input
            type="text"
            placeholder="搜索文档、工作流..."
            class="input pl-9 text-sm dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100"
          />
          <svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
        </div>
      </div>

      <!-- 右侧操作 -->
      <div class="flex items-center gap-3">
        <!-- 设置按钮 -->
        <button class="btn-ghost btn-sm" @click="showSettings = true">
          <svg class="w-5 h-5 text-gray-600 dark:text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.74 3.392.058 3.536 1.871.002.033.022.063.057.084a1.724 1.724 0 001.065 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.74 1.543-.058 3.392-1.871 3.536a1.724 1.724 0 00-2.573 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.74-3.392-.058-3.536-1.871a1.724 1.724 0 00-1.065-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.74-1.543.058-3.392 1.871-3.536.033-.002.063-.022.084-.057a1.724 1.724 0 002.573-1.066z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
        </button>

        <!-- 版本信息 -->
        <span class="text-xs text-gray-400 dark:text-gray-500">v{{ version }}</span>
      </div>
    </header>

    <!-- 主内容区域 -->
    <div class="app-main flex flex-1 overflow-hidden">
      <!-- 左侧边栏 -->
      <aside class="sidebar w-60 bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 flex-shrink-0 overflow-y-auto">
        <nav class="p-4">
          <!-- 工作空间选择 -->
          <div class="mb-4">
            <label class="text-xs font-medium text-gray-500 dark:text-gray-400 mb-2 block">工作空间</label>
            <select class="select text-sm dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100" v-model="currentWorkspace">
              <option value="default">默认工作空间</option>
            </select>
          </div>

          <!-- 导航菜单 -->
          <ul class="space-y-1">
            <li>
              <router-link
                to="/documents"
                class="flex items-center gap-2 px-3 py-2 rounded-md text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 hover:text-gray-800 dark:hover:text-gray-100 transition-colors"
                :class="{ 'bg-primary-50 dark:bg-primary-900/30 text-primary-700 dark:text-primary-300': $route.path.startsWith('/documents') }"
              >
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                <span>文档库</span>
              </router-link>
            </li>

            <li>
              <router-link
                to="/workflows"
                class="flex items-center gap-2 px-3 py-2 rounded-md text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 hover:text-gray-800 dark:hover:text-gray-100 transition-colors"
                :class="{ 'bg-primary-50 dark:bg-primary-900/30 text-primary-700 dark:text-primary-300': $route.path.startsWith('/workflows') }"
              >
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 10h16M4 14h16M4 18h16" />
                </svg>
                <span>工作流</span>
              </router-link>
            </li>

            <li>
              <router-link
                to="/analysis"
                class="flex items-center gap-2 px-3 py-2 rounded-md text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 hover:text-gray-800 dark:hover:text-gray-100 transition-colors"
                :class="{ 'bg-primary-50 dark:bg-primary-900/30 text-primary-700 dark:text-primary-300': $route.path.startsWith('/analysis') }"
              >
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.95-.083-1.867-.548-2.659a5 5 0 11-7.072 0l.547-.548A3.374 3.374 0 015.469 12H6a2 2 0 114 0v.531c0 .95.083 1.867.548 2.659z" />
                </svg>
                <span>分析历史</span>
              </router-link>
            </li>

            <li>
              <router-link
                to="/ai-qa"
                class="flex items-center gap-2 px-3 py-2 rounded-md text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 hover:text-gray-800 dark:hover:text-gray-100 transition-colors"
                :class="{ 'bg-primary-50 dark:bg-primary-900/30 text-primary-700 dark:text-primary-300': $route.path.startsWith('/ai-qa') }"
              >
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-4l-4 4v-4z" />
                </svg>
                <span>AI问答</span>
              </router-link>
            </li>

            <li>
              <router-link
                to="/settings"
                class="flex items-center gap-2 px-3 py-2 rounded-md text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 hover:text-gray-800 dark:hover:text-gray-100 transition-colors"
                :class="{ 'bg-primary-50 dark:bg-primary-900/30 text-primary-700 dark:text-primary-300': $route.path.startsWith('/settings') }"
              >
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4" />
                </svg>
                <span>设置</span>
              </router-link>
            </li>
          </ul>
        </nav>

        <!-- 快捷操作 -->
        <div class="p-4 border-t border-gray-100 dark:border-gray-700">
          <button class="btn-primary btn-sm w-full" @click="showImportDialog = true">
            <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            导入文档
          </button>
        </div>
      </aside>

      <!-- 右侧内容区域 -->
      <main class="content-area flex-1 overflow-y-auto p-6 bg-gray-50 dark:bg-gray-900">
        <router-view />
      </main>
    </div>

    <!-- 导入对话框 -->
    <ImportDialog v-if="showImportDialog" @close="showImportDialog = false" />

    <!-- 设置对话框 -->
    <SettingsDialog v-if="showSettings" @close="showSettings = false" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import ImportDialog from './components/ImportDialog.vue'
import SettingsDialog from './components/SettingsDialog.vue'
import { appApi } from './api'

// 应用版本
const version = ref('1.0.0')

// 当前工作空间
const currentWorkspace = ref('default')

// 对话框状态
const showImportDialog = ref(false)
const showSettings = ref(false)

// 主题设置
const theme = ref<'light' | 'dark' | 'system'>('light')

function normalizeTheme(value?: string): 'light' | 'dark' | 'system' {
  if (value === 'dark') {
    return 'dark'
  }
  if (value === 'system' || value === 'auto') {
    return 'system'
  }
  return 'light'
}

// 计算主题类
const themeClass = computed(() => {
  if (theme.value === 'dark') {
    return 'dark'
  }
  if (theme.value === 'system') {
    const isDark = window.matchMedia('(prefers-color-scheme: dark)').matches
    return isDark ? 'dark' : ''
  }
  return ''
})

// 加载主题设置
async function loadTheme() {
  try {
    const config = await appApi.getConfig() as any
    theme.value = normalizeTheme(config?.App?.Theme)
  } catch (e) {
    console.error('加载主题设置失败:', e)
  }
}

// 监听系统主题变化
onMounted(() => {
  loadTheme()

  // 监听系统主题变化
  const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
  mediaQuery.addEventListener('change', () => {
    if (theme.value === 'system') {
      // 触发重新计算
      theme.value = 'system'
    }
  })

  // 监听设置变化
  window.addEventListener('settings-changed', ((event: CustomEvent) => {
    if (event.detail?.theme) {
      theme.value = normalizeTheme(event.detail.theme)
    } else {
      loadTheme()
    }
  }) as EventListener)

  console.log('Policy Analyzer initialized')
})
</script>

<style scoped>
.app-container {
  user-select: none;
}

.sidebar::-webkit-scrollbar {
  width: 4px;
}

.content-area::-webkit-scrollbar {
  width: 8px;
}
</style>
