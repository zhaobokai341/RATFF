# Task 011: Server CLI i18n Language Pack

## Task Description

Implement internationalization (i18n) for `server_cli` so that changing `cfg.Language` in `config.go` switches all CLI output between Chinese and English. Language packs are stored as JSON files with unique keys mapping to translated strings.

## Implementation Details

### New Files Created

- `server_cli/lang/zh.json` - Chinese language pack (35 keys)
- `server_cli/lang/en.json` - English language pack (35 keys)
- `server_cli/i18n.go` - Language pack loader and translation functions

### Modified Files

- `server_cli/config.go` - Added `Language` field with default value `"zh"`
- `server_cli/output.go` - Changed Print functions to accept `msg string` instead of `format string, args...` to avoid govet printf warnings
- `server_cli/main.go` - Added `loadLanguagePacks()` call at startup, replaced all hardcoded strings with `T()`/`Tf()`
- `server_cli/commands.go` - Replaced all hardcoded strings with `T()`/`Tf()`
- `server_cli/helpers.go` - Replaced all hardcoded strings with `T()`/`Tf()`
- `server_cli/output_table.go` - Replaced table headers and messages with `T()`/`Tf()`
- `server_cli/output_help.go` - Replaced help title with `T()`

### Translation Functions

| Function | Usage | Example |
|----------|-------|---------|
| `T(key)` | Simple string lookup, no formatting | `T("login_success")` |
| `Tf(key, args...)` | String lookup with `fmt.Sprintf` formatting | `Tf("selected_client", id)` |

### Language Pack Fallback Logic

1. Look up key in current language (`cfg.Language`)
2. If not found, search all other loaded language packs for the key
3. If still not found, return the key itself as fallback

This design supports adding more languages in the future without code changes - just drop a new JSON file in `lang/`.

### How to Switch Language

Edit `server_cli/config.go`:

```go
Language: "zh",  // Chinese
Language: "en",  // English
```

### Language Pack Format

Each language pack is a JSON file named `<lang_code>.json` in `server_cli/lang/`:

```json
{
    "key_name": "translated string with %v or %s placeholders"
}
```

Keys must be identical across all language packs. Values may contain `fmt.Sprintf` format verbs (`%v`, `%s`, `%d`, etc.).

### Output Function Changes

All Print functions now accept a single `msg string` parameter (no variadic args). Formatting must be done via `Tf()` before calling Print:

```go
// Before (❌ Don't)
PrintError("Failed to fetch clients: %v", err)

// After (✅ Do)
PrintError(Tf("fetch_clients_failed", err))
```

### Tests

- `golangci-lint run ./server_cli/...` passes with zero warnings
- `go build ./...` succeeds
- `go test ./server_cli/...` passes