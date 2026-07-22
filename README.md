# NVR

Self-hosted NVR monorepo (see design `arch.md`).

```
nvr/
├── server/              # Go modular monolith + openapi.yaml
├── apps/
│   ├── dashboard/       # SvelteKit SPA (adapter-static → embed in Go)
│   ├── mobile/          # Expo
│   └── landing/         # Astro
├── packages/
│   ├── api-client/      # OpenAPI TS client
│   └── typescript-config/
└── deploy/
```

## Prerequisites

- Bun
- Go 1.22+
- `oapi-codegen` + `sqlc` on `PATH` for server codegen
- FFmpeg / MediaMTX at runtime

## Commands

```bash
bun install
bun run dev:dashboard
bun run dev:landing
bun run dev:mobile
bun run dev:server
bun run gen:api          # TS client from openapi.yaml
bun run lint             # Biome (lint + format check)
bun run format           # Biome format --write
cd server && make generate   # oapi-codegen + sqlc
```

## Libraries (arch.md §16)

**Go:** chi, oapi-codegen, modernc/sqlite, sqlc, golang-migrate, coder/websocket,
argon2id, minio-go, use-go/onvif, robfig/cron, goqite, koanf, gopsutil, slog.

**Dashboard:** Svelte 5, adapter-static, Tailwind v4, bits-ui, TanStack Query,
openapi-fetch client, hls.js, zod + superforms, lucide-svelte.

**Mobile:** expo-router, TanStack Query, zustand, react-native-webrtc, expo-video,
expo-notifications, expo-secure-store.

**Tooling:** Biome (lint + format) — not ESLint/Prettier.
