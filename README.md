# QAA AirType Go 版本

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go&logoColor=white)
![Platform](https://img.shields.io/badge/Platform-Windows-blue?logo=windows&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green)

**通过手机端语音输入实现电脑端远程输入的便捷工具 - Go 语言实现**

> 本项目是 [QAA-Tools/qaa-airtype](https://github.com/QAA-Tools/qaa-airtype) 的 Go 语言重构版本。  
> 原项目使用 Python + Tkinter 实现，支持局域网和 Cloudflare Workers 两种模式。  
> Go 版本专注于局域网模式，提供更好的性能和更简洁的 Web UI 体验。

这是 QAA AirType 的 Go 语言重构版本，相比 Python 版本具有更好的性能和更小的依赖。

## ✨ 主要特性

- 📱 **扫码即用**：启动程序后自动打开控制面板，扫描二维码即可连接
- 📝 **多行输入**：支持多行文本输入，自动换行
- ⌨️ **键盘发送**：手机键盘显示"发送"按钮，Enter 键直接发送
- 📜 **历史记录**：保存最近10条输入记录，支持快速重发
- 🎨 **自定义主题**：支持多种主题（默认、简洁、自动发送、智能检测）
- 🌐 **局域网模式**：无需互联网，同一 WiFi 下即可使用
- 🔽 **系统托盘**：双击托盘图标打开控制面板
- 🔒 **单实例保证**：防止重复运行
- 🎯 **Windows API**：使用 Shift+Insert 粘贴，兼容所有应用（包括终端）

## 🚀 快速开始

### 方式一：下载可执行文件

1. 从 [Releases](https://github.com/QAA-Tools/qaa-airtype/releases) 下载 `QAA-AirType-Go.exe`
2. 双击运行，自动启动服务并打开控制面板
3. 手机扫描二维码或点击地址访问输入界面
4. 在手机网页使用语音输入，按"发送"或 Enter 键，文字自动粘贴到电脑

### 方式二：从源码构建

```bash
git clone https://github.com/QAA-Tools/qaa-airtype.git
cd qaa-airtype/go

# 安装依赖
go mod download

# 构建（Windows GUI 模式）
go build -ldflags="-H windowsgui" -o ../QAA-AirType-Go.exe ./cmd/airtype

# 或直接运行（开发模式）
go run ./cmd/airtype
```

## 📋 使用说明

### 页面结构

- **`/`** - 控制面板（电脑端访问）
  - 显示二维码（指向输入界面）
  - 显示所有可用 IP 地址
  - 点击地址切换二维码
  
- **`/input`** - 输入界面（手机端访问）
  - 多行文本输入框
  - Enter 发送，Shift+Enter 换行
  - 历史记录快速重发

### 主题支持

可以通过 URL 参数使用不同主题：

- `/input` - 默认主题（手动发送）
- `/input?theme=light` - 简洁白色主题
- `/input?theme=auto` - 自动发送主题（支持定时发送）
- `/input?theme=detect` - 智能检测主题（编辑框不变化后自动发送）

### 系统托盘

- **双击托盘图标** - 打开控制面板
- **右键菜单** - 打开控制面板 / 退出程序

## 🔧 构建说明

### 生成图标

```bash
# 1. 准备 PNG 图标（推荐 256x256 或更大）
cp /path/to/icon.png go/internal/iconrender/icon.png

# 2. 生成 ICO 文件（多尺寸：16,24,32,48,64,128,256）
cd go
go run gen_icon_png.go

# 3. 生成 exe 图标资源
cd cmd/airtype
rsrc -ico app.ico -o rsrc.syso
```

### 完整构建流程

```bash
cd go

# 1. 生成图标（可选，已有图标可跳过）
go run gen_icon_png.go
cd cmd/airtype && rsrc -ico app.ico -o rsrc.syso && cd ../..

# 2. 构建 exe
go build -ldflags="-H windowsgui" -o ../QAA-AirType-Go.exe ./cmd/airtype
```

或使用 Makefile：

```bash
make build
```

## 📂 项目结构

```
go/
├── cmd/airtype/              # 主程序入口
│   ├── main.go               # Web 服务器和路由
│   ├── tray.go               # 系统托盘
│   ├── browser_windows.go    # Windows 浏览器打开
│   ├── browser_other.go      # 其他平台浏览器
│   ├── icon_windows.go       # Windows 图标生成
│   ├── icon_other.go         # 其他平台图标
│   ├── single_windows.go     # Windows 单实例
│   ├── single_other.go       # 其他平台单实例
│   ├── app.ico               # 应用图标
│   ├── rsrc.syso             # exe 图标资源
│   └── web/                  # Web 界面
│       ├── control.html      # 控制面板
│       ├── input.html        # 输入界面（默认主题）
│       ├── light.html        # 白色主题
│       ├── auto.html         # 自动发送主题
│       └── detect.html       # 智能检测主题
├── internal/
│   ├── clipboard/            # 剪贴板操作
│   ├── keyboard/             # 键盘模拟（Windows API）
│   ├── network/              # IP 地址获取
│   ├── config/               # 配置管理
│   └── iconrender/           # 图标渲染（PNG → ICO）
│       ├── icon.png          # PNG 图标源文件
│       ├── render.go         # PNG 缩放
│       └── ico.go            # ICO 生成
├── gen_icon_png.go           # PNG → ICO 生成脚本
├── go.mod                    # Go 模块定义
├── Makefile                  # 构建脚本
└── README.md                 # 本文档
```

## 🛠️ 技术栈

- **Web 服务器**: [Gin](https://github.com/gin-gonic/gin)
- **系统托盘**: [energye/systray](https://github.com/energye/systray)
- **剪贴板**: [atotto/clipboard](https://github.com/atotto/clipboard)
- **键盘模拟**: Windows API (golang.org/x/sys/windows)
- **二维码**: [skip2/go-qrcode](https://github.com/skip2/go-qrcode)
- **图标渲染**: PNG 缩放（标准库 image/png）
- **EXE 图标**: [akavel/rsrc](https://github.com/akavel/rsrc)

## 🔍 API 接口

### POST /type
发送文字到电脑

**请求**：
```json
{
  "text": "要输入的文字"
}
```

**响应**：
```json
{
  "success": true
}
```

### GET /ips
获取所有可用 IP 地址

**响应**：
```json
{
  "ips": [
    {
      "ip": "192.168.1.100",
      "url": "http://192.168.1.100:5000/",
      "main": true
    }
  ]
}
```

### GET /qr?url=xxx
生成二维码图片

**参数**：
- `url` - 要生成二维码的 URL

**响应**：PNG 图片

### GET /status
服务状态检查

**响应**：
```json
{
  "status": "running"
}
```

## 🆚 与 Python 版本对比

| 特性 | Python 版本 | Go 版本 |
|------|------------|---------|
| 运行方式 | 需要 Python 环境 | 单文件 exe |
| 文件大小 | ~120MB（打包后） | ~14MB |
| 启动速度 | 较慢 | 快速 |
| 内存占用 | ~50MB | ~20MB |
| GUI 界面 | Tkinter 窗口 | Web UI（浏览器） |
| CF 模式 | ✅ 支持 | ❌ 不支持 |
| 多行输入 | ❌ 单行 | ✅ 多行 |
| 键盘发送 | ❌ 需点击 | ✅ Enter 发送 |
| 托盘图标 | ✅ 支持 | ✅ 支持 |

## 🐛 已知问题

- macOS 和 Linux 支持有限（键盘模拟功能可能不完整）
- 不支持 Cloudflare Workers 模式（仅局域网）

## 🙏 致谢

- **Python 版本**：原始实现
- **OpenCode**：Go 版本重构和优化

## 📄 许可证

MIT License

---

<div align="center">

Made with ❤️ by QAA-Tools

</div>