# CLI Output Styling Guide

## Overview

The `server_cli` module uses `github.com/charmbracelet/lipgloss` for terminal output styling. All CLI output must use the standardized output functions defined in `output.go`, `output_table.go`, and `output_help.go`. **Never use raw `fmt.Println` or `fmt.Printf` for user-facing output.**

## Color System

| Type | Prefix | Color | ANSI Code | Usage |
|------|--------|-------|-----------|-------|
| Success | `[+]` | Green | `10` | Login success, device selected, command executed |
| Error | `[-]` | Red | `9` | Login failed, command timeout, network error |
| Info | `[*]` | Blue | `12` | Device list, connection status, general info |
| Debug | `[debug]` | Gray | `8` | Debug messages (future use) |
| Warn | `[!]` | Yellow | `11` | Warning messages (future use) |

## Available Output Functions

### Basic Output (`output.go`)

```go
// Success message (green)
PrintSuccess(T("login_success"))
PrintSuccess(Tf("selected_client", id))

// Error message (red)
PrintError(Tf("connect_failed", err))
PrintError(T("command_timeout"))

// Info message (blue)
PrintInfo(Tf("connected_devices", len(clients)))
PrintInfo(T("prompt_path_password"))

// Debug message (gray)
PrintDebug(Tf("request_payload", payload))

// Warning message (yellow)
PrintWarn(T("connection_unstable"))

// Colored prompt string
prompt := BuildPrompt(id, inCommandMode)
// Returns: "(server) >> " or "(dev-01)(console) >> " in cyan

// Bordered code block for command output
styledOutput := StyleCommandOutput(shellOutput)
// Renders output inside a rounded border box
```

### Table Output (`output_table.go`)

```go
// Print client list as a formatted table
PrintClientTable(clients)
// Automatically handles empty list case with info message
```

**Example output:**
```
[*] Connected devices (2):

┌──────────────────────────────┬────────────────┬────────────────┬──────────────────┐
│ ID                           │ IP             │ HOSTNAME       │ OS               │
├──────────────────────────────┼────────────────┼────────────────┼──────────────────┤
│ abc123def456                 │ 192.168.1.100  │ ubuntu-server  │ Linux x86_64     │
│ ghi789jkl012                 │ 192.168.1.101  │ win10-pc       │ Windows 10       │
└──────────────────────────────┴────────────────┴────────────────┴──────────────────┘
```

### Help Output (`output_help.go`)

```go
// Print formatted help with aligned commands and descriptions
PrintHelp([]HelpCommand{
    {"list", "List connected clients"},
    {"select <id>", "Select a client to control"},
    {"exit", "Exit the CLI"},
})
```

**Example output:**
```
[*] Available commands:

  list               - List connected clients
  select <id>        - Select a client to control
  exit               - Exit the CLI
```

## Performance Best Practices

1. **Styles are globally defined** - Never create `lipgloss.NewStyle()` inside render loops or frequently called functions. All styles are pre-initialized as package-level variables in `output.go`, `output_table.go`, and `output_help.go`.

2. **Avoid deep nesting** - The table rendering uses flat layout composition. Avoid excessive `JoinHorizontal` / `JoinVertical` nesting.

3. **No Bubble Tea** - This CLI uses only static output styling, not the full TUI framework. This keeps initialization time under 2ms.

## Migration Rules

When adding new output or modifying existing code:

| Before (❌ Don't) | After (✅ Do) |
|-------------------|---------------|
| `fmt.Println("[+] Success")` | `PrintSuccess(T("key_name"))` |
| `fmt.Printf("[-] Error: %v\n", err)` | `PrintError(Tf("error_key", err))` |
| `fmt.Println("[*] Info")` | `PrintInfo(T("info_key"))` |
| Manual table with `strings.Repeat` | `PrintClientTable(clients)` |
| Manual help text formatting | `PrintHelp([]HelpCommand{...})` |
| Hardcoded strings in Print functions | `T()` / `Tf()` with language pack keys |

## File Structure

```
server_cli/
├── output.go           # Color system + basic output functions
├── output_table.go     # Table rendering functions
├── output_help.go      # Help message formatting
├── commands.go         # Business logic (uses output functions)
├── helpers.go          # CLI helpers (uses output functions)
├── main.go             # Entry point (uses output functions)
└── ...
```

## Dependencies

- `github.com/charmbracelet/lipgloss` - Terminal styling library
- Pure string manipulation, no virtual terminal or canvas layer
- Automatic terminal capability detection (NO_COLOR, TTY, color depth)

## Internationalization (i18n)

All user-facing strings in `server_cli` must use the translation functions defined in `i18n.go`. **Never use hardcoded strings for user-facing output.**

### Translation Functions

| Function | Signature | Usage |
|----------|-----------|-------|
| `T(key)` | `T(key string) string` | Simple string lookup, no formatting |
| `Tf(key, args...)` | `Tf(key string, args ...interface{}) string` | String lookup with `fmt.Sprintf` formatting |

### How It Works

1. Language packs are JSON files in `server_cli/lang/` (e.g., `zh.json`, `en.json`)
2. Each pack contains the same set of keys with translated values
3. `cfg.Language` in `config.go` determines the active language
4. Fallback: if key not found in current language, searches other languages, then returns the key itself

### Rules for Adding New User-Facing Strings

1. **Add the key to ALL language packs** in `server_cli/lang/` with translations for each language
2. **Use `T()` for simple strings**, `Tf()` for strings with format arguments
3. **Call Print functions with the translated result**, not with format args:

```go
// ✅ Correct
PrintError(Tf("fetch_clients_failed", err))
PrintInfo(T("no_device_connected"))
PrintSuccess(Tf("selected_client", id))

// ❌ Wrong - Print functions no longer accept format args
PrintError("fetch_clients_failed: %v", err)
PrintInfo(T("no_device_connected", count))
```

4. **Keys must be identical across all language packs** - only values differ
5. **Values may contain `fmt.Sprintf` format verbs** (`%v`, `%s`, `%d`, etc.) when using `Tf()`

### Adding a New Language

1. Create a new JSON file: `server_cli/lang/<code>.json`
2. Copy all keys from an existing language pack
3. Translate all values to the new language
4. No code changes needed - the loader auto-detects new `.json` files

### Language Pack Structure

```
server_cli/
├── lang/
│   ├── en.json          # English
│   └── zh.json          # Chinese
├── i18n.go              # Loader + T() / Tf() functions
├── output.go            # Print functions (use translated strings)
├── output_table.go      # Table output (uses T/Tf)
├── output_help.go       # Help output (uses T/Tf)
└── ...
```

### Switching Language

Edit `server_cli/config.go`:

```go
Language: "zh",  // or "en", or any future language code
```