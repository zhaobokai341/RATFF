# 环境准备 / Environment Setup

## 安装语言和工具 / Install Languages and Tools

### 中文 / Chinese

安装[Python 3.10+](https://www.python.org)（Linux, macOS 默认自带）

安装[Go 1.11+，推荐Go 1.18+](https://go.dev/)

安装[openssl](https://github.com/openssl/openssl/wiki/Binaries)（Linux, macOS 默认自带）（WSS安全通信必需）

### English

Install [Python 3.10+](https://www.python.org) (Linux, macOS come with it by default)

Install [Go 1.11+, recommended Go 1.18+](https://go.dev/)

Install [openssl](https://github.com/openssl/openssl/wiki/Binaries) (Linux, macOS come with it by default) (Required for WSS secure communication)

## 配置环境 / Configure Environment

### 中文 / Chinese

将*Python*, *Go*, *openssl*配置好Path环境（默认自动配置）

在命令行中输入*python3*, *pip3*, *go*, *openssl*如果没有发生错误则配置成功

### English

Configure *Python*, *Go*, *openssl* in the PATH environment (automatically configured by default)

Type *python3*, *pip3*, *go*, *openssl* in the command line. If no errors occur, the configuration is successful.

## 新建虚拟环境 / Create Virtual Environment

### Python

#### 中文 / Chinese

Python新建虚拟环境：
```bash
python3 -m venv .venv
```
macOS/Linux进入虚拟环境：
```bash
source .venv/bin/activate
```
Windows进入虚拟环境：
```bash
.venv\Scripts\activate.bat
```

#### English

Create Python virtual environment:
```bash
python3 -m venv .venv
```
macOS/Linux activate virtual environment:
```bash
source .venv/bin/activate
```
Windows activate virtual environment:
```bash
.venv\Scripts\activate.bat
```

### Golang

#### 中文 / Chinese

Go新建项目：
```bash
go mod init your_project
```

#### English

Create Go project:
```bash
go mod init your_project
```

## 安装第三方库 / Install Third-party Libraries

### Python

#### 中文 / Chinese

Python需要安装以下第三方库：
- `quart` - 异步Web框架
- `requests` - HTTP请求库
- `rich` - 终端输出美化库
- `websockets` - WebSocket通信库
- `bcrypt` - 密码哈希加密库

在虚拟环境中，使用以下命令：
```bash
pip3 install quart requests rich websockets bcrypt
```
或使用：
```bash
pip3 install -r requirements.txt
```

注意：`bcrypt`库在[server_api.py](file:///home/zhaobokai/vscode/RATFF/code/server_api.py)中用于密码验证，是必需的依赖项。

#### English

Python needs to install the following third-party libraries:
- `quart` - Asynchronous web framework
- `requests` - HTTP request library
- `rich` - Terminal output beautification library
- `websockets` - WebSocket communication library
- `bcrypt` - Password hashing encryption library

In the virtual environment, use the following command:
```bash
pip3 install quart requests rich websockets bcrypt
```
Or use:
```bash
pip3 install -r requirements.txt
```

Note: The `bcrypt` library is used in [server_api.py](file:///home/zhaobokai/vscode/RATFF/code/server_api.py) for password verification and is a required dependency.

### Golang

#### 中文 / Chinese

Go需要以下第三方库：
- `github.com/gorilla/websocket` - WebSocket客户端库
- `github.com/shirou/gopsutil/v3/cpu` - CPU信息获取
- `github.com/shirou/gopsutil/v3/host` - 主机信息获取
- `github.com/shirou/gopsutil/v3/disk` - 磁盘信息获取
- `github.com/shirou/gopsutil/v3/mem` - 内存信息获取
- `github.com/shirou/gopsutil/v3/net` - 网络信息获取
- `github.com/shirou/gopsutil/v3/process` - 进程信息获取

**v3.0+ 版本说明**：客户端代码已重构为多文件结构，位于`code/client/`目录下。

在当前项目中，使用以下命令：
```bash
cd code/client
go mod init ratff-client  # 如果尚未初始化
go get github.com/gorilla/websocket github.com/shirou/gopsutil/v3/cpu github.com/shirou/gopsutil/v3/host github.com/shirou/gopsutil/v3/disk github.com/shirou/gopsutil/v3/mem github.com/shirou/gopsutil/v3/net github.com/shirou/gopsutil/v3/process
```

或者直接使用项目已有的go.mod文件（如果存在）：
```bash
cd code/client
go mod download
```

#### English

Go needs the following third-party libraries:
- `github.com/gorilla/websocket` - WebSocket client library
- `github.com/shirou/gopsutil/v3/cpu` - CPU information retrieval
- `github.com/shirou/gopsutil/v3/host` - Host information retrieval
- `github.com/shirou/gopsutil/v3/disk` - Disk information retrieval
- `github.com/shirou/gopsutil/v3/mem` - Memory information retrieval
- `github.com/shirou/gopsutil/v3/net` - Network information retrieval
- `github.com/shirou/gopsutil/v3/process` - Process information retrieval

**v3.0+ Version Note**: The client code has been refactored into a multi-file structure located in the `code/client/` directory.

In the current project, use the following command:
```bash
cd code/client
go mod init ratff-client  # If not already initialized
go get github.com/gorilla/websocket github.com/shirou/gopsutil/v3/cpu github.com/shirou/gopsutil/v3/host github.com/shirou/gopsutil/v3/disk github.com/shirou/gopsutil/v3/mem github.com/shirou/gopsutil/v3/net github.com/shirou/gopsutil/v3/process
```

Or directly use the existing go.mod file (if present):
```bash
cd code/client
go mod download
```

## 配置证书 / Configure Certificates

### 中文 / Chinese

为实现WSS安全通信，需要配置证书。

**推荐方案**：如果你想要简便且不那么重视安全性，推荐使用自签名证书：
```bash
openssl req -newkey rsa:2048 -nodes -keyout key.pem -x509 -out cert.pem -days 99999 -subj "/CN=localhost"
```

**安全提醒**：
- WSS协议提供加密通信，是项目默认和推荐的配置
- 自签名证书虽然不是权威机构签发，但能提供基本的加密保护
- 仅在特殊测试环境下才考虑其他通信方式
- 生产环境中务必使用有效的SSL/TLS证书

至此，你已经有了运行项目的环境，让我们执行下一步吧！

### English

To implement WSS secure communication, certificates need to be configured.

**Recommended Approach**: If you want simplicity and don't place high importance on security, it's recommended to use self-signed certificates:
```bash
openssl req -newkey rsa:2048 -nodes -keyout key.pem -x509 -out cert.pem -days 99999 -subj "/CN=localhost"
```

**Security Reminder**:
- WSS protocol provides encrypted communication and is the default and recommended configuration
- Self-signed certificates, while not issued by authoritative institutions, provide basic encryption protection
- Only consider alternative communication methods in special testing environments
- Production environments must use valid SSL/TLS certificates

At this point, you have the environment to run the project. Let's proceed to the next step!
