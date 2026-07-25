# RATFF - Remote Control Framework

基于WebSocket的远程控制软件框架，支持Web和CLI控制端。

## 项目结构

```
├── client/          # 客户端（被控端）
├── server_api/      # 核心API服务（WebSocket + HTTP）
├── server_web/      # Web控制端
├── server_cli/      # CLI控制端
├── shared/          # 共享代码
└── docs/            # 文档
    ├── requirements/     # 需求文档
    ├── tasks/            # 任务计划
    ├── dev-rules/        # 开发规范
    ├── ai-prompts/       # AI提示词
    └── completed-tasks/  # 已完成任务说明
```

## 快速开始

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 启动服务端

```bash
cd server_api && go run .
```

### 3. 启动客户端

```bash
cd client && go run .
```

### 4. 启动Web控制端

```bash
cd server_web && go run .
```

访问 http://localhost:8080

### 5. 使用CLI控制端

```bash
cd server_cli && go run .
```

CLI启动后会自动连接服务端，进入交互模式：
- `list` - 列出在线客户端
- `select <id>` - 选择客户端
- `command` - 进入命令执行模式
- `back` - 返回服务器模式
- `exit` - 退出

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `SERVER_URL` | 服务端地址 | `ws://localhost:9090/ws` |
| `CLIENT_ID` | 客户端ID | 自动生成 |
| `TLS_CERT` | TLS证书路径 | 空（降级WS） |
| `TLS_KEY` | TLS密钥路径 | 空（降级WS） |

## TLS配置

```bash
# 有证书 - 使用WSS
TLS_CERT=server.crt TLS_KEY=server.key go run ./server_api

# 无证书 - 自动降级WS（会有警告）
go run ./server_api
```

## 端口

- server_api: 9090
- server_web: 8080

## 注意事项

- CLI控制端使用 `__cli__` 前缀注册，不会出现在设备列表中
- 所有共享类型定义在 `shared/` 包中，各模块直接引用
- 运行测试: `go test ./...`

## 文档

- 需求文档: `docs/requirements/`
- 任务计划: `docs/tasks/`
- 开发规范: `docs/dev-rules/`
- 已完成任务: `docs/completed-tasks/`
- AI提示词: `docs/ai-prompts/`