# 跨平台编译指南 / Cross-Platform Build Guide

本指南介绍如何在不同操作系统上交叉编译生成各种平台的可执行文件。

This guide explains how to cross-compile executable files for various platforms on different operating systems.

## 🎯 支持的操作系统 / Supported OS

### 目标平台 / Target Platforms

| 操作系统 | GOOS | GOARCH | 文件扩展名 | 说明 |
|---------|------|--------|-----------|------|
| Windows | `windows` | `amd64`, `386`, `arm64` | `.exe` | 64 位/32 位/ARM |
| Linux | `linux` | `amd64`, `386`, `arm`, `arm64` | (无) | 64 位/32 位/ARM |
| macOS | `darwin` | `amd64`, `arm64` | (无) | Intel/Apple Silicon |

---

## 🐧 从 Linux 编译 / Build from Linux

### 1. 编译 Windows 版本 / Build for Windows

```bash
# Windows 64 位 (推荐)
GOOS=windows GOARCH=amd64 go build -ldflags='-H windowsgui' -o client.exe code/client.go

# Windows 32 位
GOOS=windows GOARCH=386 go build -ldflags='-H windowsgui' -o client_x86.exe code/client.go

# Windows ARM64
GOOS=windows GOARCH=arm64 go build -ldflags='-H windowsgui' -o client_arm64.exe code/client.go
```

### 2. 编译 Linux 版本 / Build for Linux

```bash
# Linux 64 位 (当前系统)
GOOS=linux GOARCH=amd64 go build -o client_linux code/client.go

# Linux 32 位
GOOS=linux GOARCH=386 go build -o client_linux_x86 code/client.go

# Linux ARM (树莓派等)
GOOS=linux GOARCH=arm GOARM=7 go build -o client_linux_arm code/client.go

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o client_linux_arm64 code/client.go
```

### 3. 编译 macOS 版本 / Build for macOS

```bash
# macOS Intel (64 位)
GOOS=darwin GOARCH=amd64 go build -o client_macos code/client.go

# macOS Apple Silicon (M1/M2)
GOOS=darwin GOARCH=arm64 go build -o client_macos_m1 code/client.go
```

---

## 🪟 从 Windows 编译 / Build from Windows

### PowerShell 环境

```powershell
# Windows 64 位 (当前系统)
$env:GOOS="windows"; $env:GOARCH="amd64"
go build -ldflags="-H windowsgui" -o client.exe code\client.go

# Linux 64 位
$env:GOOS="linux"; $env:GOARCH="amd64"
go build -o client_linux code\client.go

# macOS Intel
$env:GOOS="darwin"; $env:GOARCH="amd64"
go build -o client_macos code\client.go

# macOS Apple Silicon
$env:GOOS="darwin"; $env:GOARCH="arm64"
go build -o client_macos_m1 code\client.go
```

### CMD 环境

```cmd
REM Windows 64 位
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-H windowsgui" -o client.exe code\client.go

REM Linux 64 位
set GOOS=linux
set GOARCH=amd64
go build -o client_linux code\client.go

REM macOS Intel
set GOOS=darwin
set GOARCH=amd64
go build -o client_macos code\client.go
```

---

## 🍎 从 macOS 编译 / Build from macOS

### 1. 编译 Windows 版本 / Build for Windows

```bash
# Windows 64 位
GOOS=windows GOARCH=amd64 go build -ldflags='-H windowsgui' -o client.exe code/client.go

# Windows 32 位
GOOS=windows GOARCH=386 go build -ldflags='-H windowsgui' -o client_x86.exe code/client.go

# Windows ARM64
GOOS=windows GOARCH=arm64 go build -ldflags='-H windowsgui' -o client_arm64.exe code/client.go
```

### 2. 编译 Linux 版本 / Build for Linux

```bash
# Linux 64 位
GOOS=linux GOARCH=amd64 go build -o client_linux code/client.go

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o client_linux_arm64 code/client.go
```

### 3. 编译 macOS 版本 / Build for macOS

```bash
# macOS Intel (当前系统)
GOOS=darwin GOARCH=amd64 go build -o client_macos code/client.go

# macOS Apple Silicon (M1/M2)
GOOS=darwin GOARCH=arm64 go build -o client_macos_m1 code/client.go

# Universal Binary (同时支持 Intel 和 Apple Silicon)
GOOS=darwin GOARCH=amd64 go build -o client_intel code/client.go
GOOS=darwin GOARCH=arm64 go build -o client_arm code/client.go
lipo -create -output client_macos_universal client_intel client_arm
```

---

## ⚠️ 注意事项 / Important Notes

### 1. Windows GUI 模式 / Windows GUI Mode

- `-ldflags='-H windowsgui'` 仅对 Windows 有效
- `-ldflags='-H windowsgui'` only works for Windows
- Linux/macOS 编译时不需要此参数
- Not needed when compiling for Linux/macOS

### 2. CGO 依赖 / CGO Dependencies

如果代码使用 CGO（如 `github.com/shirou/gopsutil`）:
If your code uses CGO (e.g., `github.com/shirou/gopsutil`):

```bash
# 需要禁用 CGO 以实现纯静态编译
# Need to disable CGO for pure static compilation
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags='-H windowsgui' -o client.exe code/client.go
```
