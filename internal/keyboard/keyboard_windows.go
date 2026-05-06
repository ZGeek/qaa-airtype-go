//go:build windows

package keyboard

import (
	"time"

	"golang.org/x/sys/windows"
)

const (
	VK_SHIFT             = 0x10
	VK_INSERT            = 0x2D
	KEYEVENTF_KEYUP      = 0x0002
	KEYEVENTF_SCANCODE   = 0x0008
	KEYEVENTF_EXTENDEDKEY = 0x0001
	MAPVK_VK_TO_VSC      = 0
)

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