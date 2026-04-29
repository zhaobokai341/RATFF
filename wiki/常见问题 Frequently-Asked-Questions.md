# 常见问题 / Frequently Asked Questions

## 1. 在 Windows 电脑上，远程执行命令时会出现黑窗并关掉 / The black window pops up and closes when I run a command remotely on Windows

**问题描述 / Problem:**
Windows 执行命令时需调用 cmd，会出现黑屏闪烁。/ Windows executes commands by calling cmd, causing black screen flashes.

**解决方案 / Solution:**

### 快速修复 / Quick Fix (单文件修改 / Single File Modification)

只需 **2 个步骤** 即可完全隐藏黑屏：/ Just **2 steps** to completely hide the black screen:

#### 步骤 1: 修改 client.go / Step 1: Modify client.go

在 `client.go` 文件中添加 `runtime` 和 `syscall` 导入：/ Add `runtime` and `syscall` imports in `client.go`:

```go
import (
    // ... existing imports ...
    "runtime"
    "syscall"
)
```

#### 步骤 2: 更新命令执行函数 / Step 2: Update Command Execution Functions

在 `execute_command` 和 `execute_bg_command` 函数中添加窗口隐藏代码：/ Add window hiding code in `execute_command` and `execute_bg_command` functions:

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

我目前无法解决这个问题 / I cannot solve this problem at present

## 其他问题 / Other Issues

（待补充）/ (To be added)
