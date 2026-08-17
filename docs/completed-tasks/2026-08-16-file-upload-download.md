# 文件上传和下载功能

## 任务概述
在Web文件管理器中实现文件上传和下载功能，支持文件夹传输、断点续传、实时进度显示。

## 完成内容

### 后端新增
1. **server_web/file_transfer.go** - 核心文件传输实现
   - `sendFileCommandRaw` - 通过WebSocket发送文件命令给客户端
   - `uploadSingleFile` - 分块上传单个文件（含MD5校验）
   - `downloadSingleFile` - 分块下载单个文件（含MD5校验）
   - `downloadDirectory` - 递归下载远程目录并打包为ZIP

2. **server_web/task_manager.go** - 异步任务管理和SSE进度推送
   - `TransferTask` - 传输任务数据结构，支持进度追踪、事件广播
   - `TaskManager` - 任务管理器，支持创建/获取/删除任务
   - `handleTaskProgress` - SSE端点，实时推送传输进度
   - `handleDownloadResult` - 下载结果文件端点

3. **server_web/handlers.go** - HTTP处理器
   - `handleFileUpload` - 接收multipart表单数据，创建上传任务
   - `handleFileDownload` - 检测远程路径类型，创建下载任务

4. **server_web/main.go** - 路由注册
   - `POST /api/file/upload` - 上传文件
   - `POST /api/file/download` - 下载文件
   - `GET /api/task/progress` - SSE进度推送
   - `GET /api/file/download_result` - 下载结果文件

5. **server_web/task_manager_test.go** - 单元测试
   - 测试TaskManager创建/获取/删除任务
   - 测试TransferTask的Subscribe/Unsubscribe
   - 测试SetDone/SetError事件广播
   - 测试updateProgress进度计算
   - 测试多监听器广播

### 前端新增
1. **templates/index.html** - UI组件
   - 上传按钮（工具栏）
   - 下载按钮（文件操作列）
   - 隐藏文件选择器（支持多选和文件夹）
   - 进度浮窗（右下角，显示百分比、文件名、文件进度）

2. **static/js/app.js** - 前端逻辑
   - `triggerUpload` - 触发文件选择器
   - `handleFileSelect` - 处理文件选择，构建FormData上传
   - `downloadFile` - 发起下载请求
   - `listenTaskProgress` - 通过EventSource接收SSE进度事件

3. **static/js/i18n/messages.js** - 国际化标签
   - 中英文上传/下载相关UI标签

### 功能特性
- **非阻塞操作**：上传/下载在后台异步执行，用户可继续浏览文件
- **实时进度**：通过SSE推送进度百分比、当前文件、文件计数
- **文件夹支持**：
  - 上传：使用webkitRelativePath保留目录结构
  - 下载：递归下载目录并打包为ZIP
- **分块传输**：与server_cli一致的分块传输协议
- **MD5校验**：传输完成后验证文件完整性

## 修改文件
- `server_web/file_transfer.go` (新增)
- `server_web/task_manager.go` (新增)
- `server_web/task_manager_test.go` (新增)
- `server_web/handlers.go` (修改)
- `server_web/main.go` (修改)
- `server_web/templates/index.html` (修改)
- `server_web/static/js/app.js` (修改)
- `server_web/static/js/i18n/messages.js` (修改)