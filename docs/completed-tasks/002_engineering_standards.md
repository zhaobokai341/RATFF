# 任务完成说明 - 工程规范升级

## 任务信息

| 字段 | 值 |
|------|-----|
| 任务名称 | 工程规范升级 + WebSocket库替换 + 单元测试 + lint |
| 完成时间 | 2026-07-22 |
| 优先级 | P0 |
| 状态 | ✅ 已完成 |

## 完成内容

### 1. 文档规范调整

| 变更 | 说明 |
|------|------|
| 文件夹去数字前缀 | `01-requirements/` → `requirements/` |
| ai-prompts 精简 | 不再堆砌完整需求，改为引导式阅读 |
| 所有文档路径同步更新 | requirements, dev-rules, ai-prompts, README |

### 2. 新增工程规范

| 规范 | 位置 |
|------|------|
| 新增需求同步更新文档 | `docs/requirements/001_remote_control_framework.md` §5.5 |
| 必须写单元测试 | `docs/dev-rules/001_development_guide.md` §6.1 |
| golangci-lint 检查 | `docs/dev-rules/001_development_guide.md` §6.2 |
| 需求变更流程 | `docs/dev-rules/001_development_guide.md` §7 |

### 3. WebSocket 库替换

| 旧库 | 新库 | 原因 |
|------|------|------|
| `nhooyr.io/websocket` | `github.com/gorilla/websocket` | nhooyr 已弃用，gorilla 更成熟 |

**修改文件：**
- `server_api/handler.go` - Upgrader + 心跳逻辑
- `server_api/manager.go` - WriteMessage 替换
- `client/main.go` - DefaultDialer + 心跳逻辑

### 4. 加密算法修复

| 旧 | 新 | 原因 |
|----|----|------|
| CFB 模式 | GCM 模式 | CFB 已弃用，GCM 是 AEAD 推荐方案 |

**修改文件：** `shared/utils.go`

### 5. Lint 问题修复

| 问题 | 文件 | 修复方式 |
|------|------|----------|
| errcheck: SetReadDeadline | client/main.go, server_api/handler.go | 显式忽略 `_ =` |
| errcheck: r.Run | server_web/main.go | 添加错误检查 |
| unused: initTemplates | server_web/main.go | 删除未使用函数 |
| staticcheck: CFB弃用 | shared/utils.go | 改用 GCM |

### 6. 单元测试

| 模块 | 测试文件 | 测试数 | 状态 |
|------|----------|--------|------|
| shared | protocol_test.go | 7 | ✅ 全部通过 |

**测试覆盖：**
- NewMessage 工厂函数
- GenerateID 唯一性
- EncryptAES/DecryptAES 加解密
- 错误密钥解密
- 短密文处理
- InitLogger 初始化

### 7. golangci-lint 检查

```
golangci-lint run ./...
```
**结果：零问题通过** ✅

## 验收标准

- [x] golangci-lint run 零问题
- [x] 所有测试通过
- [x] 所有模块编译通过
- [x] 文档路径全部更新
- [x] ai-prompts 改为引导式
- [x] WebSocket 库替换完成
- [x] 加密算法升级为 GCM