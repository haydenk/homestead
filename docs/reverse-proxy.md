# Reverse Proxy

Running Homestead behind a reverse proxy allows you to expose the dashboard over HTTPS, use a custom domain, and avoid exposing the raw port directly. Homestead is a plain HTTP server, so no special upstream configuration is required.

By default, Homestead binds to `127.0.0.1:8080`. The examples below assume this default; adjust the address and port to match your runtime options.

---

## nginx

### Basic HTTP proxy

```nginx
server {
    listen 80;
    server_name homestead.example.com;

    location / {
        proxy_pass         http://127.0.0.1:8080;
        proxy_set_header   Host              $host;
        proxy_set_header   X-Real-IP         $remote_addr;
        proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
    }
}
```

### HTTPS with Let's Encrypt (Certbot)

Install Certbot and obtain a certificate, then use the following configuration:

```nginx
server {
    listen 443 ssl;
    server_name homestead.example.com;

    ssl_certificate     /etc/letsencrypt/live/homestead.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/homestead.example.com/privkey.pem;
    include             /etc/letsencrypt/options-ssl-nginx.conf;
    ssl_dhparam         /etc/letsencrypt/ssl-dhparams.pem;

    location / {
        proxy_pass         http://127.0.0.1:8080;
        proxy_set_header   Host              $host;
        proxy_set_header   X-Real-IP         $remote_addr;
        proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
    }
}

server {
    listen 80;
    server_name homestead.example.com;
    return 301 https://$host$request_uri;
}
```

Run Certbot to issue and configure the certificate automatically:

```bash
certbot --nginx -d homestead.example.com
```

### Reload nginx

```bash
nginx -t && systemctl reload nginx
```

---

## Caddy

Caddy automatically provisions and renews TLS certificates via Let's Encrypt with no additional configuration.

### Caddyfile

```caddy
homestead.example.com {
    reverse_proxy 127.0.0.1:8080
}
```

Place this in your `/etc/caddy/Caddyfile` (or include it from a separate file), then reload:

```bash
caddy validate --config /etc/caddy/Caddyfile
systemctl reload caddy
```

Caddy will handle HTTPS automatically, including HTTP-to-HTTPS redirection.

### Local domain (no public DNS)

For a homelab domain that is not publicly resolvable, use a DNS challenge provider so Caddy can still obtain a certificate. The following example uses Cloudflare:

```caddy
homestead.home.arpa {
    tls {
        dns cloudflare {env.CF_API_TOKEN}
    }
    reverse_proxy 127.0.0.1:8080
}
```

Set the `CF_API_TOKEN` environment variable in `/etc/caddy/caddy.env` (or your systemd override) and install the Cloudflare DNS module for Caddy (`xcaddy build --with github.com/caddy-dns/cloudflare`).

---

## Notes

- Homestead does not use WebSockets, so no special upgrade headers are needed.
- The health check poller runs server-side, so the proxy does not need to keep long-lived connections open.
- If Homestead is running inside Docker, replace `127.0.0.1:8080` with the container's address (e.g. `homestead:8080` when using a shared Docker network, or the host address when using host networking).
