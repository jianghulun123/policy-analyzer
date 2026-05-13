package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"policy-analyzer/logger"
)

// 静态资源嵌入（前端构建产物）
//
//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// 创建应用实例
	app := NewApp()

	// 创建Wails应用
	err := wails.Run(&options.App{
		Title:  "Policy Analyzer - 政策文件对比分析工具",
		Width:  1280,
		Height: 800,
		MinWidth: 1024,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 249, G: 250, B: 251, A: 255}, // gray-50
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		logger.Fatal("应用启动失败", logger.F("error", err.Error()))
	}
}