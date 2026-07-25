# AI 接手指南

> 按顺序阅读以下文件，不要跳过。所有项目信息、规范、进度都在对应文件中。

1. `docs/requirements/` - 需求文档
2. `docs/tasks/` - 任务计划
3. `docs/dev-rules/` - 开发规范
4. `docs/completed-tasks/` - 已完成任务
5. 本文件

## 项目简介

基于 WebSocket 的远程控制框架。

| 模块 | 说明 |
|------|------|
| `client/` | 被控端 |
| `server_api/` | 核心 WebSocket + HTTP API |
| `server_web/` | Web 控制端 |
| `server_cli/` | CLI 交互式控制端 |
| `shared/` | 共享协议和工具 |

## 铁律（必须遵守）

- `main` 函数必须放在文件底部
- 每次编写代码必须写单元测试
- 每次编写完必须用 `golangci-lint run` 检查并修复
- 新增需求必须同步更新 `docs/requirements/` 和 `docs/tasks/`
- 任务完成必须记录到 `docs/completed-tasks/`
- 单文件代码不超过 150 行，测试文件不超过 300 行
- 错误处理必须完整，不能忽略 error（测试中可用 `_ =` 显式忽略）
- 优先使用标准库和成熟第三方库，不要手写
- `server_cli` 所有用户可见输出必须使用 `output.go` 等文件中定义的输出函数，详见 `docs/dev-rules/002_cli_output_styling.md`
- WebSocket 工具函数使用 `shared/` 中的 `SetupHeartbeat`、`SendWSMessage`、`ReadWSMessage`，禁止重复定义
- 密码等敏感配置从环境变量获取，不要硬编码

## 定期刷新

每次接到新任务前，必须重新阅读 `docs/requirements/`、`docs/dev-rules/`、`docs/completed-tasks/`，确保不遗漏最新规范和进度。