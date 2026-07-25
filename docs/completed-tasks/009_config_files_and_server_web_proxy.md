# Task 009: Config Files and Server Web Proxy Architecture

## Task Description

1. Create `config.go` files for all modules (server_api, server_web, server_cli, client) to centralize HOST/PORT configuration
2. Server-side loads config from `.env` file, client-side loads from `config.go` defaults
3. Refactor `server_web` as a pure proxy layer - no own password config, path password entered via URL path
4. server_web uses cookies to track auth_token and path_prefix for request forwarding

## Implementation Details

### New Files Created

- `server_api/config.go` - Server API config (HOST, PORT, LOGIN_PATH, LOGIN_PASSWORD_HASH, JWT_SECRET), loaded from env
- `server_web/config.go` - Server Web config (HOST, PORT, APIBaseURL, WsURL), loaded from env
- `server_cli/config.go` - CLI client config (Host, Port, PathPassword, LoginPassword), hardcoded defaults
- `client/config.go` - Client config (ServerHost, ServerPort, PathPassword, ClientID), hardcoded defaults
- `.env` - Example environment file for server-side configuration

### Modified Files

- `server_api/auth.go` - Uses `cfg.JWTSecret` and `cfg.LoginPasswordHash` instead of getEnv
- `server_api/main.go` - Uses `cfg.Host`, `cfg.Port`, `cfg.PathPassword`
- `server_web/auth.go` - `verifyPasswordWithAPI` accepts pathPassword param, saves path_prefix cookie
- `server_web/main.go` - Uses `NoRoute` handler for path password routing, removes own password config
- `server_web/handlers.go` - Added `handleWebSocketWithPath`, `handleAPIProxyWithPath`, `getAuthInfo`, `buildAPIURL`, `buildWSURL`
- `server_web/websocket.go` - Uses `buildWSURL` for connection
- `server_cli/types.go` - Removed global apiBaseURL/wsURL variables
- `server_cli/api.go` - Uses `getAPIBaseURL()` from config
- `server_cli/auth.go` - Uses `getAPIBaseURL()` from config
- `server_cli/main.go` - Uses `cfg.PathPassword`, `cfg.LoginPassword`, `getWSURL()`
- `server_cli/websocket.go` - Accepts wsURL parameter
- `client/main.go` - Uses `getServerURL()` and `cfg.ClientID`
- `server_web/main_test.go` - Updated tests for new architecture

### Architecture

**server_web as proxy layer:**
- No own password/security path configuration
- Path password entered via URL: `http://host:7993/yourpath/`
- On visiting `/<path>/`, sets `path_prefix` cookie
- Login page only asks for login password
- On login success, verifies against server_api using stored path_prefix cookie
- All API/WebSocket requests read `auth_token` and `path_prefix` from cookies
- Constructs correct server_api URLs dynamically based on path_prefix

**URL routing:**
- `/login` - Login page (no auth required)
- `/<path>/` - Sets path_prefix cookie, checks auth_token, shows index
- `/<path>/ws` - WebSocket proxy with path prefix
- `/<path>/api/*` - API proxy with path prefix
- `/` - Direct access (no path prefix), checks auth_token

### Config Loading

| Module | Source | Purpose |
|--------|--------|---------|
| server_api | `.env` / env vars | Server listen addr, passwords, JWT secret |
| server_web | `.env` / env vars | Web listen addr, API/WebSocket upstream URLs |
| server_cli | `config.go` defaults | Server connection details |
| client | `config.go` defaults | Server connection details |

### Tests

All tests pass across all modules.