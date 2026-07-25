# 代码规范与开发指南

## 1. 项目结构规范

```
RATFF/
├── client/              # 客户端（被控端）
├── server_api/          # 服务端API（核心）
├── server_web/          # Web控制端
├── server_cli/          # CLI控制端
├── shared/              # 共享代码
├── docs/                # 文档
│   ├── requirements/     # 需求文档
│   ├── tasks/            # 任务计划
│   ├── dev-rules/        # 开发规范
│   ├── ai-prompts/       # AI提示词
│   └── completed-tasks/  # 已完成任务说明
└── go.mod
```

**命名规则：**
- 目录名：小写+下划线（snake_case）
- 文件名：小写+下划线（snake_case）
- 包名：与目录名一致，小写

## 2. 代码规范

### 2.1 文件行数限制
- **源代码文件不超过150行**
- **测试文件（`*_test.go`）不超过300行**
- 超过则拆分为多个文件
- 按职责拆分，而非按行数
- 当函数属于同一职责域且拆分后降低可读性时，可适当放宽限制

### 2.2 函数规范
- 单个函数不超过50行
- 函数名动词开头，如 `NewClientManager`, `handleWebSocket`
- 错误处理：尽早返回，避免嵌套

### 2.3 代码复用原则
- 共享代码放 `shared/` 目录
- 工具函数封装在 `shared/utils.go`
- 协议定义在 `shared/protocol.go`
- WebSocket 工具函数在 `shared/ws_utils.go`（`SetupHeartbeat`、`SendWSMessage`、`ReadWSMessage`）
- 优先使用标准库和成熟第三方库
- 禁止在多个模块中重复定义相同逻辑

### 2.4 日志规范
- 使用 `shared.InitLogger()` 初始化
- 日志级别：debug < info < warn < error
- 生产环境使用JSON格式
- 关键操作必须记录日志

### 2.5 错误处理
```go
// 正确示例
if err != nil {
    log.Error("操作失败: ", err)
    return err
}

// 带上下文的错误
log.WithFields(logrus.Fields{
    "client_id": id,
    "command":   cmd,
}).Error("执行失败")
```

### 2.6 并发安全
- 共享map使用 `sync.RWMutex`
- 读多写少场景用 `RLock/RUnlock`
- 写操作用 `Lock/Unlock`
- 使用 `defer` 确保解锁

## 3. 技术栈

| 功能 | 库 | 版本 |
|------|-----|------|
| HTTP | gin-gonic/gin | latest |
| WebSocket | gorilla/websocket | latest |
| 日志 | sirupsen/logrus | latest |
| CLI | urfave/cli/v2 | latest |
| UUID | google/uuid | latest |
| 限流 | golang.org/x/time | latest |
| JWT | golang-jwt/jwt/v5 | latest |
| 密码加密 | golang.org/x/crypto/bcrypt | latest |
| 终端输入 | golang.org/x/term | latest |
| 测试断言 | stretchr/testify | latest |

### 3.1 前端技术栈

**框架选择：**
- **CSS框架**：Tailwind CSS（通过CDN引入）
- **JS框架**：Vue.js 3（通过CDN引入）

**使用规范：**
- 所有HTML模板使用深色主题设计
- Vue使用 `[[ ]]` 作为分隔符，避免与Go模板的 `{{ }}` 冲突
- Tailwind配置自定义深色主题颜色
- 保持模板文件简洁，复杂逻辑移至JS文件
- 使用Vue的响应式数据管理表单状态
- 加载状态使用旋转图标提示用户

## 4. 如何添加新功能

### 4.1 添加新命令类型

1. 在 `shared/protocol.go` 添加命令常量：
```go
const CmdNewCommand CommandType = "new_command"
```

2. 在 `client/handler.go` 添加处理函数：
```go
func handleNewCommand(msg shared.Message) shared.Message {
    // 实现逻辑
    return shared.NewMessage(shared.MsgResponse, shared.CmdNewCommand, "", payload)
}
```

3. 在 `executeCommand` 的switch中添加case：
```go
case shared.CmdNewCommand:
    resp = handleNewCommand(msg)
```

### 4.2 添加新HTTP API

在 `server_api/main.go` 的 `setupRouter` 中添加路由：
```go
api.GET("/new-endpoint", handleNewEndpoint(manager))
```

在 `server_api/handler.go` 实现handler：
```go
func handleNewEndpoint(manager *ClientManager) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 实现逻辑
        c.JSON(200, result)
    }
}
```

### 4.3 添加新模块

1. 创建目录：`mkdir new_module`
2. 创建main.go
3. 复用shared库的日志和工具函数
4. 保持单文件<150行

## 5. 安全开发指南

### 5.1 认证
- 客户端注册时必须提供Token
- Token在服务端配置中验证
- 失败则拒绝连接

### 5.2 限流
- 使用 `golang.org/x/time/rate`
- 默认每秒50请求
- 可根据需要调整

### 5.3 加密
- 生产环境必须使用TLS
- 敏感数据使用AES加密
- 密钥不要硬编码

## 6. 测试指南

### 6.1 单元测试（必须）
- 每个模块的 `_test.go` 文件与源码同目录
- 测试文件命名：`xxx_test.go`
- 测试函数命名：`TestXxx`
- 使用标准库 `testing` 包
- 覆盖率目标：核心逻辑 > 80%，整体 > 60%

```bash
# 运行所有测试
go test ./...

# 运行指定模块测试
go test ./shared/...

# 带覆盖率
go test -cover ./...

# 详细输出
go test -v ./...
```

### 6.2 测试编写规范

**测试分类：**
- **单元测试**：测试纯函数逻辑，无外部依赖
- **集成测试**：使用 `httptest` 模拟 HTTP/WebSocket 交互
- **边界测试**：测试空值、错误值、异常输入

**测试模板：**
```go
func TestXxxSuccess(t *testing.T) {
    // 1. 准备测试数据
    // 2. 执行被测函数
    // 3. 断言结果
    assert.Equal(t, expected, actual)
}

func TestXxxError(t *testing.T) {
    // 测试错误路径
    assert.Error(t, err)
}

func TestXxxEdgeCase(t *testing.T) {
    // 测试边界条件
    // 使用子测试
    t.Run("case1", func(t *testing.T) { ... })
    t.Run("case2", func(t *testing.T) { ... })
}
```

**HTTP/WebSocket 测试：**
- 使用 `net/http/httptest.NewServer()` 创建测试服务器
- WebSocket 测试使用 `gorilla/websocket` 的 `DefaultDialer`
- 测试完成后务必 `defer server.Close()` 和 `defer conn.Close()`

**断言规范：**
- 使用 `github.com/stretchr/testify/assert` 进行断言
- 每个测试至少包含一个断言
- 错误消息使用 `t.Errorf("Expected %s, got %s", expected, actual)`

### 6.3 Mock 使用规范

**HTTP Mock：**
```go
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // 模拟服务端行为
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}))
defer server.Close()

// 获取 WebSocket URL
wsURL := "ws" + server.URL[4:] + "/ws"
```

**配置 Mock：**
- 测试前设置 `cfg` 全局变量
- 测试后使用 `defer` 恢复原始值
- 使用环境变量时设置后清理：`os.Setenv()` + `defer os.Unsetenv()`

### 6.4 测试覆盖率要求

| 模块类型 | 最低覆盖率 | 说明 |
|---------|---------|------|
| shared（核心库） | > 80% | 协议、工具函数必须高覆盖 |
| server_api（服务端） | > 60% | HTTP 路由、WebSocket 处理 |
| client（客户端） | > 60% | 连接、命令执行逻辑 |
| server_cli（CLI） | > 50% | 输出、交互逻辑 |
| server_web（Web） | > 30% | 代理、页面渲染 |

**提升覆盖率的方法：**
- 为核心业务逻辑编写单元测试
- 使用 `httptest` 测试 HTTP 端点
- 使用 WebSocket 测试服务器测试连接逻辑
- 测试错误路径和边界条件

### 6.5 代码检查（必须）
- 每次编写完代码后执行 `golangci-lint run`
- 修复所有 lint 问题后再提交
- 测试文件中的错误返回值也必须检查（使用 `_ =` 显式忽略）

```bash
# 安装
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 检查全部代码
golangci-lint run ./...

# 检查指定模块
golangci-lint run ./shared/...
```

### 6.6 编译检查
```bash
go build ./...
```

### 6.7 运行测试
```bash
# 启动server_api
cd server_api && go run .

# 启动client
cd client && go run .

# 启动server_web
cd server_web && go run .

# 使用server_cli
cd server_cli && go run . list
```

## 7. 需求变更流程

1. 在 `docs/requirements/` 新建或更新需求文档
2. 在 `docs/tasks/` 更新任务计划
3. 编码并写单元测试
4. `golangci-lint run` 检查
5. 在 `docs/completed-tasks/` 记录完成情况

## 8. 部署指南

### 8.1 编译
```bash
GOOS=linux GOARCH=amd64 go build -o bin/client ./client
GOOS=linux GOARCH=amd64 go build -o bin/server_api ./server_api
```

### 8.2 配置TLS
```bash
# 有证书
./server_api -cert server.crt -key server.key

# 无证书（自动降级WS，会有警告）
./server_api
```

## 9. AI辅助开发指南

当使用AI继续开发时，提供以下信息：
1. 项目根目录下的 `docs/` 文件夹
2. AI 应先阅读 `docs/ai-prompts/` 中的引导文件
3. 然后按引导顺序读取其他文档

AI可以快速理解项目结构并生成符合规范的代码。

## 10. 重要注意事项

### 10.1 CLI 连接机制
- `server_cli` 需要建立 WebSocket 连接到服务端以接收命令响应
- CLI 使用 `__cli__` 前缀的 ID 注册（如 `__cli__a1b2c3d4`）
- `ClientManager.ListClients()` 会自动过滤 `__cli__` 前缀的客户端
- 设备列表中不应出现 `__cli__` 开头的条目

### 10.2 文档注释规范
- 所有导出的类型、常量、变量、函数、方法必须添加 `//` 文档注释
- 注释以类型名或函数名开头，如 `// ClientInfo holds information about a connected client.`
- 未导出的内部函数使用小写注释，如 `// buildOSInfo constructs an OS info string`

### 10.3 代码复用原则
- 共享类型（如 `ClientInfo`、`Message`）定义在 `shared/` 包中
- 各模块直接引用 `shared.XXX`，禁止重复定义
- 发现重复定义时立即删除并替换为 `shared` 引用

### 10.4 响应匹配机制
- `server_cli` 使用 `pendingCmd` map 存储待响应的命令
- key 为 `clientID`（被控端 ID），value 为带 channel 的 `pendingCommand`
- `listenResponses` 根据响应消息的 `ClientID` 匹配并投递到对应 channel

### 10.5 CLI 输出规范
- `server_cli` 所有用户可见输出必须使用 `output.go`、`output_table.go`、`output_help.go` 中定义的函数
- 禁止直接使用 `fmt.Println` 或 `fmt.Printf` 输出用户可见信息
- 详见 `docs/dev-rules/002_cli_output_styling.md`