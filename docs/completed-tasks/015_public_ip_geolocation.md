# 公网 IP 及地理位置信息获取

> 完成日期: 2026-08-17

## 功能描述

为远程控制框架新增获取被控端公网 IP 及地理位置信息的功能，支持多 API 并发请求、智能字段映射和原始数据透传。

## 实现内容

### 1. Protocol 扩展
- `shared/protocol.go`: 新增 `CmdPublicIP` 命令类型

### 2. 智能字段映射器
- `shared/ip_geo.go`: 实现智能字段映射逻辑
  - 预定义 API 字段映射规则（ip-api.com, ipinfo.io, httpbin.org）
  - 关键词自动匹配（未知 API 降级方案）
  - 特殊格式处理（如 ipinfo.io 的 "lat,lon" 格式）
  - 标准化输出结构 `IPGeoStandard`

### 3. Client 端处理
- `client/public_ip_handler.go`: 并发请求多个 IP 地理 API
  - 同时请求 3 个 API，提高成功率
  - 返回原始 JSON 数据，不丢失任何信息
  - 错误处理和超时控制

### 4. Handler 注册
- `client/handler.go`: 在 `executeCommand` 中注册 `CmdPublicIP` 分支

### 5. CLI 端支持
- `server_cli/client/public_ip.go`: CLI 查询接口
- `server_cli/helpers.go`: 新增 `publicip` 命令和格式化输出
  - 使用 output 包函数进行样式化输出
  - 每个 API 结果独立展示
  - 标准化信息 + 原始数据同时展示
- `server_cli/lang/en.json`: 英文语言包
- `server_cli/lang/zh.json`: 中文语言包

### 6. Web 端 API
- `server_web/handlers.go`: 新增 `handleGetPublicIP` 处理函数
- `server_web/main.go`: 注册 `/api/public-ip` 路由（含路径密码前缀支持）

### 7. Web 前端
- `server_web/templates/index.html`: 新增公网 IP 按钮和模态对话框
  - 客户端列表添加"公网IP"按钮
  - 模态对话框展示各 API 结果
  - 标准化字段展示 + 原始数据折叠展示
- `server_web/static/js/app.js`: 新增前端逻辑
  - `showPublicIP()`: 打开公网 IP 对话框
  - `fetchPublicIP()`: 调用 API 获取数据
  - `extractAPISource()`: 从 URL 提取 API 来源
  - `normalizeIPData()`: 前端字段标准化映射
- `server_web/static/js/i18n/messages.js`: 新增中英文语言包键值

### 8. 单元测试
- `shared/ip_geo_test.go`: 完整的字段映射测试
  - ip-api.com 字段映射测试
  - ipinfo.io 字段映射测试（含 loc 解析）
  - httpbin.org 字段映射测试
  - 未知 API 关键词匹配测试
  - parseLocation 边界测试
  - toString/toFloat64 工具函数测试

## 使用的 API

| API | 端点 | 特点 |
|-----|------|------|
| ip-api.com | `http://ip-api.com/json/` | 字段最全，45次/分钟 |
| ipinfo.io | `http://ipinfo.io/json` | 简洁可靠，50k/月 |
| httpbin.org | `https://httpbin.org/ip` | 仅返回 IP 地址 |

## 设计亮点

1. **零耦合**: 不依赖特定 API 结构，通过映射规则适配
2. **高可用**: 多 API 并发请求，失败自动降级
3. **易扩展**: 新增 API 只需添加映射配置
4. **数据完整**: 原始数据完整保留，不会丢失任何信息
5. **智能提取**: 预定义映射 + 关键词匹配双保险
6. **向前兼容**: API 字段变化不影响核心逻辑

## CLI 使用示例

```
(dev-01)(console) >> publicip

[*] 公网 IP 信息
[*] API: http://ip-api.com/json/
[*] IP 地址: 118.81.81.81
[*] 大洲: Asia
[*] 国家: China (CN)
[*] 省份: Shanxi
[*] 城市: Taiyuan
[*] ISP: CNC Group CHINA169 Shanxi Province Network
[*] 时区: Asia/Shanghai
[*] 经纬度: 37.8706, 112.5510
[*] 原始数据:
  status: success
  continent: Asia
  country: China
  ...
```

## 测试验证

- `go test ./shared/...` - 所有测试通过
- `golangci-lint run` - 新增代码无 lint 错误
- `go build ./...` - 编译通过