import { createRouter, createWebHashHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/documents'
  },
  {
    path: '/documents',
    name: 'Documents',
    component: () => import('@/views/DocumentsView.vue'),
    meta: { title: '文档库' }
  },
  {
    path: '/documents/:id',
    name: 'DocumentDetail',
    component: () => import('@/views/DocumentDetailView.vue'),
    meta: { title: '文档详情' }
  },
  {
    path: '/workflows',
    name: 'Workflows',
    component: () => import('@/views/WorkflowsView.vue'),
    meta: { title: '工作流' }
  },
  {
    path: '/workflows/new',
    name: 'WorkflowCreate',
    component: () => import('@/views/WorkflowEditorView.vue'),
    meta: { title: '新建工作流' }
  },
  {
    path: '/workflows/:id/edit',
    name: 'WorkflowEditor',
    component: () => import('@/views/WorkflowEditorView.vue'),
    meta: { title: '编辑工作流' }
  },
  {
    path: '/analysis',
    name: 'Analysis',
    component: () => import('@/views/AnalysisView.vue'),
    meta: { title: '分析历史' }
  },
  {
    path: '/analysis/new',
    name: 'NewAnalysis',
    component: () => import('@/views/NewAnalysisView.vue'),
    meta: { title: '新建分析' }
  },
  {
    path: '/analysis/running',
    name: 'AnalysisRunning',
    component: () => import('@/views/AnalysisRunningView.vue'),
    meta: { title: '分析执行中' }
  },
  {
    path: '/analysis/:id',
    name: 'AnalysisDetail',
    component: () => import('@/views/AnalysisDetailView.vue'),
    meta: { title: '分析详情' }
  },
  {
    path: '/ai-qa',
    name: 'AiQa',
    component: () => import('@/views/AiQaView.vue'),
    meta: { title: 'AI问答' }
  },
  {
    path: '/settings',
    name: 'Settings',
    component: () => import('@/views/SettingsView.vue'),
    meta: { title: '设置' }
  }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

// 路由守卫 - 更新标题
router.beforeEach((to, _from, next) => {
  document.title = `${to.meta.title || 'Policy Analyzer'} - Policy Analyzer`
  next()
})

export default router