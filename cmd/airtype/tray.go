package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/energye/systray"
)

var (
	lastOpenMu sync.Mutex
	lastOpen   time.Time
)

func setupTray(port string) {
	systray.SetIcon(generateIcon())
	systray.SetTooltip("QAA AirType — 双击打开控制面板")

	systray.SetOnDClick(func(_ systray.IMenu) {
		openBrowserOnce(fmt.Sprintf("http://127.0.0.1:%s/", port))
	})

	mOpen := systray.AddMenuItem("打开控制面板", "在浏览器中打开控制面板")
	mOpen.Click(func() {
		openBrowserOnce(fmt.Sprintf("http://127.0.0.1:%s/", port))
	})

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("退出", "退出程序")
	mQuit.Click(func() {
		systray.Quit()
	})
}

func openBrowserOnce(url string) {
	lastOpenMu.Lock()
	defer lastOpenMu.Unlock()
	if time.Since(lastOpen) < 2*time.Second {
		return
	}
	lastOpen = time.Now()
	openBrowser(url)
}