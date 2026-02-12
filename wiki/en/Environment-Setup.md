# Environment Setup

## Install Languages and Tools
Install [Python 3.10+](https://www.python.org) (Linux, macOS come with it by default)

Install [Go 1.11+, recommended Go 1.18+](https://go.dev/)

Install [openssl](https://github.com/openssl/openssl/wiki/Binaries) (Linux, macOS come with it by default) (Required for WSS secure communication)

## Configure Environment
Configure *Python*, *Go*, *openssl* in the PATH environment (automatically configured by default)

Type *python3*, *pip3*, *go*, *openssl* in the command line. If no errors occur, the configuration is successful.

## Create Virtual Environment
### Python
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
Create Go project:
```bash
go mod init your_project
```

## Install Third-party Libraries
### Python
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
Go needs the following third-party libraries:
- `github.com/gorilla/websocket` - WebSocket client library
- `github.com/shirou/gopsutil/v3/cpu` - CPU information retrieval
- `github.com/shirou/gopsutil/v3/host` - Host information retrieval
- `github.com/shirou/gopsutil/v3/disk` - Disk information retrieval
- `github.com/shirou/gopsutil/v3/mem` - Memory information retrieval
- `github.com/shirou/gopsutil/v3/net` - Network information retrieval
- `github.com/shirou/gopsutil/v3/process` - Process information retrieval

In the current project, use the following command:
```bash
go get github.com/gorilla/websocket github.com/shirou/gopsutil/v3/cpu github.com/shirou/gopsutil/v3/host github.com/shirou/gopsutil/v3/disk github.com/shirou/gopsutil/v3/mem github.com/shirou/gopsutil/v3/net github.com/shirou/gopsutil/v3/process
```

## Configure Certificates
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