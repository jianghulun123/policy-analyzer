# Policy Analyzer - 政策文件对比分析工具

<div align="center">

![Version](https://img.shields.io/badge/version-1.0.1-blue.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)
![Platform](https://img.shields.io/badge/platform-Windows-lightgrey.svg)

**政策文件对比分析桌面应用**


</div>

---

## 📖 项目简介

Policy Analyzer 是一款基于 AI 的政策文件对比分析桌面应用，支持：

- 📄 **文档解析** - 支持 PDF、DOCX、TXT、Markdown 格式
- 🤖 **AI 分析** - 多模型支持（DeepSeek、OpenAI 等）
- 🔄 **工作流引擎** - 可自定义 YAML 工作流模板
- 📊 **报告导出** - 支持 Markdown、PDF 格式

## 🚀 快速开始

### 系统要求

- Windows 10/11 64位
- Go 1.24+
- Node.js 18+

### 安装依赖

```bash
# 后端依赖
go mod download

# 前端依赖
cd frontend
npm install
```

### 开发模式

```bash
# 启动开发服务器
wails dev
```

### 构建发布

```bash
# 构建生产版本
wails build
```

## 📁 项目结构

```
policy-analyzer/
├── ai/                 # AI 提供商管理
├── config/             # 配置管理
├── document/           # 文档解析服务
├── errors/             # 统一错误处理
├── logger/             # 结构化日志
├── report/             # 报告生成器
├── storage/            # 数据存储层
├── workflow/           # 工作流引擎
├── frontend/           # Vue 3 前端
│   ├── src/
│   │   ├── api/        # API 封装
│   │   ├── stores/     # Pinia 状态
│   │   └── views/      # 页面组件
│   └── package.json
├── main.go             # 应用入口
├── app.go              # 应用逻辑
└── wails.json          # Wails 配置
```

## 🔧 配置说明

配置文件位于 `~/.policy-analyzer/config/settings.yaml`

```yaml
providers:
  - name: DeepSeek
    type: openai
    enabled: true
    api_key: your-api-key
    base_url: https://api.deepseek.com/v1
    default_model: deepseek-chat
```

## 📋 工作流模板

工作流模板位于 `~/.policy-analyzer/templates/`

```yaml
id: standard-compare
name: 标准政策对比
description: 对比分析两份政策文件的差异
version: 1.0.1

inputs:
  - name: old_policy
    type: document
    required: true
    description: 旧版政策文件
  - name: new_policy
    type: document
    required: true
    description: 新版政策文件

steps:
  - id: compare
    name: 差异对比
    type: ai-chat
    inputs:
      prompt: "请对比以下政策文件..."
```

## 🛠️ 技术栈

### 后端
- **Go 1.24** - 核心运行时
- **Wails v2** - 桌面应用框架
- **pdfcpu** - PDF 解析
- **unioffice** - DOCX 解析
- **pongo2** - 模板引擎
- **SQLite** - 本地存储

### 前端
- **Vue 3** - UI 框架
- **TypeScript** - 类型安全
- **TailwindCSS** - 样式系统
- **Pinia** - 状态管理
- **Vue Router** - 路由

## 📝 开发进度

详见 [DEVELOPMENT_PROGRESS.md](./DEVELOPMENT_PROGRESS.md)

## 🤝 贡献指南

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

---

<div align="center">

**基于Go的AI政策分析工具**


</div>
