//go:build windows

package keyboard

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	VK_SHIFT              = 0x10
	VK_INSERT             = 0x2D
	VK_RETURN             = 0x0D
	KEYEVENTF_KEYUP       = 0x0002
	KEYEVENTF_SCANCODE    = 0x0008
	KEYEVENTF_EXTENDEDKEY = 0x0001
	MOUSEEVENTF_MOVE      = 0x0001
	MOUSEEVENTF_LEFTDOWN  = 0x0002
	MOUSEEVENTF_LEFTUP    = 0x0004
	MOUSEEVENTF_RIGHTDOWN = 0x0008
	MOUSEEVENTF_RIGHTUP   = 0x0010
	MOUSEEVENTF_WHEEL     = 0x0800
	MOUSEEVENTF_HWHEEL    = 0x01000
	MAPVK_VK_TO_VSC       = 0
	VK_CONTROL            = 0x11
	VK_LWIN               = 0x5B
	VK_TAB                = 0x09

	POINTER_INPUT_TYPE_TOUCH = 2
	POINTER_FLAG_INRANGE     = 0x00000002
	POINTER_FLAG_INCONTACT   = 0x00000004
	POINTER_FLAG_DOWN        = 0x00010000
	POINTER_FLAG_UPDATE      = 0x00020000
	POINTER_FLAG_UP          = 0x00040000
	TOUCH_MASK_CONTACTAREA   = 0x00000001
	TOUCH_MASK_ORIENTATION   = 0x00000002
	TOUCH_MASK_PRESSURE      = 0x00000004
	TOUCH_FEEDBACK_DEFAULT   = 0x00000001
	TOUCH_FEEDBACK_NONE      = 0x00000003

	INPUT_KEYBOARD    = 1
	KEYEVENTF_UNICODE = 0x0004
)

var (
	touchMu       sync.Mutex
	touchInitOnce sync.Once
	touchInitErr  error
	touchSession  touchScrollSession
)

type point struct {
	X int32
	Y int32
}

type rect struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

type pointerInfo struct {
	PointerType           uint32
	PointerID             uint32
	FrameID               uint32
	PointerFlags          uint32
	SourceDevice          windows.Handle
	HwndTarget            windows.Handle
	PtPixelLocation       point
	PtHimetricLocation    point
	PtPixelLocationRaw    point
	PtHimetricLocationRaw point
	DwTime                uint32
	HistoryCount          uint32
	InputData             int32
	DwKeyStates           uint32
	PerformanceCount      uint64
	ButtonChangeType      uint32
}

type pointerTouchInfo struct {
	PointerInfo  pointerInfo
	TouchFlags   uint32
	TouchMask    uint32
	RcContact    rect
	RcContactRaw rect
	Orientation  uint32
	Pressure     uint32
}

type touchScrollSession struct {
	active bool
	start  point
	last   point
}

type KEYBDINPUT struct {
	WVk         uint16
	WScan       uint16
	DwFlags     uint32
	Time        uint32
	DwExtraInfo uintptr
}

// INPUT 必须与 Windows 的 INPUT 结构体大小一致（x64=40 字节，x86=28 字节）。
// Windows 的 INPUT 内含一个 union，最大成员是 MOUSEINPUT，因此即使只使用
// KEYBDINPUT，也必须保留尾部填充字节，否则 SendInput 会因 cbSize 校验失败而拒绝输入。
type INPUT struct {
	Type uint32
	Ki   KEYBDINPUT
	_    [8]byte
}

func Paste() error {
	user32 := windows.NewLazySystemDLL("user32.dll")
	keybdEvent := user32.NewProc("keybd_event")
	mapVirtualKeyW := user32.NewProc("MapVirtualKeyW")

	shiftScan, _, _ := mapVirtualKeyW.Call(uintptr(VK_SHIFT), uintptr(MAPVK_VK_TO_VSC))
	insertScan, _, _ := mapVirtualKeyW.Call(uintptr(VK_INSERT), uintptr(MAPVK_VK_TO_VSC))

	keybdEvent.Call(uintptr(VK_SHIFT), shiftScan, uintptr(KEYEVENTF_SCANCODE), 0)
	time.Sleep(50 * time.Millisecond)

	keybdEvent.Call(uintptr(VK_INSERT), insertScan, uintptr(KEYEVENTF_SCANCODE|KEYEVENTF_EXTENDEDKEY), 0)
	time.Sleep(20 * time.Millisecond)

	keybdEvent.Call(uintptr(VK_INSERT), insertScan, uintptr(KEYEVENTF_SCANCODE|KEYEVENTF_EXTENDEDKEY|KEYEVENTF_KEYUP), 0)
	time.Sleep(20 * time.Millisecond)

	keybdEvent.Call(uintptr(VK_SHIFT), shiftScan, uintptr(KEYEVENTF_SCANCODE|KEYEVENTF_KEYUP), 0)

	return nil
}

func StartTouchScroll() error {
	touchMu.Lock()
	defer touchMu.Unlock()

	if touchSession.active {
		_ = injectTouch("cleanup", touchSession.last, POINTER_FLAG_UP)
		touchSession.active = false
	}

	pos, err := cursorPos()
	if err != nil {
		return err
	}

	if err := injectTouch("down", pos, POINTER_FLAG_DOWN|POINTER_FLAG_INRANGE|POINTER_FLAG_INCONTACT); err != nil {
		return err
	}

	touchSession = touchScrollSession{active: true, start: pos, last: pos}
	return nil
}

func MoveTouchScroll(offsetY float64) error {
	touchMu.Lock()
	defer touchMu.Unlock()

	if !touchSession.active {
		return nil
	}

	y := touchSession.start.Y + int32(math.Round(offsetY))
	pos := point{X: touchSession.start.X, Y: y}
	touchSession.last = pos
	return injectTouch("move", pos, POINTER_FLAG_UPDATE|POINTER_FLAG_INRANGE|POINTER_FLAG_INCONTACT)
}

func EndTouchScroll() error {
	touchMu.Lock()
	defer touchMu.Unlock()

	if !touchSession.active {
		return nil
	}

	pos := touchSession.last
	touchSession.active = false
	return injectTouch("up", pos, POINTER_FLAG_UP)
}

func MoveMouse(dx float64, dy float64) error {
	x := int32(math.Round(dx))
	y := int32(math.Round(dy))
	if x == 0 && y == 0 {
		return nil
	}

	user32 := windows.NewLazySystemDLL("user32.dll")
	mouseEvent := user32.NewProc("mouse_event")
	mouseEvent.Call(uintptr(MOUSEEVENTF_MOVE), uintptr(uint32(x)), uintptr(uint32(y)), 0, 0)
	return nil
}

func LeftClick() error {
	user32 := windows.NewLazySystemDLL("user32.dll")
	mouseEvent := user32.NewProc("mouse_event")
	mouseEvent.Call(uintptr(MOUSEEVENTF_LEFTDOWN), 0, 0, 0, 0)
	time.Sleep(20 * time.Millisecond)
	mouseEvent.Call(uintptr(MOUSEEVENTF_LEFTUP), 0, 0, 0, 0)
	return nil
}

func LeftDown() error {
	user32 := windows.NewLazySystemDLL("user32.dll")
	mouseEvent := user32.NewProc("mouse_event")
	mouseEvent.Call(uintptr(MOUSEEVENTF_LEFTDOWN), 0, 0, 0, 0)
	return nil
}

func LeftUp() error {
	user32 := windows.NewLazySystemDLL("user32.dll")
	mouseEvent := user32.NewProc("mouse_event")
	mouseEvent.Call(uintptr(MOUSEEVENTF_LEFTUP), 0, 0, 0, 0)
	return nil
}

func RightClick() error {
	user32 := windows.NewLazySystemDLL("user32.dll")
	mouseEvent := user32.NewProc("mouse_event")
	mouseEvent.Call(uintptr(MOUSEEVENTF_RIGHTDOWN), 0, 0, 0, 0)
	time.Sleep(20 * time.Millisecond)
	mouseEvent.Call(uintptr(MOUSEEVENTF_RIGHTUP), 0, 0, 0, 0)
	return nil
}

func ScrollWheel(dx float64, dy float64) error {
	user32 := windows.NewLazySystemDLL("user32.dll")
	mouseEvent := user32.NewProc("mouse_event")

	wheelY := int32(math.Round(-dy))
	wheelX := int32(math.Round(dx))
	if wheelY != 0 {
		mouseEvent.Call(uintptr(MOUSEEVENTF_WHEEL), 0, 0, uintptr(uint32(wheelY)), 0)
	}
	if wheelX != 0 {
		mouseEvent.Call(uintptr(MOUSEEVENTF_HWHEEL), 0, 0, uintptr(uint32(wheelX)), 0)
	}
	return nil
}

func ZoomWheel(delta float64) error {
	user32 := windows.NewLazySystemDLL("user32.dll")
	keybdEvent := user32.NewProc("keybd_event")
	mapVirtualKeyW := user32.NewProc("MapVirtualKeyW")
	mouseEvent := user32.NewProc("mouse_event")

	ctrlScan, _, _ := mapVirtualKeyW.Call(uintptr(VK_CONTROL), uintptr(MAPVK_VK_TO_VSC))
	keybdEvent.Call(uintptr(VK_CONTROL), ctrlScan, uintptr(KEYEVENTF_SCANCODE), 0)
	mouseEvent.Call(uintptr(MOUSEEVENTF_WHEEL), 0, 0, uintptr(uint32(int32(math.Round(delta)))), 0)
	keybdEvent.Call(uintptr(VK_CONTROL), ctrlScan, uintptr(KEYEVENTF_SCANCODE|KEYEVENTF_KEYUP), 0)
	return nil
}

func TaskView() error {
	user32 := windows.NewLazySystemDLL("user32.dll")
	keybdEvent := user32.NewProc("keybd_event")
	mapVirtualKeyW := user32.NewProc("MapVirtualKeyW")

	winScan, _, _ := mapVirtualKeyW.Call(uintptr(VK_LWIN), uintptr(MAPVK_VK_TO_VSC))
	tabScan, _, _ := mapVirtualKeyW.Call(uintptr(VK_TAB), uintptr(MAPVK_VK_TO_VSC))
	keybdEvent.Call(uintptr(VK_LWIN), winScan, uintptr(KEYEVENTF_SCANCODE), 0)
	keybdEvent.Call(uintptr(VK_TAB), tabScan, uintptr(KEYEVENTF_SCANCODE), 0)
	keybdEvent.Call(uintptr(VK_TAB), tabScan, uintptr(KEYEVENTF_SCANCODE|KEYEVENTF_KEYUP), 0)
	keybdEvent.Call(uintptr(VK_LWIN), winScan, uintptr(KEYEVENTF_SCANCODE|KEYEVENTF_KEYUP), 0)
	return nil
}

func cursorPos() (point, error) {
	user32 := windows.NewLazySystemDLL("user32.dll")
	getCursorPos := user32.NewProc("GetCursorPos")

	var pos point
	ret, _, err := getCursorPos.Call(uintptr(unsafe.Pointer(&pos)))
	if ret == 0 {
		return point{}, err
	}
	return pos, nil
}

func injectTouch(phase string, pos point, flags uint32) error {
	if err := initTouchInjection(); err != nil {
		return err
	}

	user32 := windows.NewLazySystemDLL("user32.dll")
	injectTouchInput := user32.NewProc("InjectTouchInput")

	contactSize := int32(4)
	info := pointerTouchInfo{
		PointerInfo: pointerInfo{
			PointerType:     POINTER_INPUT_TYPE_TOUCH,
			PointerID:       0,
			PointerFlags:    flags,
			PtPixelLocation: pos,
		},
	}
	if flags&POINTER_FLAG_UP == 0 {
		info.TouchMask = TOUCH_MASK_CONTACTAREA
		info.RcContact = rect{Left: pos.X - contactSize, Top: pos.Y - contactSize, Right: pos.X + contactSize, Bottom: pos.Y + contactSize}
	}

	ret, _, err := injectTouchInput.Call(1, uintptr(unsafe.Pointer(&info)))
	if ret == 0 {
		return fmt.Errorf("%s: %w", phase, err)
	}
	return nil
}

func initTouchInjection() error {
	touchInitOnce.Do(func() {
		user32 := windows.NewLazySystemDLL("user32.dll")
		initializeTouchInjection := user32.NewProc("InitializeTouchInjection")
		ret, _, err := initializeTouchInjection.Call(1, uintptr(TOUCH_FEEDBACK_NONE))
		if ret == 0 {
			touchInitErr = err
		}
	})
	return touchInitErr
}

func Enter() error {
	user32 := windows.NewLazySystemDLL("user32.dll")
	keybdEvent := user32.NewProc("keybd_event")
	mapVirtualKeyW := user32.NewProc("MapVirtualKeyW")

	enterScan, _, _ := mapVirtualKeyW.Call(uintptr(VK_RETURN), uintptr(MAPVK_VK_TO_VSC))

	keybdEvent.Call(uintptr(VK_RETURN), enterScan, uintptr(KEYEVENTF_SCANCODE), 0)
	time.Sleep(20 * time.Millisecond)

	keybdEvent.Call(uintptr(VK_RETURN), enterScan, uintptr(KEYEVENTF_SCANCODE|KEYEVENTF_KEYUP), 0)

	return nil
}

// TypeText 通过 SendInput 把文本以 Unicode 键盘事件的形式直接注入输入流，
// 不经过系统剪贴板，避免污染用户已有的剪贴板内容。
func TypeText(text string) error {
	user32 := windows.NewLazySystemDLL("user32.dll")
	sendInput := user32.NewProc("SendInput")

	// 统一换行符为 \r，兼容记事本、终端、聊天框等
	text = strings.ReplaceAll(text, "\r\n", "\r")
	text = strings.ReplaceAll(text, "\n", "\r")

	// 转为 UTF-16，自动处理代理对（如 emoji）
	codes := utf16.Encode([]rune(text))

	const batchSize = 64
	for start := 0; start < len(codes); start += batchSize {
		end := start + batchSize
		if end > len(codes) {
			end = len(codes)
		}

		batch := codes[start:end]
		inputs := make([]INPUT, 0, len(batch)*2)
		for _, code := range batch {
			inputs = append(inputs, INPUT{
				Type: INPUT_KEYBOARD,
				Ki: KEYBDINPUT{
					WScan:   code,
					DwFlags: KEYEVENTF_UNICODE,
				},
			})
			inputs = append(inputs, INPUT{
				Type: INPUT_KEYBOARD,
				Ki: KEYBDINPUT{
					WScan:   code,
					DwFlags: KEYEVENTF_UNICODE | KEYEVENTF_KEYUP,
				},
			})
		}

		sent, _, _ := sendInput.Call(
			uintptr(len(inputs)),
			uintptr(unsafe.Pointer(&inputs[0])),
			unsafe.Sizeof(INPUT{}),
		)
		if sent == 0 {
			return fmt.Errorf("SendInput failed: %v", windows.GetLastError())
		}

		if end < len(codes) {
			time.Sleep(10 * time.Millisecond)
		}
	}

	return nil
}
