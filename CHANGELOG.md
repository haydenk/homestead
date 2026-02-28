# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
