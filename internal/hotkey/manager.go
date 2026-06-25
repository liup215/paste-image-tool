package hotkey

import (
	"fmt"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	procRegisterHotKey   = user32.NewProc("RegisterHotKey")
	procUnregisterHotKey = user32.NewProc("UnregisterHotKey")
	procGetMessageW      = user32.NewProc("GetMessageW")
	procPeekMessageW     = user32.NewProc("PeekMessageW")
	procTranslateMessage = user32.NewProc("TranslateMessage")
	procDispatchMessageW = user32.NewProc("DispatchMessageW")
	procGetCurrentThreadId = kernel32.NewProc("GetCurrentThreadId")
)

// Modifier 修饰键
type Modifier uint32

const (
	ModAlt   Modifier = 0x0001
	ModCtrl  Modifier = 0x0002
	ModShift Modifier = 0x0004
	ModWin   Modifier = 0x0008
)

// Manager 热键管理器
type Manager struct {
	hwnd       uintptr
	hotkeys    map[int]*Hotkey
	running    bool
	stopChan   chan struct{}
	onHotkey   func(id int)
}

// Hotkey 热键定义
type Hotkey struct {
	ID       int
	Modifier Modifier
	Key      uint32
}

// NewManager 创建热键管理器
func NewManager() *Manager {
	return &Manager{
		hotkeys:  make(map[int]*Hotkey),
		stopChan: make(chan struct{}),
	}
}

// Register 注册热键
func (m *Manager) Register(id int, modifier Modifier, key uint32) error {
	// 创建消息窗口
	if m.hwnd == 0 {
		hwnd, err := createMessageWindow()
		if err != nil {
			return err
		}
		m.hwnd = hwnd
	}

	// 注册热键
	ret, _, err := procRegisterHotKey.Call(
		m.hwnd,
		uintptr(id),
		uintptr(modifier),
		uintptr(key),
	)

	if ret == 0 {
		return fmt.Errorf("注册热键失败: %v", err)
	}

	// 保存热键
	m.hotkeys[id] = &Hotkey{
		ID:       id,
		Modifier: modifier,
		Key:      key,
	}

	return nil
}

// Unregister 注销热键
func (m *Manager) Unregister(id int) error {
	if _, exists := m.hotkeys[id]; !exists {
		return fmt.Errorf("热键 %d 未注册", id)
	}

	ret, _, err := procUnregisterHotKey.Call(m.hwnd, uintptr(id))
	if ret == 0 {
		return fmt.Errorf("注销热键失败: %v", err)
	}

	delete(m.hotkeys, id)
	return nil
}

// Start 开始监听热键
func (m *Manager) Start() {
	if m.running {
		return
	}
	m.running = true

	go m.messageLoop()
}

// Stop 停止监听热键
func (m *Manager) Stop() {
	if !m.running {
		return
	}
	m.running = false
	
	// 注销所有已注册的热键
	for id := range m.hotkeys {
		procUnregisterHotKey.Call(m.hwnd, uintptr(id))
	}
	m.hotkeys = make(map[int]*Hotkey)
	
	close(m.stopChan)
}

// SetOnHotkey 设置热键回调
func (m *Manager) SetOnHotkey(callback func(id int)) {
	m.onHotkey = callback
}

// messageLoop 消息循环 - 使用 PeekMessage 非阻塞方式
func (m *Manager) messageLoop() {
	var msg MSG
	const PM_REMOVE = 0x0001

	for m.running {
		// 使用 PeekMessage 非阻塞检查消息
		ret, _, _ := procPeekMessageW.Call(
			uintptr(unsafe.Pointer(&msg)),
			m.hwnd,
			0,
			0,
			PM_REMOVE,
		)

		if ret != 0 {
			// 检查是否是热键消息
			if msg.Message == 0x0312 { // WM_HOTKEY
				if m.onHotkey != nil {
					m.onHotkey(int(msg.WParam))
				}
			}

			// 翻译和分发消息
			procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
			procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
		}

		// 短暂休眠避免 CPU 占用过高
		time.Sleep(10 * time.Millisecond)
	}
}

// ParseHotkey 解析热键字符串
// 格式: "Ctrl+Shift+V"
func ParseHotkey(s string) (Modifier, uint32, error) {
	parts := strings.Split(s, "+")
	if len(parts) == 0 {
		return 0, 0, fmt.Errorf("无效的热键格式")
	}

	var modifier Modifier
	var key uint32

	for _, part := range parts {
		part = strings.TrimSpace(part)
		partUpper := strings.ToUpper(part)

		switch partUpper {
		case "CTRL", "CONTROL":
			modifier |= ModCtrl
		case "ALT":
			modifier |= ModAlt
		case "SHIFT":
			modifier |= ModShift
		case "WIN", "WINDOWS":
			modifier |= ModWin
		default:
			// 解析按键
			key = parseKey(partUpper)
			if key == 0 {
				return 0, 0, fmt.Errorf("未知的按键: %s", part)
			}
		}
	}

	if key == 0 {
		return 0, 0, fmt.Errorf("未指定按键")
	}

	return modifier, key, nil
}

// parseKey 解析按键字符串
func parseKey(s string) uint32 {
	// 字母
	if len(s) == 1 && s[0] >= 'A' && s[0] <= 'Z' {
		return uint32(s[0])
	}

	// 数字
	if len(s) == 1 && s[0] >= '0' && s[0] <= '9' {
		return uint32(s[0])
	}

	// F1-F12
	if strings.HasPrefix(s, "F") {
		n := 0
		for i := 1; i < len(s); i++ {
			if s[i] >= '0' && s[i] <= '9' {
				n = n*10 + int(s[i]-'0')
			}
		}
		if n >= 1 && n <= 12 {
			return uint32(0x70 + n - 1) // VK_F1 = 0x70
		}
	}

	// 特殊按键
	switch s {
	case "SPACE":
		return 0x20
	case "ENTER", "RETURN":
		return 0x0D
	case "TAB":
		return 0x09
	case "ESC", "ESCAPE":
		return 0x1B
	case "BACK", "BACKSPACE":
		return 0x08
	case "DELETE", "DEL":
		return 0x2E
	case "INSERT", "INS":
		return 0x2D
	case "HOME":
		return 0x24
	case "END":
		return 0x23
	case "PAGEUP", "PGUP":
		return 0x21
	case "PAGEDOWN", "PGDN":
		return 0x22
	case "UP":
		return 0x26
	case "DOWN":
		return 0x28
	case "LEFT":
		return 0x25
	case "RIGHT":
		return 0x27
	}

	return 0
}

// createMessageWindow 创建隐藏的消息窗口用于接收热键消息
func createMessageWindow() (uintptr, error) {
	user32 := syscall.NewLazyDLL("user32.dll")
	
	// 注册窗口类
	className := syscall.StringToUTF16Ptr("HotkeyMessageWindow")
	wndProc := syscall.NewCallback(func(hwnd uintptr, msg uint32, wParam uintptr, lParam uintptr) uintptr {
		// 窗口过程函数
		if msg == 0x0312 { // WM_HOTKEY
			// 热键消息，返回 1 表示已处理
			return 1
		}
		// 调用默认窗口过程
		defWindowProc := user32.NewProc("DefWindowProcW")
		ret, _, _ := defWindowProc.Call(hwnd, uintptr(msg), wParam, lParam)
		return ret
	})
	
	// 注册窗口类
	var wndClass struct {
		style       uint32
		wndProc     uintptr
		clsExtra    int32
		wndExtra    int32
		hInstance   uintptr
		hIcon       uintptr
		hCursor     uintptr
		hbrBackground uintptr
		lpszMenuName  *uint16
		lpszClassName *uint16
	}
	
	wndClass.wndProc = wndProc
	wndClass.lpszClassName = className
	
	procRegisterClass := user32.NewProc("RegisterClassW")
	ret, _, _ := procRegisterClass.Call(uintptr(unsafe.Pointer(&wndClass)))
	if ret == 0 {
		// 类可能已存在，继续
	}
	
	// 创建隐藏窗口
	procCreateWindow := user32.NewProc("CreateWindowExW")
	hwnd, _, _ := procCreateWindow.Call(
		0,                                    // dwExStyle
		uintptr(unsafe.Pointer(className)),  // lpClassName
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("HotkeyWindow"))), // lpWindowName
		0,                                    // dwStyle
		0, 0, 0, 0,                          // x, y, width, height
		0,                                    // hWndParent
		0,                                    // hMenu
		0,                                    // hInstance
		0,                                    // lpParam
	)
	
	if hwnd == 0 {
		return 0, fmt.Errorf("创建消息窗口失败")
	}
	
	return hwnd, nil
}

// MSG 结构体
type MSG struct {
	HWnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct {
		X, Y int32
	}
}
