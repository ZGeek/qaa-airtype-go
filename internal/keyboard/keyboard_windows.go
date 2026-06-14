//go:build windows

package keyboard

import (
	"fmt"
	"math"
	"sync"
	"time"
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
	MAPVK_VK_TO_VSC       = 0

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
