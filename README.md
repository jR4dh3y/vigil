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

One command starts the full stack via Turbo (MediaMTX, Go API, dashboard, landing, mobile):

```bash
# Needs Go, FFmpeg, and MediaMTX at .bin/mediamtx or on PATH
bun run dev
```

| Process | Filter / script | Notes |
|---------|-----------------|--------|
| MediaMTX | `bun run dev:mediamtx` | Uses `deploy/mediamtx.yml` |
| Go API (`nvrd`) | `bun run dev:server` | `:8080`, data in `server/data`, recordings in `./recordings` |
| Dashboard | `bun run dev:dashboard` | Vite on `:5173`, proxies `/api` → Go |
| Landing | `bun run dev:landing` | Astro |
| Mobile | `bun run dev:mobile` | Expo |

Open the dashboard at http://localhost:5173

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
