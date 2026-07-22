# syntax=docker/dockerfile:1
FROM oven/bun:1 AS dashboard
WORKDIR /src
COPY package.json bun.lock turbo.json ./
COPY packages ./packages
COPY apps/dashboard ./apps/dashboard
RUN bun install
RUN bun run build --filter=@nvr/dashboard

FROM golang:1.24-bookworm AS server
WORKDIR /src/server
COPY server/ ./
RUN CGO_ENABLED=0 go build -o /out/nvrd ./cmd/nvrd

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates ffmpeg \
  && rm -rf /var/lib/apt/lists/*
COPY --from=server /out/nvrd /usr/local/bin/nvrd
ENV NVR_HTTP_ADDR=:8080 NVR_DATA_DIR=/var/lib/nvr/data NVR_RECORDINGS_DIR=/var/lib/nvr/recordings
VOLUME ["/var/lib/nvr/data", "/var/lib/nvr/recordings"]
EXPOSE 8080
ENTRYPOINT ["nvrd"]
