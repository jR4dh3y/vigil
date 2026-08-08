# Operations

This document describes how to deploy and operate Vigil. It covers Docker, the MediaMTX configuration, the archive behavior, and the security notes.

## Deployment Models

You can run Vigil in two ways:

- **Docker Compose**: the recommended way. Two services run in containers.
- **Single binary**: the Go binary and the MediaMTX binary run directly.

### Docker Compose

The Docker Compose file is in `deploy/docker-compose.yml`. It defines two services:

- `mediamtx`: the MediaMTX media server.
- `nvr`: the Go backend.

Both services use `network_mode: host`. There are no port mappings because the host network exposes the ports directly.

The named volumes are:

- `nvr-data` at `/var/lib/nvr/data`.
- `nvr-recordings` at `/var/lib/nvr/recordings`.

The `nvr` service depends on `mediamtx`. Both services restart unless stopped.

### The Docker Image

The Dockerfile is in `deploy/Dockerfile`. It has three build stages plus two
runtime stages:

1. **dashboard**: builds the SvelteKit dashboard with Bun.
2. **server**: copies the dashboard into the Go binary and builds `nvrd` (full).
3. **server-slim**: builds `nvrd` with the `slim` build tag (no dashboard).
4. **runtime**: the default final stage, a full image with the embedded
   dashboard.
5. **runtime-slim**: a headless image with no embedded dashboard, for hosted
   dashboard deployments.

The runtime images install FFmpeg for snapshots and probing, declare volumes
for the data and recordings directories, and expose port `8080`.

The default build target is `runtime`, which embeds the dashboard. The
`docker-compose.yml` service and a bare `docker build` both produce the full
image. To build the headless image for a hosted dashboard:

```bash
cd deploy && docker build --target runtime-slim -t nvr-slim ..
```

The preferred deployment is MediaMTX as a sidecar service. The Dockerfile
comments say baking MediaMTX into the image is optional.

For the headless image, set `NVR_HOSTED_DASHBOARD_URL` (and
`NVR_PUBLIC_URL`/`NVR_CORS_ORIGINS` as needed) so the connection page deep-links
to the hosted dashboard. See [configuration](./configuration.md).

### Hosted dashboard (headless) model

When the dashboard is hosted separately (for example on Vercel) and the
recorder is a headless server, the two are different origins:

1. The recorder runs the slim/headless image and serves the API plus a small
   connection page.
2. The connection page deep-links to the hosted dashboard with a `?server=`
   query parameter that carries the recorder's public HTTPS URL.
3. The hosted dashboard connects to that URL over HTTPS, probes `/health`, and
   authenticates with a bearer session token that the recorder issues in the
   `X-Session-Token` header.

The recorder must be reachable from the browser over an **HTTPS tunnel** (for
example a reverse proxy or a Tailscale/cloudflared tunnel). Configure
`NVR_PUBLIC_URL` to that HTTPS origin and add the hosted dashboard origin to the
CORS allow-list (it is included automatically from `NVR_HOSTED_DASHBOARD_URL`).
See [dashboard](./dashboard.md) for the client-side connection flow.

## The MediaMTX Configuration

The MediaMTX configuration is in `deploy/mediamtx.yml`.

### Authentication

The configuration is a local and development profile. It uses open LAN auth. It defines two internal users named `any` with empty passwords.

The Go backend still mints stream tokens. MediaMTX does not enforce these tokens in this profile.

A production alternative is commented out. It uses `authMethod: http` with the backend auth hook.

### The Listeners

The MediaMTX listeners are:

- API at `:9997`.
- Playback at `:9996`.
- HLS at `:8888`.
- WebRTC at `:8889`.
- RTSP at `:8554`.

WebRTC uses local UDP at `:8189`.

### The Paths

The `paths` map is empty. The backend creates the paths at runtime through the MediaMTX control API.

The path defaults configure the recording. Each segment is one minute of fMP4.

### The Segment Hook

The configuration sets the `runOnRecordSegmentComplete` hook. After each segment, MediaMTX runs a `curl` command that POSTs to the backend:

```text
POST http://127.0.0.1:8080/internal/mediamtx/segment-complete
```

The body contains the path, the file path, and the segment duration. The backend indexes the segment.

## The Recordings Layout

The recorded files are stored in this layout:

```text
recordings/<camera_id>/<YYYY-MM-DD>/<HH-MM-SS-microseconds>.mp4
```

The recordings directory is runtime data. It is ignored by Git. Refer to [database](./database.md) for the index and the recovery model.

## The Archive Behavior

Vigil uploads unarchived recordings to Google Drive at 00:00 UTC each day. The archive runs in batches of 100. You can also trigger an archive from the dashboard.

The archive retries are idempotent. A failed upload does not corrupt the state. The retention preserves the pending recording metadata while a Drive connection is stored.

## Cleanup

The retention job prunes the old index rows. The default retention period is 7 days. You can change it with the `NVR_RETENTION_DAYS` variable or in the dashboard settings.

The retention job does not delete the media files. It deletes the index rows. See [database](./database.md) for the details.

## The Local Media Tools

The `tools/mediamtx/run.sh` launcher starts MediaMTX. It uses the local binary in `.bin/mediamtx` first. It falls back to `mediamtx` from the PATH. It passes `deploy/mediamtx.yml` as the configuration.

The `.bin/mediamtx` binary is version 1.19.2. The Docker image pins MediaMTX to version 1.12.2.

## Security Notes

- Set a long `NVR_SECRETS_KEY` in production.
- The session cookie is not marked `Secure`. Shield the service with a reverse proxy for HTTPS.
- Use host networking for WebRTC on a LAN box.
- The example environment file and the setup guide contain example values. Do not copy example credentials into production.
- In the hosted-dashboard model, the recorder must be reachable over HTTPS and
  `NVR_PUBLIC_URL` must match that HTTPS origin so the browser can reach the
  API, HLS, and WebRTC signaling. Plain-HTTP recorder URLs are rejected when the
  dashboard itself is served over HTTPS.
- `NVR_CORS_ORIGINS` accepts only exact HTTPS origins (plus localhost
  development origins). Do not add broad wildcards; add only the origins the
  hosted dashboard is actually served from.
- The remote dashboard authenticates with a bearer session token stored in the
  browser's `localStorage`. Unlike the HttpOnly cookie used by the embedded UI,
  this token is readable by script on the dashboard origin, so keep the hosted
  dashboard to trusted origins and HTTPS.
- `NVR_ADMIN_PASSWORD`, when used for first-start bootstrap, is argon2id-hashed
  and never persisted. Remove it from the environment after the first admin is
  created.

## Troubleshooting

- If the server does not start, check the log level and the health endpoint.
- If a camera does not record, check that MediaMTX runs and that the camera streams over RTSP.
- If Google Drive does not connect, verify the OAuth redirect URI and the `NVR_SECRETS_KEY`.
- If the disk fills, check the retention job and the recordings directory.