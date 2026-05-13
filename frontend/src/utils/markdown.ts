import { marked } from 'marked'
import DOMPurify from 'dompurify'

// 配置 marked 启用 GFM (GitHub Flavored Markdown) 表格支持
marked.setOptions({
  gfm: true,
  breaks: true
})

// 唯一标识符前缀，避免与内容冲突
const MERMAID_MARKER_PREFIX = '%%MERMAID_BLOCK_'
const MERMAID_MARKER_SUFFIX = '_%%'

/**
 * 提取并替换 Mermaid 代码块为唯一标记
 * 返回替换后的内容和 Mermaid 块映射
 */
function extractMermaidBlocks(content: string): { content: string; mermaidBlocks: string[] } {
  const mermaidBlocks: string[] = []
  let index = 0

  const processedContent = content.replace(/```mermaid\s*([\s\S]*?)```/gi, (_, code) => {
    const marker = `${MERMAID_MARKER_PREFIX}${index}${MERMAID_MARKER_SUFFIX}`
    mermaidBlocks.push(code.trim())
    index++
    return marker
  })

  return { content: processedContent, mermaidBlocks }
}

function escapeHtml(str: string): string {
  const div = document.createElement('div')
  div.textContent = str
  return div.innerHTML
}

/**
 * 将 Mermaid 标记替换为渲染后的 HTML
 */
function replaceMermaidMarkers(html: string, mermaidBlocks: string[]): string {
  for (let i = 0; i < mermaidBlocks.length; i++) {
    const marker = `${MERMAID_MARKER_PREFIX}${i}${MERMAID_MARKER_SUFFIX}`
    const mermaidDiv = `<div class="mermaid-chart" data-mermaid-code="${encodeURIComponent(mermaidBlocks[i])}"><pre class="mermaid-code"><code>${escapeHtml(mermaidBlocks[i])}</code></pre></div>`

    // 替换可能被包裹在 <p> 标签中的标记
    html = html.split(`<p>${marker}</p>`).join(mermaidDiv)
    html = html.split(marker).join(mermaidDiv)
  }

  return html
}

// 统一 Markdown 渲染入口，支持 Mermaid 图表
export function renderMarkdown(content: string): string {
  if (!content) return ''

  // 提取 Mermaid 代码块
  const { content: processedContent, mermaidBlocks } = extractMermaidBlocks(content)

  // 渲染 Markdown
  let html = marked.parse(processedContent) as string
  html = DOMPurify.sanitize(html, {
    ADD_TAGS: ['mermaid', 'div'],
    ADD_ATTR: ['class', 'data-mermaid-code']
  })

  // 替换 Mermaid 标记为占位 div
  html = replaceMermaidMarkers(html, mermaidBlocks)

  return html
}

/**
 * 异步渲染 Markdown，支持 Mermaid 图表动态渲染
 */
export async function renderMarkdownAsync(content: string): Promise<string> {
  if (!content) return ''

  const { content: processedContent, mermaidBlocks } = extractMermaidBlocks(content)

  // 渲染 Markdown
  let html = marked.parse(processedContent) as string
  html = DOMPurify.sanitize(html, {
    ADD_TAGS: ['mermaid', 'div'],
    ADD_ATTR: ['class', 'data-mermaid-code']
  })

  // 如果没有 Mermaid 块，直接返回
  if (mermaidBlocks.length === 0) {
    return html
  }

  // 加载 Mermaid
  await loadMermaid()

  // 渲染 Mermaid 图表并替换标记
  const mermaid = (window as any).mermaid
  for (let i = 0; i < mermaidBlocks.length; i++) {
    const marker = `${MERMAID_MARKER_PREFIX}${i}${MERMAID_MARKER_SUFFIX}`
    const containerId = `mermaid-${Date.now()}-${i}`

    try {
      const { svg } = await mermaid.render(containerId, mermaidBlocks[i])
      const mermaidDiv = `<div class="mermaid-chart">${svg}</div>`

      html = html.split(`<p>${marker}</p>`).join(mermaidDiv)
      html = html.split(marker).join(mermaidDiv)
    } catch {
      // 渲染失败，显示原始代码
      const errorDiv = `<div class="mermaid-chart mermaid-error"><pre><code>${escapeHtml(mermaidBlocks[i])}</code></pre></div>`
      html = html.split(`<p>${marker}</p>`).join(errorDiv)
      html = html.split(marker).join(errorDiv)
    }
  }

  return html
}

/**
 * 动态加载 Mermaid 库
 */
async function loadMermaid(): Promise<void> {
  if ((window as any).mermaid) return

  return new Promise((resolve, reject) => {
    const script = document.createElement('script')
    script.src = 'https://cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.min.js'
    script.async = true

    script.onload = () => {
      try {
        (window as any).mermaid.initialize({
          startOnLoad: false,
          theme: document.documentElement.classList.contains('dark') ? 'dark' : 'default',
          securityLevel: 'loose'
        })
        resolve()
      } catch {
        reject(new Error('Failed to initialize Mermaid'))
      }
    }

    script.onerror = () => reject(new Error('Failed to load Mermaid'))
    document.head.appendChild(script)
  })
}

/**
 * 在 DOM 中渲染所有 Mermaid 图表
 * 用于处理通过 renderMarkdown 渲染后的静态占位符
 */
export async function renderMermaidInDOM(): Promise<void> {
  const mermaidElements = document.querySelectorAll('.mermaid-chart[data-mermaid-code]')
  if (mermaidElements.length === 0) return

  await loadMermaid()
  const mermaid = (window as any).mermaid
  if (!mermaid) return

  mermaidElements.forEach(async (el, index) => {
    const code = decodeURIComponent(el.getAttribute('data-mermaid-code') || '')
    if (!code) return

    try {
      const id = `mermaid-rendered-${Date.now()}-${index}`
      const { svg } = await mermaid.render(id, code)
      el.innerHTML = svg
      el.removeAttribute('data-mermaid-code')
    } catch {
      // 渲染失败保持原样
      console.warn('Mermaid render failed for:', code)
    }
  })
}
