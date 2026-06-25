package input

import (
	"fmt"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

var (
	user32                  = syscall.NewLazyDLL("user32.dll")
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	procGetForegroundWindow = user32.NewProc("GetForegroundWindow")
	procGetWindowThreadProc = user32.NewProc("GetWindowThreadProcessId")
	procAttachThreadInput   = user32.NewProc("AttachThreadInput")
	procGetCurrentThreadId  = kernel32.NewProc("GetCurrentThreadId")
	procSendInput           = user32.NewProc("SendInput")
	procVkKeyScanW          = user32.NewProc("VkKeyScanW")
	procMapVirtualKeyW      = user32.NewProc("MapVirtualKeyW")
)

// Injector 输入注入器
type Injector struct{}

// NewInjector 创建输入注入器
func NewInjector() *Injector {
	return &Injector{}
}

// InjectText 向当前焦点窗口注入文本
func (i *Injector) InjectText(text string) error {
	// 获取当前前台窗口
	hwnd := getForegroundWindow()
	if hwnd == 0 {
		return fmt.Errorf("无法获取前台窗口")
	}

	// 将文本转换为 UTF-16
	utf16Text := utf16.Encode([]rune(text))

	// 发送每个字符
	for _, char := range utf16Text {
		if err := i.sendUnicodeChar(char); err != nil {
			return err
		}
	}

	return nil
}

// sendUnicodeChar 发送 Unicode 字符
func (i *Injector) sendUnicodeChar(char uint16) error {
	// INPUT 结构体
	const INPUT_KEYBOARD = 1
	const KEYEVENTF_UNICODE = 0x0004
	const KEYEVENTF_KEYUP = 0x0002

	// 构造输入事件
	inputs := make([]INPUT, 2)

	// 按下键
	inputs[0].Type = INPUT_KEYBOARD
	inputs[0].Ki.WScan = char
	inputs[0].Ki.DwFlags = KEYEVENTF_UNICODE

	// 释放键
	inputs[1].Type = INPUT_KEYBOARD
	inputs[1].Ki.WScan = char
	inputs[1].Ki.DwFlags = KEYEVENTF_UNICODE | KEYEVENTF_KEYUP

	// 发送输入
	ret, _, err := procSendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		unsafe.Sizeof(inputs[0]),
	)

	if ret == 0 {
		return fmt.Errorf("SendInput 失败: %v", err)
	}

	return nil
}

// sendAsciiChar 发送 ASCII 字符（备选方案）
func (i *Injector) sendAsciiChar(char byte) error {
	// 获取虚拟键码
	var vk uint16
	if char >= 'a' && char <= 'z' {
		vk = uint16(char - 'a' + 0x41) // 转换为大写虚拟键码
	} else if char >= 'A' && char <= 'Z' {
		vk = uint16(char - 'A' + 0x41)
	} else if char >= '0' && char <= '9' {
		vk = uint16(char)
	} else {
		// 特殊字符映射
		switch char {
		case ':':
			vk = 0xBA // VK_OEM_1
		case '\\':
			vk = 0xDC // VK_OEM_5
		case '/':
			vk = 0xBF // VK_OEM_2
		case '.':
			vk = 0xBE // VK_OEM_PERIOD
		case '-':
			vk = 0xBD // VK_OEM_MINUS
		case '_':
			vk = 0xBD // VK_OEM_MINUS (需要 Shift)
		default:
			// 对于其他字符，使用 Unicode 输入
			return i.sendUnicodeChar(uint16(char))
		}
	}

	// 构造输入事件
	const INPUT_KEYBOARD = 1
	inputs := make([]INPUT, 2)

	// 按下键
	inputs[0].Type = INPUT_KEYBOARD
	inputs[0].Ki.WVk = vk

	// 释放键
	inputs[1].Type = INPUT_KEYBOARD
	inputs[1].Ki.WVk = vk
	inputs[1].Ki.DwFlags = 0x0002 // KEYEVENTF_KEYUP

	// 发送输入
	ret, _, err := procSendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		unsafe.Sizeof(inputs[0]),
	)

	if ret == 0 {
		return fmt.Errorf("SendInput 失败: %v", err)
	}

	return nil
}

// getForegroundWindow 获取前台窗口句柄
func getForegroundWindow() uintptr {
	hwnd, _, _ := procGetForegroundWindow.Call()
	return hwnd
}

// INPUT 结构体对应 Windows API
type INPUT struct {
	Type uint32
	Ki   KEYBDINPUT
	_    [8]byte // padding to match union size
}

// KEYBDINPUT 结构体
type KEYBDINPUT struct {
	WVk         uint16
	WScan       uint16
	DwFlags     uint32
	Time        uint32
	DwExtraInfo uintptr
}
