package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/QAA-Tools/qaa-airtype/go/internal/config"
	"github.com/QAA-Tools/qaa-airtype/go/internal/keyboard"
	"github.com/QAA-Tools/qaa-airtype/go/internal/network"
	"github.com/energye/systray"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/skip2/go-qrcode"
)

//go:embed web
var webFS embed.FS

var scrollUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	runtimeConfigMu sync.RWMutex
	runtimeConfig   config.Config
)

type scrollRequest struct {
	Phase   string  `json:"phase"`
	OffsetY float64 `json:"offsetY"`
}

type touchpadRequest struct {
	Action string  `json:"action"`
	DX     float64 `json:"dx"`
	DY     float64 `json:"dy"`
}

func main() {
	if !ensureSingleInstance() {
		return
	}

	cfg := config.Load()
	setRuntimeConfig(cfg)
	systray.Run(onReady(cfg), onExit)
}

func setRuntimeConfig(cfg config.Config) {
	runtimeConfigMu.Lock()
	runtimeConfig = cfg
	runtimeConfigMu.Unlock()
}

func getRuntimeConfig() config.Config {
	runtimeConfigMu.RLock()
	cfg := runtimeConfig
	runtimeConfigMu.RUnlock()
	return cfg
}

func onReady(cfg config.Config) func() {
	return func() {
		ln, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", cfg.Port))
		if err != nil {
			showError(
				"QAA AirType — 启动失败",
				fmt.Sprintf("HTTP 服务监听端口 %s 失败，该端口可能已被其他程序占用。\n\n%v", cfg.Port, err),
			)
			systray.Quit()
			return
		}

		setupTray(cfg.Port)
		go startWebServer(ln, cfg)
		go openWhenReady(cfg.Port)
	}
}

func onExit() {
}

func startWebServer(ln net.Listener, cfg config.Config) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/", indexHandler)
	r.GET("/input", inputHandler)
	r.POST("/type", typeHandler)
	r.GET("/status", statusHandler)
	r.GET("/qr", qrHandler)
	r.GET("/ips", ipsHandler)
	r.GET("/config", getConfigHandler)
	r.POST("/config", configHandler)
	r.POST("/scroll", scrollHandler)
	r.GET("/scroll-ws", scrollWebSocketHandler)
	r.POST("/touchpad", touchpadHandler)
	r.GET("/touchpad-ws", touchpadWebSocketHandler)
	r.NoRoute(staticHandler)

	fmt.Println("\n====================================")
	fmt.Println("  QAA AirType (Go Version)")
	fmt.Println("====================================")
	fmt.Printf("  Port: %s\n", cfg.Port)
	fmt.Println()
	fmt.Println("  Available URLs:")

	ips := network.GetAllIPs()
	for _, ip := range ips {
		fmt.Printf("  - http://%s:%s/\n", ip, cfg.Port)
	}

	fmt.Println("\n====================================")
	fmt.Printf("  Open http://localhost:%s in browser\n", cfg.Port)
	fmt.Println("====================================")

	if err := r.RunListener(ln); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
		os.Exit(1)
	}
}

func openWhenReady(port string) {
	addr := fmt.Sprintf("http://127.0.0.1:%s/", port)
	client := &http.Client{Timeout: 500 * time.Millisecond}
	for i := 0; i < 30; i++ {
		time.Sleep(100 * time.Millisecond)
		resp, err := client.Get(addr)
		if err == nil {
			resp.Body.Close()
			openBrowserOnce(addr)
			return
		}
	}
}

func indexHandler(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", mustLoadFile("web/control.html"))
}

func inputHandler(c *gin.Context) {
	theme := c.Query("theme")

	if theme != "" && theme != "default" {
		data, err := webFS.ReadFile(fmt.Sprintf("web/%s.html", theme))
		if err == nil {
			c.Data(http.StatusOK, "text/html; charset=utf-8", data)
			return
		}
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", mustLoadFile("web/input.html"))
}

func mustLoadFile(path string) []byte {
	data, err := webFS.ReadFile(path)
	if err != nil {
		return []byte("Failed to load page")
	}
	return data
}

func typeHandler(c *gin.Context) {
	var req struct {
		Text string `json:"text"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false})
		return
	}

	if req.Text == "" {
		if err := keyboard.Enter(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
		return
	}

	// 等待用户在手机端点发送后把焦点切回电脑端输入框
	time.Sleep(100 * time.Millisecond)

	if err := keyboard.TypeText(req.Text); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func statusHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "running",
	})
}

func qrHandler(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		c.String(http.StatusBadRequest, "Missing url parameter")
		return
	}

	png, err := qrcode.Encode(url, qrcode.Medium, 256)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to generate QR code")
		return
	}

	c.Data(http.StatusOK, "image/png", png)
}

func ipsHandler(c *gin.Context) {
	ips := network.GetAllIPs()
	cfg := config.Load()

	type IPInfo struct {
		IP   string `json:"ip"`
		URL  string `json:"url"`
		Main bool   `json:"main"`
	}

	mainIP := network.GetHostIP()
	var result []IPInfo

	for _, ip := range ips {
		result = append(result, IPInfo{
			IP:   ip,
			URL:  fmt.Sprintf("http://%s:%s/", ip, cfg.Port),
			Main: ip == mainIP,
		})
	}

	c.JSON(http.StatusOK, gin.H{"ips": result})
}

func getConfigHandler(c *gin.Context) {
	c.JSON(http.StatusOK, getRuntimeConfig())
}

func configHandler(c *gin.Context) {
	var req struct {
		Port              *string  `json:"port"`
		IP                *string  `json:"ip"`
		Sensitivity       *float64 `json:"sensitivity"`
		ScrollSensitivity *float64 `json:"scrollSensitivity"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config"})
		return
	}

	cfg := config.Load()
	if req.Port != nil {
		cfg.Port = *req.Port
	}
	if req.IP != nil {
		cfg.IP = *req.IP
	}
	if req.Sensitivity != nil {
		cfg.Sensitivity = *req.Sensitivity
	} else if req.ScrollSensitivity != nil {
		cfg.Sensitivity = *req.ScrollSensitivity
	}
	cfg.ScrollSensitivity = 0
	if cfg.Sensitivity <= 0 {
		cfg.Sensitivity = 1.5
	} else if cfg.Sensitivity > 5 {
		cfg.Sensitivity = 5
	}

	if err := config.Save(cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save config"})
		return
	}
	setRuntimeConfig(cfg)

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func scrollHandler(c *gin.Context) {
	var req scrollRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false})
		return
	}

	if err := applyScroll(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func scrollWebSocketHandler(c *gin.Context) {
	conn, err := scrollUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var req scrollRequest
		if err := json.Unmarshal(data, &req); err != nil {
			_ = conn.WriteJSON(gin.H{"success": false, "error": "Invalid scroll message"})
			continue
		}

		if err := applyScroll(req); err != nil {
			_ = conn.WriteJSON(gin.H{"success": false, "error": err.Error()})
			continue
		}
	}
}

func applyScroll(req scrollRequest) error {
	cfg := getRuntimeConfig()
	switch req.Phase {
	case "start":
		return keyboard.StartTouchScroll()
	case "move":
		return keyboard.MoveTouchScroll(req.OffsetY * cfg.Sensitivity)
	case "end", "cancel":
		return keyboard.EndTouchScroll()
	default:
		return fmt.Errorf("unknown scroll phase: %s", req.Phase)
	}
}

func touchpadHandler(c *gin.Context) {
	var req touchpadRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false})
		return
	}

	if err := applyTouchpad(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func touchpadWebSocketHandler(c *gin.Context) {
	conn, err := scrollUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var req touchpadRequest
		if err := json.Unmarshal(data, &req); err != nil {
			_ = conn.WriteJSON(gin.H{"success": false, "error": "Invalid touchpad message"})
			continue
		}

		if err := applyTouchpad(req); err != nil {
			_ = conn.WriteJSON(gin.H{"success": false, "error": err.Error()})
			continue
		}
	}
}

func applyTouchpad(req touchpadRequest) error {
	cfg := getRuntimeConfig()
	switch req.Action {
	case "move":
		return keyboard.MoveMouse(req.DX*cfg.Sensitivity, req.DY*cfg.Sensitivity)
	case "leftClick":
		return keyboard.LeftClick()
	case "leftDown":
		return keyboard.LeftDown()
	case "leftUp":
		return keyboard.LeftUp()
	case "rightClick":
		return keyboard.RightClick()
	default:
		return fmt.Errorf("unknown touchpad action: %s", req.Action)
	}
}

func staticHandler(c *gin.Context) {
	path := c.Request.URL.Path

	data, err := webFS.ReadFile("web" + path)
	if err != nil {
		c.String(http.StatusNotFound, "Not found")
		return
	}

	contentType := "text/html; charset=utf-8"
	if len(path) > 4 && path[len(path)-4:] == ".css" {
		contentType = "text/css; charset=utf-8"
	} else if len(path) > 3 && path[len(path)-3:] == ".js" {
		contentType = "application/javascript; charset=utf-8"
	}

	c.Data(http.StatusOK, contentType, data)
}

func loadTemplate() error {
	return nil
}
