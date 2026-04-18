# * Build frontend
FROM node:24-alpine AS web

WORKDIR /app

# Enable pnpm via corepack (version pinned in package.json -> packageManager)
RUN corepack enable

# Install dependencies first to leverage layer cache
COPY web/package.json web/pnpm-lock.yaml ./
RUN --mount=type=cache,id=pnpm-store,target=/root/.local/share/pnpm/store \
    pnpm install --frozen-lockfile

# Build static assets
COPY web/ ./
RUN pnpm run build


# * Build Go binary
FROM golang:1.25-bookworm AS backend

ENV CGO_ENABLED=1 \
    GOOS=linux

WORKDIR /app

# Cache module downloads
COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source and build
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-s -w" -o /app/api-proxy ./cmd/server


# * Final image
FROM debian:bookworm-slim AS runtime

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates tzdata curl \
    && rm -rf /var/lib/apt/lists/*

RUN useradd --system --uid 10001 --home /app --shell /usr/sbin/nologin app

WORKDIR /app

# Copy binary and static assets
COPY --from=backend /app/api-proxy /app/api-proxy
COPY --from=web /app/dist /app/web

# Persistent SQLite data
RUN mkdir -p /app/data && chown -R app:app /app
VOLUME ["/app/data"]

USER app

EXPOSE 80 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl --fail --silent --show-error --max-time 3 http://127.0.0.1/api/healthz || exit 1

ENV PROXY_ENDPOINT=http://localhost:80

ENTRYPOINT ["/app/api-proxy"]
CMD ["-b", "80", "-p", "8080", "-d", "/app/web"]
