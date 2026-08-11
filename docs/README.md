# Vigil Documentation

Vigil is a self-hosted network video recorder (NVR). It records video from RTSP/ONVIF cameras to your own disk. It provides live view, timeline playback, and event alerts. It has a Go backend (`nvrd`), a SvelteKit dashboard, an Expo mobile app, and an Astro marketing website.

This documentation describes the completed codebase as implemented in this repository. Write all documentation in Simplified Technical English (STE).

## Documents

- [Architecture](./architecture.md): system design, parts, and data flows.
- [Quickstart](./quickstart.md): prerequisites, development run, first admin, and production build.
- [Configuration](./configuration.md): environment variables, defaults, and the secrets key.
- [Setup And Authentication](./setup-and-authentication.md): first-start bootstrap, sessions, cookies, and roles.
- [API](./api.md): REST endpoints, the WebSocket, and the MediaMTX hooks.
- [Database](./database.md): SQLite schema, migrations, queries, and recovery.
- [Backend](./backend.md): the Go services and packages.
- [Dashboard](./dashboard.md): the SvelteKit web client.
- [Mobile](./mobile.md): the Expo mobile client.
- [Landing](./landing.md): the Astro marketing website.
- [Operations](./operations.md): Docker, MediaMTX, archive, and security.
- [Development](./development.md): repository layout, build, code generation, and linting.

## Quick Start

For the full setup flow, see [Quickstart](./quickstart.md).

Run the development tasks from the repository root:

```bash
bun install
bun run dev
```

The server runs on `:8080`. On a new database, open the dashboard and create the first administrator. See [Setup And Authentication](./setup-and-authentication.md).

Run the production binary:

```bash
bun run build:bin
./server/bin/nvrd
```

This builds the dashboard, embeds it into the Go binary, and serves both the
API and dashboard from one process.

## API Families

The backend exposes four HTTP surfaces:

- System: `GET /health`, `GET /system/version`, `GET /system/disk`, `GET /system/status`.
- Authentication: `GET /auth/status`, `POST /auth/setup`, `POST /auth/login`, `POST /auth/logout`, `GET /auth/me`.
- Management: cameras, recordings, events, settings, users, and Google Drive under `/api/v1`.
- The WebSocket at `/api/v1/ws` for server-push events.

Protected routes accept the `nvr_session` cookie, the `Authorization: Bearer` header, or the `X-Session-Token` header. Management routes additionally require the applicable role. See [API](./api.md) and [Setup And Authentication](./setup-and-authentication.md).