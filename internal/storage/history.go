package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// HistoryEntry 历史记录条目
type HistoryEntry struct {
	Path      string    `json:"path"`
	Name      string    `json:"name"`
	Format    string    `json:"format"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"createdAt"`
}

// HistoryManager 历史记录管理器
type HistoryManager struct {
	entries    []HistoryEntry
	maxEntries int
	historyPath string
}

// NewHistoryManager 创建历史记录管理器
func NewHistoryManager(maxEntries int) (*HistoryManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	historyPath := filepath.Join(homeDir, ".paste-image-tool", "history.json")

	hm := &HistoryManager{
		maxEntries:  maxEntries,
		historyPath: historyPath,
		entries:     []HistoryEntry{},
	}

	// 加载历史记录
	hm.Load()

	return hm, nil
}

// Load 从文件加载历史记录
func (hm *HistoryManager) Load() error {
	if _, err := os.Stat(hm.historyPath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(hm.historyPath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &hm.entries)
}

// Save 保存历史记录到文件
func (hm *HistoryManager) Save() error {
	// 确保目录存在
	dir := filepath.Dir(hm.historyPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(hm.entries, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(hm.historyPath, data, 0644)
}

// Add 添加历史记录
func (hm *HistoryManager) Add(entry HistoryEntry) error {
	// 添加到开头
	hm.entries = append([]HistoryEntry{entry}, hm.entries...)

	// 限制数量
	if len(hm.entries) > hm.maxEntries {
		hm.entries = hm.entries[:hm.maxEntries]
	}

	return hm.Save()
}

// GetAll 获取所有历史记录
func (hm *HistoryManager) GetAll() []HistoryEntry {
	return hm.entries
}

// GetRecent 获取最近 N 条历史记录
func (hm *HistoryManager) GetRecent(n int) []HistoryEntry {
	if n > len(hm.entries) {
		n = len(hm.entries)
	}
	return hm.entries[:n]
}

// Delete 删除指定路径的历史记录
func (hm *HistoryManager) Delete(path string) error {
	for i, entry := range hm.entries {
		if entry.Path == path {
			hm.entries = append(hm.entries[:i], hm.entries[i+1:]...)
			return hm.Save()
		}
	}
	return nil
}

// Clear 清空历史记录
func (hm *HistoryManager) Clear() error {
	hm.entries = []HistoryEntry{}
	return hm.Save()
}

// UpdateMaxEntries 更新最大条目数
func (hm *HistoryManager) UpdateMaxEntries(max int) {
	hm.maxEntries = max
	if len(hm.entries) > max {
		hm.entries = hm.entries[:max]
		hm.Save()
	}
}
