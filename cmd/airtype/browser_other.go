//go:build !windows

package main

import (
	"os/exec"
	"runtime"
)

func openBrowser(url string) {
	var cmd *exec.Cmd
	
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		cmd = exec.Command("cmd", "/c", "start", url)
	}
	
	cmd.Start()
}