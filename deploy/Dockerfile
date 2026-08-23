# syntax=docker/dockerfile:1

# --- Dashboard (SvelteKit static SPA) ---
# Static assets are architecture-independent, so this stage runs on the
# builder's native platform instead of emulating the target arch.
FROM --platform=$BUILDPLATFORM oven/bun:1 AS dashboard
WORKDIR /src
COPY package.json bun.lock turbo.json ./
COPY packages ./packages
COPY apps/dashboard ./apps/dashboard
RUN bun install --frozen-lockfile
RUN bun run build --filter=@nvr/dashboard

# --- Go build stage (full, embeds the dashboard) ---
# CGO_ENABLED=0 lets us cross-compile natively for the target arch via
# TARGETARCH instead of running the whole toolchain under QEMU.
FROM --platform=$BUILDPLATFORM golang:1.26-bookworm AS server
ARG TARGETARCH
WORKDIR /src/server
COPY server/ ./
# Embed built SPA so one process serves API + UI.
COPY --from=dashboard /src/apps/dashboard/build/ ./internal/ui/dist/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -ldflags="-s -w" -o /out/nvrd ./cmd/nvrd

# --- Go build stage (slim/headless, no dashboard) ---
# Does not depend on the dashboard build.
FROM --platform=$BUILDPLATFORM golang:1.26-bookworm AS server-slim
ARG TARGETARCH
WORKDIR /src/server
COPY server/ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -tags slim -ldflags="-s -w" -o /out/nvrd ./cmd/nvrd

# --- Runtime (slim/headless) ---
# Build with: docker build --target runtime-slim
FROM debian:bookworm-slim AS runtime-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates ffmpeg curl \
  && rm -rf /var/lib/apt/lists/*
COPY --from=server-slim /out/nvrd /usr/local/bin/nvrd
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

# --- Runtime (default: full embedded UI) ---
# The final stage is the default build target: bare, compose, and default
# docker build all produce the full embedded-dashboard image.
FROM debian:bookworm-slim AS runtime
RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates ffmpeg curl \
  && rm -rf /var/lib/apt/lists/*
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