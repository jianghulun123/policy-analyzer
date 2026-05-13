package config

// Config 应用配置结构
type Config struct {
	App       AppConfig        `yaml:"app"`
	Providers []ProviderConfig `yaml:"providers"`
	Workflows WorkflowConfig   `yaml:"workflows"`
	Storage   StorageConfig    `yaml:"storage"`
	Privacy   PrivacyConfig    `yaml:"privacy"`
}

// AppConfig 应用配置
type AppConfig struct {
	Version  string         `yaml:"version"`
	Theme    string         `yaml:"theme"`    // light/dark/auto
	Language string         `yaml:"language"` // zh-CN/en-US
	Defaults DefaultsConfig `yaml:"defaults"`
	UI       UIConfig       `yaml:"ui"`
}

// DefaultsConfig 默认配置
type DefaultsConfig struct {
	Workspace string `yaml:"workspace"`
	Workflow  string `yaml:"workflow"`
}

// UIConfig UI配置
type UIConfig struct {
	SidebarCollapsed bool   `yaml:"sidebar_collapsed"`
	FontSize         int    `yaml:"font_size"`
	CodeTheme        string `yaml:"code_theme"`
}

// ProviderConfig AI提供商配置
type ProviderConfig struct {
	Name         string        `yaml:"name"`          // 显示名称
	Type         string        `yaml:"type"`          // openai/claude/deepseek/openai-compatible
	Endpoint     string        `yaml:"endpoint"`      // API地址
	APIKey       string        `yaml:"api_key"`       // 密钥（加密存储）
	DefaultModel string        `yaml:"default_model"` // 默认模型
	Models       []ModelConfig `yaml:"models"`        // 可用模型列表
	Timeout      int           `yaml:"timeout"`       // 超时时间（秒）
	Enabled      bool          `yaml:"enabled"`       // 是否启用
}

// ModelConfig 模型配置
type ModelConfig struct {
	ID            string `yaml:"id"`             // 模型ID
	Name          string `yaml:"name"`           // 显示名称
	MaxTokens     int    `yaml:"max_tokens"`     // 最大输出Token数
	ContextWindow int    `yaml:"context_window"` // 上下文窗口大小（总Token数）
}

// WorkflowConfig 工作流配置
type WorkflowConfig struct {
	TemplateDirs []string `yaml:"template_dirs"`
	AutoUpdate   bool     `yaml:"auto_update"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	BaseDir          string   `yaml:"base_dir"`
	MaxFileSize      int      `yaml:"max_file_size"` // MB
	AllowedExts      []string `yaml:"allowed_extensions"`
	LogAutoCleanup   bool     `yaml:"log_auto_cleanup"`
	LogRetentionDays int      `yaml:"log_retention_days"`
}

// PrivacyConfig 隐私配置
type PrivacyConfig struct {
	Telemetry       bool `yaml:"telemetry"`
	AutoUpdateCheck bool `yaml:"auto_update_check"`
	CrashReport     bool `yaml:"crash_report"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	homeDir := "~/.policy-analyzer" // 实际使用时替换

	return &Config{
		App: AppConfig{
			Version:  "1.0.1",
			Theme:    "light",
			Language: "zh-CN",
			Defaults: DefaultsConfig{
				Workspace: "default",
				Workflow:  "standard-compare",
			},
			UI: UIConfig{
				SidebarCollapsed: false,
				FontSize:         14,
				CodeTheme:        "github",
			},
		},
		Providers: []ProviderConfig{
			{
				Name:         "DeepSeek",
				Type:         "deepseek",
				Endpoint:     "https://api.deepseek.com/v1",
				APIKey:       "",
				DefaultModel: "deepseek-chat",
				Models: []ModelConfig{
					{ID: "deepseek-chat", Name: "DeepSeek Chat", MaxTokens: 8192, ContextWindow: 64000},
					{ID: "deepseek-reasoner", Name: "DeepSeek R1", MaxTokens: 8192, ContextWindow: 64000},
				},
				Timeout: 60,
				Enabled: true,
			},
			{
				Name:         "OpenAI",
				Type:         "openai",
				Endpoint:     "https://api.openai.com/v1",
				APIKey:       "",
				DefaultModel: "gpt-4",
				Models: []ModelConfig{
					{ID: "gpt-4", Name: "GPT-4", MaxTokens: 8192, ContextWindow: 128000},
					{ID: "gpt-3.5-turbo", Name: "GPT-3.5 Turbo", MaxTokens: 4096, ContextWindow: 16384},
				},
				Timeout: 60,
				Enabled: true,
			},
			{
				Name:         "本地模型(Ollama)",
				Type:         "openai-compatible",
				Endpoint:     "http://localhost:11434/v1",
				APIKey:       "",
				DefaultModel: "qwen:14b",
				Models: []ModelConfig{
					{ID: "qwen:14b", Name: "Qwen 14B", MaxTokens: 8192, ContextWindow: 32768},
					{ID: "llama3:8b", Name: "Llama 3 8B", MaxTokens: 4096, ContextWindow: 8192},
				},
				Timeout: 120,
				Enabled: false,
			},
		},
		Workflows: WorkflowConfig{
			TemplateDirs: []string{"./templates/workflows"},
			AutoUpdate:   true,
		},
		Storage: StorageConfig{
			BaseDir:          homeDir,
			MaxFileSize:      100, // 100MB
			AllowedExts:      []string{".pdf", ".docx", ".txt", ".md", ".png", ".jpg", ".jpeg"},
			LogAutoCleanup:   true,
			LogRetentionDays: 7,
		},
		Privacy: PrivacyConfig{
			Telemetry:       false,
			AutoUpdateCheck: true,
			CrashReport:     false,
		},
	}
}
