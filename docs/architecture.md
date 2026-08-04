# Architecture

This document describes the design of Vigil. It gives the system overview, the main parts, and the data flows between them.

## System overview

Vigil is a self-hosted network video recorder. It records video from IP cameras to a local disk. It has one backend and three frontends.

The backend is a Go binary called `nvrd`. It runs the REST API, the WebSocket, the camera management, the recording index, and the background jobs. It serves the dashboard as static files.

The backend uses MediaMTX as the media server. MediaMTX does the ingest, the live output, and the recording. It is an external process. The backend does not start or stop MediaMTX. The backend talks to MediaMTX over its HTTP API.

The backend uses FFmpeg as a tool. It runs FFmpeg to make camera snapshots and to probe camera streams.

The data is stored in three places:

- **SQLite**: the metadata. This is the database for users, cameras, recordings, and events.
- **Recordings disk**: the recorded media files.
- **Google Drive**: the optional archive tier for offsite backups.

## The parts

### `nvrd` (the Go backend)

`nvrd` is a Go binary. It is the center of the system. It provides:

- The REST API under `/api/v1`.
- The WebSocket at `/api/v1/ws`.
- Authentication and role-based access control (RBAC).
- Camera management.
- MediaMTX integration.
- The recording index.
- The event system.
- Background jobs.
- The embedded dashboard.

The main entry point is `server/cmd/nvrd/main.go`. It loads the configuration, opens the database, creates the services, and starts the HTTP server.

### MediaMTX (the media server)

MediaMTX is an external media server. It owns the media plane:

- It pulls streams from the cameras over RTSP.
- It serves live video over WebRTC (WHEP) and HLS.
- It records the video as fMP4 segments.
- It serves time-ranged playback from the recorded segments.

`nvrd` configures MediaMTX over its HTTP API. It creates one MediaMTX path for each camera. The path is named `cam_<uuid-without-dashes>`.

MediaMTX calls back into `nvrd` through two HTTP hooks:

- `POST /internal/mediamtx/auth`: `nvrd` validates stream tokens.
- `POST /internal/mediamtx/segment-complete`: `nvrd` indexes a completed recording segment.

### FFmpeg (the media tool)

The backend runs FFmpeg as a tool. It uses FFmpeg to:

- Make camera snapshots.
- Probe camera streams during setup.

FFmpeg is not used for the continuous recording path. MediaMTX does the recording.

### The dashboard (SvelteKit)

The dashboard is a single-page application. It is built with SvelteKit and embedded into the Go binary. It is a client of the REST API. It provides:

- The live camera grid.
- Timeline playback.
- Camera setup.
- The events feed.
- Settings.
- User administration.

The dashboard is served by `nvrd` at the HTTP root. It is a static client. It has no business logic. It renders the API state.

### The mobile app (Expo)

The mobile app is an Expo application. It is a client of the REST API. It provides:

- Live view.
- Event playback.
- Event alerts.
- Camera status.

The mobile app is not an admin surface. It does not manage users or storage. It connects to a recorder over the network.

### The landing page (Astro)

The landing page is a static marketing website. It is built with Astro. It is not part of the installed system. It describes the product and gives install instructions.

### The shared API client

The shared API client is a TypeScript package. It is generated from the OpenAPI contract. Both the dashboard and the mobile app use it. It gives typed access to the API.

## The OpenAPI contract

The file `server/api/openapi.yaml` is the single source of truth for the API. It defines the REST API under `/api/v1`.

Two code generators use this file:

- `oapi-codegen` generates the Go server stubs.
- `openapi-typescript` generates the TypeScript client types.

The contract is the hinge of the repository. A change to the API contract becomes a build error in both the Go server and the TypeScript clients.

## Data flows

### Live view

The client asks the API for a live stream. The API returns a WHEP URL, an HLS URL, and a short-lived token. The client connects to MediaMTX. MediaMTX calls the auth hook to validate the token. The stream flows over WebRTC or HLS.

### Recording

MediaMTX records the video as one-minute fMP4 segments. When a segment completes, MediaMTX calls the segment-complete hook. `nvrd` indexes the segment in SQLite. The recording is continuous by default. Retention prunes the old index rows.

### Timeline playback

The dashboard asks the API for the recordings in a time range. The API returns coverage bars from SQLite. The user seeks to a time. The API issues a playback token and a MediaMTX playback URL. MediaMTX streams the requested range.

### Archive

The archive job uploads unarchived recordings to Google Drive. It uses the S3-style archive seam. The row is marked as archived. Playback of archived footage is not implemented.

### Events

The background jobs emit events. The event service writes them to SQLite and publishes them to the event bus. The WebSocket pushes them to connected clients. The dashboard and mobile app poll or receive event updates.

## Design principles

- **Modular monolith**: the backend is one binary with clear packages. The packages communicate through services and the event bus.
- **Contract-first**: the OpenAPI contract drives both the Go server and the TypeScript clients.
- **Events for decoupling**: cross-domain side effects go through the event bus.
- **Crash-only design**: on boot, the backend reconciles state. It does not keep state only in memory.
- **Filesystem as backup truth**: the recordings files can be re-indexed by scanning the disk.

## The service boundaries

The backend packages are:

- `internal/api`: the HTTP handlers.
- `internal/auth`: sessions, tokens, and roles.
- `internal/camera`: camera management and probing.
- `internal/config`: environment configuration.
- `internal/event`: the event bus and event service.
- `internal/jobs`: the cron scheduler.
- `internal/media`: the MediaMTX client and stream tokens.
- `internal/notify`: the notification seam (not implemented).
- `internal/recording`: the segment index and retention.
- `internal/secrets`: encryption at rest.
- `internal/storage`: the archive seam and Google Drive.
- `internal/store`: the SQLite data layer.
- `internal/ui`: the embedded dashboard.

Refer to [backend](./backend.md) for the details of each package.