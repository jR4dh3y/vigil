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

## Deployment artifacts

Vigil ships two server artifacts with different trust boundaries:

- **Full (default)** — the Go binary embeds the SvelteKit dashboard. One
  process serves the API and the UI on the same origin, so the browser
  authenticates with HttpOnly cookies. This is the recommended default and
  what `make build`, `docker build`, and `docker compose up` produce.
- **Slim / headless** — the binary omits the embedded dashboard and instead
  serves a small connection page that deep-links to a separately hosted
  dashboard with a `?server=` parameter. Build it with `make build-slim` inside
  `server/` or `docker build --target runtime-slim` in `deploy/`. The hosted
  dashboard and the recorder are different origins, so auth uses a bearer
  session token rather than cookies, and the recorder must be reachable over an
  HTTPS tunnel.

See [`docs/operations.md`](docs/operations.md) and
[`docs/dashboard.md`](docs/dashboard.md) for the hosted-dashboard setup.

## Google Drive Archive

Set `NVR_SECRETS_KEY` on the server, then enter the Google OAuth web-client ID,
secret, and redirect URL in **Settings → Google Drive**. The secret is encrypted
at rest, and the redirect URL must exactly match the authorized redirect URI
configured for the Google OAuth web client. `NVR_GOOGLE_CLIENT_ID`,
`NVR_GOOGLE_CLIENT_SECRET`, and `NVR_GOOGLE_REDIRECT_URL` remain available as an
environment-based fallback for unattended deployments.

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
bun run build:bin
./server/bin/nvrd       # Serves API + Dashboard on :8080
```

The build command builds the dashboard, copies it into the Go embed directory,
and produces `server/bin/nvrd`.

Runtime state defaults to `./data` and `./recordings` relative to the directory
where `nvrd` is launched, not the binary's directory. Set `NVR_DATA_DIR` and
`NVR_RECORDINGS_DIR` to absolute paths for a fixed deployment location.

### Docker
```bash
cd deploy && docker compose up --build
# Or, for the headless/hosted-dashboard image (no embedded UI):
docker build --target runtime-slim -f deploy/Dockerfile -t nvr-slim ..
```

## Codegen & Check

```bash
bun run gen:api && bun run gen:server
bun run check
```
