# 跨平台编译指南 / Cross-Platform Build Guide

本指南介绍如何在不同操作系统上交叉编译生成各种平台的可执行文件。

**重要提示**：从v3.0版本开始，客户端代码已重构为多文件结构，位于`code/client/`目录下。编译时需要进入该目录或使用正确的路径。

This guide explains how to cross-compile executable files for various platforms on different operating systems.

**Important Note**: Starting from v3.0, the client code has been refactored into a multi-file structure located in the `code/client/` directory. You need to enter this directory or use the correct path when compiling.

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
cd code/client

# Windows 64 位 (推荐)
GOOS=windows GOARCH=amd64 go build -ldflags='-H windowsgui' -o client.exe .

# Windows 32 位
GOOS=windows GOARCH=386 go build -ldflags='-H windowsgui' -o client_x86.exe .

# Windows ARM64
GOOS=windows GOARCH=arm64 go build -ldflags='-H windowsgui' -o client_arm64.exe .
```

### 2. 编译 Linux 版本 / Build for Linux

```bash
cd code/client

# Linux 64 位 (当前系统)
GOOS=linux GOARCH=amd64 go build -o client_linux .

# Linux 32 位
GOOS=linux GOARCH=386 go build -o client_linux_x86 .

# Linux ARM (树莓派等)
GOOS=linux GOARCH=arm GOARM=7 go build -o client_linux_arm .

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o client_linux_arm64 .
```

### 3. 编译 macOS 版本 / Build for macOS

```bash
cd code/client

# macOS Intel (64 位)
GOOS=darwin GOARCH=amd64 go build -o client_macos .

# macOS Apple Silicon (M1/M2)
GOOS=darwin GOARCH=arm64 go build -o client_macos_m1 .
```

---

## 🪟 从 Windows 编译 / Build from Windows

### PowerShell 环境

```powershell
cd code\client

# Windows 64 位 (当前系统)
$env:GOOS="windows"; $env:GOARCH="amd64"
go build -ldflags="-H windowsgui" -o client.exe .

# Linux 64 位
$env:GOOS="linux"; $env:GOARCH="amd64"
go build -o client_linux .

# macOS Intel
$env:GOOS="darwin"; $env:GOARCH="amd64"
go build -o client_macos .

# macOS Apple Silicon
$env:GOOS="darwin"; $env:GOARCH="arm64"
go build -o client_macos_m1 .
```

### CMD 环境

```cmd
cd code\client

REM Windows 64 位
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-H windowsgui" -o client.exe .

REM Linux 64 位
set GOOS=linux
set GOARCH=amd64
go build -o client_linux .

REM macOS Intel
set GOOS=darwin
set GOARCH=amd64
go build -o client_macos .
```

---

## 🍎 从 macOS 编译 / Build from macOS

### 1. 编译 Windows 版本 / Build for Windows

```bash
cd code/client

# Windows 64 位
GOOS=windows GOARCH=amd64 go build -ldflags='-H windowsgui' -o client.exe .

# Windows 32 位
GOOS=windows GOARCH=386 go build -ldflags='-H windowsgui' -o client_x86.exe .

# Windows ARM64
GOOS=windows GOARCH=arm64 go build -ldflags='-H windowsgui' -o client_arm64.exe .
```

### 2. 编译 Linux 版本 / Build for Linux

```bash
cd code/client

# Linux 64 位
GOOS=linux GOARCH=amd64 go build -o client_linux .

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o client_linux_arm64 .
```

### 3. 编译 macOS 版本 / Build for macOS

```bash
cd code/client

# macOS Intel (当前系统)
GOOS=darwin GOARCH=amd64 go build -o client_macos .

# macOS Apple Silicon (M1/M2)
GOOS=darwin GOARCH=arm64 go build -o client_macos_m1 .

# Universal Binary (同时支持 Intel 和 Apple Silicon)
GOOS=darwin GOARCH=amd64 go build -o client_intel .
GOOS=darwin GOARCH=arm64 go build -o client_arm .
lipo -create -output client_macos_universal client_intel client_arm
```

---

## ⚠️ 注意事项 / Important Notes

### 1. Windows GUI 模式 / Windows GUI Mode

- `-ldflags='-H windowsgui'` 仅对 Windows 有效
- `-ldflags='-H windowsgui'` only works for Windows
- Linux/macOS 编译时不需要此参数
- Not needed when compiling for Linux/macOS