package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config 应用配置
type Config struct {
	// 保存目录
	SaveDirectory string `json:"saveDirectory"`
	// 文件名格式
	FilenameTemplate string `json:"filenameTemplate"`
	// 图片格式: png, jpg, webp
	ImageFormat string `json:"imageFormat"`
	// JPEG 质量 1-100
	JPEGQuality int `json:"jpegQuality"`
	// 全局快捷键
	Hotkey string `json:"hotkey"`
	// 启动时最小化到托盘
	StartMinimized bool `json:"startMinimized"`
	// 路径格式: absolute, relative, markdown, html, url
	PathFormat string `json:"pathFormat"`
	// 是否启用历史记录
	EnableHistory bool `json:"enableHistory"`
	// 历史记录最大数量
	MaxHistory int `json:"maxHistory"`
	// 是否自动压缩大尺寸图片
	AutoCompress bool `json:"autoCompress"`
	// 最大尺寸限制 (像素)
	MaxDimension int `json:"maxDimension"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	defaultSaveDir := filepath.Join(homeDir, "Pictures", "PasteImages")

	return &Config{
		SaveDirectory:    defaultSaveDir,
		FilenameTemplate: "paste_{date}_{time}.png",
		ImageFormat:      "png",
		JPEGQuality:      85,
		Hotkey:           "Ctrl+Shift+Insert",  // 默认热键，避免与常用工具冲突
		StartMinimized:   false,
		PathFormat:       "absolute",
		EnableHistory:    true,
		MaxHistory:       100,
		AutoCompress:     true,
		MaxDimension:     1920,
	}
}

// ConfigManager 配置管理器
type ConfigManager struct {
	config     *Config
	configPath string
}

// NewConfigManager 创建配置管理器
func NewConfigManager() (*ConfigManager, error) {
	// 获取配置目录
	configDir, err := getConfigDir()
	if err != nil {
		return nil, err
	}

	// 确保配置目录存在
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}

	configPath := filepath.Join(configDir, "config.json")

	cm := &ConfigManager{
		configPath: configPath,
	}

	// 加载配置
	if err := cm.Load(); err != nil {
		// 如果加载失败，使用默认配置
		cm.config = DefaultConfig()
		// 保存默认配置
		_ = cm.Save()
	}

	return cm, nil
}

// Load 从文件加载配置
func (cm *ConfigManager) Load() error {
	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		return err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}

	cm.config = &cfg
	return nil
}

// Save 保存配置到文件
func (cm *ConfigManager) Save() error {
	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cm.configPath, data, 0644)
}

// Get 获取当前配置
func (cm *ConfigManager) Get() *Config {
	return cm.config
}

// Update 更新配置
func (cm *ConfigManager) Update(cfg *Config) error {
	cm.config = cfg
	return cm.Save()
}

// getConfigDir 获取配置目录
func getConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".paste-image-tool"), nil
}
