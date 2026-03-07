# Configuration

Homestead is configured through a single TOML file. The repository includes `config.example.toml` as a starting point — copy it to `config.toml` (which is gitignored) and edit it to suit your setup:

```bash
cp config.example.toml config.toml
```

By default the binary looks for `config.toml` in the working directory, but any path can be provided via the `-config` flag or `HOMESTEAD_CONFIG` environment variable.

---

## Global settings

These keys appear at the top level of the config file.

| Key | Type | Default | Description |
|---|---|---|---|
| `title` | string | `"Homestead"` | Page title shown in the browser tab and header |
| `subtitle` | string | `""` | Secondary heading displayed beneath the title |
| `logo` | string | `"🏠"` | Logo shown in the header — emoji or an image URL |
| `theme` | string | `"dark"` | UI theme: `"dark"`, `"light"`, or omit to follow the system preference |
| `columns` | integer | `4` | Number of cards per row (1–6) |
| `check_interval` | integer | `30` | Seconds between background health checks |
| `footer` | string | `""` | Text displayed in the page footer |

```toml
title          = "Homestead"
subtitle       = "My self-hosted dashboard"
logo           = "🏠"
theme          = "dark"
columns        = 4
check_interval = 30
footer         = "Running on Proxmox"
```

---

## Sections

Sections group related service cards. Each section appears as a labeled group on the dashboard. A config file may contain any number of sections using TOML array-of-tables syntax.

| Key | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Label displayed above the group |
| `icon` | string | no | Emoji or image URL shown next to the section name |
| `items` | array | no | Service cards belonging to this section |

```toml
[[sections]]
name = "Media"
icon = "🎬"
```

---

## Items

Items are the individual service cards inside a section. Each item represents one link.

| Key | Type | Required | Default | Description |
|---|---|---|---|---|
| `title` | string | yes | — | Card heading |
| `url` | string | yes | — | URL the card links to |
| `description` | string | no | `""` | Short text shown beneath the title |
| `icon` | string | no | `""` | Emoji or image URL displayed on the card |
| `tags` | array of strings | no | `[]` | Labels used by the search filter |
| `status_check` | boolean | no | `false` | Enable background health monitoring for this URL |
| `color` | string | no | `""` | Hex accent color applied to the card border (e.g. `"#00a4dc"`) |
| `target` | string | no | `"_blank"` | Link target: `"_blank"` (new tab) or `"_self"` (same tab) |

```toml
[[sections.items]]
title        = "Jellyfin"
url          = "https://jellyfin.lan"
description  = "Open source media server"
icon         = "📺"
tags         = ["media", "streaming"]
status_check = true
color        = "#00a4dc"
target       = "_blank"
```

---

## Health check behavior

When `status_check = true` is set on an item, Homestead performs background HTTP checks against the item's URL on the configured `check_interval`.

- A `HEAD` request is sent first.
- If the server responds with `405 Method Not Allowed`, a `GET` request is made as a fallback.
- Any HTTP response with a status code below 500 is treated as **up** — this includes auth-protected pages (401/403).
- Self-signed TLS certificates are accepted.
- The request timeout is 10 seconds.
- Up to 3 redirects are followed automatically.

---

## Complete example

```toml
title          = "Homestead"
subtitle       = "My self-hosted dashboard"
logo           = "🏠"
theme          = "dark"
columns        = 4
check_interval = 30
footer         = "Running on Proxmox"


# ── Networking ──────────────────────────────────────────────────
[[sections]]
name = "Networking"
icon = "🌐"

  [[sections.items]]
  title        = "Router"
  url          = "http://router.lan"
  description  = "Home router admin panel"
  icon         = "🔀"
  tags         = ["networking"]
  status_check = true

  [[sections.items]]
  title        = "Pi-hole"
  url          = "http://pi.hole/admin"
  description  = "DNS ad blocker"
  icon         = "🕳️"
  tags         = ["networking", "dns"]
  status_check = true


# ── Media ────────────────────────────────────────────────────────
[[sections]]
name = "Media"
icon = "🎬"

  [[sections.items]]
  title        = "Jellyfin"
  url          = "http://jellyfin.local:8096"
  description  = "Open source media server"
  icon         = "📺"
  tags         = ["media"]
  status_check = true
  color        = "#00a4dc"

  [[sections.items]]
  title        = "Sonarr"
  url          = "http://sonarr.local:8989"
  description  = "TV series management"
  icon         = "📡"
  tags         = ["media", "arr"]
  status_check = true


# ── Infrastructure ───────────────────────────────────────────────
[[sections]]
name = "Infrastructure"
icon = "🖥️"

  [[sections.items]]
  title        = "Proxmox"
  url          = "https://proxmox.vm:8006"
  description  = "Virtualisation platform"
  icon         = "🔵"
  tags         = ["infra", "vms"]
  status_check = true
  color        = "#e57000"

  [[sections.items]]
  title        = "Portainer"
  url          = "http://portainer.local:9000"
  description  = "Docker container manager"
  icon         = "🐳"
  tags         = ["infra", "docker"]
  status_check = true
```

---

## Using image URLs as icons

The `icon` field on both sections and items accepts an image URL in addition to an emoji. This is useful when a service provides its own logo.

```toml
[[sections.items]]
title = "Grafana"
url   = "http://grafana.local:3000"
icon  = "https://grafana.local/public/img/grafana_icon.svg"
```

The image is displayed at a fixed size inside the card.

---

## Live reload

You can reload the configuration from disk without restarting the process:

```bash
curl -X POST http://localhost:8080/api/reload
```

A `204 No Content` response indicates success. The updated config takes effect immediately and health checks restart automatically.
