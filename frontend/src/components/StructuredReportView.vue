<template>
  <div class="structured-report-view space-y-6">
    <!-- 执行摘要区域 - 始终显示 -->
    <section class="report-hero">
      <div class="report-hero__copy">
        <p class="report-eyebrow">执行摘要</p>
        <h3 class="report-title">政策变化一页总览</h3>
        <div class="report-summary markdown-body prose max-w-none" v-html="renderedSummary"></div>
      </div>
      <div v-if="headlineMetrics.length > 0" class="report-metrics">
        <article v-for="metric in headlineMetrics" :key="metric.title + metric.value" class="metric-card">
          <p class="metric-card__title">{{ metric.title }}</p>
          <p class="metric-card__value">{{ metric.value || '--' }}</p>
          <p class="metric-card__delta">{{ metric.delta || '变化待补充' }}</p>
          <p class="metric-card__insight">{{ metric.insight || '暂无说明' }}</p>
        </article>
      </div>
    </section>

    <!-- 五法解读：措辞变化 -->
    <section v-if="showSupplementalSections && wordingChanges.length > 0" class="card report-section">
      <div class="report-section__head">
        <div>
          <p class="report-section__eyebrow">五法解读</p>
          <h3 class="report-section__title">措辞变化分析</h3>
        </div>
      </div>
      <div class="report-table-wrap">
        <table class="report-table">
          <thead>
            <tr>
              <th>议题</th>
              <th>旧版表述</th>
              <th>新版表述</th>
              <th>信号解读</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in wordingChanges" :key="`${row.topic}-${row.old}`">
              <td>{{ row.topic }}</td>
              <td>{{ row.old }}</td>
              <td>{{ row.new }}</td>
              <td>{{ row.signal }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <!-- 五法解读：排序变化 -->
    <section v-if="showSupplementalSections && rankingChanges && (rankingChanges.rising?.length > 0 || rankingChanges.falling?.length > 0 || rankingChanges.new?.length > 0 || rankingChanges.disappeared?.length > 0)" class="card report-section">
      <div class="report-section__head">
        <div>
          <p class="report-section__eyebrow">五法解读</p>
          <h3 class="report-section__title">重点任务排序变化</h3>
        </div>
      </div>
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div v-if="rankingChanges.rising?.length > 0" class="p-4 bg-green-50 rounded-lg">
          <h4 class="font-medium text-green-800 mb-2">↑ 上升议题</h4>
          <ul class="text-sm text-green-700 space-y-1">
            <li v-for="item in rankingChanges.rising" :key="item">{{ item }}</li>
          </ul>
        </div>
        <div v-if="rankingChanges.falling?.length > 0" class="p-4 bg-orange-50 rounded-lg">
          <h4 class="font-medium text-orange-800 mb-2">↓ 下降议题</h4>
          <ul class="text-sm text-orange-700 space-y-1">
            <li v-for="item in rankingChanges.falling" :key="item">{{ item }}</li>
          </ul>
        </div>
        <div v-if="rankingChanges.new?.length > 0" class="p-4 bg-blue-50 rounded-lg">
          <h4 class="font-medium text-blue-800 mb-2">★ 新增议题</h4>
          <ul class="text-sm text-blue-700 space-y-1">
            <li v-for="item in rankingChanges.new" :key="item">{{ item }}</li>
          </ul>
        </div>
        <div v-if="rankingChanges.disappeared?.length > 0" class="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
          <h4 class="font-medium text-gray-800 dark:text-gray-100 mb-2">✕ 消失议题</h4>
          <ul class="text-sm text-gray-700 dark:text-gray-300 space-y-1">
            <li v-for="item in rankingChanges.disappeared" :key="item">{{ item }}</li>
          </ul>
        </div>
      </div>
    </section>

    <!-- 五法解读：新提法 -->
    <section v-if="showSupplementalSections && newPhrases.length > 0" class="card report-section">
      <div class="report-section__head">
        <div>
          <p class="report-section__eyebrow">五法解读</p>
          <h3 class="report-section__title">新提法识别</h3>
        </div>
      </div>
      <div class="report-table-wrap">
        <table class="report-table">
          <thead>
            <tr>
              <th>新提法</th>
              <th>政策含义</th>
              <th>潜在影响</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in newPhrases" :key="row.phrase">
              <td class="font-medium text-primary-700">{{ row.phrase }}</td>
              <td>{{ row.meaning }}</td>
              <td>{{ row.impact }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <!-- 五法解读：消失提法 -->
    <section v-if="showSupplementalSections && disappearedPhrases.length > 0" class="card report-section">
      <div class="report-section__head">
        <div>
          <p class="report-section__eyebrow">五法解读</p>
          <h3 class="report-section__title">消失提法识别</h3>
        </div>
      </div>
      <div class="report-table-wrap">
        <table class="report-table">
          <thead>
            <tr>
              <th>消失提法</th>
              <th>旧版语境</th>
              <th>信号解读</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in disappearedPhrases" :key="row.phrase">
              <td class="font-medium text-gray-700 dark:text-gray-200">{{ row.phrase }}</td>
              <td>{{ row.context }}</td>
              <td>{{ row.signal }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <!-- 五法解读：重心迁移 -->
    <section v-if="showSupplementalSections && focusShift" class="card report-section">
      <div class="report-section__head">
        <div>
          <p class="report-section__eyebrow">五法解读</p>
          <h3 class="report-section__title">政策重心迁移判断</h3>
        </div>
      </div>
      <div class="space-y-4">
        <div v-if="focusShift.macro_tone" class="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
          <h4 class="font-medium text-gray-800 dark:text-gray-100 mb-2">宏观政策基调</h4>
          <p class="text-gray-700 dark:text-gray-300">{{ focusShift.macro_tone }}</p>
        </div>
        <div v-if="focusShift.priority_directions?.length > 0" class="p-4 bg-primary-50 rounded-lg">
          <h4 class="font-medium text-primary-800 mb-2">重点发展方向</h4>
          <ol class="text-sm text-primary-700 space-y-1 list-decimal list-inside">
            <li v-for="item in focusShift.priority_directions" :key="item">{{ item }}</li>
          </ol>
        </div>
        <div v-if="focusShift.risk_areas?.length > 0" class="p-4 bg-red-50 rounded-lg">
          <h4 class="font-medium text-red-800 mb-2">风险防范领域</h4>
          <ul class="text-sm text-red-700 space-y-1">
            <li v-for="item in focusShift.risk_areas" :key="item">{{ item }}</li>
          </ul>
        </div>
      </div>
    </section>

    <!-- 关键变化对比表 -->
    <section v-if="comparisonTable.length > 0" class="card report-section">
      <div class="report-section__head">
        <div>
          <p class="report-section__eyebrow">关键变化对比</p>
          <h3 class="report-section__title">新版与旧版差异表</h3>
        </div>
      </div>
      <div class="report-table-wrap">
        <table class="report-table">
          <thead>
            <tr>
              <th>维度</th>
              <th>旧版</th>
              <th>新版</th>
              <th>影响</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in comparisonTable" :key="`${row.dimension}-${row.old}-${row.new}`">
              <td>{{ row.dimension }}</td>
              <td>{{ row.old }}</td>
              <td>{{ row.new }}</td>
              <td>{{ row.impact }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <!-- 趋势表格 - Markdown 格式 -->
    <section v-if="trendMarkdown" class="card report-section">
      <div class="report-section__head">
        <div>
          <p class="report-section__eyebrow">数据表格</p>
          <h3 class="report-section__title">政策重心变化评分</h3>
        </div>
      </div>
      <div class="markdown-body prose max-w-none" v-html="renderMarkdown(trendMarkdown)"></div>
    </section>

    <!-- 图表区域 -->
    <section v-if="trendPoints.length > 0" class="grid grid-cols-1 xl:grid-cols-2 gap-6">
      <div class="card report-section">
        <div class="report-section__head">
          <div>
            <p class="report-section__eyebrow">图表解读</p>
            <h3 class="report-section__title">关键变化对比柱状图</h3>
          </div>
        </div>
        <div ref="barChartRef" class="report-chart"></div>
      </div>

      <div class="card report-section">
        <div class="report-section__head">
          <div>
            <p class="report-section__eyebrow">政策重心</p>
            <h3 class="report-section__title">新版与旧版雷达图</h3>
          </div>
        </div>
        <div ref="radarChartRef" class="report-chart"></div>
      </div>
    </section>

    <!-- 行动建议 -->
    <section v-if="actionSection" class="card report-section">
      <div class="report-section__head">
        <div>
          <p class="report-section__eyebrow">行动建议</p>
          <h3 class="report-section__title">建议优先落地事项</h3>
        </div>
      </div>
      <p v-if="actionSection.highlight" class="report-body-section__highlight mb-3">{{ actionSection.highlight }}</p>
      <div class="markdown-body prose max-w-none" v-html="actionSection.html"></div>
    </section>

    <!-- 分析亮点 -->
    <section v-if="highlightItems.length > 0" class="card report-section">
      <div class="report-section__head">
        <div>
          <p class="report-section__eyebrow">分析亮点</p>
          <h3 class="report-section__title">四个章节导读</h3>
        </div>
      </div>
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <article v-for="item in highlightItems" :key="item.title" class="highlight-card">
          <p class="highlight-card__title">{{ item.title }}</p>
          <p class="highlight-card__text">{{ item.content }}</p>
        </article>
      </div>
    </section>

    <!-- 正文分析 - 只要有内容就显示 -->
    <section v-if="sectionItems.length > 0" class="card report-section">
      <div class="report-section__head">
        <div>
          <p class="report-section__eyebrow">正文分析</p>
          <h3 class="report-section__title">完整政策分析报告</h3>
        </div>
      </div>
      <div class="space-y-5">
        <article v-for="item in sectionItems" :key="item.title" class="report-body-section">
          <div class="report-body-section__head">
            <h4>{{ item.title }}</h4>
            <p v-if="item.highlight" class="report-body-section__highlight">{{ item.highlight }}</p>
          </div>
          <div class="markdown-body prose max-w-none" v-html="item.html"></div>
        </article>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch, nextTick } from 'vue'
import * as echarts from 'echarts'
import type { ECharts, EChartsOption } from 'echarts'
import type { StructuredReport, WordingChange, RankingChanges, NewPhrase, DisappearedPhrase, FocusShift } from '@/types/report'
import { renderMarkdown } from '@/utils/markdown'

interface SectionItem {
  key: keyof StructuredReport['sections']
  title: string
  highlight: string
  html: string
}

type SectionKey = keyof StructuredReport['sections']

const props = defineProps<{
  report: StructuredReport | null
  fallbackSummary?: string
}>()

const barChartRef = ref<HTMLDivElement | null>(null)
const radarChartRef = ref<HTMLDivElement | null>(null)
let barChart: ECharts | null = null
let radarChart: ECharts | null = null
let themeObserver: MutationObserver | null = null

// 增强容错：显示摘要优先使用结构化数据
const displaySummary = computed(() => {
  return props.report?.executive_summary || props.fallbackSummary || '暂无摘要'
})

const renderedSummary = computed(() => renderMarkdown(displaySummary.value))

const headlineMetrics = computed(() => props.report?.headline_metrics || [])
const comparisonTable = computed(() => props.report?.comparison_table || [])
const trendPoints = computed(() => props.report?.trend_points || [])

// 五法解读数据
const wordingChanges = computed<WordingChange[]>(() => props.report?.wording_changes || [])
const rankingChanges = computed<RankingChanges | null>(() => props.report?.ranking_changes || null)
const newPhrases = computed<NewPhrase[]>(() => props.report?.new_phrases || [])
const disappearedPhrases = computed<DisappearedPhrase[]>(() => props.report?.disappeared_phrases || [])
const focusShift = computed<FocusShift | null>(() => props.report?.focus_shift || null)

// 生成趋势表格的 Markdown
const trendMarkdown = computed(() => {
  const points = trendPoints.value
  if (!points || points.length === 0) return ''

  const header = '| 维度 | 旧版分值 | 新版分值 | 变化 |\n| --- | --- | --- | --- |'
  const rows = points.map(p => `| ${p.name} | ${p.old} | ${p.new} | ${p.delta > 0 ? '+' : ''}${p.delta} |`).join('\n')
  return header + '\n' + rows
})

const highlightItems = computed(() => {
  const highlights = props.report?.section_highlights
  return [
    { title: '总体变化概览', content: highlights?.overview || '' },
    { title: '关键变化拆解', content: highlights?.key_changes || '' },
    { title: '影响评估', content: highlights?.impacts || '' },
    { title: '应对建议', content: highlights?.actions || '' }
  ].filter(item => item.content)
})

const actionSection = computed(() => {
  const sections = props.report?.sections
  const highlights = props.report?.section_highlights
  const html = renderMarkdown(sections?.actions || '')
  if (html.trim() === '') return null
  return {
    key: 'actions' as SectionKey,
    title: '应对建议',
    highlight: highlights?.actions || '',
    html
  }
})

const showSupplementalSections = computed(() => false)

const sectionItems = computed<SectionItem[]>(() => {
  const sections = props.report?.sections
  const highlights = props.report?.section_highlights
  return [
    { key: 'overview' as SectionKey, title: '总体变化概览', highlight: highlights?.overview || '', html: renderMarkdown(sections?.overview || '') },
    { key: 'key_changes' as SectionKey, title: '关键变化拆解', highlight: highlights?.key_changes || '', html: renderMarkdown(sections?.key_changes || '') },
    { key: 'impacts' as SectionKey, title: '影响评估', highlight: highlights?.impacts || '', html: renderMarkdown(sections?.impacts || '') }
  ].filter(item => item.html.trim() !== '')
})

onMounted(() => {
  renderCharts()
  window.addEventListener('resize', resizeCharts)
  // 异步渲染 Mermaid 图表
  renderMermaidCharts()

  themeObserver = new MutationObserver(() => {
    disposeCharts()
    renderCharts()
    renderMermaidCharts()
  })
  themeObserver.observe(document.documentElement, {
    attributes: true,
    attributeFilter: ['class']
  })
})

// 渲染 Mermaid 图表
async function renderMermaidCharts() {
  await nextTick()

  // 查找所有带有 data-mermaid-code 属性的元素
  const mermaidElements = document.querySelectorAll('.mermaid-chart[data-mermaid-code]')
  if (mermaidElements.length === 0) return

  // 动态加载 Mermaid
  if (!(window as any).mermaid) {
    const script = document.createElement('script')
    script.src = 'https://cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.min.js'
    document.head.appendChild(script)

    await new Promise<void>((resolve) => {
      script.onload = () => {
        (window as any).mermaid.initialize({
          startOnLoad: false,
          theme: document.documentElement.classList.contains('dark') ? 'dark' : 'default',
          securityLevel: 'loose'
        })
        resolve()
      }
      script.onerror = () => resolve() // 加载失败也继续
    })
  }

  const mermaid = (window as any).mermaid
  if (!mermaid) return

  mermaidElements.forEach(async (el, index) => {
    const code = decodeURIComponent(el.getAttribute('data-mermaid-code') || '')
    if (!code) return

    try {
      const id = `mermaid-rendered-${Date.now()}-${index}`
      const { svg } = await mermaid.render(id, code)
      el.innerHTML = svg
    } catch {
      // 渲染失败保持原样
    }
  })
}

onBeforeUnmount(() => {
  window.removeEventListener('resize', resizeCharts)
  themeObserver?.disconnect()
  disposeCharts()
})

watch(trendPoints, () => {
  renderCharts()
}, { deep: true })

function renderCharts() {
  if (!barChartRef.value || !radarChartRef.value || trendPoints.value.length === 0) return

  // 检测是否为深色模式
  const isDark = document.documentElement.classList.contains('dark')

  if (!barChart) {
    barChart = echarts.init(barChartRef.value, isDark ? 'dark' : undefined)
  }
  if (!radarChart) {
    radarChart = echarts.init(radarChartRef.value, isDark ? 'dark' : undefined)
  }

  const names = trendPoints.value.map(item => item.name)
  const oldValues = trendPoints.value.map(item => item.old)
  const newValues = trendPoints.value.map(item => item.new)
  const maxValue = Math.max(...oldValues, ...newValues, 10)

  // 深色/浅色模式适配颜色
  const textColor = isDark ? '#e2e8f0' : '#334155'
  const axisColor = isDark ? '#94a3b8' : '#475569'
  const splitLineColor = isDark ? '#334155' : '#e2e8f0'

  const barOption: EChartsOption = {
    color: [isDark ? '#64748b' : '#94a3b8', '#0f766e'],
    tooltip: { trigger: 'axis' },
    legend: { top: 0, textStyle: { color: textColor } },
    grid: { left: 30, right: 16, top: 40, bottom: 30, containLabel: true },
    xAxis: {
      type: 'category',
      data: names,
      axisLabel: { color: axisColor, interval: 0, rotate: names.length > 4 ? 20 : 0 }
    },
    yAxis: {
      type: 'value',
      axisLabel: { color: axisColor },
      splitLine: { lineStyle: { color: splitLineColor } }
    },
    series: [
      { name: '旧版', type: 'bar', barMaxWidth: 26, data: oldValues, itemStyle: { borderRadius: [6, 6, 0, 0] } },
      { name: '新版', type: 'bar', barMaxWidth: 26, data: newValues, itemStyle: { borderRadius: [6, 6, 0, 0] } }
    ]
  }

  const radarOption: EChartsOption = {
    color: [isDark ? '#94a3b8' : '#334155', '#dc2626'],
    tooltip: {},
    legend: { top: 0, textStyle: { color: textColor } },
    radar: {
      radius: '64%',
      splitNumber: 4,
      indicator: names.map(name => ({ name, max: maxValue })),
      axisName: { color: axisColor },
      splitLine: { lineStyle: { color: isDark ? '#374151' : '#dbeafe' } },
      splitArea: { areaStyle: { color: isDark ? ['rgba(55,65,81,0.3)', 'rgba(75,85,99,0.3)'] : ['rgba(226,232,240,0.18)', 'rgba(191,219,254,0.18)'] } }
    },
    series: [{
      type: 'radar',
      data: [
        { value: oldValues, name: '旧版', areaStyle: { color: isDark ? 'rgba(148,163,184,0.2)' : 'rgba(51,65,85,0.12)' } },
        { value: newValues, name: '新版', areaStyle: { color: 'rgba(220,38,38,0.16)' } }
      ]
    }]
  }

  barChart.setOption(barOption)
  radarChart.setOption(radarOption)
}

function resizeCharts() {
  barChart?.resize()
  radarChart?.resize()
}

function disposeCharts() {
  barChart?.dispose()
  radarChart?.dispose()
  barChart = null
  radarChart = null
}
</script>

<style scoped>
.report-table-wrap {
  overflow-x: auto;
}
.report-table {
  width: 100%;
  border-collapse: collapse;
}
.report-table th {
  @apply bg-gray-50 dark:bg-gray-700 px-4 py-3 text-left font-semibold text-gray-700 dark:text-gray-200 border-b-2 border-gray-200 dark:border-gray-600;
}
.report-table td {
  @apply px-4 py-3 border-b border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-300;
}
.report-table tr:hover td {
  @apply bg-gray-50 dark:bg-gray-700/50;
}

/* 深色模式适配 */
:deep(.report-hero) {
  @apply dark:bg-gray-800/50 dark:border-gray-700;
}
:deep(.report-eyebrow) {
  @apply dark:text-teal-400;
}
:deep(.report-title) {
  @apply dark:text-gray-100;
}
:deep(.report-summary) {
  @apply dark:text-gray-300;
}
:deep(.metric-card) {
  @apply dark:bg-gray-700/50 dark:border-gray-600;
}
:deep(.metric-card__title) {
  @apply dark:text-gray-400;
}
:deep(.metric-card__value) {
  @apply dark:text-gray-100;
}
:deep(.metric-card__delta) {
  @apply dark:text-teal-400;
}
:deep(.metric-card__insight) {
  @apply dark:text-gray-400;
}
:deep(.report-section) {
  @apply dark:bg-gray-800 dark:border-gray-700;
}
:deep(.report-section__eyebrow) {
  @apply dark:text-gray-400;
}
:deep(.report-section__title) {
  @apply dark:text-gray-100;
}
:deep(.highlight-card) {
  @apply dark:bg-gray-700/50 dark:border-gray-600;
}
:deep(.highlight-card__title) {
  @apply dark:text-gray-100;
}
:deep(.highlight-card__text) {
  @apply dark:text-gray-400;
}
:deep(.report-body-section) {
  @apply dark:bg-gray-800 dark:border-gray-700;
}
:deep(.report-body-section__head h4) {
  @apply dark:text-gray-100;
}
:deep(.report-body-section__highlight) {
  @apply dark:text-gray-400;
}

/* Markdown 深色模式 */
:deep(.markdown-body) {
  @apply dark:text-gray-300;
}
:deep(.markdown-body h1) {
  @apply dark:text-gray-100 dark:border-gray-700;
}
:deep(.markdown-body h2) {
  @apply dark:text-gray-100;
}
:deep(.markdown-body h3) {
  @apply dark:text-gray-200;
}
:deep(.markdown-body code) {
  @apply dark:bg-gray-700 dark:text-gray-200;
}
:deep(.markdown-body pre) {
  @apply dark:bg-gray-900 dark:text-gray-200;
}
:deep(.markdown-body blockquote) {
  @apply dark:border-teal-500 dark:text-gray-400;
}
:deep(.markdown-body th) {
  @apply dark:bg-gray-700 dark:text-gray-200;
}
:deep(.markdown-body td) {
  @apply dark:border-gray-700;
}

/* 五法解读深色模式 */
:deep(.bg-green-50) {
  @apply dark:bg-green-900/30;
}
:deep(.text-green-800) {
  @apply dark:text-green-300;
}
:deep(.text-green-700) {
  @apply dark:text-green-400;
}
:deep(.bg-orange-50) {
  @apply dark:bg-orange-900/30;
}
:deep(.text-orange-800) {
  @apply dark:text-orange-300;
}
:deep(.text-orange-700) {
  @apply dark:text-orange-400;
}
:deep(.bg-blue-50) {
  @apply dark:bg-blue-900/30;
}
:deep(.text-blue-800) {
  @apply dark:text-blue-300;
}
:deep(.text-blue-700) {
  @apply dark:text-blue-400;
}
:deep(.bg-gray-50) {
  @apply dark:bg-gray-700/50;
}
:deep(.text-gray-800) {
  @apply dark:text-gray-200;
}
:deep(.text-gray-700) {
  @apply dark:text-gray-300;
}
:deep(.bg-primary-50) {
  @apply dark:bg-primary-900/30;
}
:deep(.text-primary-800) {
  @apply dark:text-primary-300;
}
:deep(.text-primary-700) {
  @apply dark:text-primary-400;
}
:deep(.bg-red-50) {
  @apply dark:bg-red-900/30;
}
:deep(.text-red-800) {
  @apply dark:text-red-300;
}
:deep(.text-red-700) {
  @apply dark:text-red-400;
}
:deep(.text-primary-700) {
  @apply dark:text-teal-400;
}
:deep(.text-gray-700) {
  @apply dark:text-gray-400;
}

/* Mermaid 图表样式 */
:deep(.mermaid-chart) {
  @apply bg-gray-50 dark:bg-gray-800 rounded-lg p-4 overflow-x-auto;
}
:deep(.mermaid-chart svg) {
  @apply max-w-full h-auto;
}
:deep(.mermaid-error) {
  @apply bg-red-50 dark:bg-red-900/30 p-4 rounded-lg text-red-600 dark:text-red-400;
}

/* 图表容器样式 */
.report-chart {
  @apply h-64 w-full;
}
</style>
