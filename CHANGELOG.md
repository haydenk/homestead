# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]


## [0.2.0] - 2026-03-07

### Changed

- **Release tagging** — Tags now use a `v` prefix (e.g. `v0.2.0`) to comply with Go module versioning requirements

## [0.1.1] - 2026-03-07

### Added

- **Code of Conduct** — Added `CODE_OF_CONDUCT.md` (Contributor Covenant v2.1)
- **Security Policy** — Added `SECURITY.md` with supported versions and vulnerability reporting instructions
- **Dependabot** — Automated dependency updates for Go modules, Docker base images, and GitHub Actions

### Changed

- **Go runtime** — Updated from `golang:1.22-alpine` to `golang:1.26-alpine`
- **Alpine base image** — Updated from `3.19` to `3.23`
- **`github.com/BurntSushi/toml`** — Updated from `1.4.0` to `1.6.0`
- **`docker/build-push-action`** — Updated from `6.19.2` to `7.0.0`
- **`docker/metadata-action`** — Updated from `5.10.0` to `6.0.0`
- **`actions/github-script`** — Updated from `7.1.0` to `8.0.0`
- **`actions/stale`** — Updated from `9.1.0` to `10.2.0`
- **`jdx/mise-action`** — Updated from `3.6.1` to `3.6.3`

## [0.1.0] - 2026-03-06

### Added

- **WebSocket status updates** — New `GET /ws` endpoint pushes health check results to all connected clients in real time; clients receive an immediate snapshot on connect and a fresh push after each check round
- **Dual-stack listening** — Server now binds to both IPv4 (`0.0.0.0`) and IPv6 (`[::]`) by default, with graceful fallback if either family is unavailable

### Changed

- **Default bind address** — Changed from `127.0.0.1` (loopback only) to all interfaces; pass `-host 127.0.0.1` to restore the previous behaviour
- **Status delivery** — The dashboard UI now receives status updates via WebSocket instead of HTTP polling, reducing latency and eliminating periodic requests
- **Static file routing** — Index route changed to an exact match (`GET /{$}`) so requests for static assets (`/style.css`, `/app.js`) are served correctly by the file handler

### Removed

- **`GET /api/status`** — Replaced by the WebSocket endpoint
- **`GET /api/status/{id}`** — Replaced by the WebSocket endpoint
- **`GET /api/config`** — Removed

### Fixed

- **Empty state always visible** — Added `.empty-state[hidden] { display: none }` to prevent `display: flex` from overriding the `hidden` attribute when the search box is empty

## [0.0.1] - 2026-02-28

### Added

- **Service Dashboard** — Centralized hub for self-hosted applications organized into named sections with icons
- **Real-time Health Checks** — Background HTTP/HTTPS status monitoring with a configurable interval (`check_interval`)
  - Uses `HEAD` first, falls back to `GET` on `405 Method Not Allowed`
  - Any response with `statusCode < 500` is treated as up (supports auth-protected endpoints)
  - Accepts self-signed TLS certificates
  - 10-second request timeout per check
- **Search & Filter** — Full-text search across service titles, descriptions, and tags; activate with `/` or `Ctrl+K`
- **Keyboard Shortcuts** — `/` or `Ctrl+K` to open search, `R` to refresh statuses, `Esc` to close search
- **Light / Dark Theme** — Automatic system-preference detection with explicit `"light"` / `"dark"` override in config; per-card custom accent colors via `color` field
- **Live Config Reload** — Reload configuration from disk and restart health checks without restarting the process via `POST /api/reload`
- **TOML Configuration** — Single `config.toml` file with support for dashboard title, subtitle, logo, column count, footer, theme, sections, and items
- **Runtime Overrides** — Config path, bind address, and port configurable via CLI flags (`-config`, `-host`, `-port`) or environment variables (`HOMESTEAD_CONFIG`, `HOMESTEAD_HOST`, `HOMESTEAD_PORT`)
- **REST API** — Endpoints for config (`GET /api/config`), all statuses (`GET /api/status`), single item status (`GET /api/status/{id}`), config reload (`POST /api/reload`), and health probe (`GET /api/health`)
- **Single Binary** — Self-contained Go binary with one external dependency (`BurntSushi/toml`)
- **Docker Support** — Multi-stage Alpine-based Docker image with cross-platform builds for `linux/amd64` and `linux/arm64`
- **System Package Support** — `deb`, `rpm`, and `apk` packages via nfpm; installs binary, config, and a hardened systemd service running as a non-root `homestead` user
- **CI/CD Pipelines** — GitHub Actions workflows for continuous integration (test on push/PR) and automated release (binaries, packages, Docker image, GitHub Release) on semver tags
