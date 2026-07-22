
## 1. Review of your stack — verdict and the things to watch

The stack is good. A Go modular monolith orchestrating MediaMTX/FFmpeg, with SQLite for metadata and an S3-compatible archive tier, is basically the same shape as Frigate and Viseron, and it's the right shape for self-hosting. A few honest flags before the design:

- **The hardest problems here are not web problems.** They're media-plane problems: codec compatibility (H.265 cameras vs. browsers), WebRTC through NAT/Docker, disk-full behavior, and clock drift on timelines. Budget most of your design care there, not in the CRUD.
- **Don't split the media role between FFmpeg and MediaMTX ad hoc.** Pick one owner per job or you'll end up with two half-implementations. My recommendation below: MediaMTX owns *ingest, live output, and recording*; FFmpeg is a *utility* invoked by Go for thumbnails, clip export, and optional transcode. This is the single most important decision in the project.
- **SQLite is fine** for this workload (a segment-index insert per camera per minute is nothing), but you must treat it correctly: WAL mode, one writer, `busy_timeout`. Design for it from day one.
- **Google Drive as an archive target** - would be fine as a cron job that runs periodically for eg, send the whole of recorded media to Google Drive at midnight every day.
- **Three frontends is a lot for one person.** The architecture below deliberately makes web and mobile thin so they share one OpenAPI-generated client. Ship the dashboard first; mobile is phase 4-ish.

---

## 2. System architecture overview

One deployable unit: the Go binary is the brain, MediaMTX is the media muscle running alongside it (supervised child process or sidecar container), FFmpeg is a tool the Go binary shells out to.

```
                ┌─────────────────────────── Self-hosted box ───────────────────────────┐
                │                                                                       │
 IP Cameras ────┼─RTSP──► ┌──────────┐  WebRTC(WHEP)/HLS ────────────► Browser / Mobile │
 (any vendor)   │         │ MediaMTX │──records fMP4 segments──► [Recordings disk]      │
                │         └────┬─────┘                                 │                │
                │    REST ctrl │ ▲ auth-hook                           │                │
                │              ▼ │                                     ▼                │
                │         ┌──────────────────────────────────────────────────┐          │
                │         │                Go backend (monolith)             │          │
                │         │  REST + WebSocket API │ auth │ camera mgmt       │          │
                │         │  recording index │ events │ jobs │ notifications │          │
                │         │  serves built SvelteKit dashboard as static files│          │
                │         └───────┬──────────────┬──────────────┬────────────┘          │
                │                 │              │              │                       │
                │            [SQLite]      exec [FFmpeg]   S3 API ──► FBS-Core /        │
                │                          (thumbs, clips,           any S3 archive     │
                │                           transcode)                                  │
                └───────────────────────────────────────────────────────────────────────┘

  Astro landing/docs = separate static site, publicly hosted, not part of the install.
```

Key stance: **clients never talk to MediaMTX with standing credentials.** The Go API issues short-lived stream tokens; MediaMTX validates every stream request by calling back into the Go API (its `externalAuthentication` hook). Backend stays the sole source of truth for authz.

---

## 3. The media plane (the load-bearing decision)

**MediaMTX owns:**
- **Ingest** — one MediaMTX "path" per enabled camera stream, configured by Go via MediaMTX's REST API (`on-demand` pull from the camera's RTSP URL). MediaMTX handles reconnects.
- **Live output** — WebRTC via WHEP (sub-second latency, primary for web and mobile) with HLS as automatic fallback (works everywhere, ~3–6 s latency).
- **Recording** — MediaMTX's native segmented recording (fMP4, e.g. 60 s segments) to the recordings disk. Its `runOnRecordSegmentComplete` hook POSTs to the Go API, which indexes the segment in SQLite. No transcoding on the hot path — stream copy only.
- **Playback** — MediaMTX's built-in playback server serves time-ranged VOD from recorded segments; Go fronts it (authz + token), the dashboard timeline requests ranges from it.

**FFmpeg (invoked by Go) owns:** thumbnail/poster generation per segment or event, clip export (concat + trim to a single MP4 for download/sharing), optional transcode-on-archive (e.g. re-encode to lower bitrate before upload), and probing (`ffprobe`) camera streams during setup to detect codec/resolution.

**The H.265 caveat you must design for:** many cameras default to H.265, and browser support (especially over WebRTC) is inconsistent. Handle it in setup UX: probe the camera, and if it's H.265, either (a) instruct the user to switch the camera's substream to H.264, (b) use the camera's substream for live view and record the H.265 main stream, or (c) offer an opt-in FFmpeg transcode path with a visible CPU-cost warning. Option (b) — **substream for live, mainstream for record** — is the standard NVR trick and should be your default model: every camera has up to two stream profiles.

---

## 4. Repository structure

One monorepo:

```
nvr/
├── server/                  # Go backend (its own go.mod)
│   ├── cmd/nvrd/            # main: wiring, config load, MediaMTX supervision
│   ├── internal/
│   │   ├── api/             # HTTP handlers, WS hub, middleware (thin layer)
│   │   ├── auth/            # sessions, tokens, RBAC
│   │   ├── camera/          # camera domain: drivers, probing, health
│   │   ├── media/           # MediaMTX client, stream tokens, FFmpeg runner
│   │   ├── recording/       # segment index, retention, timeline queries, clip export
│   │   ├── event/           # event ingestion, rules, in-proc bus
│   │   ├── notify/          # notifier interface + webhook/push impls
│   │   ├── storage/         # StorageProvider interface + s3/local impls, archive logic
│   │   ├── jobs/            # SQLite-backed job queue + workers + cron
│   │   ├── store/           # sqlc-generated queries, migrations
│   │   └── config/
│   ├── migrations/          # numbered .sql files
│   └── api/openapi.yaml     # THE contract — single source of truth
├── apps/
│   ├── dashboard/           # SvelteKit (adapter-static, embedded into Go binary)
│   ├── mobile/              # Expo
│   └── landing/             # Astro + Starlight docs
├── packages/
│   └── api-client/          # TS client generated from openapi.yaml (shared web+mobile)
└── deploy/                  # docker-compose.yml, Dockerfile, example configs
```

The OpenAPI spec is the hinge of the whole repo: Go handlers are generated/validated against it (oapi-codegen), and the TS client both frontends use is generated from it (openapi-typescript). Contract drift becomes a build error instead of a runtime bug.

---

## 5. Backend module boundaries

Each `internal/` package above is a module with a narrow public surface; dependencies flow one way (api → domain packages → store), never sideways between domains — cross-domain reactions go through the event bus. The ones worth spelling out:

- **camera** — CRUD, credentials (encrypted at rest), the `Driver` interface (see §13), stream-profile management, periodic health checks. Emits `camera.online/offline`.
- **media** — the only package that knows MediaMTX exists. Translates "camera enabled" into MediaMTX path config, answers MediaMTX's auth callbacks, mints stream tokens, wraps FFmpeg invocations with timeouts and zombie cleanup.
- **recording** — owns the segment index. Answers timeline queries ("what exists between t1–t2 for camera X"), enforces retention, builds export clips. Never touches MediaMTX directly.
- **event** — small in-process pub/sub (Go channels, fan-out). Everything interesting (motion, camera offline, disk low, archive done) is an event; WS pushes, notifications, and future AI all hang off this bus. This is your main internal decoupling mechanism.
- **jobs** — SQLite-backed durable queue (survives restarts) + cron scheduler. All slow/failable work goes here, never in request handlers.

---

## 6. Frontend responsibilities

- **Astro (landing + docs):** fully static, marketing + install guide + user docs (Starlight). Zero coupling to the app; deployed to any static host.
- **SvelteKit (dashboard):** built with `adapter-static` and **embedded in the Go binary** (`go:embed`) — the self-hoster runs one process and gets the UI. Pure SPA client of the REST/WS API: live grid (WHEP player w/ HLS fallback), timeline playback, camera setup wizard (probe → pick profiles → test), events feed, settings, user admin. No business logic; it renders API state.
- **Expo (mobile):** the "check on my house" subset — live view, event notifications (deep-link into playback at the event timestamp), recent events, arm/disarm-style toggles. Push via Expo's push service. Explicitly not an admin surface (no user management, no storage config) — keeps scope survivable.

---

## 7. Data flows

**Live view:** client asks `POST /cameras/{id}/live` → API checks authz, returns WHEP URL + short-lived (~60 s, one-time) stream token → player connects to MediaMTX → MediaMTX calls Go's auth hook to validate the token → stream flows peer-to-peer-ish over WebRTC. HLS fallback follows the same token dance.

**Recording:** camera enabled → media module creates MediaMTX path with record on → MediaMTX writes 60 s fMP4 segments → segment-complete hook → Go inserts a row (camera, path, start, duration, size) and queues a thumbnail job. Recording is **continuous by default** with retention doing the forgetting; event-only recording is a later optimization (motion-gated retention: keep event-adjacent segments longer than the base window).

**Timeline playback:** dashboard queries `GET /cameras/{id}/recordings?from&to` → renders coverage bars + event markers from SQLite (no disk touch) → user seeks → API mints a playback token → MediaMTX playback server streams the requested range.

**Archive:** rule matches (e.g. "event-flagged segments older than 3 days") → archive job uploads segment + thumbnail to FBS-Core via S3 API → row updated with archive location → local file becomes eligible for pruning. Playback of archived footage = presigned-URL redirect or a restore-to-local job, depending on how hot you want it.

---

## 8. Database schema (high level)

- `users` (role: admin/operator/viewer), `sessions`, `api_keys`
- `cameras` (name, driver, host, encrypted creds, enabled), `camera_permissions` (user × camera, for viewer scoping)
- `stream_profiles` (camera_id, role: live|record, rtsp_url, codec, resolution — the substream/mainstream model)
- `recordings` (camera_id, started_at, duration, path, size, codec, archived_at, archive_location, thumbnail) — the big table; index on `(camera_id, started_at)`
- `events` (camera_id nullable, type, severity, started/ended, metadata JSON, thumbnail, acknowledged)
- `event_rules` (match conditions → actions: notify/flag/webhook)
- `notification_channels` + `push_tokens` (Expo tokens per user/device)
- `storage_providers` (type, encrypted config), `archive_rules`, `jobs` (queue), `settings`, `audit_log`

SQLite specifics: WAL mode, `busy_timeout=5000`, foreign keys on, one connection for writes / pool for reads, litestream-friendly layout (DB on its own path, not on the recordings volume).

---

## 9. API boundaries

REST (all under `/api/v1`, defined in OpenAPI): `auth/*`, `users/*`, `cameras/*` (+ `/probe`, `/live`, `/snapshot`), `recordings/*` (+ `/export`), `events/*`, `storage/*`, `notifications/*`, `system/*` (health, disk, version). Plus two **internal-only** endpoints bound to localhost for MediaMTX hooks: stream auth callback and segment-complete.

One WebSocket at `/api/v1/ws`, server-push only (commands go over REST): event stream, camera status changes, job progress, disk warnings. Keeps the WS protocol trivial.

---

## 10. Background jobs

All on the SQLite-backed queue with retry/backoff, run by an in-process worker pool: **retention pruner** (delete expired segments, DB rows, then empty dirs — and a panic mode that prunes oldest-first when disk crosses a threshold, because disk-full is the #1 NVR failure), **archive uploader**, **thumbnail generator**, **camera health prober** (drives online/offline events), **clip exporter**, **disk monitor**, **event-rule evaluator**, and **push-token pruner**. Cron layer (robfig/cron) enqueues the periodic ones.

---

## 11. Storage architecture

Three tiers: **SQLite** (metadata, small, on fast disk, backed up via Litestream — optionally *to FBS-Core*, which is a nice dogfood), **recordings volume** (large, layout `recordings/{camera_id}/{yyyy-mm-dd}/{ts}.mp4` — filesystem layout mirrors the DB index so a corrupted DB can be re-indexed by scanning disk; build that reconciliation command early, it's your disaster recovery), and **archive tier** behind a `StorageProvider` interface (S3-compatible first — FBS-Core, MinIO, B2, R2 all free — local/rclone/Drive later).

---

## 12. AuthN / AuthZ

- **Web:** username/password (Argon2id) → opaque session token in an HttpOnly cookie, sessions in SQLite (revocable, no JWT statefulness problems).
- **Mobile:** same login → long-lived refresh token in SecureStore + short-lived access token.
- **Streams:** separate short-lived signed tokens per stream request, validated via MediaMTX's auth callback — so leaking a stream URL leaks nothing durable.
- **Integrations:** hashed API keys with scopes.
- **Authz:** three roles (admin / operator / viewer) + per-camera grants for viewers. Do not build full RBAC-with-custom-roles; this covers homes and small businesses.
- First-boot: no users → setup wizard creates admin. Everything sensitive (camera creds, storage creds) encrypted at rest with a key from config/env.

---

## 13. Events, notifications, and extensibility

**Event pipeline:** producers (camera drivers via ONVIF events, health prober, disk monitor, jobs, future detectors) → in-proc bus → persist to `events` → rule evaluation → sinks (WS push, notifiers). Debounce/coalesce at ingestion (motion events chatter badly — merge events < N seconds apart).

**Notifiers** behind one interface: Expo push and generic webhook (which also gets you Home Assistant / ntfy / Discord integration nearly for free).

**Extension points — four Go interfaces, compiled-in registry (no dynamic plugins; keep it a monolith):**
1. `camera.Driver` — `Probe`, `StreamProfiles`, `Snapshot`, `SubscribeEvents`, `Capabilities`. Ship `generic-rtsp` and `onvif` (ONVIF alone covers most vendors incl. motion events + PTZ); vendor-specific drivers (Hikvision/Dahua quirks) are additive files later.
2. `storage.Provider` — `Put/Get/Delete/List/PresignURL`. Ship `s3` and `local`.
3. `notify.Notifier` — ship push/webhook.
4. `detect.Detector` — the future-AI seam: input = frames or segments, output = events on the bus (`event.object_detected` with bounding boxes in metadata JSON). First real impl can literally be "POST frames to an external Frigate-style sidecar or ONNX service." Because detectors only *produce events*, all the downstream machinery (rules, notifications, timeline markers) already works — design the event metadata schema now so this slots in.

---

## 14. Deployment

Primary: **Docker Compose, two services** — `nvr` (Go binary with embedded dashboard; either supervising MediaMTX as a child process inside the same image — my preference, one container — or MediaMTX as a second service) — plus volumes for `data/` (SQLite) and `recordings/`. WebRTC needs care in Docker: expose the UDP port range or document `network_mode: host` (simplest and what most self-hosters on a LAN should use). Secondary: single static binary + mediamtx binary via systemd for the no-Docker crowd — your stack (pure-Go SQLite driver included) makes this nearly free, and it's a real differentiator for Pi/NAS users. Reverse proxy (Caddy) optional in front for TLS; document but don't bundle. Target hardware honesty: with no transcoding, a Pi 4 / N100 handles 4–8 cameras comfortably; transcoding changes that math — say so in docs.

---

## 15. Patterns and principles

- **Modular monolith, hexagonal-lite:** interfaces at the four extension seams and at the store layer; everywhere else, plain functions. No DI framework — wire dependencies by hand in `cmd/nvrd/main.go`.
- **Orchestrate, don't reimplement:** Go never touches media bytes on the hot path; it configures MediaMTX and runs FFmpeg. Resist any temptation to parse RTSP in Go.
- **Contract-first:** OpenAPI drives both server stubs and the shared TS client.
- **Events for decoupling, direct calls for workflows:** cross-domain side effects (notify on motion) go via the bus; in-domain sequences stay explicit function calls. Don't event-source; SQLite rows are the state.
- **Crash-only design:** MediaMTX or the Go process can die anytime; on boot, reconcile (rescan MediaMTX paths, re-index orphan segments, resume jobs). No state that only lives in memory.
- **Filesystem as backup truth** for recordings (re-indexable), DB as truth for everything else.

---

## 16. Recommended packages and libraries

**Go backend**

| Concern | Pick | Note |
|---|---|---|
| Router | `go-chi/chi` | stdlib-compatible, boring, perfect fit |
| OpenAPI | `oapi-codegen` | generates chi server stubs + types from your spec |
| SQLite driver | `modernc.org/sqlite` | pure Go → static cross-compiled binaries (Pi/NAS story); swap to `mattn/go-sqlite3` only if profiling demands it |
| Queries / migrations | `sqlc` + `golang-migrate` (or `pressly/goose`) | typed queries, no ORM |
| WebSocket | `coder/websocket` | minimal, context-aware |
| Passwords | `alexedwards/argon2id` | |
| S3 (FBS-Core) | `minio-go` | lighter than aws-sdk-go-v2 for pure S3-compat targets |
| ONVIF | `use-go/onvif` (evaluate `IOTechSystems/onvif` too) | the ecosystem here is rough; wrap it behind your Driver interface immediately |
| Cron | `robfig/cron/v3` | |
| Job queue | hand-rolled table + worker pool, or `maragudk/goqite` | River/asynq need Postgres/Redis — skip |
| Logging | stdlib `log/slog` | |
| Config | `knadh/koanf` | file + env, lighter than Viper |
| Expo push | plain HTTP to Expo's push API | no SDK needed |
| System stats | `shirou/gopsutil` | disk monitor |

**SvelteKit dashboard:** Svelte 5 + `adapter-static`, Tailwind CSS, `bits-ui`/shadcn-svelte for components, TanStack Query (svelte) for server state, generated `openapi-fetch` client from `packages/api-client`, `hls.js` for HLS fallback, a small hand-written WHEP client for WebRTC (it's ~100 lines against MediaMTX; existing libs are thin anyway), **custom canvas/SVG timeline** (this is your signature UI — off-the-shelf timeline libs won't give you coverage bars + event markers + scrub the way you want), `zod` + Superforms for forms, `lucide-svelte` icons.

**Expo mobile:** `expo-router`, TanStack Query + the same generated API client, `zustand` for local state, `react-native-webrtc` for WHEP live view with `expo-video` HLS fallback, `expo-notifications`, `expo-secure-store`.

**Astro:** Astro look at the /home/radhey/code/fbs/fbs-landing for a template

**Media:** MediaMTX (control via its REST API + hooks) and FFmpeg/ffprobe invoked as subprocesses — pin versions in the Docker image.

---
