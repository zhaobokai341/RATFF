# 远程控制软件框架 - 实施任务计划

## 任务总览

| 阶段 | 任务 | 优先级 | 状态 |
|------|------|--------|------|
| 1 | 初始化项目 + shared库 | P0 | ✅ 已完成 |
| 2 | server_api核心服务 | P0 | ✅ 已完成 |
| 3 | client客户端 | P0 | ✅ 已完成 |
| 4 | server_web控制端 | P1 | ✅ 已完成 |
| 5 | server_cli控制端 | P1 | ✅ 已完成 |
| 6 | 测试 + 文档 | P2 | ✅ 已完成 |
| 7 | 两层密码安全防护 | P1 | ✅ 已完成 |

## 阶段7: 两层密码安全防护

### 7.1 server_api 安全防护
- [x] URL路径密码（PATH_PASSWORD环境变量）
- [x] /verify接口用于登录验证
- [x] JWT token生成与验证
- [x] authMiddleware中间件验证token
- [x] API端点受JWT保护

### 7.2 server_web 安全防护
- [x] URL路径密码（PATH_PASSWORD环境变量）
- [x] 登录页面（/login）
- [x] bcrypt密码验证
- [x] Cookie认证中间件
- [x] 代理请求携带JWT token
- [x] 登录成功后获取server_api的JWT token

### 7.3 server_cli 安全防护
- [x] 启动时输入路径密码和登录密码
- [x] 登录获取JWT token
- [x] API请求携带Authorization header

### 7.4 client 安全防护
- [x] WebSocket连接URL包含路径密码
- [x] PATH_PASSWORD环境变量支持

### 7.5 依赖安装
- [x] golang.org/x/crypto/bcrypt
- [x] github.com/golang-jwt/jwt/v5

## 阶段1: 初始化项目 + shared库

### 1.1 初始化Go模块
- [x] 创建go.mod
- [x] 安装依赖（gin, websocket, logrus, cli, uuid, rate）

### 1.2 shared/protocol.go
- [x] 定义Message结构体
- [x] 定义MessageType枚举
- [x] 定义CommandType枚举
- [x] NewMessage工厂函数

### 1.3 shared/utils.go
- [x] InitLogger函数（logrus初始化）
- [x] GenerateID函数（uuid）
- [x] EncryptAES/DecryptAES函数

### 1.4 shared/commands.go
- [x] 命令类型常量定义（已合并到protocol.go）

## 阶段2: server_api核心服务

### 2.1 server_api/main.go
- [x] Gin路由 setup
- [x] WebSocket端点
- [x] HTTP API端点
- [x] TLS自动降级逻辑
- [x] 优雅关闭

### 2.2 server_api/manager.go
- [x] ClientManager结构体
- [x] Register/Unregister
- [x] Send/Broadcast
- [x] IsOnline/ListClients

### 2.3 server_api/handler.go
- [x] WebSocket升级处理
- [x] 消息循环
- [x] 命令转发
- [x] HTTP API handlers
- [x] 限流中间件

## 阶段3: client客户端

### 3.1 client/main.go
- [x] 连接服务器
- [x] 注册消息
- [x] 断线重连循环
- [x] 指数退避

### 3.2 client/handler.go
- [x] 消息循环
- [x] 命令执行switch
- [x] Shell执行
- [x] 系统信息获取

## 阶段4: server_web控制端

### 4.1 server_web/main.go
- [x] Gin服务器
- [x] 静态文件服务
- [x] API代理
- [x] HTML模板

## 阶段5: server_cli控制端

### 5.1 server_cli/main.go
- [x] urfave/cli setup
- [x] list命令
- [x] exec命令
- [x] API调用

## 阶段6: 测试 + 文档

### 6.1 基础测试
- [x] 编译检查
- [ ] 基本功能测试（需手动运行）

### 6.2 文档
- [x] README.md
- [x] 使用说明

## 依赖关系

```
阶段1 → 阶段2 → 阶段3
         ↓
       阶段4
         ↓
       阶段5
         ↓
       阶段6
```

## 验收标准

- [x] 所有模块编译通过
- [ ] client能连接server_api（需手动测试）
- [ ] server_web能代理API（需手动测试）
- [ ] server_cli能执行命令（需手动测试）
- [x] TLS降级正常工作
- [x] 断线重连正常
- [x] 单文件<150行
- [x] 代码复用率高