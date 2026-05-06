//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

var (
	shell32          = syscall.NewLazyDLL("shell32.dll")
	shellExecuteW    = shell32.NewProc("ShellExecuteW")
)

func openBrowser(url string) {
	shellExecuteW.Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("open"))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(url))),
		0,
		0,
		1, // SW_SHOWNORMAL
	)
}