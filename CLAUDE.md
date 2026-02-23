# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

wxcli is a WeChat/Weixin CLI tool for managing draftbox and material operations for subscription accounts. It handles authentication, access token caching, and provides commands for draft and material CRUD operations.

## Build and Test Commands

```bash
# Build the CLI
go build -o bin/wxcli ./src/cmd/wxcli

# Run tests
CGO_ENABLED=0 go test ./...

# Run a single test
CGO_ENABLED=0 go test -v ./src/internal/auth/...

# Build release binary (uses goreleaser)
goreleaser build --snapshot --clean
```

## Code Architecture

### Entry Point
- `src/cmd/wxcli/main.go` - Simple main that delegates to `cmd.Execute()`

### Command Layer (`src/internal/cmd/`)
- Uses [kong](https://github.com/alecthomas/kong) CLI parser for declarative command definitions
- `root.go` - CLI struct with subcommands (Auth, Config, Draft, Material), context setup, output mode handling
- `auth.go` - Auth subcommands: set, login, status, clear, keyring
- `draft_cmd.go` - Draft subcommands: add, get, list, delete
- `material_cmd.go` - Material subcommands: get, list, delete, count, upload, add-news, update-news
- `json.go` - JSON output helper functions
- `exit_codes.go` - Exit code definitions

### Core Packages

**`src/internal/auth/`**
- `token_manager.go` - Handles access token caching and refresh with 5-minute skew window
- `RequireAppID()` - Validates AppID is configured

**`src/internal/config/`**
- `config.go` - Reads/writes JSON5 config file (AppID, Name, KeyringBackend)
- Uses JSON5 for config parsing, regular JSON for writing

**`src/internal/secrets/`**
- `store.go` - Cross-platform secrets storage
- Linux: stores AppSecret/access_token in config.json (0600 permissions)
- macOS/Windows: uses keyring library with auto/keychain/file backend selection

**`src/internal/draft/`**
- `client.go` - WeChat draft API client (add, get, list, delete)

**`src/internal/material/`**
- `client.go` - WeChat permanent material API client (get, list, delete, count, upload, add-news, update-news)

**`src/internal/httpclient/`**
- `transport.go` - Retry transport with backoff for 429/5xx errors

**`src/internal/markup/`**
- Markdown to HTML conversion and CSS inlining for draft content

### Output Formatting
- Three modes: human (default), JSON (`--json`), plain (`--plain`)
- Context-based mode detection via `outfmt` package

### Error Handling
- `errfmt.APIError` - Wraps WeChat API error responses with errcode/errmsg
- Error output redacts `secret=` and `access_token=` query params

## Key Behaviors

- Token refresh happens automatically when token expires within 5 minutes
- Base URL can be overridden with `--base-url` flag for testing
- Content can be read from stdin using `-` for `--content` flag
- Markdown content is auto-detected and converted to HTML
- CSS files can be inlined into HTML content via `--css-path`
