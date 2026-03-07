# Installation

## Requirements

- Go 1.22 or later (for building from source)
- Docker (for container deployment)
- [nfpm](https://nfpm.goreleaser.com/) (for building system packages)

---

## From source

```bash
git clone https://github.com/haydenk/homestead.git
cd homestead
go build -o homestead .
cp config.example.toml config.toml   # create your local config
$EDITOR config.toml
./homestead -config config.toml
```

The resulting `homestead` binary has no runtime dependencies and embeds the web UI.

---

## Docker

### Build the image

```bash
docker build -t homestead .
```

The Dockerfile uses a multi-stage build (`golang:1.22-alpine` -> `alpine:3.19`) to produce a minimal image.

### Run the container

```bash
docker run -d \
  --name homestead \
  -p 8080:8080 \
  -v $(pwd)/config.toml:/app/config/config.toml \
  homestead
```

The container reads its configuration from `/app/config/config.toml`. Mount your config file at that path.

### Environment variables in Docker

```bash
docker run -d \
  --name homestead \
  -p 9000:9000 \
  -e HOMESTEAD_PORT=9000 \
  -e HOMESTEAD_CONFIG=/app/config/config.toml \
  -v $(pwd)/config.toml:/app/config/config.toml \
  homestead
```

---

## System package (deb / rpm / apk)

Build packages with [nfpm](https://nfpm.goreleaser.com/):

```bash
nfpm package --config packaging/nfpm.yaml
```

### What the package installs

| Path | Description |
|---|---|
| `/usr/bin/homestead` | Binary |
| `/etc/homestead/config.toml` | Default configuration |

### Systemd service

The package registers and enables a hardened systemd service that runs as a non-root `homestead` user.

```bash
# Enable and start
systemctl enable --now homestead

# Check status
systemctl status homestead

# View logs
journalctl -u homestead -f

# Restart after config change
systemctl restart homestead
# or use the live reload API (no restart required):
curl -X POST http://localhost:8080/api/reload
```

The unit file is installed at `/lib/systemd/system/homestead.service`.

---

## Runtime options

Configuration can be overridden at startup via command-line flags or environment variables. Environment variables take precedence over flags.

| Flag | Environment variable | Default | Description |
|---|---|---|---|
| `-config` | `HOMESTEAD_CONFIG` | `config.toml` | Path to the TOML config file |
| `-host` | `HOMESTEAD_HOST` | *(all interfaces)* | Bind address |
| `-port` | `HOMESTEAD_PORT` | `8080` | Listen port |

By default (no `-host` set), Homestead listens on all IPv4 and IPv6 interfaces simultaneously. The legacy `PORT` environment variable is also accepted for compatibility with container platforms, but `HOMESTEAD_PORT` takes precedence if both are set.

```bash
# Default: listens on 0.0.0.0 and [::] (all interfaces, IPv4 + IPv6)
./homestead -config /etc/homestead/config.toml

# Restrict to loopback only
./homestead -host 127.0.0.1 -port 8080

# Custom port via environment variable
HOMESTEAD_PORT=9000 ./homestead
```
