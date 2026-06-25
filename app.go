package main

import (
	"context"
	"fmt"
	"image"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/22569/paste-image-tool/internal/clipboard"
	"github.com/22569/paste-image-tool/internal/config"
	"github.com/22569/paste-image-tool/internal/hotkey"
	"github.com/22569/paste-image-tool/internal/input"
	"github.com/22569/paste-image-tool/internal/storage"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// App 应用服务
type App struct {
	ctx        context.Context
	config     *config.ConfigManager
	clipboard  *clipboard.Watcher
	storage    *storage.ImageSaver
	injector   *input.Injector
	hotkey     *hotkey.Manager
	history    *storage.HistoryManager
	onImageSaved func(string)
	lastSavedPath string  // 最后保存的图片路径
	wailsApp   *application.App // Wails 应用实例
}

// NewApp 创建应用
func NewApp() *App {
	app := &App{}
	
	// 初始化配置管理器（在 Wails 启动前）
	cfgManager, err := config.NewConfigManager()
	if err != nil {
		log.Printf("加载配置失败: %v", err)
		// 创建最小配置管理器
		homeDir, _ := os.UserHomeDir()
		configDir := filepath.Join(homeDir, ".paste-image-tool")
		os.MkdirAll(configDir, 0755)
		cfgManager = &config.ConfigManager{}
		cfgManager.Update(config.DefaultConfig())
	}
	app.config = cfgManager
	
	return app
}

// SetApp 设置 Wails 应用实例
func (a *App) SetApp(app *application.App) {
	a.wailsApp = app
}

// InitInBackground 在后台初始化组件
func (a *App) InitInBackground() {
	log.Printf("InitInBackground 被调用...")
	a.initComponents()
}

// OnStartup 应用启动时调用 (Wails v3 生命周期方法)
func (a *App) OnStartup(ctx context.Context) {
	log.Printf("OnStartup 被调用...")
	a.ctx = ctx
	
	// 组件已在 InitInBackground 中初始化
	// 这里不需要再次初始化
}

// initComponents 初始化组件
func (a *App) initComponents() {
	log.Printf("initComponents 开始初始化...")
	
	cfg := a.config.Get()
	if cfg == nil {
		log.Printf("错误: 配置为 nil")
		return
	}
	
	log.Printf("当前配置: Hotkey=%s, SaveDirectory=%s", cfg.Hotkey, cfg.SaveDirectory)

	// 停止旧的热键管理器（如果存在）
	if a.hotkey != nil {
		log.Printf("停止旧的热键管理器...")
		a.hotkey.Stop()
		a.hotkey = nil
	}
	
	// 停止旧的剪切板监听（如果存在）
	if a.clipboard != nil {
		log.Printf("停止旧的剪切板监听...")
		a.clipboard.Stop()
		a.clipboard = nil
	}

	// 初始化历史记录管理器
	historyManager, err := storage.NewHistoryManager(cfg.MaxHistory)
	if err != nil {
		log.Printf("初始化历史记录管理器失败: %v", err)
	}
	a.history = historyManager

	// 初始化剪切板监听器（只监听，不自动保存）
	cbWatcher, err := clipboard.NewWatcher()
	if err != nil {
		log.Printf("初始化剪切板监听器失败: %v", err)
		return
	}
	a.clipboard = cbWatcher

	// 初始化存储
	a.storage = storage.NewImageSaver(cfg)

	// 初始化输入注入器
	a.injector = input.NewInjector()

	// 初始化热键管理器
	log.Printf("初始化热键管理器...")
	a.hotkey = hotkey.NewManager()
	a.hotkey.SetOnHotkey(func(id int) {
		if id == 1 {
			log.Printf("热键触发，执行粘贴...")
			if err := a.PasteLastImage(); err != nil {
				log.Printf("粘贴失败: %v", err)
			}
		}
	})

	// 注册热键
	modifier, key, err := hotkey.ParseHotkey(cfg.Hotkey)
	if err != nil {
		log.Printf("解析热键失败: %v", err)
	} else {
		if err := a.hotkey.Register(1, modifier, key); err != nil {
			log.Printf("注册热键失败: %v", err)
		} else {
			log.Printf("热键已注册: %s", cfg.Hotkey)
			a.hotkey.Start()
			log.Printf("热键管理器已启动")
		}
	}

	// 启动剪切板监听
	a.clipboard.Start()
}

// handleImageFound 处理发现的图片
func (a *App) handleImageFound(img image.Image) {
	// 保存图片
	path, err := a.storage.SaveImage(img)
	if err != nil {
		log.Printf("保存图片失败: %v", err)
		return
	}

	// 获取图片信息
	info, err := a.storage.GetImageInfo(path)
	if err == nil && a.history != nil {
		// 添加到历史记录
		entry := storage.HistoryEntry{
			Path:      path,
			Name:      info.Format,
			Format:    info.Format,
			Width:     info.Width,
			Height:    info.Height,
			Size:      info.Size,
			CreatedAt: time.Now(),
		}
		if err := a.history.Add(entry); err != nil {
			log.Printf("添加历史记录失败: %v", err)
		}
	}

	// 通知前端
	if a.onImageSaved != nil {
		a.onImageSaved(path)
	}

	log.Printf("图片已保存: %s", path)
}

// PasteLastImage 粘贴剪切板中的图片路径（异步执行，立即返回）
// 流程：读取剪切板图片 -> 保存 -> 插入路径
func (a *App) PasteLastImage() error {
	log.Printf("开始粘贴图片...")
	
	// 从剪切板读取图片
	img, err := a.clipboard.ReadImage()
	if err != nil {
		log.Printf("读取剪切板失败: %v", err)
		return fmt.Errorf("读取剪切板失败: %w", err)
	}

	if img == nil {
		log.Printf("剪切板中没有图片")
		return fmt.Errorf("剪切板中没有图片，请先截图")
	}

	log.Printf("读取到图片，开始异步处理...")
	
	// 异步执行保存和注入，避免阻塞热键响应
	go func(imageToSave image.Image) {
		// 保存图片
		path, err := a.storage.SaveImage(imageToSave)
		if err != nil {
			log.Printf("保存图片失败: %v", err)
			return
		}

		log.Printf("图片已保存到: %s", path)

		// 格式化路径
		formattedPath := a.storage.GetFormattedPath(path)
		log.Printf("格式化后的路径: %s", formattedPath)

		// 注入文本
		log.Printf("开始注入文本...")
		if err := a.injector.InjectText(formattedPath); err != nil {
			log.Printf("注入文本失败: %v", err)
			return
		}

		log.Printf("粘贴完成")
	}(img)
	
	return nil
}

// GetConfig 获取配置
func (a *App) GetConfig() *config.Config {
	return a.config.Get()
}

// UpdateConfig 更新配置
func (a *App) UpdateConfig(cfg *config.Config) error {
	log.Printf("UpdateConfig 被调用")
	
	if cfg == nil {
		log.Printf("错误: 配置为 nil")
		return fmt.Errorf("配置不能为空")
	}
	
	if a.config == nil {
		log.Printf("错误: 配置管理器未初始化")
		return fmt.Errorf("配置管理器未初始化")
	}
	
	log.Printf("更新配置: SaveDirectory=%s, ImageFormat=%s, Hotkey=%s", 
		cfg.SaveDirectory, cfg.ImageFormat, cfg.Hotkey)
	
	if err := a.config.Update(cfg); err != nil {
		log.Printf("配置更新失败: %v", err)
		return err
	}

	log.Printf("配置更新成功，重新初始化组件...")
	
	// 重新初始化组件
	a.initComponents()

	return nil
}

// GetSaveDirectory 获取保存目录
func (a *App) GetSaveDirectory() string {
	return a.config.Get().SaveDirectory
}

// SetSaveDirectory 设置保存目录
func (a *App) SetSaveDirectory(dir string) error {
	if a.config == nil {
		return fmt.Errorf("配置管理器未初始化")
	}
	cfg := a.config.Get()
	if cfg == nil {
		return fmt.Errorf("配置未加载")
	}
	cfg.SaveDirectory = dir
	return a.config.Save()
}

// GetHotkey 获取热键
func (a *App) GetHotkey() string {
	return a.config.Get().Hotkey
}

// SetHotkey 设置热键
func (a *App) SetHotkey(hotkeyStr string) error {
	if a.config == nil {
		return fmt.Errorf("配置管理器未初始化")
	}
	cfg := a.config.Get()
	if cfg == nil {
		return fmt.Errorf("配置未加载")
	}
	cfg.Hotkey = hotkeyStr
	return a.config.Save()
}

// GetRecentImages 获取最近保存的图片
func (a *App) GetRecentImages() []map[string]interface{} {
	if a.history == nil {
		return []map[string]interface{}{}
	}

	entries := a.history.GetAll()
	result := make([]map[string]interface{}, len(entries))
	for i, entry := range entries {
		result[i] = map[string]interface{}{
			"Path":      entry.Path,
			"Name":      entry.Name,
			"Format":    entry.Format,
			"Width":     entry.Width,
			"Height":    entry.Height,
			"Size":      entry.Size,
			"CreatedAt": entry.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return result
}

// OpenDirectory 打开保存目录
func (a *App) OpenDirectory() error {
	// 使用系统默认程序打开目录
	cmd := fmt.Sprintf("explorer \"%s\"", a.config.Get().SaveDirectory)
	return exec.Command("cmd", "/c", cmd).Start()
}

// DeleteImage 删除图片
func (a *App) DeleteImage(path string) error {
	// 删除文件
	if err := os.Remove(path); err != nil {
		return err
	}

	// 从历史记录中删除
	if a.history != nil {
		return a.history.Delete(path)
	}

	return nil
}

// ClearHistory 清空历史记录
func (a *App) ClearHistory() error {
	if a.history == nil {
		return nil
	}
	return a.history.Clear()
}

// SelectDirectory 选择目录对话框
func (a *App) SelectDirectory() (string, error) {
	// 使用 OpenFileDialog 并设置 CanChooseDirectories
	if a.wailsApp == nil {
		return "", fmt.Errorf("应用未初始化")
	}
	
	dialog := a.wailsApp.Dialog.OpenFile()
	dialog.SetTitle("选择图片保存目录")
	dialog.CanChooseDirectories(true)
	dialog.CanChooseFiles(false)
	
	result, err := dialog.PromptForSingleSelection()
	if err != nil {
		return "", err
	}
	
	return result, nil
}

// OnShutdown 应用关闭时调用 (Wails v3 生命周期方法)
func (a *App) OnShutdown() {
	if a.clipboard != nil {
		a.clipboard.Stop()
	}
	if a.hotkey != nil {
		a.hotkey.Stop()
	}
}
