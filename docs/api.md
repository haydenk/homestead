# API Reference

Homestead exposes a small API for querying health check statuses in real-time and triggering live reloads.

---

## Endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/` | Web UI |
| `WS` | `/ws` | Real-time status updates (WebSocket) |
| `POST` | `/api/reload` | Reload config from disk and restart checks |
| `GET` | `/api/health` | Liveness probe |

---

## WS /ws

Upgrades to a WebSocket connection and streams health check status updates in real time.

- On connect, the server immediately sends the current snapshot of all statuses.
- After each check round completes, the server pushes an updated snapshot to all connected clients.
- The client does not need to send any messages; the connection is server-to-client only.

**Message format** — JSON object keyed by item ID, same structure as the status object below.

```json
{
  "s0-i0": {
    "id": "s0-i0",
    "url": "https://jellyfin.lan",
    "up": true,
    "statusCode": 200,
    "responseTimeMs": 145,
    "lastChecked": "2024-02-24T16:30:00Z"
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
| `error` | string | Error message if the request failed; omitted on success |

### Example (browser)

```js
const proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
const ws = new WebSocket(`${proto}//${location.host}/ws`);

ws.onmessage = (e) => {
  const statuses = JSON.parse(e.data);
  console.log(statuses);
};

ws.onclose = () => setTimeout(connect, 3000); // auto-reconnect
```

---

## POST /api/reload

Re-reads the configuration file from disk, applies the new configuration, clears stale health check results, and triggers an immediate re-check of all items. Connected WebSocket clients receive the updated statuses automatically.

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

IDs are stable as long as the order of sections and items in the config file does not change. After a `/api/reload`, IDs are recalculated and all connected WebSocket clients receive a fresh snapshot.
