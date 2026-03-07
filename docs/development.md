# Development

## Requirements

- Go 1.22 or later
- [mise](https://mise.jdx.dev/) (optional — manages tool versions via `.mise.toml`)

---

## Getting started

```bash
git clone https://github.com/haydenk/homestead.git
cd homestead

# Install tool versions declared in .mise.toml (optional)
mise install

# Copy the example config (once)
cp config.example.toml config.toml

# Run locally (listens on all IPv4 + IPv6 interfaces by default)
go run . -config config.toml

# Build a binary
go build -o homestead .

# Run tests
go test ./...

# Format
go fmt ./...
```

The server listens on all IPv4 and IPv6 interfaces (`0.0.0.0` and `[::]`) on port `8080` by default. Use `-host` to restrict to a specific address and `-port` to change the port.

---

## Dev container

A VS Code dev container is included in `.devcontainer/` with Go, mise, and the TOML formatter pre-configured. Open the repository in VS Code and choose **Reopen in Container** when prompted.

---

## Project structure

```
.
├── main.go                      # Entry point — flag parsing, wiring, graceful shutdown
├── config.toml                  # Example configuration
├── go.mod / go.sum              # Module definition and dependency lock
├── Dockerfile                   # Multi-stage build
├── internal/
│   ├── config/
│   │   ├── config.go            # Config structs and TOML loader
│   │   └── config_test.go
│   ├── checker/
│   │   ├── checker.go           # Background health check engine
│   │   └── checker_test.go
│   └── server/
│       ├── server.go            # HTTP server setup, routing, dual-stack listener
│       ├── hub.go               # WebSocket connection hub and broadcaster
│       ├── handlers.go          # HTTP and WebSocket handler implementations
│       └── handlers_test.go
├── web/
│   ├── index.html               # UI template (served as embedded FS)
│   ├── app.js                   # Frontend logic (vanilla JS, ES2020+)
│   └── style.css                # Styles with CSS custom properties for theming
├── packaging/
│   ├── nfpm.yaml                # Package build config (deb, rpm, apk)
│   ├── homestead.service        # systemd unit file
│   ├── postinstall.sh           # Package post-install hook
│   └── preremove.sh             # Package pre-remove hook
├── docs/                        # This documentation
└── .github/workflows/           # CI and release pipelines
```

---

## mise tasks

Tasks are defined in `.mise/tasks/` and can be run with `mise run <task>`.

| Task | Description |
|---|---|
| `run` | Build and run the binary |
| `dev` | Run with live reload during development |
| `test` | Run the test suite |
| `tidy` | Run `go mod tidy` |
| `clean` | Remove build artifacts |
| `install` | Install the binary system-wide |
| `uninstall` | Remove the installed binary |
| `docker/build` | Build the Docker image |
| `docker/run` | Run the Docker container |

---

## Branch model

This project follows [git flow](https://nvie.com/posts/a-successful-git-branching-model/). The two long-lived branches are:

| Branch | Purpose |
|---|---|
| `master` | Always reflects the latest production release. Protected — no direct pushes. |
| `develop` | Integration branch for completed features. CI must pass before merge. |

### Feature development

```bash
# 1. Branch from develop
git checkout develop
git pull origin develop
git checkout -b feature/my-feature

# 2. Work, commit, push
git push -u origin feature/my-feature

# 3. Open a PR → develop
#    CI must pass. Merge when ready.
```

### Release procedure

```bash
# 1. Cut a release branch from develop
git checkout develop
git pull origin develop
git checkout -b release/0.2.0

# 2. Update CHANGELOG.md — move [Unreleased] items under ## [0.2.0] - YYYY-MM-DD
# 3. Commit and push
git add CHANGELOG.md
git commit -m "chore: prepare release 0.2.0"
git push -u origin release/0.2.0

# 4. Open a PR: release/0.2.0 → master
#    CI must pass. Merging triggers the automated release pipeline:
#      - Tags master with 0.2.0
#      - Builds binaries, packages, and Docker image
#      - Creates a GitHub Release with the changelog entry
#      - Back-merges master into develop automatically
```

### Hotfix procedure

Use hotfixes for urgent fixes to production that cannot wait for the next planned release.

```bash
# 1. Branch from master (not develop)
git checkout master
git pull origin master
git checkout -b hotfix/0.1.2

# 2. Apply the fix, update CHANGELOG.md
git add .
git commit -m "fix: description of the fix"
git push -u origin hotfix/0.1.2

# 3. Open a PR: hotfix/0.1.2 → master
#    Merging triggers the same automated pipeline as a release:
#      - Tags master with 0.1.2
#      - Builds and publishes all artifacts
#      - Back-merges master into develop automatically
```

---

## Architecture

### Config loading (`internal/config`)

`config.Load` reads the TOML file using `BurntSushi/toml`, applies defaults for any missing fields, and assigns stable IDs to each item (`s{section}-i{item}`). The resulting `*Config` value is immutable during normal operation and replaced atomically on live reload.

### Health checker (`internal/checker`)

`Checker` runs a goroutine-per-item health check on a configurable interval. Results are stored in a `map[string]*Status` guarded by a `sync.RWMutex`. Checks use `HEAD` with a `GET` fallback and accept self-signed TLS certificates. `UpdateConfig` replaces the config and prunes results for items that no longer exist. After each check round, a signal is sent on the `Notify()` channel so the server can push updates to WebSocket clients.

### HTTP server (`internal/server`)

`Server` wraps the Go standard library `net/http` server. Routes are registered on a `http.ServeMux` using the Go 1.22 method+path pattern syntax. The web UI is served from an embedded filesystem (`//go:embed web`), so the binary is fully self-contained.

A `hub` manages all active WebSocket connections. A `runHub` goroutine subscribes to `checker.Notify()` and broadcasts the latest status snapshot to every connected client after each check round. New clients receive an immediate snapshot on connect so there is no wait for the next tick.

When no `-host` flag is set, `Run()` opens two listeners — `tcp4` on `0.0.0.0` and `tcp6` on `[::]` — so the server accepts both IPv4 and IPv6 connections on all platforms regardless of OS dual-stack settings.

---

## Adding a new API endpoint

1. Add a handler method on `*Server` in `internal/server/handlers.go`.
2. Register the route in `registerRoutes()` in `internal/server/server.go`.
3. Add a test in `internal/server/handlers_test.go`.

---

## Running tests

```bash
# All packages
go test ./...

# With verbose output
go test -v ./...

# A single package
go test ./internal/checker/...

# With race detector
go test -race ./...
```

---

## Building packages

Packages are built with [nfpm](https://nfpm.goreleaser.com/). Install it and run:

```bash
nfpm package --config packaging/nfpm.yaml --packager deb   # Debian / Ubuntu
nfpm package --config packaging/nfpm.yaml --packager rpm   # RHEL / Fedora
nfpm package --config packaging/nfpm.yaml --packager apk   # Alpine
```

Package metadata is defined in `nfpm.yaml`. Pre/post install scripts are in `packaging/`.
