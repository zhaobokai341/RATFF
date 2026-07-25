# Task 010: Server Web Routing Refactor

## Task Description

Fix 404 errors on `/api/clients` and other routes by replacing `NoRoute` handler with explicit route registration. Support both root-level routes (no path prefix) and path-prefixed routes (with path password in URL).

## Problem

Previous implementation used `r.NoRoute(handlePathRouter)` to catch all unmatched routes and parse path passwords from URL. This caused `/api/clients` and other root-level routes to return 404 because they were not explicitly registered.

## Solution

Replaced `NoRoute` with explicit route registration for both root-level and path-prefixed routes.

## Route Structure

### Root-level routes (no path prefix)
| Method | Route | Handler | Description |
|--------|-------|---------|-------------|
| GET | `/` | handleRoot | Index page (requires auth_token cookie) |
| GET | `/ws` | handleWebSocketRoot | WebSocket proxy (no path prefix) |
| GET | `/api/clients` | handleAPIClientsRoot | Get client list |
| POST | `/api/command` | handleAPICommandRoot | Send command |

### Path-prefixed routes (with path password)
| Method | Route | Handler | Description |
|--------|-------|---------|-------------|
| GET | `/:pathPassword` | handlePathIndex | Set path_prefix cookie, show index |
| GET | `/:pathPassword/` | handlePathIndex | Same as above |
| GET | `/:pathPassword/index.html` | handlePathIndex | Same as above |
| GET | `/:pathPassword/ws` | handlePathWebSocket | WebSocket proxy with path prefix |
| GET | `/:pathPassword/api/clients` | handlePathAPIClients | Get client list with path prefix |
| POST | `/:pathPassword/api/command` | handleAPICommand | Send command with path prefix |

### Auth routes
| Method | Route | Handler | Description |
|--------|-------|---------|-------------|
| GET | `/login` | handleLoginPage | Login page |
| POST | `/login` | handleLogin | Verify password, set cookies |
| GET | `/logout` | handleLogout | Clear cookies |

## Modified Files

- `server_web/main.go` - Replaced `NoRoute` with explicit routes, added `handleRoot` and `handlePathIndex`
- `server_web/handlers.go` - Added `handleWebSocketRoot`, `handleAPIClientsRoot`, `handleAPICommandRoot`, `handlePathWebSocket`, `handlePathAPIClients`, `handlePathAPICommand`, moved `upgrader` here
- `server_web/websocket.go` - Removed duplicate `upgrader` declaration, removed unused `handleWebSocket`
- `server_web/main_test.go` - Updated tests for new route structure

## Key Behavior

1. Visiting `/:pathPassword/` automatically sets `path_prefix` cookie
2. All API/WebSocket handlers read `auth_token` and `path_prefix` from cookies
3. Root-level routes work without path prefix (empty path_prefix)
4. Path-prefixed routes set the path_prefix cookie and forward to server_api with correct URL