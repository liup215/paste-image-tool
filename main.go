package main

import (
	"embed"
	_ "embed"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

func init() {
	// 设置日志输出到文件
	homeDir, _ := os.UserHomeDir()
	logDir := filepath.Join(homeDir, ".paste-image-tool")
	os.MkdirAll(logDir, 0755)
	logFile := filepath.Join(logDir, "app.log")
	
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(f)
	}
	log.Printf("=== 应用启动 ===")
	log.Printf("日志文件: %s", logFile)
}

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var iconBytes []byte

func main() {
	// 创建应用实例
	appInstance := NewApp()

	// 创建 Wails 应用
	app := application.New(application.Options{
		Name:        "Paste Image Tool",
		Description: "剪切板图片粘贴工具",
		Services: []application.Service{
			application.NewService(appInstance),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
		Icon: iconBytes,
	})

	// 设置应用实例以便访问对话框等功能
	appInstance.SetApp(app)
	
	// 延迟初始化组件（等应用完全启动后）
	go func() {
		time.Sleep(1000 * time.Millisecond)
		appInstance.InitInBackground()
	}()

	// 创建主窗口
	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "Paste Image Tool",
		Width:            900,
		Height:           650,
		BackgroundColour: application.NewRGB(45, 45, 58), // #2d2d3a
		URL:              "/",
	})

	// 拦截关闭按钮事件：隐藏到托盘而不是退出
	window.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		window.Hide()
		e.Cancel()
	})

	// --- 系统托盘设置 ---
	systemTray := app.SystemTray.New()
	systemTray.SetIcon(iconBytes)
	systemTray.AttachWindow(window)

	trayMenu := app.NewMenu()
	toggleItem := trayMenu.Add("Hide Paste Image Tool")
	toggleItem.OnClick(func(ctx *application.Context) {
		if window.IsVisible() {
			window.Hide()
		} else {
			window.Show().Focus()
		}
	})

	// 动态更新菜单标签
	window.RegisterHook(events.Common.WindowHide, func(e *application.WindowEvent) {
		if toggleItem != nil {
			toggleItem.SetLabel("Show Paste Image Tool")
		}
	})
	window.RegisterHook(events.Common.WindowShow, func(e *application.WindowEvent) {
		if toggleItem != nil {
			toggleItem.SetLabel("Hide Paste Image Tool")
		}
	})

	trayMenu.AddSeparator()
	trayMenu.Add("Quit").OnClick(func(ctx *application.Context) {
		app.Quit()
	})

	systemTray.SetMenu(trayMenu)
	// --- 系统托盘设置结束 ---

	// 运行应用
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
