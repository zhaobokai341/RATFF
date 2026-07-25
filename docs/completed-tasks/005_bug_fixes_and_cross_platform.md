# 任务完成情况：Bug 修复与跨平台支持

## 修复内容

### 1. Web UI 列表显示 `[object Object]`
- **问题**: JS 遍历 `data.clients` 时直接渲染对象而非访问属性
- **修复**: `forEach(c => ...)` 访问 `c.id`, `c.ip`, `c.hostname`, `c.os_info`

### 2. IP 显示为 `unknown`
- **问题**: `BuildClientInfo` 无法获取连接的真实 IP
- **修复**: 服务端在 `handleWebSocket` 中从 `c.Request.RemoteAddr` 提取 IP 并注入 `ClientInfo`

### 3. 执行命令后 WebSocket 超时断开
- **问题**: `Broadcast` 将响应广播给所有客户端（包括发送者），导致客户端收到自己的响应后误解析，连接断开
- **修复**: `Broadcast(msg, excludeID)` 排除发送者，避免回环

### 4. 跨平台 Shell 执行
- **问题**: 使用 `exec.Command(parts[0], parts[1:]...)` 无法正确处理管道符、重定向等 Shell 语法
- **修复**: 
  - Linux/macOS: `exec.Command("sh", "-c", cmdStr)`
  - Windows: `exec.Command("cmd", "/C", cmdStr)`

## 修改文件

| 文件 | 改动 |
|------|------|
| `server_web/templates/index.html` | 修复 JS 渲染对象问题 |
| `server_api/handler.go` | 提取 RemoteAddr 作为 IP；Broadcast 排除发送者 |
| `server_api/manager.go` | Broadcast 增加 excludeID 参数 |
| `client/handler.go` | 跨平台 Shell 执行（sh/cmd） |

## 验证结果

- ✅ `go build ./...` 编译通过
- ✅ `golangci-lint run` 零问题
- ✅ `go test ./...` 全部通过