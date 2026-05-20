# 常见问题 / Frequently Asked Questions

## 1. 在 Windows 电脑上，远程执行命令时会出现黑窗并关掉 / The black window pops up and closes when I run a command remotely on Windows

**问题描述 / Problem:**
Windows 执行命令时需调用 cmd，会出现黑屏闪烁。/ Windows executes commands by calling cmd, causing black screen flashes.

**解决方案 / Solution:**

如果需要在Windows上完全隐藏命令执行时的黑屏，需要修改客户端代码：

#### 步骤 1: 修改 client/main.go / Step 1: Modify client/main.go

在 `client/main.go` 文件中添加 `runtime` 和 `syscall` 导入：/ Add `runtime` and `syscall` imports in `client/main.go`:

```go
import (
    // ... existing imports ...
    "runtime"
    "syscall"
)
```

#### 步骤 2: 更新命令执行函数 / Step 2: Update Command Execution Functions

在 [execute_command.go](file:///home/zhaobokai/vscode/RATFF/code/client/execute_command.go) 的 `execute_command` 和 `execute_bg_command` 函数中添加窗口隐藏代码：

**execute_command 函数:**
```go
cmd := exec.Command(shell, flag, command)

// Hide window for Windows child processes
if runtime.GOOS == "windows" {
    cmd.SysProcAttr = &syscall.SysProcAttr{
        HideWindow:    true,
        CreationFlags: 0x08000000,
    }
}
```

**execute_bg_command 函数:**
```go
cmd := exec.Command(shell, flag, command)

// Hide window for Windows child processes
if runtime.GOOS == "windows" {
    cmd.SysProcAttr = &syscall.SysProcAttr{
        HideWindow:    true,
        CreationFlags: 0x08000000,
    }
}
```

## 2.Linux和Windows进行通讯，除英文外的字符出现乱码问题 / Characters other than English appear garbled when communicating between Linux and Windows

**当前状态 / Current Status:**

从v3.0版本开始，已在Windows平台上实现了UTF-8编码支持：

```go
// 在 execute_command.go 中
if host_info.OS == "windows" {
    shell, flag = "cmd", "/C"
    // 强制设置 UTF-8 输出
    command = fmt.Sprintf("chcp %d >nul & ", 65001) + command
}

// 处理可能存在的无效 UTF-8 序列
output_str := strings.ToValidUTF8(string(output), "")
```

**注意事项 / Notes:**
- 该方案通过在命令前添加 `chcp 65001` 来强制Windows使用UTF-8编码
- 使用 `strings.ToValidUTF8()` 处理可能存在的无效UTF-8序列
- 大多数情况下可以正常显示中文等非英文字符
- 如果仍有乱码问题，可能是由于特定命令或系统配置导致

## 3. v3.0版本的命令协议变更 / v3.0 Command Protocol Changes

**问题 / Question:**
升级至v3.0后，某些命令无法正常工作。/ After upgrading to v3.0, some commands don't work properly.

**解答 / Answer:**

v3.0版本对命令协议进行了优化，以下命令前缀已更改：

| 旧协议 (v2.x) | 新协议 (v3.0+) | 说明 |
|--------------|---------------|------|
| `command:` | `cmd:` | 执行命令 |
| `background:` | `bg:` | 后台执行命令 |
| [compress](file:///home/zhaobokai/vscode/RATFF/code/client/execute_command.go#L157-L176) | `compress:` | 压缩文件 |
| [extract](file:///home/zhaobokai/vscode/RATFF/code/client/execute_command.go#L179-L243) | `extract:` | 解压文件 |

**服务端和客户端会自动同步更新**，无需手动修改。但如果您有自定义的客户端实现，需要注意这些变更。

## 4. 如何编译v3.0版本的客户端？ / How to compile v3.0 client?

**问题 / Question:**
按照旧文档编译客户端时出错。/ Getting errors when compiling client following old documentation.

**解答 / Answer:**

从v3.0版本开始，客户端代码结构已变更：

**旧版本 (v2.x):**
```bash
go build -o client code/client.go
```

**新版本 (v3.0+):**
```bash
cd code/client
go build -o client .
```

详细说明请参考 [跨平台编译指南](跨平台编译.md)。

## 其他问题 / Other Issues

（待补充）/ (To be added)
