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

# Run locally
go run . -config config.toml -host 0.0.0.0

# Build a binary
go build -o homestead .

# Run tests
go test ./...

# Format
go fmt ./...
```

The server binds to `http://0.0.0.0:8080` by default when using the command above. Adjust with `-host` and `-port` as needed.

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
│       ├── server.go            # HTTP server setup and route registration
│       ├── handlers.go          # API handler implementations
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

## Architecture

### Config loading (`internal/config`)

`config.Load` reads the TOML file using `BurntSushi/toml`, applies defaults for any missing fields, and assigns stable IDs to each item (`s{section}-i{item}`). The resulting `*Config` value is immutable during normal operation and replaced atomically on live reload.

### Health checker (`internal/checker`)

`Checker` runs a goroutine-per-item health check on a configurable interval. Results are stored in a `map[string]*Status` guarded by a `sync.RWMutex`. Checks use `HEAD` with a `GET` fallback and accept self-signed TLS certificates. `UpdateConfig` replaces the config and prunes results for items that no longer exist.

### HTTP server (`internal/server`)

`Server` wraps the Go standard library `net/http` server. Routes are registered on a `http.ServeMux` using the Go 1.22 method+path pattern syntax (e.g. `GET /api/config`). The web UI is served from an embedded filesystem (`//go:embed web`), so the binary is fully self-contained.

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
