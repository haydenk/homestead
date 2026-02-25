# ── Build stage ───────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /build

# Cache dependency download separately from source
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w -X main.version=$(git describe --tags --always 2>/dev/null || echo dev)" \
    -o homestead .

# ── Final stage ───────────────────────────────────────────────────────────────
FROM alpine:3.19

# ca-certificates: needed for HTTPS health checks
# tzdata: correct timestamps in logs
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -S homestead && \
    adduser  -S homestead -G homestead

WORKDIR /app

COPY --from=builder /build/homestead .

# Ship a default config; users mount their own at /app/config/config.toml
COPY config.toml ./config/config.toml

RUN chown -R homestead:homestead /app

USER homestead

EXPOSE 8080

# /app/config is the volume for a custom config.toml
VOLUME ["/app/config"]

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/api/health || exit 1

ENTRYPOINT ["./homestead"]
CMD ["--config", "/app/config/config.toml"]
