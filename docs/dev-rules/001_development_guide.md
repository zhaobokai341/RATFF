# Code Standards and Development Guide

## 1. Project Structure

```
RATFF/
├── client/              # Client (controlled endpoint)
├── server_api/          # Server API (core)
├── server_web/          # Web controller
│   ├── lang/             # Language packs (zh.json, en.json)
│   ├── static/           # Static resources
│   │   └── js/           # JavaScript (includes lang-switcher.js)
│   └── templates/        # HTML templates
├── server_cli/          # CLI controller
│   ├── lang/             # Language packs (zh.json, en.json)
│   └── translator.go     # Translator encapsulation
├── shared/              # Shared code
│   └── translations.go   # Translation functions (stateless)
├── docs/                # Documentation
│   ├── requirements/     # Requirements documents
│   ├── tasks/            # Task plans
│   ├── dev-rules/        # Development standards
│   ├── ai-prompts/       # AI prompts
│   └── completed-tasks/  # Completed task descriptions
└── go.mod
```

**Naming Rules:**
- Directory names: lowercase + underscore (snake_case)
- File names: lowercase + underscore (snake_case)
- Package names: same as directory name, lowercase

## 2. Code Standards

### 2.1 File Line Limits
- **Source code files must not exceed 150 lines**
- **Test files (`*_test.go`) must not exceed 300 lines**
- Split into multiple files when exceeded
- Split by responsibility, not by line count
- When a function belongs to the same responsibility domain and splitting would reduce readability, the limit can be appropriately relaxed

### 2.2 Function Standards
- Single function should not exceed 50 lines
- Function names start with a verb, e.g., `NewClientManager`, `handleWebSocket`
- Error handling: return early, avoid nesting

### 2.3 Code Reuse Principles
- Shared code goes in `shared/` directory
- Utility functions encapsulated in `shared/utils.go`
- Protocol definitions in `shared/protocol.go`
- WebSocket utility functions in `shared/ws_utils.go` (`SetupHeartbeat`, `SendWSMessage`, `ReadWSMessage`)
- Translation functions in `shared/translations.go` (`LoadLanguagePacks`, `T`, `Tf`)
- Client info utilities in `shared/client_info.go` (`BuildClientInfo`, `GenerateClientID`, `CalculateBackoff`)
- Prefer standard libraries and mature third-party libraries
- Prohibit duplicating the same logic across multiple modules

### 2.4 Logging Standards
- Use `shared.InitLogger()` to initialize
- Log levels: debug < info < warn < error
- Production environment uses JSON format
- Critical operations must be logged

### 2.5 Error Handling

```go
// Correct example
if err != nil {
    log.Error("Operation failed: ", err)
    return err
}

// With context
log.WithFields(logrus.Fields{
    "client_id": id,
    "command":   cmd,
}).Error("Execution failed")
```

### 2.6 Error Handling Rules (Mandatory)

**2.6.1 No Ignored Errors**
- Every `error` return value must be checked
- In test files, use `_ =` to explicitly indicate intentional ignoring
- Never use blank identifier `_` in production code to swallow errors

```go
// ❌ Wrong - silently ignores error
file.Close()
os.Chdir(dir)
os.Getwd()

// ✅ Correct - handles error
if err := file.Close(); err != nil {
    return fmt.Errorf("close file failed: %v", err)
}

// ✅ Test file - explicitly ignored
_ = conn.Close()
```

**2.6.2 Resource Cleanup Must Be Verified**
- `Close()`, `Shutdown()`, `Cleanup()` calls must check return values
- `defer` is acceptable only when the cleanup error is truly non-critical
- For critical resources (files, connections), return errors on cleanup failure

```go
// ❌ Wrong - cleanup error ignored
defer file.Close()

// ✅ Correct - cleanup in goroutine with error handling (for long-running)
go func() {
    _ = cmd.Wait()
    if err := file.Close(); err != nil {
        log.Error("Failed to close output file: ", err)
    }
}()

// ✅ Correct - immediate cleanup with error check
if err := session.File.Close(); err != nil {
    return shared.NewMessage(shared.MsgError, cmd, clientID,
        map[string]interface{}{"error": fmt.Sprintf("close file failed: %v", err)})
}
```

**2.6.3 Complete Exception Coverage**
Every feature must handle all possible failure scenarios:

| Scenario | Required Handling |
|----------|------------------|
| File not found | Check before operation, return clear error |
| Permission denied | Catch and return, do not crash |
| Network timeout | Set timeout, return error with context |
| Connection lost | Detect, clean up resources, attempt reconnect |
| Invalid input | Validate early, return descriptive error |
| Disk full | Check before write, return error |
| Channel full | Use `select { default: }` to avoid blocking |
| Goroutine leak | Ensure exit on connection close or context cancel |

**2.6.4 Error Messages Must Be Descriptive**
- Include operation context: what was being done
- Include the original error: `fmt.Sprintf("operation failed: %v", err)`
- Use i18n keys for user-facing errors in CLI

```go
// ❌ Wrong - no context
return err

// ✅ Correct - with context
return fmt.Errorf("upload chunk %d failed: %v", chunkIndex, err)

// ✅ CLI user-facing
PrintError(Tf("upload_chunk_failed", chunkIndex))
```

**2.6.5 Channel Operations Must Not Block Indefinitely**
- Use `select` with `default` or `timeout` for sends
- Never send to a channel without a receiver guarantee

```go
// ❌ Wrong - may block forever if channel is full
pc.ch <- msg

// ✅ Correct - non-blocking send
select {
case pc.ch <- msg:
default:
    // Channel full, message dropped
}
```

**2.6.6 Goroutine Lifecycle Management**
- Every goroutine must have a clear exit condition
- Use `context.Context` or done channels for cancellation
- Ticker/Timer must be stopped with `defer`

```go
// ✅ Correct - exits on connection close
go func() {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    for range ticker.C {
        if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
            return // Connection closed, exit goroutine
        }
    }
}()
```

### 2.7 Concurrency Safety
- Shared maps use `sync.RWMutex`
- Read-heavy scenarios use `RLock/RUnlock`
- Write operations use `Lock/Unlock`
- Use `defer` to ensure unlocking

## 3. Technology Stack

| Feature | Library | Version |
|---------|---------|---------|
| HTTP | gin-gonic/gin | latest |
| WebSocket | gorilla/websocket | latest |
| Logging | sirupsen/logrus | latest |
| CLI | urfave/cli/v2 | latest |
| UUID | google/uuid | latest |
| Rate Limiting | golang.org/x/time | latest |
| JWT | golang-jwt/jwt/v5 | latest |
| Password Hashing | golang.org/x/crypto/bcrypt | latest |
| Terminal Input | golang.org/x/term | latest |
| Test Assertions | stretchr/testify | latest |
| Terminal Styling | charmbracelet/lipgloss | latest |
| Shell Parsing | google/shlex | latest |

### 3.1 Frontend Technology Stack

**Framework Selection:**
- **CSS Framework**: Tailwind CSS (via CDN)
- **JS Framework**: Vue.js 3 (via CDN)

**Usage Standards:**
- All HTML templates use dark theme design
- Vue uses `[[ ]]` as delimiter to avoid conflict with Go template's `{{ }}`
- Tailwind configures custom dark theme colors
- Keep template files concise, move complex logic to JS files
- Use Vue's reactive data management for form state
- Loading state uses spinning icon to prompt users

## 4. How to Add New Features

### 4.1 Add New Command Types

1. Add command constant in `shared/protocol.go`:
```go
const CmdNewCommand CommandType = "new_command"
```

2. Add handler function in `client/handler.go`:
```go
func handleNewCommand(msg shared.Message) shared.Message {
    // Implementation
    return shared.NewMessage(shared.MsgResponse, shared.CmdNewCommand, "", payload)
}
```

3. Add case in `executeCommand` switch:
```go
case shared.CmdNewCommand:
    resp = handleNewCommand(msg)
```

### 4.2 Add New HTTP API

Add route in `server_api/main.go`'s `setupRouter`:
```go
api.GET("/new-endpoint", handleNewEndpoint(manager))
```

Implement handler in `server_api/handler.go`:
```go
func handleNewEndpoint(manager *ClientManager) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Implementation
        c.JSON(200, result)
    }
}
```

### 4.3 Add New Module

1. Create directory: `mkdir new_module`
2. Create main.go
3. Reuse shared library's logging and utility functions
4. Keep single file < 150 lines

## 5. Security Development Guide

### 5.1 Authentication
- Client registration must provide Token
- Token verified in server configuration
- Reject connection on failure

### 5.2 Rate Limiting
- Use `golang.org/x/time/rate`
- Default 50 requests per second
- Adjustable as needed

### 5.3 Encryption
- Production environment must use TLS
- Sensitive data encrypted with AES
- Keys must not be hardcoded

## 6. Testing Guide

### 6.1 Unit Tests (Required)
- Each module's `_test.go` file in the same directory as source
- Test file naming: `xxx_test.go`
- Test function naming: `TestXxx`
- Use standard library `testing` package
- Coverage targets: core logic > 80%, overall > 60%

```bash
# Run all tests
go test ./...

# Run specific module tests
go test ./shared/...

# With coverage
go test -cover ./...

# Verbose output
go test -v ./...
```

### 6.2 Test Writing Standards

**Test Categories:**
- **Unit tests**: Test pure function logic, no external dependencies
- **Integration tests**: Use `httptest` to simulate HTTP/WebSocket interaction
- **Boundary tests**: Test empty values, error values, abnormal input

**Test Template:**
```go
func TestXxxSuccess(t *testing.T) {
    // 1. Prepare test data
    // 2. Execute tested function
    // 3. Assert results
    assert.Equal(t, expected, actual)
}

func TestXxxError(t *testing.T) {
    // Test error paths
    assert.Error(t, err)
}

func TestXxxEdgeCase(t *testing.T) {
    // Test boundary conditions
    // Use sub-tests
    t.Run("case1", func(t *testing.T) { ... })
    t.Run("case2", func(t *testing.T) { ... })
}
```

**HTTP/WebSocket Testing:**
- Use `net/http/httptest.NewServer()` to create test server
- WebSocket tests use `gorilla/websocket`'s `DefaultDialer`
- Always `defer server.Close()` and `defer conn.Close()` after testing

**Assertion Standards:**
- Use `github.com/stretchr/testify/assert` for assertions
- Each test includes at least one assertion
- Error messages use `t.Errorf("Expected %s, got %s", expected, actual)`

### 6.3 Mock Usage Standards

**HTTP Mock:**
```go
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // Simulate server behavior
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}))
defer server.Close()

// Get WebSocket URL
wsURL := "ws" + server.URL[4:] + "/ws"
```

**Configuration Mock:**
- Set `cfg` global variable before testing
- Restore original value with `defer` after testing
- When using environment variables, clean up after setting: `os.Setenv()` + `defer os.Unsetenv()`

### 6.4 Test Coverage Requirements

| Module Type | Minimum Coverage | Description |
|-------------|-----------------|-------------|
| shared (core library) | > 80% | Protocol, utility functions must have high coverage |
| server_api (server) | > 60% | HTTP routes, WebSocket handling |
| client (client) | > 60% | Connection, command execution logic |
| server_cli (CLI) | > 50% | Output, interaction logic |
| server_web (Web) | > 30% | Proxy, page rendering |

**Methods to Improve Coverage:**
- Write unit tests for core business logic
- Use `httptest` to test HTTP endpoints
- Use WebSocket test server to test connection logic
- Test error paths and boundary conditions

### 6.5 Code Checking (Required)
- Execute `golangci-lint run` after writing code
- Fix all lint issues before committing
- Error return values in test files must also be checked (use `_ =` to explicitly ignore)

```bash
# Install
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Check all code
golangci-lint run ./...

# Check specific module
golangci-lint run ./shared/...
```

### 6.6 Build Check
```bash
go build ./...
```

### 6.7 Running Tests
```bash
# Start server_api
cd server_api && go run .

# Start client
cd client && go run .

# Start server_web
cd server_web && go run .

# Use server_cli
cd server_cli && go run .
```

## 7. Requirement Change Process

1. Create or update requirement document in `docs/requirements/`
2. Update task plan in `docs/tasks/`
3. Code and write unit tests
4. Run `golangci-lint run` check
5. Record completion in `docs/completed-tasks/`

## 8. Deployment Guide

### 8.1 Build
```bash
GOOS=linux GOARCH=amd64 go build -o bin/client ./client
GOOS=linux GOARCH=amd64 go build -o bin/server_api ./server_api
```

### 8.2 Configure TLS
```bash
# With certificate
./server_api -cert server.crt -key server.key

# Without certificate (auto-degrade WS, will have warning)
./server_api
```

## 9. AI-Assisted Development Guide

When using AI to continue development, provide:
1. The `docs/` folder in the project root directory
2. AI should first read the guide files in `docs/ai-prompts/`
3. Then read other documents in the guided order

AI can quickly understand the project structure and generate code that meets the standards.

## 10. Important Notes

### 10.1 CLI Connection Mechanism
- `server_cli` needs to establish WebSocket connection to server to receive command responses
- CLI registers with `__cli__` prefix ID (e.g., `__cli__a1b2c3d4`)
- `ClientManager.ListClients()` automatically filters `__cli__` prefixed clients
- Device list should not contain entries starting with `__cli__`

### 10.2 Documentation Comment Standards
- All exported types, constants, variables, functions, and methods must have `//` documentation comments
- Comments start with the type name or function name, e.g., `// ClientInfo holds information about a connected client.`
- Non-exported internal functions use lowercase comments, e.g., `// buildOSInfo constructs an OS info string`

### 10.3 Code Reuse Principles
- Shared types (like `ClientInfo`, `Message`) are defined in the `shared/` package
- Each module directly references `shared.XXX`, duplicate definitions are prohibited
- When duplicate definitions are found, delete immediately and replace with `shared` reference

### 10.4 Response Matching Mechanism
- `server_cli` uses `pendingCmd` map to store pending command responses
- Key is `clientID` (controlled endpoint ID), value is `pendingCommand` with channel
- `listenResponses` matches based on response message's `ClientID` and delivers to corresponding channel

### 10.5 CLI Output Standards
- All user-visible output in `server_cli` must use functions defined in `output.go`, `output_table.go`, and `output_help.go`
- Direct use of `fmt.Println` or `fmt.Printf` for user-visible information is prohibited
- See `docs/dev-rules/002_cli_output_styling.md` for details

### 10.6 Internationalization (i18n) Standards

**Backend Translation Functions:**
- Translation functions defined in `shared/translations.go`
  - `LoadLanguagePacks(langDir string)` - Load all `.json` language packs in specified directory
  - `T(langCode, key string)` - Look up translation, return key itself if not found
  - `Tf(langCode, key string, args ...interface{})` - Formatted translation
- Each module should call `shared.LoadLanguagePacks("lang")` at startup to load language packs
- Language pack file naming: `<lang_code>.json` (e.g., `zh.json`, `en.json`), stored in module's `lang/` directory
- All language pack keys must be identical, values may contain `fmt.Sprintf` format placeholders (`%v`, `%s`, `%d`, etc.)

**CLI Module Translator Encapsulation:**
- `server_cli` uses `Translator` struct to encapsulate language parameter, avoiding passing `langCode` on each call
- `Translator` type and `T()`/`Tf()` package-level functions defined in `server_cli/translator.go`
- All user-visible strings must be obtained through `T()` or `Tf()`, hardcoding is prohibited

**Web Module Language Switching:**
- Language state stored in cookie `app_lang`
- `translator.go` defines language middleware `languageMiddleware()`, reads language from cookie and injects into gin context
- `T(c *gin.Context, key)` and `Tf(c *gin.Context, key, args...)` get current language from context for translation
- Provides `/api/lang` (GET/POST) API for getting/setting language
- Language switch button encapsulated in `static/js/lang-switcher.js`, imported as Vue component for each page
- Language switching **does not affect global theme** (dark mode remains unchanged)
- HTML template's `<html lang="">` attribute should dynamically reflect current language

**Frontend i18n Implementation:**
- Define `messages` object in HTML template's `<script>`, containing text for each language
- Read current language from Go template variable `{{.lang}}`, select corresponding `messages` sub-object
- Replace all hardcoded text with Vue data binding (e.g., `${ labels.refresh }`)
- Cross-page reusable components (like language switcher) should be independent files, imported via `<script>` tag

**Adding New Language:**
1. Create new `<lang_code>.json` file in corresponding module's `lang/` directory
2. Copy all keys from existing language pack, fill in new language translation values
3. Frontend `messages` object adds corresponding language text
4. `lang-switcher.js` adds new `<option>` option