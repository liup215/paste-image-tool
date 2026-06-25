package clipboard

import (
	"bytes"
	"image"
	"image/png"
	"sync"
	"time"

	"golang.design/x/clipboard"
)

// Watcher 剪切板监听器
type Watcher struct {
	isRunning    bool
	stopChan     chan struct{}
	mu           sync.Mutex
	onImageFound func(image.Image)
	lastContent  []byte
	checkInterval time.Duration
}

// NewWatcher 创建剪切板监听器
func NewWatcher() (*Watcher, error) {
	// 初始化剪切板
	if err := clipboard.Init(); err != nil {
		return nil, err
	}

	return &Watcher{
		stopChan:      make(chan struct{}),
		checkInterval: 500 * time.Millisecond,
	}, nil
}

// SetOnImageFound 设置图片发现回调
func (w *Watcher) SetOnImageFound(callback func(image.Image)) {
	w.mu.Lock()
	w.onImageFound = callback
	w.mu.Unlock()
}

// Start 开始监听
func (w *Watcher) Start() error {
	w.mu.Lock()
	if w.isRunning {
		w.mu.Unlock()
		return nil
	}
	w.isRunning = true
	w.mu.Unlock()

	// 启动监听循环
	go w.watchLoop()

	return nil
}

// Stop 停止监听
func (w *Watcher) Stop() {
	w.mu.Lock()
	if !w.isRunning {
		w.mu.Unlock()
		return
	}
	w.isRunning = false
	w.mu.Unlock()

	close(w.stopChan)
}

// IsRunning 返回是否正在运行
func (w *Watcher) IsRunning() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.isRunning
}

// watchLoop 监听循环
func (w *Watcher) watchLoop() {
	ticker := time.NewTicker(w.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopChan:
			return
		case <-ticker.C:
			w.checkClipboard()
		}
	}
}

// checkClipboard 检查剪切板内容
func (w *Watcher) checkClipboard() {
	// 读取剪切板图片
	data := clipboard.Read(clipboard.FmtImage)
	if data == nil || len(data) == 0 {
		return
	}

	// 检查是否与上次内容相同
	if bytes.Equal(data, w.lastContent) {
		return
	}

	// 更新上次内容
	w.lastContent = data

	// 解码图片
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		// 尝试其他格式
		img, _, err = image.Decode(bytes.NewReader(data))
		if err != nil {
			return
		}
	}

	// 调用回调
	w.mu.Lock()
	callback := w.onImageFound
	w.mu.Unlock()

	if callback != nil {
		callback(img)
	}
}

// ReadImage 立即读取剪切板中的图片
func (w *Watcher) ReadImage() (image.Image, error) {
	data := clipboard.Read(clipboard.FmtImage)
	if data == nil || len(data) == 0 {
		return nil, nil
	}

	// 尝试解码图片
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		// 尝试其他格式
		img, _, err = image.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
	}

	return img, nil
}

// WriteText 向剪切板写入文本
func (w *Watcher) WriteText(text string) error {
	clipboard.Write(clipboard.FmtText, []byte(text))
	return nil
}

// ReadText 从剪切板读取文本
func (w *Watcher) ReadText() string {
	data := clipboard.Read(clipboard.FmtText)
	if data == nil {
		return ""
	}
	return string(data)
}
