# Vigil /ˈvɪdʒəl/

Self-hosted Network Video Recorder (NVR) with live viewing, timeline playback, and event alerts.

```
vigil/
├── server/          # Go monolith (nvrd) + OpenAPI
├── apps/
│   ├── dashboard/   # SvelteKit SPA (embedded in Go)
│   └── mobile/      # Expo mobile app
├── packages/
│   └── api-client/  # OpenAPI TS client
└── deploy/          # Docker & MediaMTX configs
```

## Features

- **Live & Playback:** Low-latency HLS/WHEP streams, timeline segment playback.
- **Management:** RTSP camera CRUD, setup wizard, role-based auth, user control.
- **Alerts & System:** Camera state/disk monitoring, retention pruning, event logs.

## Quick Start

### Dev Mode
```bash
bun run dev          # Starts MediaMTX, Go API (:8080), and Dashboard (:5173)
```

### Production Binary
```bash
bun run build --filter=@nvr/dashboard
cp -a apps/dashboard/build/. server/internal/ui/dist/
cd server && go build -o bin/nvrd ./cmd/nvrd
./bin/nvrd           # Serves API + Dashboard on :8080
```

### Docker
```bash
cd deploy && docker compose up --build
```

## Codegen & Check

```bash
bun run gen:api && bun run gen:server
bun run check
```
