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
- **Offsite Archive:** Optional encrypted Google Drive connection with nightly and on-demand uploads.

## Google Drive Archive

Set `NVR_SECRETS_KEY`, `NVR_GOOGLE_CLIENT_ID`, `NVR_GOOGLE_CLIENT_SECRET`, and
`NVR_GOOGLE_REDIRECT_URL` on the server, then connect an account from
**Settings → Google Drive**. The redirect URL must exactly match the authorized
redirect URI configured for the Google OAuth web client.

Vigil uploads pending recordings at 00:00 UTC each day. Archive retries are
idempotent, and retention preserves pending recording metadata while a Drive
connection is stored. See [`deploy/nvr.example.env`](deploy/nvr.example.env) for
the required OAuth scopes and an example callback URL.

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
