# 项目配置 / Project Configuration

## 中文 / Chinese

本项目支持中文和英文两种语言，请根据需要选择相应的文档版本：
- 中文文档：[准备环境](zh/准备环境.md) | [运行](zh/运行.md)
- 英文文档：[Environment Setup](en/Environment-Setup.md) | [Running](en/Running.md)

## English / 英文

This project supports both Chinese and English languages. Please choose the corresponding documentation version according to your needs:
- Chinese Documentation: [Environment Setup](zh/准备环境.md) | [Running](zh/运行.md)
- English Documentation: [Environment Setup](en/Environment-Setup.md) | [Running](en/Running.md)

---

## server_api.py

### 中文 / Chinese

打开`server_api.py`，找到与它类似的代码：
```python
# 服务器配置
HOST = "0.0.0.0" # 服务器地址
PORT = 8765 # 服务器端口
WEB_HOST = "0.0.0.0" # Web服务器监听地址
WEB_PORT = 5000 # Web服务器监听端口
SSL_CERT = "cert.pem" # SSL证书
SSL_KEY = "key.pem" # SSL密钥
LANGUAGE = "zh" # 语言
SECURITY_PATH = "fuck" # 安全路径
SECURITY_PASSWORD_HASH = b"$2b$04$T8NZ.WUIuO05WyVpLrQYiOdgqc2zbx7E9ysF03696dYvwGohCFzwC" # 安全密码哈希（bcrypt格式，默认密码为fuck）
```
下面是对这些参数的说明：
|参数|说明|
|---|---|
|HOST/PORT|WebSocket 服务器启动时所在的IP地址和端口|
|WEB_HOST/WEB_PORT|Quart API 服务器启动时所在的IP地址和端口（被server.py和server_web.py连接）|
|SSL_CERT/SSL_KEY|WebSocket 服务器启动时所需的SSL证书和密钥的路径|
|LANGUAGE|界面语言设置，'zh'为中文，'en'为英文|
|SECURITY_PATH|API 服务器的安全访问路径|
|SECURITY_PASSWORD_HASH|API 服务器的安全密码bcrypt哈希值（默认密码为fuck）|

**重要提示**：
- `WEB_HOST/WEB_PORT` 是内部API服务器的地址，[server.py](file:///home/zhaobokai/vscode/RATFF/code/server.py) 和 [server_web.py](file:///home/zhaobokai/vscode/RATFF/code/server_web.py) 需要连接到这个地址
- `HOST/PORT` 是客户端连接的WebSocket服务器地址
- SSL证书用于WSS加密通信，详见 [环境准备](准备环境.md)

自行修改配置即可

还有两个文件`server.py`和`server_web.py`

分别用于通过命令行控制和通过网页控制

自行选择即可

### English / 英文

Enter the `RATFF/code` directory.

Open `server_api.py` and find code similar to this:
```python
# Server configuration
HOST = "0.0.0.0" # Server address
PORT = 8765 # Server port
WEB_HOST = "0.0.0.0" # Web server listen address
WEB_PORT = 5000 # Web server listen port
SSL_CERT = "cert.pem" # SSL certificate
SSL_KEY = "key.pem" # SSL key
LANGUAGE = "zh" # Language
SECURITY_PATH = "fuck" # Security path
SECURITY_PASSWORD_HASH = b"$2b$04$T8NZ.WUIuO05WyVpLrQYiOdgqc2zbx7E9ysF03696dYvwGohCFzwC" # Security password hash (bcrypt format, default password is fuck)
```
Below are explanations of these parameters:
|Parameter|Description|
|---|---|
|HOST/PORT|IP address and port where WebSocket server runs|
|WEB_HOST/WEB_PORT|IP address and port where Quart API server runs (connected by server.py and server_web.py)|
|SSL_CERT/SSL_KEY|Path to SSL certificate and key required when WebSocket server starts|
|LANGUAGE|Interface language setting, 'zh' for Chinese, 'en' for English|
|SECURITY_PATH|Security access path for API server|
|SECURITY_PASSWORD_HASH|Security password bcrypt hash value for API server (default password is fuck)|

**Important Notes**:
- `WEB_HOST/WEB_PORT` is the internal API server address that [server.py](file:///home/zhaobokai/vscode/RATFF/code/server.py) and [server_web.py](file:///home/zhaobokai/vscode/RATFF/code/server_web.py) need to connect to
- `HOST/PORT` is the WebSocket server address that clients connect to
- SSL certificates are used for WSS encrypted communication, see [Environment Setup](Environment-Setup.md) for details

Modify the configuration as needed.

There are two other files `server.py` and `server_web.py`

Used for command-line control and web control respectively.

Choose as needed.

---

## server.py

### 中文 / Chinese

打开`server.py`，找到与它类似的代码：
```python
# 服务器配置
LANGUAGE = "zh"
API_SITE = "http://127.0.0.1:5000"  # 服务器地址
API_PATH = "fuck"  # API路径
APT_PASSWORD = "fuck"  # 访问密码
REQUEST_TIMEOUT = 10  # 请求超时时间
```
以下是对这些参数的说明：
|参数|说明|
|---|---|
|LANGUAGE|界面语言设置，'zh'为中文，'en'为英文|
|API_SITE|配置所连接 quart 服务器的地址（格式`http/https://{HOST}:{PORT}`）（要与`server_api.py`中的WEB_HOST/WEB_PORT对应）|
|API_PATH|所连接的服务器所需的安全路径（要与`server_api.py`中的SECURITY_PATH相同）|
|APT_PASSWORD|访问密码（明文，用于生成cookie）|
|REQUEST_TIMEOUT|HTTP请求超时时间（秒）|

所有参数均与`server_api.py`配置所对应的参数相关联

自行修改即可

### English / 英文

Open `server.py` and find code similar to this:
```python
# Server configuration
LANGUAGE = "zh"
API_SITE = "http://127.0.0.1:5000"  # Server address
API_PATH = "fuck"  # API path
APT_PASSWORD = "fuck"  # Access password
REQUEST_TIMEOUT = 10  # Request timeout
```
Below are explanations of these parameters:
|Parameter|Description|
|---|---|
|LANGUAGE|Interface language setting, 'zh' for Chinese, 'en' for English|
|API_SITE|Configure the address of the connected quart server (format `http/https://{HOST}:{PORT}`) (should correspond to WEB_HOST/WEB_PORT in `server_api.py`)|
|API_PATH|Security path required by the connected server (should match SECURITY_PATH in `server_api.py`)|
|APT_PASSWORD|Access password (plain text, used to generate cookie)|
|REQUEST_TIMEOUT|HTTP request timeout (seconds)|

All parameters are related to the corresponding configuration in `server_api.py`

Modify as needed.

---

## server_web.py

### 中文 / Chinese

打开`server_web.py`，找到与它类似的代码：
```python
LANGUAGE = "zh"
WEB_HOST = "0.0.0.0"
WEB_PORT = 8000
API_SITE = 'http://127.0.0.1:5000'
SECURITY_PATH = 'fuck'  # 安全路径
SECURITY_PASSWORD_HASH = b'$2b$04$T8NZ.WUIuO05WyVpLrQYiOdgqc2zbx7E9ysF03696dYvwGohCFzwC'  # 密码哈希值（bcrypt格式）
REQUESTS_TIMEOUT = 10  # 请求超时时间
```
以下是对这些参数的说明：
|参数|说明|
|---|---|
|LANGUAGE|界面语言设置，'zh'为中文，'en'为英文|
|WEB_HOST/WEB_PORT|quart 服务器运行时所在的IP地址和端口|
|API_SITE|配置所连接 quart 服务器的地址（格式`http/https://{HOST}:{PORT}`）（要与`server_api.py`中的WEB_HOST/WEB_PORT对应）|
|SECURITY_PATH|quart 服务器启动时的安全路径（要与`server_api.py`所对应的配置相同）|
|SECURITY_PASSWORD_HASH|quart 服务器启动时的安全密码bcrypt哈希值（默认密码为fuck）（要与`server_api.py`所对应的配置相同）|
|REQUESTS_TIMEOUT|HTTP请求超时时间（秒）|

**注意**：`server_web.py`和`server_api.py`都使用相同的bcrypt密码哈希格式。

自行修改即可

### English / 英文

Open `server_web.py` and find code similar to this:
```python
LANGUAGE = "zh"
WEB_HOST = "0.0.0.0"
WEB_PORT = 8000
API_SITE = 'http://127.0.0.1:5000'
SECURITY_PATH = 'fuck'  # Security path
SECURITY_PASSWORD_HASH = b'$2b$04$T8NZ.WUIuO05WyVpLrQYiOdgqc2zbx7E9ysF03696dYvwGohCFzwC'  # Password hash value (bcrypt format)
REQUESTS_TIMEOUT = 10  # Request timeout
```
Below are explanations of these parameters:
|Parameter|Description|
|---|---|
|LANGUAGE|Interface language setting, 'zh' for Chinese, 'en' for English|
|WEB_HOST/WEB_PORT|IP address and port where quart server runs|
|API_SITE|Configure the address of the connected quart server (format `http/https://{HOST}:{PORT}`) (should correspond to WEB_HOST/WEB_PORT in `server_api.py`)|
|SECURITY_PATH|Security path when quart server starts (should match the corresponding configuration in `server_api.py`)|
|SECURITY_PASSWORD_HASH|Security password bcrypt hash value when quart server starts (default password is fuck) (should match the corresponding configuration in `server_api.py`)|
|REQUESTS_TIMEOUT|HTTP request timeout (seconds)|

**Note**: Both `server_web.py` and `server_api.py` use the same bcrypt password hash format.

Modify as needed.

---

## client.go

### 中文 / Chinese

回到上一级，打开`client.go`
找到与它类似的代码
``go
const (
	HOST string = "127.0.0.1"
	PORT int = 8765
	INSECURESKIPVERIFY bool = true
)
```
以下是对这些参数的说明：
|参数|说明|
|---|---|
|HOST/PORT|要连接的服务器地址和端口（要与`server_api.py`中对应的HOST/PORT配置相同）|
|INSECURESKIPVERIFY|是否跳过证书验证，在开发环境中设为true，生产环境中建议设为false|

修改配置时请注意：
1. 确保客户端的HOST/PORT与服务器端配置匹配
2. 在生产环境中，建议将INSECURESKIPVERIFY设置为false以启用证书验证

### English / 英文

Go back one level and open `client.go`
Find code similar to this:
```go
const (
	HOST string = "127.0.0.1"
	PORT int = 8765
	INSECURESKIPVERIFY bool = true
)
```
Below are explanations of these parameters:
|Parameter|Description|
|---|---|
|HOST/PORT|Address and port of the server to connect to (should match HOST/PORT configuration in `server_api.py`)|
|INSECURESKIPVERIFY|Whether to skip certificate verification, set to true in development environment, recommend false in production environment|

When modifying configuration, please note:
1. Ensure client's HOST/PORT matches server-side configuration
2. In production environment, it's recommended to set INSECURESKIPVERIFY to false to enable certificate verification

---

## client目录 / client Directory

### 中文 / Chinese

从v3.0版本开始，客户端代码已重构为多文件结构，位于`code/client/`目录下：

**目录结构**：
```
code/client/
├── main.go              # 主程序入口和WebSocket通信循环
├── execute_command.go   # 命令执行和系统信息获取
└── other_struct.go      # 辅助结构体（文件复制、压缩等）
```

**主要配置**：打开`main.go`，找到与它类似的代码：
```go
const (
	HOST               string = "127.0.0.1"  // 服务器地址
	PORT               int    = 8765         // 服务器端口
	INSECURESKIPVERIFY bool   = true         // 跳过证书验证
	VERSION            string = "3.0-beta.1" // 版本号，不要手动修改
)
```

以下是对这些参数的说明：
|参数|说明|
|---|---|
|HOST/PORT|要连接的服务器地址和端口（要与`server_api.py`中对应的HOST/PORT配置相同）|
|INSECURESKIPVERIFY|是否跳过证书验证，在开发环境中设为true，生产环境中建议设为false|
|VERSION|客户端版本号，会自动添加到系统信息中显示|

**命令协议变更**（v3.0）：
- `command:` → `cmd:` （执行命令）
- `background:` → `bg:` （后台执行命令）
- [compress](file:///home/zhaobokai/vscode/RATFF/code/client/execute_command.go#L157-L176) → `compress:` （压缩文件）
- [extract](file:///home/zhaobokai/vscode/RATFF/code/client/execute_command.go#L179-L243) → `extract:` （解压文件）

修改配置时请注意：
1. 确保客户端的HOST/PORT与服务器端配置匹配
2. 在生产环境中，建议将INSECURESKIPVERIFY设置为false以启用证书验证
3. VERSION参数不应手动修改，由项目维护者统一管理

### English / 英文

Starting from v3.0, the client code has been refactored into a multi-file structure located in the `code/client/` directory:

**Directory Structure**:
```
code/client/
├── main.go              # Main program entry and WebSocket communication loop
├── execute_command.go   # Command execution and system information retrieval
└── other_struct.go      # Helper structures (file copy, compression, etc.)
```

**Main Configuration**: Open `main.go` and find code similar to this:
```go
const (
	HOST               string = "127.0.0.1"  // Server address
	PORT               int    = 8765         // Server port
	INSECURESKIPVERIFY bool   = true         // Skip certificate verification
	VERSION            string = "3.0-beta.1" // Version number, do not modify manually
)
```

Below are explanations of these parameters:
|Parameter|Description|
|---|---|
|HOST/PORT|Address and port of the server to connect to (should match HOST/PORT configuration in `server_api.py`)|
|INSECURESKIPVERIFY|Whether to skip certificate verification, set to true in development environment, recommend false in production environment|
|VERSION|Client version number, automatically added to system information display|

**Command Protocol Changes** (v3.0):
- `command:` → `cmd:` (Execute command)
- `background:` → `bg:` (Execute command in background)
- [compress](file:///home/zhaobokai/vscode/RATFF/code/client/execute_command.go#L157-L176) → `compress:` (Compress file)
- [extract](file:///home/zhaobokai/vscode/RATFF/code/client/execute_command.go#L179-L243) → `extract:` (Extract file)

When modifying configuration, please note:
1. Ensure client's HOST/PORT matches server-side configuration
2. In production environment, it's recommended to set INSECURESKIPVERIFY to false to enable certificate verification
3. VERSION parameter should not be modified manually, managed uniformly by project maintainers

---

## WSS/WS协议切换说明 / WSS/WS Protocol Switching Instructions

### 中文 / Chinese

**重要安全提醒**：本项目默认使用WSS（WebSocket Secure）协议，提供加密通信。

### 完整的WSS降级方案（同时修改客户端和服务端）

**警告**：降级到WS协议会失去加密保护，仅在受控测试环境中使用！

[完整的WSS降级方案](WSS降级.md)

### English / 英文

**Important Security Reminder**: This project uses WSS (WebSocket Secure) protocol by default, providing encrypted communication.

[WSS Downgrade Scheme](WSS-Downgrade.md)
