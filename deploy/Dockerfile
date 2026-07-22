# syntax=docker/dockerfile:1

# --- Dashboard (SvelteKit static SPA) ---
FROM oven/bun:1 AS dashboard
WORKDIR /src
COPY package.json bun.lock turbo.json ./
COPY packages ./packages
COPY apps/dashboard ./apps/dashboard
RUN bun install --frozen-lockfile || bun install
RUN bun run build --filter=@nvr/dashboard

# --- Go server (embed dashboard into binary) ---
FROM golang:1.24-bookworm AS server
WORKDIR /src/server
COPY server/ ./
# Embed built SPA so one process serves API + UI.
COPY --from=dashboard /src/apps/dashboard/build/ ./internal/ui/dist/
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/nvrd ./cmd/nvrd

# --- Runtime ---
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates ffmpeg curl \
  && rm -rf /var/lib/apt/lists/*
# Optional: copy mediamtx if you bake it into the image (prefer sidecar compose).
COPY --from=server /out/nvrd /usr/local/bin/nvrd
ENV NVR_HTTP_ADDR=:8080 \
    NVR_DATA_DIR=/var/lib/nvr/data \
    NVR_RECORDINGS_DIR=/var/lib/nvr/recordings \
    NVR_RETENTION_DAYS=7 \
    NVR_MEDIAMTX_API_URL=http://127.0.0.1:9997 \
    NVR_MEDIAMTX_WEBRTC_URL=http://127.0.0.1:8889 \
    NVR_MEDIAMTX_HLS_URL=http://127.0.0.1:8888
VOLUME ["/var/lib/nvr/data", "/var/lib/nvr/recordings"]
EXPOSE 8080
ENTRYPOINT ["nvrd"]
