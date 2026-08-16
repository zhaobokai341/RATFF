# 014 - Folder Transfer Support for Upload/Download

## 概述

为 `server_cli` 的上传和下载功能增加文件夹传输支持。之前仅支持单文件传输，现在支持递归上传/下载整个文件夹。

## 修改内容

### 1. `server_cli/client_upload.go`
- 将 `uploadFile` 重命名为入口函数，检测路径类型后分发
- 提取 `uploadSingleFile` 作为单文件上传核心逻辑
- 新增 `uploadDirectory` 函数：使用 `filepath.Walk` 递归遍历本地文件夹，计算相对路径，对每个文件调用 `uploadSingleFile`
- 移除了 `upload_dir_not_supported` 的硬拒绝

### 2. `server_cli/client_download.go`
- 将 `downloadFile` 重命名为入口函数，先探测远程路径是否为目录
- 提取 `downloadSingleFile` 作为单文件下载核心逻辑
- 新增 `isRemoteDirectory` 函数：使用 `file_list` 命令探测远程路径类型
- 新增 `listRemoteDir` 函数：获取远程目录内容
- 新增 `downloadDirectory` 函数：递归下载远程文件夹
- 新增 `walkRemoteDir` 函数：递归遍历远程目录结构，收集所有文件路径

### 3. `server_cli/client_helpers.go`
- 重构 `waitForCommandResponseWithMsg`，提取公共的 `sendCommandRaw`
- 新增 `waitForCommandResponseRaw`：返回原始消息而不打印错误（用于探测目录类型）

### 4. 语言文件 (`server_cli/lang/zh.json`, `server_cli/lang/en.json`)
- 新增 12 个 i18n 翻译键，覆盖文件夹上传/下载的各个阶段

## 技术方案

无协议变更。文件夹传输在 CLI 层通过递归遍历 + 复用现有单文件 chunk 传输协议实现。客户端 `file_handler.go` 无需修改（`handleFileUploadStart` 已有 `MkdirAll` 创建目录）。

## 测试结果

- `go build ./...` 通过
- `go vet ./server_cli/...` 通过
- `golangci-lint run ./server_cli/...` 通过
- `go test ./...` 全部通过