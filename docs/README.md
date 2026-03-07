# Homestead Documentation

Homestead is a lightweight, self-hosted personal dashboard for monitoring and accessing your home lab services. It is built with Go and vanilla JavaScript — no framework dependencies, no bloat.

## Pages

- [Installation](installation.md) — Build from source, Docker, and system packages
- [Configuration](configuration.md) — Full reference for `config.toml`
- [Icons](icons.md) — Emoji icon reference with ASCII text alternatives
- [API](api.md) — REST API reference
- [Development](development.md) — Contributing and local development guide

## Preview

| Dark | Light |
|------|-------|
| ![Dashboard dark mode](assets/dashboard-dark.png) | ![Dashboard light mode](assets/dashboard-light.png) |

## Quick start

```bash
git clone https://github.com/haydenk/homestead.git
cd homestead
go build -o homestead .
cp config.example.toml config.toml   # create your local config
$EDITOR config.toml
./homestead -config config.toml
```

Open `http://127.0.0.1:8080` in your browser.

## Features

- **Service dashboard** — Centralized hub for all your self-hosted applications organized into sections
- **Real-time health checks** — Background HTTP/HTTPS status monitoring with configurable intervals
- **Search and filter** — Full-text search across service titles, descriptions, and tags (`/` or `Ctrl+K`)
- **Keyboard shortcuts** — `/` or `Ctrl+K` to search, `R` to refresh, `Esc` to close
- **Light / dark theme** — Automatic or explicit theme selection with custom accent colors per card
- **Live config reload** — Reload configuration without restarting via `POST /api/reload`
- **Minimal footprint** — Single binary, one external Go dependency

## Tech stack

| Layer | Technology |
|---|---|
| Backend | Go standard library + `BurntSushi/toml` |
| Frontend | HTML5, vanilla JavaScript (ES2020+), CSS3 |
| Packaging | Docker (Alpine), systemd, nfpm (deb/rpm/apk) |
