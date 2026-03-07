---
template: home.html
title: Homestead
hide:
  - navigation
  - toc
hero:
  title: Homestead
  subtitle: A lightweight, self-hosted personal dashboard for monitoring and accessing your home lab services. Built with Go and vanilla JavaScript — no framework dependencies, no bloat.
  install_button: Get Started
  install_button_link: installation/
  source_button: View on GitHub
  source_button_link: https://github.com/haydenk/homestead
---

## Features

<div class="grid cards" markdown>

-   :material-view-dashboard: **Service Dashboard**

    ---

    Centralized hub for all your self-hosted applications organized into sections with custom icons and accent colors.

-   :material-heart-pulse: **Real-time Health Checks**

    ---

    Background HTTP/HTTPS status monitoring pushed to the UI over WebSocket — no polling from the browser.

-   :material-magnify: **Search & Filter**

    ---

    Full-text search across service titles, descriptions, and tags. Press `/` or `Ctrl+K` to open instantly.

-   :material-theme-light-dark: **Light / Dark Theme**

    ---

    Automatic or explicit theme selection. Each service card can have its own hex accent color.

-   :material-reload: **Live Config Reload**

    ---

    Reload configuration from disk without restarting the process via `POST /api/reload`.

-   :material-feather: **Minimal Footprint**

    ---

    Single binary, ~1,000 lines of code, two external Go dependencies. Embeds the web UI — no separate assets to manage.

</div>

## Quick start

```bash
git clone https://github.com/haydenk/homestead.git
cd homestead
go build -o homestead .
cp config.example.toml config.toml
$EDITOR config.toml
./homestead -config config.toml
```

Open `http://127.0.0.1:8080` in your browser.

[Full installation guide](installation.md){ .md-button }

## Preview

| Dark | Light |
|------|-------|
| ![Dashboard dark mode](assets/dashboard-dark.png) | ![Dashboard light mode](assets/dashboard-light.png) |

## Tech stack

| Layer | Technology |
|-------|-----------|
| Backend | Go standard library + `BurntSushi/toml` + `gorilla/websocket` |
| Frontend | HTML5, vanilla JavaScript (ES2020+), CSS3 custom properties |
| Packaging | Docker (Alpine), systemd, nfpm (deb/rpm/apk) |
