//go:build !windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func ensureSingleInstance() bool {
	lockFile := filepath.Join(os.TempDir(), "qaa-airtype-go.lock")
	
	f, err := os.OpenFile(lockFile, os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		fmt.Println("程序已在运行")
		return false
	}
	
	f.Close()
	
	return true
}

func showError(title, message string) {
	fmt.Printf("%s: %s\n", title, message)
}