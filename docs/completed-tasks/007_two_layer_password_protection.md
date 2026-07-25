# 任务 007: 两层密码安全防护

## 任务描述

为 server_api 和 server_web 配置两层加密防护：
1. URL路径密码：API根目录从 `/` 改为 `/<path_password>`
2. 登录密码：使用bcrypt加密存储，验证后返回JWT临时token

## 实现细节

### server_api
- 新增 `auth.go`：JWT生成/验证、bcrypt密码验证、authMiddleware
- 新增 `http_handlers.go`：HTTP API handlers（从handler.go拆分）
- 修改 `main.go`：添加路径密码保护的路由组
- 修改 `handler.go`：拆分HTTP handlers到独立文件

### server_web
- 新增 `auth.go`：Cookie验证、登录页面、bcrypt验证、JWT获取
- 新增 `websocket.go`：WebSocket连接和响应监听
- 新增 `handlers.go`：代理handlers和命令执行
- 修改 `main.go`：添加路径密码和登录路由
- 新增 `templates/login.html`：登录页面模板

### server_cli
- 新增 `auth.go`：登录获取JWT token
- 新增 `api.go`：认证HTTP请求封装
- 新增 `websocket.go`：WebSocket连接管理
- 新增 `commands.go`：客户端操作命令
- 新增 `helpers.go`：CLI辅助函数
- 新增 `types.go`：共享类型定义
- 修改 `main.go`：启动时输入密码并登录

### client
- 修改 `main.go`：支持PATH_PASSWORD环境变量构建WebSocket URL

## 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| PATH_PASSWORD | URL路径密码 | 空字符串 |
| LOGIN_PASSWORD_HASH | bcrypt加密的登录密码 | 空字符串 |
| JWT_SECRET | JWT签名密钥 | default-jwt-secret-change-in-production |

## 文件行数检查

所有文件均不超过150行，符合开发规范。

## 依赖

- golang.org/x/crypto/bcrypt
- github.com/golang-jwt/jwt/v5