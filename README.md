# wxcli

Weixin draftbox CLI for subscription accounts. Manages auth, access_token caching, and draftbox operations.

## Features

- Draftbox: add/get/list/delete
- Auth: AppID/AppSecret in config + keyring, access_token caching + refresh
- Output modes: human (default), JSON, plain
- Retry/backoff for 429/5xx
- Cross‑platform keyring backend selection

## Install

```bash
go build -o bin/wxcli ./cmd/wxcli
```

## Configuration

AppID is stored in config.json; AppSecret/access_token are stored in keyring.

```bash
wxcli config path
```

## Auth

Set AppID/AppSecret:

```bash
wxcli auth set --appid YOUR_APPID --appsecret YOUR_SECRET
```

Check status:

```bash
wxcli auth status
wxcli auth status --json
wxcli auth status --plain
```

Clear secrets:

```bash
wxcli auth clear
```

Keyring backend:

```bash
wxcli auth keyring
wxcli auth keyring auto
wxcli auth keyring keychain
wxcli auth keyring file
```

## Draftbox

Add a draft:

```bash
wxcli draft add --title "Hello" --content "<p>Hi</p>" --thumb-media-id MEDIA_ID
```

Get a draft:

```bash
wxcli draft get MEDIA_ID
```

List drafts:

```bash
wxcli draft list --offset 0 --count 10 --no-content 1
```

Delete a draft:

```bash
wxcli draft delete MEDIA_ID
```

## Output Modes

- Default: human‑readable
- `--json`: JSON output to stdout
- `--plain`: key=value lines (script‑friendly)

Examples:

```bash
wxcli draft list --json
wxcli auth status --plain
```

## Base URL Override (testing)

```bash
wxcli --base-url http://127.0.0.1:9000 draft list --json
```

## Tests

```bash
CGO_ENABLED=0 go test ./...
```

## Security Notes

- Never store AppSecret/access_token in config.json
- Keyring backend can be configured via `wxcli auth keyring` or env vars
- Error output redacts `secret=` and `access_token=` query params

## Environment Variables

- `WXCLI_KEYRING_BACKEND`: `auto|keychain|file`
- `WXCLI_KEYRING_PASSWORD`: password for file backend (non‑interactive)
