# NVR

Self-hosted network video recorder monorepo (design: `arch.md`).

```
nvr/
├── server/              # Go modular monolith (nvrd) + openapi.yaml
├── apps/
│   ├── dashboard/       # SvelteKit SPA (adapter-static → embed in Go)
│   ├── mobile/          # Expo (not required for core)
│   └── landing/         # Astro marketing (not required for core)
├── packages/
│   └── api-client/      # OpenAPI TypeScript client
└── deploy/              # Docker, MediaMTX config, example env
```

## What works

| Area | Features |
|------|----------|
| Auth | First-boot setup, login/logout, session cookies, roles |
| Cameras | CRUD, RTSP probe, enable/disable → MediaMTX paths |
| Live | HLS (+ WHEP fallback), short-lived stream tokens |
| Recordings | Segment index hook, timeline query, playback tokens |
| Events | Online/offline + disk alerts, acknowledge, list |
| Jobs | Health probe, retention prune, disk monitor |
| System | Disk usage, status, site settings |
| Users | Admin user management |
| UI | Live grid, cameras, timeline, events, settings — embedded in `nvrd` |

## Prerequisites

- Bun
- Go 1.22+
- FFmpeg / ffprobe
- MediaMTX (see `deploy/mediamtx.yml` or `docker compose`)
- `oapi-codegen` + `sqlc` on `PATH` for codegen

## Dev

```bash
# Terminal 1 — MediaMTX
./.bin/mediamtx deploy/mediamtx.yml   # or download bluenviron/mediamtx

# Terminal 2 — API
export NVR_DATA_DIR=./server/data
export NVR_RECORDINGS_DIR=./recordings
export NVR_MEDIAMTX_API_URL=http://127.0.0.1:9997
export NVR_MEDIAMTX_WEBRTC_URL=http://127.0.0.1:8889
export NVR_MEDIAMTX_HLS_URL=http://127.0.0.1:8888
bun run dev:server
# or: cd server && go run ./cmd/nvrd

# Terminal 3 — Dashboard (hot reload; proxies /api → :8080)
bun run dev:dashboard
```

Open http://localhost:5173

## Production-style single binary

```bash
bun run build --filter=@nvr/dashboard
rm -rf server/internal/ui/dist && mkdir -p server/internal/ui/dist
cp -a apps/dashboard/build/. server/internal/ui/dist/
cd server && go build -o bin/nvrd ./cmd/nvrd
./bin/nvrd   # serves API + SPA on :8080
```

## Docker

```bash
cd deploy && docker compose up --build
```

Compose runs MediaMTX + `nvrd` with host networking for WebRTC.

## Codegen

```bash
bun run gen:api          # TS client from openapi.yaml
bun run gen:server       # oapi-codegen + sqlc
bun run lint && bun run format && bun run check
```

## Env

See `deploy/nvr.example.env`.

## Out of scope (later)

- Expo mobile + push
- Astro landing/docs
- S3 archive tier
- ONVIF / vendor drivers
- Full event rules + webhooks
