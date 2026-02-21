# wxcli

Weixin CLI for subscription accounts. Manages auth, access_token caching, and draftbox/material operations.

## Features

- Draftbox: add/get/list/delete
- Material: get/list/delete/count/upload/update-news for permanent materials
- Auth: AppID in config; AppSecret/access_token in keyring (Linux: config file), access_token caching + refresh
- Output modes: human (default), JSON, plain
- Retry/backoff for 429/5xx
- Keyring backend selection on macOS/Windows

## Install

### macOS (Homebrew Cask)

```bash
brew tap inforix/wxcli
brew install --cask wxcli
```

### Linux (deb/rpm)

Download the `.deb` or `.rpm` from GitHub Releases, then install:

```bash
# Debian/Ubuntu
sudo dpkg -i wxcli_<version>_linux_amd64.deb

# RHEL/CentOS/Fedora
sudo rpm -i wxcli_<version>_linux_amd64.rpm
```

### Build from source

```bash
go build -o bin/wxcli ./src/cmd/wxcli
```

## Configuration

AppID is stored in config.json. On macOS/Windows, AppSecret/access_token are stored in the keyring; on Linux they are stored in config.json.

```bash
wxcli config path
```

## Auth

Set AppID/AppSecret:

```bash
wxcli auth set --appid YOUR_APPID --appsecret YOUR_SECRET
```

Interactive login:

```bash
wxcli auth login
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

Keyring backend (macOS/Windows):

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

Read HTML content from stdin:

```bash
npx markdown-to-html-cli --source article.md --style=./style.css | \
  wxcli draft add --title "Hello" --content - --thumb-media-id MEDIA_ID
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

## Material

Get a permanent material (news/video JSON or binary media):

```bash
wxcli material get MEDIA_ID
```

Save binary material to a file:

```bash
wxcli material get MEDIA_ID --output ./downloads
```

Upload a permanent material (image/voice/video/thumb):

```bash
wxcli material upload --type image --file ./path/to/image.jpg
wxcli material upload --type video --file ./path/to/video.mp4 --title "Video" --description "Intro"
```

Upload a permanent news material:

```bash
wxcli material add-news --title "Hello" --content "<p>Hi</p>" --thumb-media-id MEDIA_ID
```

Update a permanent news material:

```bash
wxcli material update-news MEDIA_ID --index 0 --title "Hello" --content "<p>Hi</p>" --thumb-media-id MEDIA_ID
```

List permanent materials (news titles per article are shown):

```bash
wxcli material list --type image --offset 0 --count 10
```

Delete a permanent material:

```bash
wxcli material delete MEDIA_ID
```

Get material counts:

```bash
wxcli material count
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

- On macOS/Windows, AppSecret/access_token are stored in the keyring, not config.json
- On Linux, AppSecret/access_token are stored in config.json (0600)
- Keyring backend can be configured via `wxcli auth keyring` or env vars on macOS/Windows
- Error output redacts `secret=` and `access_token=` query params

## Environment Variables (macOS/Windows)

- `WXCLI_KEYRING_BACKEND`: `auto|keychain|file`
- `WXCLI_KEYRING_PASSWORD`: password for file backend (non‑interactive)
