# 远程控制软件框架需求文档

## 1. 项目概述

构建一个基于WebSocket的远程控制软件框架，包含客户端（被控端）和服务端（控制端），支持多种控制方式（Web、CLI）。

## 2. 项目结构

```
RATFF/
├── client/              # 客户端（被控端）
├── server_api/          # 服务端API接口（核心业务逻辑）
├── server_web/          # Web控制端（提供Web界面）
├── server_cli/          # CLI控制端（命令行控制端）
├── shared/              # 共享代码库（协议定义、工具函数等）
├── docs/                # 文档
│   ├── 01-requirements/    # 需求文档
│   ├── 02-tasks/           # 任务计划
│   ├── 03-dev-rules/       # 开发规范
│   ├── 04-ai-prompts/      # AI提示词
│   └── 05-completed-tasks/ # 已完成任务说明
└── go.mod
```

## 3. 技术栈

| 功能 | 技术选型 | 说明 |
|------|----------|------|
| HTTP框架 | `github.com/gin-gonic/gin` | 高并发、成熟稳定 |
| WebSocket | `nhooyr.io/websocket` | 现代、支持内置心跳 |
| 日志 | `github.com/sirupsen/logrus` | 生产级、结构化日志 |
| CLI | `github.com/urfave/cli/v2` | 成熟稳定 |
| JSON | `encoding/json` | 标准库 |
| 加密 | `crypto/aes` | 标准库 |
| UUID | `github.com/google/uuid` | 生成唯一ID |
| 限流 | `golang.org/x/time/rate` | 标准库扩展 |

## 4. 功能需求

### 4.1 核心功能

- [x] WebSocket长连接通信
- [x] 客户端注册与发现
- [x] 命令下发与执行
- [x] 结果返回
- [x] 心跳保活
- [x] 断线重连

### 4.2 控制端功能

- [ ] Web界面控制（server_web）
- [x] CLI交互式控制端（server_cli）
- [x] 客户端列表查看（设备ID、IP、主机名、系统信息）
- [x] 命令执行
- [x] 删除指定客户端

### 4.3 客户端功能

- [x] 连接服务端
- [x] 注册自身信息（设备ID、IP、主机名、系统信息）
- [x] 接收并执行命令
- [x] 返回执行结果
- [x] 断线无限重连
- [x] 收到退出命令后正常退出

### 4.4 支持的命令类型

- `shell_exec` - 执行Shell命令
- `system_info` - 获取系统信息
- `exit` - 退出客户端
- `screen_capture` - 屏幕截图（待实现）
- `file_list` - 列出文件（待实现）
- `file_upload` - 上传文件（待实现）
- `file_download` - 下载文件（待实现）

### 4.5 CLI交互端功能（server_cli）

- [x] 交互式命令行界面（非参数式）
- [x] 启动后显示已连接服务器
- [x] 多级命令状态：
  - `(server) >>` - 服务器模式，可执行 list/select/help/clear/delete/exit
  - `(<id>)(console) >>` - 设备控制台，可执行 command/exit/back/help
  - `(<id>)(command) >>` - 命令执行模式，输入具体命令后执行并返回结果
- [x] 命令说明：
  - `help` - 显示帮助列表
  - `list` - 显示已连接设备列表
  - `select <id>` - 选择设备进入控制台
  - `clear` - 清空控制台内容
  - `delete <id>` - 删除指定客户端并发送退出命令
  - `exit` - 退出CLI
  - `back` - 从设备控制台返回服务器模式
  - `command` - 进入命令执行模式
  - 任意其他输入在 command 模式下作为系统命令执行
- [x] 错误提示：设备不存在、无效命令等
- [x] 美化输出：使用 lipgloss 实现彩色文本、表格边框、格式化帮助信息

### 4.6 客户端信息字段

- `device_id` - 设备唯一标识
- `ip` - 设备IP地址
- `hostname` - 设备主机名
- `os_info` - 操作系统信息（OS名称、架构等）

## 5. 非功能需求

### 5.1 代码质量

- [ ] 代码尽可能多复用（shared库、工具函数）
- [ ] 代码简洁优雅
- [ ] 代码解耦，单一职责
- [ ] 单个文件代码行数不超过150行
- [ ] 优先使用标准库和成熟第三方库，避免手写容易出错的代码

### 5.2 安全性

- [x] 支持TLS/WSS加密传输
- [x] 无证书时警告并自动降级为WS
- [x] Token认证机制（JWT临时token）
- [x] 请求限流防DDoS
- [x] 命令执行权限控制（白名单）
- [x] 操作审计日志
- [x] URL路径密码防护（PATH_PASSWORD环境变量）
- [x] 登录密码bcrypt加密存储（LOGIN_PASSWORD_HASH环境变量）
- [x] server_api JWT token验证中间件
- [x] server_web Cookie验证中间件
- [x] server_cli 登录获取token后携带访问
- [x] client WebSocket连接带路径密码
- [x] 两层加密防护：路径密码 + 登录密码

### 5.3 稳定性

- [ ] WebSocket内置心跳机制（PingPongPeriod参数）
- [ ] 客户端断线指数退避重连
- [ ] 服务端优雅关闭（Graceful Shutdown）
- [ ] 连接超时控制
- [ ] 并发安全（sync.RWMutex）
- [ ] Panic恢复（Gin.Recovery）

### 5.4 可维护性

- [ ] 模块独立，易于测试
- [ ] 命名清晰，注释简洁
- [ ] 配置外部化
- [ ] 日志分级（debug/info/warn/error）

### 5.5 工程规范

- [ ] 新增需求时同步更新 `docs/requirements/` 和 `docs/tasks/`
- [ ] 每个模块必须编写单元测试
- [ ] 每次编写完代码后用 `golangci-lint run` 检查并修复问题
- [ ] 任务完成后在 `docs/completed-tasks/` 记录

## 6. 通信协议

### 6.1 消息格式（JSON）

```json
{
  "id": "uuid",
  "type": "register|heartbeat|command|response|error",
  "command": "screen_capture|shell_exec|...",
  "client_id": "client-uuid",
  "payload": {},
  "timestamp": 1234567890
}
```

### 6.2 消息类型

- `register` - 客户端注册
- `heartbeat` - 心跳保活
- `command` - 控制命令
- `response` - 执行结果
- `error` - 错误信息

## 7. 部署架构

```
控制端(Web/CLI) → server_api(WebSocket) → client(被控端)
                      ↑
                 server_web(HTTP代理)
```

- server_api: 端口9090（WebSocket + HTTP API）
- server_web: 端口8080（Web界面 + API代理）
- server_cli: 命令行工具
- client: 被控端程序

## 8. 优先级

P0（必须）: 核心WebSocket通信、客户端管理、命令路由
P1（重要）: 安全性、稳定性、Web/CLI控制端
P2（可选）: 文件传输、屏幕截图等具体命令实现