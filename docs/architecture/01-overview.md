# RATFF 技术架构文档 / Technical Architecture

基于 WebSocket 的远程控制软件框架，由受控端（client）和多个控制端（Web/CLI）组成，支持跨平台远程管理。

A WebSocket-based remote control software framework consisting of a client (controlled endpoint) and multiple controllers (Web/CLI), supporting cross-platform remote management.

## 1. 项目概述 / Project Overview

### 1.1 核心设计目标 / Core Design Goals

| 目标 / Goal | 说明 / Description |
|-------------|-------------------|
| **模块化架构 / Modular Architecture** | server_api 作为核心枢纽，server_web 和 server_cli 作为独立控制端 / server_api as core hub, server_web and server_cli as independent controllers |
| **协议统一 / Unified Protocol** | 所有组件通过 shared 库共享统一的 JSON 消息协议 / All components share a unified JSON message protocol via the shared library |
| **安全优先 / Security First** | 双层密码保护、JWT 认证、TLS 传输加密、速率限制 / Two-layer password protection, JWT auth, TLS encryption, rate limiting |
| **高可用性 / High Availability** | 指数退避重连、心跳保活、优雅关闭、并发安全 / Exponential backoff reconnect, heartbeat keep-alive, graceful shutdown, concurrency safety |

### 1.2 部署架构 / Deployment Architecture

```mermaid
graph TD
    subgraph 控制层 / Controllers
        Web[server_web<br/>Web UI - Port 7993<br/>Vue.js + Tailwind CSS]
        CLI[server_cli<br/>CLI Interactive<br/>Terminal UI]
    end

    subgraph 核心 API 层 / Core API Layer
        API[server_api<br/>WebSocket Hub + HTTP API<br/>Port 6341<br/>JWT Auth + Rate Limiting<br/>Client Registry]
    end

    subgraph 受控端层 / Agent Layer
        Client[client<br/>WebSocket Client + Command Executor<br/>Shell + File Ops + Screen Capture<br/>System Info Collection]
    end

    Web -->|HTTP/WS Proxy| API
    CLI -->|WebSocket| API
    API -->|WebSocket| Client
```

## 2. 通信协议 / Communication Protocol

### 2.1 消息格式 / Message Format

所有组件之间通过 JSON 格式的消息进行通信：

All components communicate via JSON-formatted messages:

```json
{
  "id": "uuid",
  "type": "register|heartbeat|command|response|error",
  "command": "screen_capture|shell_exec|file_list|...",
  "client_id": "client-uuid",
  "payload": {},
  "timestamp": 1234567890
}
```

### 2.2 消息类型 / Message Types

| 类型 / Type | 说明 / Description | 方向 / Direction |
|-------------|-------------------|------------------|
| `register` | 客户端注册，携带系统信息 / Client registration with system info | client → server |
| `heartbeat` | 心跳保活（Ping/Pong）/ Heartbeat keep-alive | 双向 / Bidirectional |
| `command` | 控制命令 / Control command | server → client |
| `response` | 执行结果 / Execution result | client → server |
| `error` | 错误信息 / Error message | 双向 / Bidirectional |

### 2.3 命令类型 / Command Types

| 命令 / Command | 说明 / Description |
|----------------|-------------------|
| `shell_exec` | 执行 Shell 命令（同步）/ Execute shell command (synchronous) |
| `shell_exec_bg` | 后台执行 Shell 命令 / Execute shell command in background |
| `system_info` | 获取系统信息 / Get system information |
| `screen_capture` | 屏幕截图 / Screen capture |
| `file_list` | 列出目录文件 / List directory files |
| `file_upload` | 上传文件（分块）/ Upload file (chunked) |
| `file_download` | 下载文件（分块）/ Download file (chunked) |
| `file_move` | 移动文件 / Move file |
| `file_copy` | 复制文件 / Copy file |
| `file_delete` | 删除文件 / Delete file |
| `cd` | 切换工作目录 / Change working directory |
| `pwd` | 获取当前目录 / Get current directory |
| `public_ip` | 获取公网 IP / Get public IP address |
| `exit` | 退出客户端 / Exit client |

### 2.4 心跳机制 / Heartbeat Mechanism

| 参数 / Parameter | 值 / Value |
|------------------|-----------|
| **Ping 间隔 / Ping Interval** | 30 秒 / 30 seconds |
| **Pong 超时 / Pong Timeout** | 60 秒 / 60 seconds |
| **实现 / Implementation** | `gorilla/websocket` 内置 Ping/Pong / Built-in Ping/Pong |
| **并发安全 / Concurrency Safety** | `sync.Mutex` 保证写入安全 / `sync.Mutex` for write safety |

## 3. 安全架构 / Security Architecture

### 3.1 双层密码保护 / Two-Layer Password Protection

```mermaid
graph LR
    A[第一层 / Layer 1<br/>路径密码<br/>PATH_PASSWORD] --> B[URL 路径前缀保护<br/>如 /mypath/ws<br/>URL path prefix protection]
    C[第二层 / Layer 2<br/>登录密码<br/>LOGIN_PASSWORD] --> D[Bcrypt 哈希存储<br/>签发 JWT Token<br/>Bcrypt hash + JWT Token]
```

| 层级 / Layer | 机制 / Mechanism | 说明 / Description |
|--------------|-----------------|-------------------|
| **第一层 / Layer 1** | 路径密码 / Path Password | URL 路径前缀保护，客户端 WebSocket 连接时必须携带 / URL path prefix, required for client WebSocket connection |
| **第二层 / Layer 2** | 登录密码 / Login Password | Bcrypt 哈希存储，登录成功后签发 JWT Token / Bcrypt hashed, JWT Token issued after login |

### 3.2 认证流程 / Authentication Flow

```mermaid
sequenceDiagram
    participant C as 控制端 / Controller
    participant A as server_api
    participant DB as 密码存储 / Password Store

    C->>A: POST /verify (登录密码 / Login Password)
    A->>DB: 验证 Bcrypt Hash / Verify Bcrypt Hash
    DB-->>A: 验证结果 / Verification Result
    A-->>C: 签发 JWT Token / Issue JWT Token
    C->>A: 后续请求携带 Token / Subsequent requests with Token
    A->>A: authMiddleware 验证 Token / Validate Token
```

### 3.3 速率限制 / Rate Limiting

| 类型 / Type | 速率 / Rate | 适用范围 / Scope |
|-------------|------------|-----------------|
| **全局限流 / Global Rate Limit** | 20 req/s | 非 API 端点（WebSocket 升级、静态资源）/ Non-API endpoints (WS upgrade, static resources) |
| **客户端限流 / Per-Client Rate Limit** | 10 req/s | API 端点（命令执行、文件操作）/ API endpoints (command execution, file operations) |

### 3.4 TLS 支持 / TLS Support

- 通过环境变量 `TLS_CERT` 和 `TLS_KEY` 启用 / Enable via environment variables
- 无证书时自动降级为普通 WebSocket 并记录警告 / Auto-degrade to plain WebSocket with warning if no certificate
- 生产环境下未配置 TLS 将拒绝启动 / Production environment refuses to start without TLS

## 4. 各模块详细设计 / Module Design Details

### 4.1 shared（共享库 / Shared Library）

**职责 / Responsibility**：提供所有组件共享的基础设施 / Provide shared infrastructure for all components

| 文件 / File | 职责 / Responsibility |
|-------------|----------------------|
| `protocol.go` | 消息结构体、类型枚举、消息工厂函数 / Message struct, type enums, factory functions |
| `ws_utils.go` | WebSocket 连接封装、心跳机制、安全读写 / WebSocket wrapper, heartbeat, safe read/write |
| `utils.go` | 日志初始化、UUID 生成、客户端 ID 生成 / Logger init, UUID generation, client ID generation |
| `client_info.go` | 客户端信息收集（OS、主机名、IP）/ Client info collection (OS, hostname, IP) |
| `ip_geo.go` | 公网 IP 地理位置查询 / Public IP geolocation lookup |
| `translations.go` | i18n 多语言支持 / i18n multi-language support |

### 4.2 client（受控端 / Client Agent）

**职责 / Responsibility**：运行在被控设备上，接收并执行命令 / Runs on target device, receives and executes commands

| 文件 / File | 职责 / Responsibility |
|-------------|----------------------|
| `main.go` | 连接管理、注册、消息循环、重连逻辑 / Connection management, registration, message loop, reconnect |
| `config.go` | 环境变量配置加载 / Environment variable configuration loading |
| `handler.go` | 命令路由、Shell 执行、目录切换 / Command routing, shell execution, directory change |
| `file_handler.go` | 文件上传/下载分块处理 / File upload/download chunk handling |
| `file_operations.go` | 文件列表、移动、复制、删除操作 / File list, move, copy, delete operations |
| `screen_capture.go` | 屏幕截图功能 / Screen capture functionality |
| `systeminfo_handler.go` | 系统信息收集 / System information collection |
| `public_ip_handler.go` | 公网 IP 查询 / Public IP lookup |

**重连策略 / Reconnect Strategy**：指数退避算法，初始 1 秒，最大 60 秒 / Exponential backoff, initial 1s, max 60s

### 4.3 server_api（核心 API 服务器 / Core API Server）

**职责 / Responsibility**：WebSocket Hub + HTTP API，管理所有客户端连接 / WebSocket Hub + HTTP API, manages all client connections

| 文件 / File | 职责 / Responsibility |
|-------------|----------------------|
| `main.go` | Gin 路由配置、TLS 启动、优雅关闭 / Gin routing, TLS startup, graceful shutdown |
| `config.go` | 环境变量配置加载、安全警告 / Environment config loading, security warnings |
| `manager.go` | 客户端注册表、命令分发、广播 / Client registry, command dispatch, broadcast |
| `handler.go` | WebSocket 处理、HTTP API 端点 / WebSocket handling, HTTP API endpoints |
| `auth.go` | JWT 认证中间件、密码验证 / JWT auth middleware, password verification |
| `http_handlers.go` | 验证端点、客户端列表端点 / Verify endpoint, client list endpoint |

**端口 / Port**：默认 6341 / Default 6341

**关键中间件 / Key Middleware**：

| 中间件 / Middleware | 功能 / Function |
|---------------------|----------------|
| `authMiddleware` | JWT Token 验证 / JWT Token verification |
| `rateLimitMiddleware` | 全局速率限制 / Global rate limiting |
| `apiRateLimitMiddleware` | 每客户端速率限制 / Per-client rate limiting |

### 4.4 server_web（Web 控制端 / Web Controller）

**职责 / Responsibility**：提供浏览器-based 的图形化控制界面 / Provides browser-based graphical control interface

| 文件 / File | 职责 / Responsibility |
|-------------|----------------------|
| `main.go` | Gin 路由配置、静态资源、模板渲染 / Gin routing, static resources, template rendering |
| `config.go` | 环境变量配置加载 / Environment variable configuration loading |
| `handlers.go` | 登录页面、API 代理、文件操作代理 / Login page, API proxy, file operation proxy |
| `auth.go` | Cookie 认证中间件、登录/登出 / Cookie auth middleware, login/logout |
| `websocket.go` | WebSocket 代理（连接 server_api）/ WebSocket proxy (connects to server_api) |
| `websocket_proxy.go` | WebSocket 双向消息转发 / WebSocket bidirectional message forwarding |
| `file_transfer.go` | 文件上传/下载代理 / File upload/download proxy |
| `task_manager.go` | 异步任务管理（进度跟踪）/ Async task management (progress tracking) |
| `translator.go` | i18n 翻译中间件 / i18n translation middleware |

**端口 / Port**：默认 7993 / Default 7993

**前端技术栈 / Frontend Tech Stack**：

| 技术 / Technology | 用途 / Purpose |
|-------------------|---------------|
| Vue 3 | 响应式 UI 框架 / Reactive UI framework |
| Tailwind CSS | 原子化 CSS 框架 / Utility-first CSS framework |
| 自研 i18n / Custom i18n | 多语言支持（中/英）/ Multi-language support (CN/EN) |

### 4.5 server_cli（CLI 控制端 / CLI Controller）

**职责 / Responsibility**：交互式命令行控制工具 / Interactive command-line control tool

| 文件/目录 / File/Directory | 职责 / Responsibility |
|---------------------------|----------------------|
| `main.go` | 主循环、密码输入、模式切换 / Main loop, password input, mode switching |
| `config.go` | 配置加载 / Configuration loading |
| `translator.go` | i18n 翻译 / i18n translation |
| `types.go` | 类型定义 / Type definitions |
| `wrappers.go` | 打印函数封装 / Print function wrappers |
| `helpers.go` | 提示符构建、模式处理 / Prompt building, mode handling |
| `api/` | API 客户端封装（认证、WebSocket、请求）/ API client wrapper (auth, WebSocket, requests) |
| `client/` | 客户端操作（列表、选择、命令执行）/ Client operations (list, select, command execution) |
| `output/` | 格式化输出（表格、系统信息渲染、样式）/ Formatted output (tables, system info rendering, styles) |
| `lang/` | 语言包（en.json, zh.json）/ Language packs |

**交互模式 / Interaction Modes**：

| 模式 / Mode | 提示符 / Prompt | 可用命令 / Available Commands |
|-------------|----------------|------------------------------|
| **Server 模式** | `(server) >>` | list, select, delete, exit |
| **Console 模式** | `(<id>)(console) >>` | command, cd, bg, exit, back |
| **Command 模式** | `(<id>)(command) >>` | 直接输入命令执行 / Direct command input |

## 5. 文件结构 / File Structure

```
RATFF/
├── client/                    # 受控端（被控设备运行）/ Client agent (runs on target device)
│   ├── main.go               # 入口：连接、注册、消息循环、重连 / Entry: connect, register, message loop, reconnect
│   ├── config.go             # 配置：SERVER_HOST, SERVER_PORT, PATH_PASSWORD / Config
│   ├── handler.go            # 命令路由：shell_exec, cd, pwd, exit / Command routing
│   ├── file_handler.go       # 文件上传/下载分块处理 / File upload/download chunk handling
│   ├── file_operations.go    # 文件操作：list, move, copy, delete / File operations
│   ├── screen_capture.go     # 屏幕截图 / Screen capture
│   ├── systeminfo_handler.go # 系统信息收集 / System info collection
│   ├── public_ip_handler.go  # 公网 IP 查询 / Public IP lookup
│   └── *_test.go             # 单元测试 / Unit tests
│
├── server_api/               # 核心 API 服务器 / Core API server
│   ├── main.go               # 入口：Gin 路由、TLS、优雅关闭 / Entry: Gin routing, TLS, graceful shutdown
│   ├── config.go             # 配置：HOST, PORT, PATH_PASSWORD, JWT_SECRET / Config
│   ├── manager.go            # 客户端管理器：注册、注销、发送命令 / Client manager: register, unregister, send commands
│   ├── handler.go            # WebSocket 处理、HTTP API 端点 / WebSocket handling, HTTP API endpoints
│   ├── auth.go               # JWT 认证、密码验证 / JWT auth, password verification
│   ├── http_handlers.go      # 验证端点、客户端列表 / Verify endpoint, client list
│   └── *_test.go             # 单元测试 / Unit tests
│
├── server_web/               # Web 控制端 / Web controller
│   ├── main.go               # 入口：Gin 路由、静态资源、模板 / Entry: Gin routing, static resources, templates
│   ├── config.go             # 配置：HOST, PORT, API_URL, WS_URL / Config
│   ├── handlers.go           # HTTP 处理器：登录、API 代理 / HTTP handlers: login, API proxy
│   ├── auth.go               # Cookie 认证、登录/登出 / Cookie auth, login/logout
│   ├── websocket.go          # WebSocket 代理 / WebSocket proxy
│   ├── websocket_proxy.go    # WebSocket 双向转发 / WebSocket bidirectional forwarding
│   ├── file_transfer.go      # 文件传输代理 / File transfer proxy
│   ├── task_manager.go       # 异步任务管理 / Async task management
│   ├── translator.go         # i18n 翻译中间件 / i18n translation middleware
│   ├── static/               # 静态资源 / Static resources
│   │   └── js/               # Vue 3, Tailwind, i18n 消息 / i18n messages
│   ├── templates/            # HTML 模板 / HTML templates
│   │   ├── index.html        # 主界面 / Main interface
│   │   └── login.html        # 登录页 / Login page
│   ├── lang/                 # 语言包 / Language packs
│   └── *_test.go             # 单元测试 / Unit tests
│
├── server_cli/               # CLI 控制端 / CLI controller
│   ├── main.go               # 入口：密码输入、交互循环 / Entry: password input, interactive loop
│   ├── config.go             # 配置 / Config
│   ├── translator.go         # i18n 翻译 / i18n translation
│   ├── types.go              # 类型定义 / Type definitions
│   ├── wrappers.go           # 打印函数封装 / Print function wrappers
│   ├── helpers.go            # 提示符、模式处理 / Prompt, mode handling
│   ├── api/                  # API 客户端封装 / API client wrapper
│   │   ├── auth.go           # 登录认证 / Login authentication
│   │   ├── client_api.go     # HTTP API 请求 / HTTP API requests
│   │   └── websocket.go      # WebSocket 连接管理 / WebSocket connection management
│   ├── client/               # 客户端操作 / Client operations
│   │   ├── mgmt.go           # 客户端管理（列表、选择、删除）/ Client management (list, select, delete)
│   │   ├── operations.go     # 命令执行 / Command execution
│   │   ├── shell.go          # Shell 交互 / Shell interaction
│   │   ├── transfer.go       # 文件传输 / File transfer
│   │   ├── screen_capture.go # 屏幕截图 / Screen capture
│   │   ├── systeminfo.go     # 系统信息 / System info
│   │   └── public_ip.go      # 公网 IP / Public IP
│   ├── output/               # 格式化输出 / Formatted output
│   │   ├── table.go          # 表格渲染 / Table rendering
│   │   ├── styles.go         # 样式定义（lipgloss）/ Style definitions (lipgloss)
│   │   └── systeminfo_render.go # 系统信息渲染 / System info rendering
│   └── lang/                 # 语言包（en.json, zh.json）/ Language packs
│
├── shared/                   # 共享库 / Shared library
│   ├── protocol.go           # 消息协议定义 / Message protocol definitions
│   ├── ws_utils.go           # WebSocket 工具 / WebSocket utilities
│   ├── utils.go              # 通用工具 / General utilities
│   ├── client_info.go        # 客户端信息 / Client info
│   ├── ip_geo.go             # IP 地理位置 / IP geolocation
│   ├── translations.go       # i18n 翻译 / i18n translation
│   └── *_test.go             # 单元测试 / Unit tests
│
├── docs/                     # 文档 / Documentation
│   ├── architecture/         # 技术架构文档（本目录）/ Technical architecture docs (this directory)
│   ├── requirements/         # 需求文档 / Requirements docs
│   ├── tasks/                # 任务计划 / Task plans
│   ├── dev-rules/            # 开发规范 / Development rules
│   ├── ai-prompts/           # AI 提示词 / AI prompts
│   └── completed-tasks/      # 已完成任务记录 / Completed task records
│
├── go.mod                    # Go 模块定义 / Go module definition
├── go.sum                    # 依赖校验 / Dependency checksums
├── .gitignore                # Git 忽略规则 / Git ignore rules
└── SECURITY.md               # 安全策略 / Security policy
```

## 6. 技术栈 / Technology Stack

### 6.1 后端（Go）/ Backend (Go)

| 库 / Library | 用途 / Purpose |
|--------------|---------------|
| `github.com/gin-gonic/gin` | HTTP 框架，高并发、成熟稳定 / HTTP framework, high concurrency, mature and stable |
| `github.com/gorilla/websocket` | WebSocket 连接，广泛使用 / WebSocket connection, widely used |
| `github.com/sirupsen/logrus` | 生产级结构化日志 / Production-grade structured logging |
| `github.com/golang-jwt/jwt/v5` | JWT Token 认证 / JWT Token authentication |
| `golang.org/x/crypto/bcrypt` | 密码加密 / Password hashing |
| `golang.org/x/term` | 终端密码输入掩码 / Terminal password input masking |
| `golang.org/x/time/rate` | 请求速率限制 / Request rate limiting |
| `github.com/google/uuid` | UUID 生成 / UUID generation |
| `github.com/shirou/gopsutil/v3` | 系统信息收集 / System information collection |
| `github.com/kbinani/screenshot` | 屏幕截图 / Screen capture |
| `github.com/charmbracelet/lipgloss` | CLI 终端样式 / CLI terminal styling |
| `github.com/google/shlex` | Shell 命令解析 / Shell command parsing |
| `github.com/stretchr/testify` | 测试断言 / Test assertions |

### 6.2 前端（server_web）/ Frontend (server_web)

| 技术 / Technology | 用途 / Purpose |
|-------------------|---------------|
| Vue 3 | 响应式 UI 框架 / Reactive UI framework |
| Tailwind CSS | 原子化 CSS 框架 / Utility-first CSS framework |
| 自研 i18n / Custom i18n | 多语言支持（中/英）/ Multi-language support (CN/EN) |

### 6.3 通信协议 / Communication Protocols

| 协议 / Protocol | 用途 / Purpose |
|-----------------|---------------|
| WebSocket | 长连接双向通信 / Long-lived bidirectional communication |
| HTTP/JSON | RESTful API |
| TLS/WSS | 传输层加密 / Transport layer encryption |

## 7. 关键实现细节 / Key Implementation Details

### 7.1 客户端注册流程 / Client Registration Flow

```mermaid
sequenceDiagram
    participant C as client
    participant A as server_api
    participant M as ClientManager

    C->>A: 1. 建立 WebSocket 连接 / Establish WebSocket connection
    C->>A: 2. 发送 register 消息（client_id + 系统信息）/ Send register message
    A->>M: 3. 注册到内存注册表 / Register to in-memory registry
    M-->>C: 4. 注册完成 / Registration complete
    C->>C: 5. 进入消息循环，等待命令 / Enter message loop, wait for commands
```

1. 客户端启动，通过环境变量获取服务器地址和路径密码 / Client starts, gets server address and path password from environment variables
2. 建立 WebSocket 连接到 `ws://host:port/path/ws` / Establish WebSocket connection
3. 发送 `register` 消息，携带 `client_id` 和系统信息 / Send `register` message with `client_id` and system info
4. server_api 的 `ClientManager` 将客户端注册到内存注册表 / `ClientManager` registers client to in-memory registry
5. 客户端进入消息循环，等待接收命令 / Client enters message loop, waits for commands

### 7.2 命令执行流程 / Command Execution Flow

```mermaid
sequenceDiagram
    participant C as 控制端 / Controller
    participant A as server_api
    participant M as ClientManager
    participant T as client

    C->>A: 1. 发送命令（HTTP API 或 WebSocket）/ Send command
    A->>M: 2. 查找目标客户端 / Find target client
    M->>T: 3. 通过 WebSocket 转发 command 消息 / Forward command via WebSocket
    T->>T: 4. executeCommand 路由到处理器 / Route to handler
    T->>M: 5. 返回 response 消息 / Return response message
    M->>A: 6. 响应返回给控制端 / Response back to controller
    A-->>C: 7. 返回执行结果 / Return execution result
```

1. 控制器（Web/CLI）通过 HTTP API 或 WebSocket 发送命令 / Controller sends command via HTTP API or WebSocket
2. server_api 的 `ClientManager` 查找目标客户端 / `ClientManager` finds target client
3. 通过 WebSocket 将 `command` 消息发送给客户端 / Send `command` message to client via WebSocket
4. 客户端的 `executeCommand` 函数根据命令类型路由到对应处理器 / `executeCommand` routes to handler by command type
5. 处理器执行完成后返回 `response` 消息 / Handler returns `response` message after execution
6. server_api 将响应返回给控制器 / server_api returns response to controller

### 7.3 文件传输（分块机制）/ File Transfer (Chunked Mechanism)

文件上传和下载采用分块传输机制，避免大文件占用过多内存：

File upload and download use chunked transfer mechanism to avoid large files consuming too much memory:

**上传流程 / Upload Flow**：

```mermaid
sequenceDiagram
    participant C as 控制端 / Controller
    participant T as client

    C->>T: file_upload_start - 初始化上传 / Initialize upload
    T-->>C: 创建临时文件 / Create temp file
    loop 分块传输 / Chunked transfer
        C->>T: file_upload_chunk - 写入数据块 / Write data chunk
    end
    C->>T: file_upload_complete - 完成上传 / Complete upload
    T-->>C: 清理临时文件 / Cleanup temp file
```

| 步骤 / Step | 命令 / Command | 说明 / Description |
|-------------|---------------|-------------------|
| 1 | `file_upload_start` | 初始化上传，创建临时文件 / Initialize upload, create temp file |
| 2 | `file_upload_chunk` | 分块写入数据 / Write data in chunks |
| 3 | `file_upload_complete` | 完成上传，清理临时文件 / Complete upload, cleanup temp file |

**下载流程 / Download Flow**：

| 步骤 / Step | 命令 / Command | 说明 / Description |
|-------------|---------------|-------------------|
| 1 | `file_download_start` | 初始化下载，获取文件大小 / Initialize download, get file size |
| 2 | `file_download_chunk` | 分块读取文件数据 / Read file data in chunks |
| 3 | `file_download_complete` | 完成下载，清理资源 / Complete download, cleanup resources |

### 7.4 并发安全 / Concurrency Safety

| 资源 / Resource | 保护机制 / Protection |
|-----------------|----------------------|
| WebSocket 写入 / WebSocket writes | `sync.Mutex` |
| 客户端注册表 / Client registry | `sync.RWMutex` |
| 工作目录 / Working directory | `sync.RWMutex` |
| Channel 发送 / Channel sends | `select { default: }` 防止阻塞 / Prevent blocking |

### 7.5 优雅关闭 / Graceful Shutdown

```mermaid
graph LR
    A[收到 SIGINT/SIGTERM<br/>Receive signal] --> B[context.WithTimeout<br/>5 秒超时 / 5s timeout]
    B --> C[http.Server.Shutdown<br/>优雅关闭 / Graceful shutdown]
    C --> D[asyncWriter.Close<br/>刷新日志 / Flush logs]
```

- 监听 `SIGINT` 和 `SIGTERM` 信号 / Listen for `SIGINT` and `SIGTERM` signals
- 使用 `context.WithTimeout` 设置关闭超时 / Use `context.WithTimeout` for shutdown timeout
- 调用 `http.Server.Shutdown()` 优雅关闭 HTTP 服务 / Call `http.Server.Shutdown()` for graceful shutdown
- 客户端收到 `exit` 命令后调用 `os.Exit(0)` 退出 / Client calls `os.Exit(0)` on `exit` command

### 7.6 日志系统 / Logging System

| 特性 / Feature | 说明 / Description |
|----------------|-------------------|
| **框架 / Framework** | `logrus` 结构化日志 / Structured logging |
| **级别 / Levels** | `info`, `warn`, `error`, `debug` |
| **异步写入 / Async Write** | `AsyncWriter` 实现异步日志 / Async log writing |
| **格式 / Format** | 生产环境 JSON，开发环境文本 / JSON for production, text for development |

## 8. 工程规范 / Engineering Standards

### 8.1 代码规范 / Code Standards

| 规范 / Standard | 要求 / Requirement |
|-----------------|-------------------|
| 单文件行数 / File line limit | 不超过 150 行（测试文件不超过 300 行）/ Max 150 lines (test files max 300 lines) |
| 职责划分 / Responsibility | 单一职责原则，每个文件职责清晰 / Single responsibility, clear file职责 |
| 库选择 / Library choice | 优先使用标准库和成熟第三方库 / Prefer standard library and mature third-party libraries |
| 错误处理 / Error handling | 所有错误必须检查，禁止静默忽略 / All errors must be checked, no silent ignoring |
| 代码检查 / Code linting | 使用 `golangci-lint` 进行代码检查 / Use `golangci-lint` for code checking |

### 8.2 测试规范 / Testing Standards

| 规范 / Standard | 要求 / Requirement |
|-----------------|-------------------|
| 单元测试 / Unit tests | 每个模块必须有单元测试 / Every module must have unit tests |
| 断言库 / Assertion library | 使用 `github.com/stretchr/testify` / Use `github.com/stretchr/testify` |
| 文件命名 / File naming | 测试文件以 `_test.go` 结尾 / Test files end with `_test.go` |

### 8.3 文档规范 / Documentation Standards

| 目录 / Directory | 内容 / Content |
|------------------|---------------|
| `docs/requirements/` | 需求文档 / Requirements docs |
| `docs/tasks/` | 任务计划 / Task plans |
| `docs/dev-rules/` | 开发规范 / Development rules |
| `docs/completed-tasks/` | 已完成任务记录 / Completed task records |
| `docs/ai-prompts/` | AI 提示词 / AI prompts |

## 9. 国际化（i18n）/ Internationalization

项目支持多语言，目前支持中文和英文：

The project supports multiple languages, currently Chinese and English:

| 组件 / Component | 实现方式 / Implementation |
|------------------|--------------------------|
| **server_cli** | 通过 `lang/en.json` 和 `lang/zh.json` 加载翻译 / Load translations via language packs |
| **server_web** | 通过 `lang/en.json` 和 `lang/zh.json` 加载翻译 / Load translations via language packs |
| **shared** | 通过 `translations.go` 提供通用翻译功能 / Provide general translation via `translations.go` |
| **前端 / Frontend** | 自研 i18n 消息系统，通过 `messages.js` 管理翻译 / Custom i18n system via `messages.js` |

## 10. 扩展方向 / Future Extensions

| 方向 / Direction | 说明 / Description |
|------------------|-------------------|
| **更多命令类型 / More Commands** | 进程管理、服务管理、网络诊断等 / Process management, service management, network diagnostics |
| **客户端分组 / Client Grouping** | 客户端分组和批量操作 / Client grouping and batch operations |
| **审计日志 / Audit Log** | 命令执行历史记录和审计 / Command execution history and audit |
| **前端增强 / Frontend Enhancement** | 终端模拟器、文件管理器、任务调度 / Terminal emulator, file manager, task scheduling |
| **数据持久化 / Data Persistence** | 数据库持久化客户端信息 / Database persistence for client information |
| **插件系统 / Plugin System** | 插件系统支持自定义命令 / Plugin system for custom commands |