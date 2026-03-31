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
HOST = '0.0.0.0'
PORT = 8765
WEB_HOST = '0.0.0.0'
WEB_PORT = 5000
SSL_CERT = 'cert.pem'
SSL_KEY = 'key.pem'
LANGUAGE = 'zh'
SECURITY_PATH = 'fuck'
SECURITY_PASSWORD_HASH = b'$2b$04$T8NZ.WUIuO05WyVpLrQYiOdgqc2zbx7E9ysF03696dYvwGohCFzwC'
```
下面是对这些参数的说明：
|参数|说明|
|---|---|
|HOST/PORT|websocket 服务器启动时所在的IP地址和端口|
|WEB_HOST/WEB_PORT|quart 服务器启动时所在的IP地址和端口|
|SSL_CERT/SSL_KEY|websocket 服务器启动时所需的证书密钥的路径|
|LANGUAGE|界面语言设置，'zh'为中文，'en'为英文|
|SECURITY_PATH|quart 服务器启动时的安全路径|
|SECURITY_PASSWORD_HASH|quart 服务器启动时的安全密码bcrypt哈希值（默认密码为fuck）|

自行修改配置即可

还有两个文件`server.py`和`server_web.py`

分别用于通过命令行控制和通过网页控制

自行选择即可

### English / 英文

Enter the `RATFF/code` directory.

Open `server_api.py` and find code similar to this:
```python
# Server configuration
HOST = '0.0.0.0'
PORT = 8765
WEB_HOST = '0.0.0.0'
WEB_PORT = 5000
SSL_CERT = 'cert.pem'
SSL_KEY = 'key.pem'
LANGUAGE = 'zh'
SECURITY_PATH = 'fuck'
SECURITY_PASSWORD_HASH = b'$2b$04$T8NZ.WUIuO05WyVpLrQYiOdgqc2zbx7E9ysF03696dYvwGohCFzwC'
```
Below are explanations of these parameters:
|Parameter|Description|
|---|---|
|HOST/PORT|IP address and port where websocket server runs|
|WEB_HOST/WEB_PORT|IP address and port where quart server runs|
|SSL_CERT/SSL_KEY|Path to certificate key required when websocket server starts|
|LANGUAGE|Interface language setting, 'zh' for Chinese, 'en' for English|
|SECURITY_PATH|Security path when quart server starts|
|SECURITY_PASSWORD_HASH|Security password bcrypt hash value when quart server starts (default password is fuck)|

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
API_SITE = "http://localhost:5000"  # 服务器地址
API_PATH = "fuck"  # API路径
APT_PASSWORD = "fuck"  # 访问密码
```
以下是对这些参数的说明：
|参数|说明|
|---|---|
|LANGUAGE|界面语言设置，'zh'为中文，'en'为英文|
|API_SITE|配置所连接 quart 服务器的地址（格式`http/https://{HOST}:{PORT}`）|
|API_PATH/APT_PASSWORD|所连接的服务器所需的安全路径和密码|

所有参数均与`server_api.py`配置所对应的参数相同

自行修改即可

### English / 英文

Open `server.py` and find code similar to this:
```python
# Server configuration
LANGUAGE = "zh"
API_SITE = "http://localhost:5000"  # Server address
API_PATH = "fuck"  # API path
APT_PASSWORD = "fuck"  # Access password
```
Below are explanations of these parameters:
|Parameter|Description|
|---|---|
|LANGUAGE|Interface language setting, 'zh' for Chinese, 'en' for English|
|API_SITE|Configure the address of the connected quart server (format `http/https://{HOST}:{PORT}`)|
|API_PATH/APT_PASSWORD|Security path and password required by the connected server|

All parameters correspond to the configuration in `server_api.py`

Modify as needed.

---

## server_web.py

### 中文 / Chinese

打开`server_web.py`，找到与它类似的代码：
```python
LANGUAGE = "zh"
WEB_HOST = "0.0.0.0"
WEB_PORT = 8000
API_SITE = 'http://localhost:5000'
SECURITY_PATH = 'fuck'  # 安全路径
SECURITY_PASSWORD_HASH = '6ac3c336e4094835293a3fed8a4b5fedde1b5e2626d9838fed50693bba00af0e'  # 密码哈希值
```
以下是对这些参数的说明
|参数|说明|
|---|---|
|LANGUAGE|界面语言设置，'zh'为中文，'en'为英文|
|WEB_HOST/WEB_PORT|quart 服务器运行时所在的IP地址和端口|
|API_SITE|配置所连接 quart 服务器的地址（格式`http/https://{HOST}:{PORT}`）（要与`server_api.py`所对应的配置相同）|
|SECURITY_PATH|quart 服务器启动时的安全路径（要与`server_api.py`所对应的配置相同）|
|SECURITY_PASSWORD_HASH|quart 服务器启动时的安全密码SHA256哈希值（默认密码为fuck）（要与`server_api.py`所对应的配置相同）|

注意：`server_web.py`中的密码哈希格式与`server_api.py`不同，前者使用SHA256字符串格式，后者使用bcrypt字节格式。

自行修改即可

### English / 英文

Open `server_web.py` and find code similar to this:
```python
LANGUAGE = "zh"
WEB_HOST = "0.0.0.0"
WEB_PORT = 8000
API_SITE = 'http://localhost:5000'
SECURITY_PATH = 'fuck'  # Security path
SECURITY_PASSWORD_HASH = '6ac3c336e4094835293a3fed8a4b5fedde1b5e2626d9838fed50693bba00af0e'  # Password hash value
```
Below are explanations of these parameters:
|Parameter|Description|
|---|---|
|LANGUAGE|Interface language setting, 'zh' for Chinese, 'en' for English|
|WEB_HOST/WEB_PORT|IP address and port where quart server runs|
|API_SITE|Configure the address of the connected quart server (format `http/https://{HOST}:{PORT}`) (should match the corresponding configuration in `server_api.py`)|
|SECURITY_PATH|Security path when quart server starts (should match the corresponding configuration in `server_api.py`)|
|SECURITY_PASSWORD_HASH|Security password SHA256 hash value when quart server starts (default password is fuck) (should match the corresponding configuration in `server_api.py`)|

Note: The password hash format in `server_web.py` is different from `server_api.py`. The former uses SHA256 string format, while the latter uses bcrypt byte format.

Modify as needed.

---

## client.go

### 中文 / Chinese

回到上一级，打开`client.go`
找到与它类似的代码
```go
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

## WSS/WS协议切换说明 / WSS/WS Protocol Switching Instructions

### 中文 / Chinese

**重要安全提醒**：本项目默认使用WSS（WebSocket Secure）协议，提供加密通信。

### 完整的WSS降级方案（同时修改客户端和服务端）

**警告**：降级到WS协议会失去加密保护，仅在受控测试环境中使用！

[完整的WSS降级方案](WSS降级.md)

### English / 英文

**Important Security Reminder**: This project uses WSS (WebSocket Secure) protocol by default, providing encrypted communication.

[WSS Downgrade Scheme](WSS-Downgrade.md)
