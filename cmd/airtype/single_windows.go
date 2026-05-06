//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

var (
	kernel32     = syscall.NewLazyDLL("kernel32.dll")
	user32       = syscall.NewLazyDLL("user32.dll")
	createMutex  = kernel32.NewProc("CreateMutexW")
	getLastError = kernel32.NewProc("GetLastError")
	msgBox       = user32.NewProc("MessageBoxW")
)

const errorAlreadyExists = 183

func ensureSingleInstance() bool {
	name, _ := syscall.UTF16PtrFromString("Global\\QAA-AirType-Go-SingleInstance")
	createMutex.Call(0, 0, uintptr(unsafe.Pointer(name)))
	code, _, _ := getLastError.Call()
	if code == errorAlreadyExists {
		showError("QAA AirType", "程序已在运行，请查看系统托盘。")
		return false
	}
	return true
}

func showError(title, message string) {
	t, _ := syscall.UTF16PtrFromString(title)
	m, _ := syscall.UTF16PtrFromString(message)
	const mbIconError = 0x10
	msgBox.Call(0, uintptr(unsafe.Pointer(m)), uintptr(unsafe.Pointer(t)), mbIconError)
}