# Backend

This document describes the Go backend. The backend is the Go binary called `nvrd`. It is in the `server/` directory.

## Overview

`nvrd` is a Go module at `github.com/nvr/nvr/server`. It is a modular monolith. It provides the REST API, the WebSocket, camera management, the recording index, the event system, and the background jobs.

The backend is the center of the system. It does not process media bytes. It configures MediaMTX and runs FFmpeg as a tool.

## The entry point

The main entry point is `server/cmd/nvrd/main.go`. The `main` function does this sequence:

1. Creates a root context for the interrupt and termination signals.
2. Loads the configuration.
3. Sets up the JSON logger.
4. Creates the data and recordings directories.
5. Opens the SQLite database and applies the migrations.
6. Creates the domain services.
7. Reads the settings that override the runtime configuration.
8. Creates the API server and the HTTP router.
9. Registers the MediaMTX hooks and the WebSocket.
10. Registers the generated OpenAPI routes.
11. Serves the embedded dashboard.
12. Starts the HTTP server.
13. On shutdown, stops the scheduler and the HTTP server.

## The configuration

The configuration comes from environment variables. The package is `server/internal/config`. There is no configuration file. The `Load` function reads the variables and returns a `Config` value.

The configuration variables are:

| Variable | Purpose | Default |
|---|---|---|
| `NVR_HTTP_ADDR` | The HTTP listen address. | `:8080` |
| `NVR_DATA_DIR` | The data directory for the database. | `./data` |
| `NVR_RECORDINGS_DIR` | The recordings directory. | `./recordings` |
| `NVR_SECRETS_KEY` | The encryption key for secrets. | empty |
| `NVR_LOG_LEVEL` | The log level. | `info` |
| `NVR_MEDIAMTX_API_URL` | The MediaMTX control API URL. | `http://127.0.0.1:9997` |
| `NVR_MEDIAMTX_WEBRTC_URL` | The MediaMTX WebRTC URL. | `http://127.0.0.1:8889` |
| `NVR_MEDIAMTX_HLS_URL` | The MediaMTX HLS URL. | `http://127.0.0.1:8888` |
| `NVR_MEDIAMTX_PLAYBACK_URL` | The MediaMTX playback URL. | empty |
| `NVR_RETENTION_DAYS` | The recording retention period. | `7` |
| `NVR_GOOGLE_CLIENT_ID` | The Google OAuth client ID. | empty |
| `NVR_GOOGLE_CLIENT_SECRET` | The Google OAuth client secret. | empty |
| `NVR_GOOGLE_REDIRECT_URL` | The Google OAuth redirect URL. | empty |

The backend stores some settings in the database. These settings override the environment variables. The settings are the recordings directory, the recording enabled flag, and the retention period.

## The service packages

The packages are in `server/internal/`. Each package has a narrow public surface.

### `internal/api`

The API package is the HTTP adapter layer. It provides the REST handlers, the middleware, and the WebSocket.

The main file is `server/internal/api/server.go`. It defines the `Server` type. The `Server` holds all the domain services.

The handler files are:

- `handlers.go`: health, version, and authentication.
- `cameras.go`: camera CRUD, live, snapshot, and probe.
- `recordings.go`: recording list and playback.
- `events.go`: event list and acknowledge.
- `settings.go`: settings get and update.
- `system.go`: disk and status.
- `users.go`: user list, create, and delete.
- `storage_gdrive.go`: Google Drive administration.

The middleware is in `middleware.go`. The `SessionMiddleware` resolves the session token and attaches the user to the request context. The `CORSMiddleware` allows local development origins.

The authorization helpers are in `authz.go`. They are `requireUser`, `requireOperator`, and `requireAdmin`.

The WebSocket is in `ws.go`. Refer to [api](./api.md) for the WebSocket details.

### `internal/auth`

The auth package manages authentication and roles. It provides password hashing, session tokens, and role constants.

The password functions use Argon2id. Refer to `password.go`.

The session functions are in `session.go`. A session token is 32 random bytes. The backend stores only the SHA-256 hash of the token. The session TTL is 30 days.

The cookie functions are in `cookie.go`. The cookie name is `nvr_session`. The cookie is `HttpOnly`, `Path=/`, `SameSite=Lax`, and `Secure=false`. The backend also accepts the `Authorization: Bearer` header and the `X-Session-Token` header.

The service is in `service.go`. It creates sessions, resolves users from tokens, and deletes sessions.

The roles are `admin`, `operator`, and `viewer`. On first boot, the setup endpoint creates the first user as `admin`.

### `internal/camera`

The camera package manages cameras and stream profiles. It provides camera CRUD, encrypted credentials, and stream probing.

The service is in `service.go`. It lists, gets, creates, updates, and deletes cameras. It manages the stream profiles. It encrypts the camera passwords.

The driver is in `driver.go`. The `Driver` type probes a camera. The `GenericRTSPDriver` in `generic_rtsp.go` uses `ffprobe` to get the camera stream metadata.

The probing result includes the codec and resolution. It flags H.265/HEVC streams.

### `internal/media`

The media package is the MediaMTX client. It is the only package that knows MediaMTX exists.

The service is in `service.go`. It issues live and playback streams. It mints short-lived tokens. It manages the MediaMTX paths.

The MediaMTX client is in `mediamtx.go`. It uses the MediaMTX v3 API to upsert and delete paths. A path is named `cam_<uuid-without-dashes>`.

The token store is in `tokens.go`. Tokens are path-bound and held in memory. They are reusable until expiry.

The auth handler is in `auth.go`. It accepts the MediaMTX external-auth callback and validates the token.

The snapshot function is in `snapshot.go`. It runs FFmpeg to capture one JPEG frame.

### `internal/recording`

The recording package owns the segment index. It indexes completed segments, answers timeline queries, and enforces retention.

The service is in `service.go`. It lists recordings, indexes segments, and prunes old rows. The retention period defaults to 7 days.

The segment-complete handler is in `hook.go`. It accepts the MediaMTX callback and indexes the segment. It tolerates JSON or form data with varied field names.

The path helper is in `path.go`. It converts a MediaMTX path name to a camera ID.

There is no clip-export function. Playback is a short-lived MediaMTX URL.

### `internal/event`

The event package persists events and publishes them to subscribers.

The bus is in `bus.go`. Subscribers receive buffered channels. A slow subscriber can lose events.

The service is in `service.go`. It emits, gets, lists, and acknowledges events. The event types are camera online, camera offline, disk low, and archive complete.

### `internal/jobs`

The jobs package is the cron scheduler. It is not a durable queue.

The scheduler is in `scheduler.go`. It uses `robfig/cron`. The jobs are:

- Camera health probes every minute.
- Retention prune every hour.
- Disk check every five minutes.
- Google Drive archive at midnight UTC.

The archive job is in `archive_gdrive.go`.

There is no thumbnail generator.

### `internal/notify`

The notify package is a placeholder. It has only a package comment. There is no notifier implementation.

### `internal/secrets`

The secrets package provides encryption at rest. It uses AES-256-GCM. The key is the SHA-256 of the secrets key. Without a key, values are stored as plaintext for development.

### `internal/storage`

The storage package is the archive seam. It defines the `Provider` interface for object storage. It has `Put`, `Get`, `Delete`, `List`, and `PresignURL`.

There is no local or S3 implementation. Google Drive is the concrete archive tier in `internal/storage/gdrive`.

The recording index adapter is in `recording_index.go`. It connects the recording service to the Drive archive.

### `internal/storage/gdrive`

The gdrive package is the Google Drive archive. It manages the OAuth lifecycle and the archive uploads.

The service is in `service.go`. It validates configuration, begins OAuth, handles the callback, and disconnects.

The archive functions are in `archive.go`. The `ArchivePending` function is single-flight. It uploads older unarchived rows in batches. It marks missing files as skipped.

The upload function is in `upload.go`. It creates files in the `NVR Archives` folder. It uses the `nvr_recording_id` property for idempotent retries.

### `internal/store`

The store package is the SQLite data layer. It opens the database, applies migrations, and provides the generated queries.

The database is opened in `open.go`. It uses `modernc.org/sqlite`. It enables WAL, foreign keys, and a busy timeout. It sets one maximum open connection.

The migrations are applied in `migrate.go`. Refer to [database](./database.md) for the schema.

The queries are generated by sqlc. The query files are in `server/internal/store/queries/`.

### `internal/ui`

The ui package embeds the dashboard. The `Handler` function serves the built dashboard files. The `/api/` paths return 404. Unknown routes serve `index.html`.

### `internal/deps`

The deps package pins libraries in `go.mod` with blank imports. It has no runtime API. Some pinned libraries have no implementation yet.

## The routers and hooks

The backend registers three kinds of routes:

- The generated OpenAPI routes under `/api/v1`.
- The WebSocket at `/api/v1/ws`.
- The MediaMTX hooks at `/internal/mediamtx/auth` and `/internal/mediamtx/segment-complete`.

The MediaMTX hooks are internal. They are not part of the public API.

## Build and code generation

The backend uses two code generators:

- `oapi-codegen` generates the server stubs from the OpenAPI contract.
- `sqlc` generates the database queries.

The commands are in `server/Makefile`. The `make generate` command runs both generators.

Refer to [api](./api.md) for the API contract and to [development](./development.md) for the build commands.