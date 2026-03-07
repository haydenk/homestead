# API Reference

Homestead exposes a small REST API for reading configuration, querying health check statuses, and triggering live reloads.

All endpoints return `application/json` unless otherwise noted.

---

## Endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/` | Web UI |
| `GET` | `/api/config` | Full configuration as JSON |
| `GET` | `/api/status` | Health check statuses for all items |
| `GET` | `/api/status/{id}` | Health check status for a single item |
| `POST` | `/api/reload` | Reload config from disk and restart checks |
| `GET` | `/api/health` | Liveness probe |

---

## GET /api/config

Returns the currently loaded configuration as JSON. Useful for debugging or building integrations.

**Response**

```json
{
  "title": "Homestead",
  "subtitle": "My self-hosted dashboard",
  "logo": "🏠",
  "theme": "dark",
  "columns": 4,
  "checkInterval": 30,
  "footer": "Running on Proxmox",
  "sections": [
    {
      "name": "Media",
      "icon": "🎬",
      "items": [
        {
          "id": "s0-i0",
          "title": "Jellyfin",
          "url": "https://jellyfin.lan",
          "description": "Open source media server",
          "icon": "📺",
          "tags": ["media"],
          "target": "_blank",
          "statusCheck": true,
          "color": "#00a4dc"
        }
      ]
    }
  ]
}
```

---

## GET /api/status

Returns a map of health check statuses keyed by item ID. Only items with `status_check = true` appear here.

**Response**

```json
{
  "s0-i0": {
    "id": "s0-i0",
    "url": "https://jellyfin.lan",
    "up": true,
    "statusCode": 200,
    "responseTimeMs": 145,
    "lastChecked": "2024-02-24T16:30:00Z",
    "error": ""
  },
  "s0-i1": {
    "id": "s0-i1",
    "url": "https://sonarr.lan",
    "up": false,
    "statusCode": 0,
    "responseTimeMs": 10001,
    "lastChecked": "2024-02-24T16:30:01Z",
    "error": "dial tcp: connection refused"
  }
}
```

### Status object fields

| Field | Type | Description |
|---|---|---|
| `id` | string | Item ID in the format `s{section}-i{item}` (e.g. `s0-i0`) |
| `url` | string | URL that was checked |
| `up` | boolean | `true` if the HTTP status code was less than 500 |
| `statusCode` | integer | HTTP status code returned, or `0` on connection failure |
| `responseTimeMs` | integer | Round-trip time in milliseconds |
| `lastChecked` | string | RFC 3339 timestamp of the last check |
| `error` | string | Error message if the request failed; empty string on success |

---

## GET /api/status/{id}

Returns the health check status for a single item. The `id` corresponds to the `id` field in `/api/config` (e.g. `s0-i0`).

**Response** — same structure as a single entry from `/api/status`.

**404 Not Found** — returned if no item with that ID exists or if the item does not have `status_check = true`.

```json
{"error": "not found"}
```

---

## POST /api/reload

Re-reads the configuration file from disk, applies the new configuration, clears stale health check results, and triggers an immediate re-check of all items.

No request body is required.

**Success** — `204 No Content`

**Failure** — `500 Internal Server Error` with a JSON error body if the config file cannot be parsed.

```bash
curl -X POST http://localhost:8080/api/reload
```

---

## GET /api/health

Simple liveness probe intended for Docker health checks, Kubernetes readiness probes, and uptime monitors.

**Response** — always `200 OK`

```json
{
  "status": "ok",
  "time": "2024-02-24T16:30:00Z"
}
```

**Docker health check example**

```dockerfile
HEALTHCHECK --interval=30s --timeout=5s \
  CMD wget -qO- http://localhost:8080/api/health || exit 1
```

---

## Item IDs

Item IDs are assigned at load time based on the position of the item in the config file:

- Format: `s{section_index}-i{item_index}`
- Indices are zero-based
- Example: the first item in the second section is `s1-i0`

IDs are stable as long as the order of sections and items in the config file does not change. After a `/api/reload`, IDs are recalculated, so any externally cached IDs should be refreshed from `/api/config`.
