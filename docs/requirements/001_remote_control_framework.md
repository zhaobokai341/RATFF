# Remote Control Software Framework Requirements

## 1. Project Overview

Build a WebSocket-based remote control software framework consisting of a client (controlled endpoint) and a server (controller), supporting multiple control methods (Web, CLI).

## 2. Project Structure

```
RATFF/
├── client/              # Client (controlled endpoint)
├── server_api/          # Server API (core business logic)
├── server_web/          # Web controller (provides Web interface)
├── server_cli/          # CLI controller (command-line controller)
├── shared/              # Shared codebase (protocol definitions, utility functions, etc.)
├── docs/                # Documentation
│   ├── requirements/    # Requirements documents
│   ├── tasks/           # Task plans
│   ├── dev-rules/       # Development standards
│   ├── ai-prompts/      # AI prompts
│   └── completed-tasks/ # Completed task descriptions
└── go.mod
```

## 3. Technology Stack

| Feature | Technology | Description |
|---------|------------|-------------|
| HTTP Framework | `github.com/gin-gonic/gin` | High concurrency, mature and stable |
| WebSocket | `github.com/gorilla/websocket` | Modern, widely used |
| Logging | `github.com/sirupsen/logrus` | Production-grade, structured logging |
| CLI | `github.com/urfave/cli/v2` | Mature and stable |
| JSON | `encoding/json` | Standard library |
| Encryption | `crypto/aes` | Standard library |
| UUID | `github.com/google/uuid` | Generate unique IDs |
| Rate Limiting | `golang.org/x/time/rate` | Standard library extension |
| JWT | `github.com/golang-jwt/jwt/v5` | Token authentication |
| Password Hashing | `golang.org/x/crypto/bcrypt` | Password encryption |
| Terminal Input | `golang.org/x/term` | Password input masking |
| Test Assertions | `github.com/stretchr/testify` | Test assertion library |
| Terminal Styling | `github.com/charmbracelet/lipgloss` | CLI output styling |
| Shell Parsing | `github.com/google/shlex` | Shell command argument parsing |

## 4. Functional Requirements

### 4.1 Core Features

- [x] WebSocket long connection communication
- [x] Client registration and discovery
- [x] Command dispatch and execution
- [x] Result return
- [x] Heartbeat keep-alive
- [x] Disconnection reconnection

### 4.2 Controller Features

- [x] Web interface control (server_web)
- [x] CLI interactive controller (server_cli)
- [x] Client list view (device ID, IP, hostname, system info)
- [x] Command execution
- [x] Delete specified client

### 4.3 Client Features

- [x] Connect to server
- [x] Register own information (device ID, IP, hostname, system info)
- [x] Receive and execute commands
- [x] Return execution results
- [x] Infinite reconnection on disconnect
- [x] Graceful exit on exit command

### 4.4 Supported Command Types

- `shell_exec` - Execute shell commands
- `shell_exec_bg` - Execute shell commands in background
- `system_info` - Get system information
- `cd` - Change working directory
- `exit` - Exit client
- `screen_capture` - Screen capture (to be implemented)
- `file_list` - List files (to be implemented)
- `file_upload` - Upload file (to be implemented)
- `file_download` - Download file (to be implemented)

### 4.5 CLI Interactive Controller Features (server_cli)

- [x] Interactive command-line interface (non-parameter-based)
- [x] Display connected server on startup
- [x] Multi-level command states:
  - `(server) >>` - Server mode, can execute list/select/help/clear/delete/cd/exit
  - `(<id>)(console) >>` - Device console, can execute command/cd/bg/exit/back/help
  - `(<id>)(command) >>` - Command execution mode, input specific command to execute and return result
- [x] Command descriptions:
  - `help` - Display help list
  - `list` - Display connected device list
  - `select <id>` - Select device to enter console
  - `cd <id> <dir>` - Change working directory of remote client
  - `clear` - Clear console content
  - `delete <id>` - Delete specified client and send exit command
  - `exit` - Exit CLI
  - `back` - Return from device console to server mode
  - `command` - Enter command execution mode
  - `bg <cmd> [file]` - Execute command in background on remote client
  - Any other input in command mode executes as a system command
- [x] Error prompts: device not found, invalid command, etc.
- [x] Styled output: Use lipgloss for colored text, table borders, formatted help messages

### 4.6 Client Information Fields

- `device_id` - Unique device identifier
- `ip` - Device IP address
- `hostname` - Device hostname
- `os_info` - Operating system information (OS name, architecture, etc.)

## 5. Non-Functional Requirements

### 5.1 Code Quality

- [x] Maximize code reuse (shared library, utility functions)
- [x] Code is clean and elegant
- [x] Code is decoupled, single responsibility
- [x] Source code files must not exceed 150 lines
- [x] Test files (`*_test.go`) must not exceed 300 lines
- [x] Prefer standard libraries and mature third-party libraries, avoid error-prone handwritten code

### 5.2 Security

- [x] Support TLS/WSS encrypted transport
- [x] Warn and auto-degrade to WS without certificates
- [x] Token authentication mechanism (JWT temporary token)
- [x] Rate limiting to prevent DDoS
- [x] Command execution permission control (whitelist)
- [x] Operation audit logging
- [x] URL path password protection (PATH_PASSWORD environment variable)
- [x] Login password bcrypt encrypted storage (LOGIN_PASSWORD_HASH environment variable)
- [x] server_api JWT token verification middleware
- [x] server_web Cookie verification middleware
- [x] server_cli obtains token on login and carries it for access
- [x] client WebSocket connection with path password
- [x] Two-layer encryption: path password + login password

### 5.3 Stability

- [x] WebSocket built-in heartbeat mechanism (manual ping/pong with ticker)
- [x] Client exponential backoff reconnection
- [x] Server graceful shutdown
- [x] Connection timeout control
- [x] Concurrency safety (sync.RWMutex)
- [x] Panic recovery (Gin.Recovery)

### 5.4 Maintainability

- [x] Modular, easy to test
- [x] Clear naming, concise comments
- [x] Externalized configuration
- [x] Log level classification (debug/info/warn/error)

### 5.5 Engineering Standards

- [x] Update `docs/requirements/` and `docs/tasks/` when adding new requirements
- [x] Every module must have unit tests
- [x] After writing code, use `golangci-lint run` to check and fix issues
- [x] Record completed tasks in `docs/completed-tasks/`

### 5.6 Error Handling Standards

- [x] All error return values must be checked, no silent ignoring
- [x] Resource cleanup (`Close()`, `Shutdown()`) must verify return values
- [x] Every feature must handle all failure scenarios (file not found, permission denied, timeout, connection lost, disk full, etc.)
- [x] Error messages must include operation context and original error
- [x] Channel sends must use `select { default: }` to prevent blocking
- [x] Goroutines must have clear exit conditions, no leaks
- [x] CLI user-facing errors must use i18n translation keys

## 6. Communication Protocol

### 6.1 Message Format (JSON)

```json
{
  "id": "uuid",
  "type": "register|heartbeat|command|response|error",
  "command": "screen_capture|shell_exec|...",
  "client_id": "client-uuid",
  "payload": {},
  "timestamp": 1234567890
}
```

### 6.2 Message Types

- `register` - Client registration
- `heartbeat` - Heartbeat keep-alive
- `command` - Control command
- `response` - Execution result
- `error` - Error message

## 7. Deployment Architecture

```
Controller (Web/CLI) → server_api (WebSocket) → client (controlled endpoint)
                          ↑
                     server_web (HTTP proxy)
```

- server_api: Port 6341 (WebSocket + HTTP API)
- server_web: Port 7993 (Web interface + API proxy)
- server_cli: Command-line tool
- client: Controlled endpoint program

## 8. Priority

P0 (Must-have): Core WebSocket communication, client management, command routing
P1 (Important): Security, stability, Web/CLI controllers
P2 (Optional): File transfer, screen capture, and other specific command implementations