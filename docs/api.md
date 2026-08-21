# API

This document describes the Vigil API. It covers the REST API, the WebSocket, and the MediaMTX hooks.

The API contract is in `server/api/openapi.yaml`. It is an OpenAPI 3.1 document. The base URL is `/api/v1`.

The contract drives two code generators. `oapi-codegen` generates the Go server stubs. `openapi-typescript` generates the TypeScript client types.

## Authentication

The API uses session tokens. The security scheme is `cookieAuth`. The cookie name is `nvr_session`.

The backend accepts the session in three forms:

- The `nvr_session` cookie.
- The `Authorization: Bearer` header.
- The `X-Session-Token` header.

The cookie takes precedence when more than one form is present.

A session expires after 30 days. The session token is required for most endpoints.

## Authorization roles

The API has three roles:

- `admin`: full access.
- `operator`: camera management and probing.
- `viewer`: read-only access.

The authorization rules are:

- User administration, settings writes, and Drive administration require `admin`.
- Camera create, update, delete, discovery, and probe require `operator` or `admin`.
- Camera reads, live, snapshots, recordings, events, and system reads require any authenticated user.

## The REST endpoints

### System

| Method | Path | Purpose |
|---|---|---|
| GET | `/health` | The health of the service. |
| GET | `/system/version` | The version and commit. |
| GET | `/system/disk` | The disk usage. |
| GET | `/system/status` | The system status. |

### Authentication

| Method | Path | Purpose |
|---|---|---|
| GET | `/auth/status` | If setup is required and the session is valid. |
| POST | `/auth/setup` | Create the first administrator. |
| POST | `/auth/login` | Log in. |
| POST | `/auth/logout` | Log out. |
| GET | `/auth/me` | The current user. |

### Cameras

| Method | Path | Purpose |
|---|---|---|
| GET | `/cameras` | List the cameras. |
| POST | `/cameras` | Create a camera. |
| POST | `/cameras/discover` | Discover and authenticate RTSP/ONVIF cameras using transient credentials. |
| POST | `/cameras/discover/streams` | Detect RTSP streams for a selected ONVIF camera. |
| POST | `/cameras/probe` | Probe a camera stream. |
| GET | `/cameras/{id}` | Get a camera. |
| PATCH | `/cameras/{id}` | Update a camera. |
| DELETE | `/cameras/{id}` | Delete a camera. |
| POST | `/cameras/{id}/live` | Get a live stream session. |
| GET | `/cameras/{id}/snapshot` | Get a camera snapshot. |
| GET | `/cameras/{id}/recordings` | List the recordings in a range. |
| POST | `/cameras/{id}/playback` | Get a playback session. |
| GET | `/recordings/{id}/content` | Stream a tokenized Drive archive with byte ranges. |

### Events

| Method | Path | Purpose |
|---|---|---|
| GET | `/events` | List the events. |
| POST | `/events/{id}/acknowledge` | Acknowledge an event. |

### Settings

| Method | Path | Purpose |
|---|---|---|
| GET | `/settings` | Get the settings. |
| PATCH | `/settings` | Update the settings. |

### Google Drive storage

| Method | Path | Purpose |
|---|---|---|
| GET | `/storage/gdrive/status` | Get the Drive status. |
| PUT | `/storage/gdrive/configuration` | Set the Drive configuration. |
| POST | `/storage/gdrive/connect` | Begin the OAuth connection. |
| GET | `/storage/gdrive/callback` | The OAuth callback. |
| DELETE | `/storage/gdrive/disconnect` | Disconnect Drive. |
| POST | `/storage/gdrive/archive` | Run the archive now. |

The archive response reports `uploaded`, `deleted`, `deleteFailed`, `failed`,
and `skipped`. `deleted` counts local MP4s removed after a durable Drive
upload. `deleteFailed` means the Drive upload succeeded but local cleanup needs
another reconciliation pass; it is not an upload failure.

Archived playback uses `GET /recordings/{id}/content?token=...`. It is normally
opened by the native video element from a playback session and supports HTTP
`Range` requests; clients should not construct its token themselves.

### Users

| Method | Path | Purpose |
|---|---|---|
| GET | `/users` | List the users. |
| POST | `/users` | Create a user. |
| DELETE | `/users/{id}` | Delete a user. |

## The live stream session

The live endpoint returns a live stream session. The session has:

- `cameraId`: the camera ID.
- `hlsUrl`: the HLS URL.
- `whepUrl`: the WHEP URL.
- `token`: the short-lived stream token.
- `expiresAt`: the token expiry time.

The client appends the token to the HLS and WHEP URLs. The token expires after about 60 seconds. The client refreshes the session before the token expires.

## The playback session

The playback endpoint returns a playback session. The session has:

- `cameraId`: the camera ID.
- `recordingId`: the selected recording ID.
- `playbackUrl`: a MediaMTX URL for local video or a Vigil URL for Drive video.
- `token`: the short-lived playback token.
- `expiresAt`: the token expiry time.
- `source`: `local` or `gdrive`.
- `startOffsetSec`: the seek position inside a Drive segment (zero for local playback).
- `nextRecordingStart`: the next indexed segment start. The field is omitted when there is no following segment. Drive clients use it to continue across one-minute files.

The playback request has a `start` time and an optional `durationSec`. The default duration is 60 seconds.

## The WebSocket

The WebSocket is at `GET /api/v1/ws`. It is not in the OpenAPI contract. It is mounted directly in the entry point.

The WebSocket is server-push only. Commands go over the REST API. The server authenticates the request, upgrades the connection, and subscribes to the event bus.

The server sends JSON text frames. The frame shape is:

```json
{
  "type": "event",
  "data": {}
}
```

The server writes with a five-second timeout. A reader goroutine drains incoming frames to detect disconnects.

The dashboard and mobile app do not use the WebSocket. They poll the REST API. Refer to the dashboard and mobile documents.

## The MediaMTX hooks

The MediaMTX hooks are internal. They are not part of the public API.

### The auth hook

`POST /internal/mediamtx/auth`

MediaMTX calls this hook to validate a stream request. The handler accepts the MediaMTX external-auth JSON. It validates the token from the password, the token field, or the token query string. It returns 200 on success and 401 on failure.

### The segment-complete hook

`POST /internal/mediamtx/segment-complete`

MediaMTX calls this hook when a recording segment completes. The handler accepts JSON or form data. It tolerates varied field names. It indexes the segment in SQLite.

## Error responses

Errors return a JSON object with an `error` string and an optional `code`. The HTTP status codes are standard:

- `400`: bad request.
- `401`: unauthenticated.
- `403`: forbidden.
- `404`: not found.
- `409`: conflict.
- `504`: timeout.

## The generated client

The TypeScript client is in `packages/api-client`. The `gen` script runs `openapi-typescript` on the contract. The output is `src/generated/schema.ts`.

The client wraps `openapi-fetch`. The `createApiClient` function creates a client for a base URL. It defaults credentials to `include`.

Refer to [development](./development.md) for the code generation commands.
