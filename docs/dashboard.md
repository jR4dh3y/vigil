# Dashboard

This document describes the web dashboard. The dashboard is in `apps/dashboard/`. It is a SvelteKit single-page application.

## Overview

The dashboard is the main admin surface for Vigil. It provides live view, timeline playback, camera setup, system alerts, settings, and user administration.

The dashboard is a static client. It renders the API state. It has no business logic. The Go backend embeds the built dashboard files.

The dashboard is built with Svelte 5 runes mode. It uses Tailwind CSS 4. It uses TanStack Query for server state.

## Camera discovery

The add-camera flow asks for the camera or NVR username and password before
scanning. The server first checks ONVIF devices and then scans local IPv4
networks for Dahua-compatible RTSP channels such as
`/cam/realmonitor?channel=1&subtype=0`. Credentials are used transiently to
validate each result and are not returned in discovered URLs.

If a device does not respond to ONVIF or the Dahua RTSP channel pattern, the
user can continue with manual RTSP stream URLs.

## Build and configuration

The dashboard uses the `adapter-static` adapter. The configuration is in `apps/dashboard/vite.config.ts`. The build output goes to `build/`. The SPA fallback is `index.html`.

The dashboard has no `svelte.config.js`. The adapter and compiler configuration are in `vite.config.ts`.

The dev server runs on port `5173`. It proxies:

- `/api` to the Go backend on `:8080`.
- `/mtx-hls` to MediaMTX HLS on `:8888`.
- `/mtx-webrtc` to MediaMTX WebRTC on `:8889`.

The API base URL is resolved at runtime in `src/lib/connection/`. In embedded
mode it is the relative `/api/v1`. When a remote server is configured it is the
absolute HTTPS base URL of that recorder.

## The routes

The dashboard uses SvelteKit file-based routing. The routes are in `apps/dashboard/src/routes/`.

| URL | Purpose |
|---|---|
| `/` | The live camera grid. |
| `/cameras` | The camera list. |
| `/cameras/new` | Create a camera. |
| `/cameras/[id]` | Camera detail and edit. |
| `/cameras/[id]/timeline` | Timeline playback. |
| `/events` | The system alerts feed. It remains available by direct URL but is not in the primary navigation. |
| `/settings` | System settings and storage. |
| `/settings/users` | User administration. |
| `/login` | Login. |
| `/setup` | First-time setup. |

The root layout disables server rendering and enables prerendering. The dynamic camera routes disable prerendering.

## Authentication and navigation

The `AuthGate` component guards the routes. It queries `GET /auth/status`. It redirects based on the auth state:

- If setup is required, it redirects to `/setup`.
- If the user is not authenticated, it redirects to `/login`.
- If the user is authenticated, it shows the app shell.

The shell has a sidebar with three primary sections: Live, Cameras, and Settings. The system alerts feed remains available at `/events` for direct access. The Settings section has a Users tab for admins.

## Embedded vs hosted connection

The dashboard has two connection modes, selected at runtime in
`src/lib/connection/`:

- **Embedded (same-origin).** When the dashboard is served by the Go binary
  with the embedded UI, it uses the relative `/api/v1` base and HttpOnly
  cookies (`credentials: "include"`). This is the default. There is no remote
  server configuration and no bearer token.
- **Hosted (remote).** When the dashboard is hosted separately (for example on
  Vercel), it connects to a recorder over HTTPS and authenticates with a bearer
  session token. Remote requests use `credentials: "omit"` so the dashboard's
  cookies are never sent to the recorder.

The connection gate (`ServerGate`) probes same-origin `/api/v1/health` first. It
accepts the embedded mode only when the response matches the Vigil JSON contract
`{"status":"ok"}`. Otherwise it shows the login/setup shell with a server URL
field.

### Connection UI and deep links

The login/setup shell displays the active remote server and lets you change it.
Typing a server address runs it through the same normalization, HTTPS policy,
health validation, and persistence path as any other input.

The dashboard also accepts a `server` query parameter as a deep link. The
slim/headless recorder's connection page emits a link like
`https://hosted-dashboard.example/?server=https://recorder.example.com`. The
parameter prefills the server field and is validated exactly like manual input —
it never bypasses normalization, the HTTPS policy, or the health check. The
saved server and session token persist in the browser's `localStorage` (keys
`nvr_remote_server` and `nvr_session`), degrading to in-memory state when storage
is unavailable.

## The live grid

The home page shows a three-by-three grid of live camera tiles. Each tile requests a live stream with `POST /cameras/{id}/live`. The tile plays the stream.

The tile prefers HLS. The `HlsPlayer` component uses `hls.js`. It falls back to native HLS where available. On a fatal HLS error, the tile switches to WHEP.

The `WhepPlayer` component creates a WebRTC peer connection and negotiates over HTTP. It sends the SDP offer to the WHEP URL and applies the SDP answer.

The live query refreshes before the token expires. The refresh interval is about ten seconds before expiry.

## Timeline playback

The live playback timeline opens one local calendar day at 100% zoom, from midnight through 11:59 PM. Its date control sits beside the zoom controls and opens any older retained day. The calendar requests `GET /recordings/days` for its visible six-week grid. Days with local recordings are purple, days archived to Google Drive are green, and mixed days show both markers; the text legend and accessible day labels carry the same meaning without relying on color alone. The camera timeline also provides 1-hour, 24-hour, and 7-day presets.

The route queries `GET /cameras/{id}/recordings`. The response has the segments and the coverage bars.

The timeline fits the selected range to the panel at 100%. Vertical scrolling zooms around the timestamp under the pointer. After zooming in, horizontal scrolling or a touch swipe pans through the range. Hovering, tapping, clicking, or dragging shows the exact local timestamp. Keyboard arrows move by one visible tick; Home and End jump to the range boundaries.

A seek calls `POST /cameras/{id}/playback`. The `PlaybackPlayer` loads either a tokenized MediaMTX MP4 URL or a tokenized Vigil Drive-proxy URL in the same browser video element. Drive playback forwards byte ranges for seeking and displays a small source label. When a Drive MP4 ends, the player uses `nextRecordingStart` from the session to request the following indexed segment. Relative proxy URLs are resolved against the active recorder so hosted-dashboard Bearer mode also works.

## Camera setup

The create and edit forms are in the `CameraForm` component. The form fields are:

- Name.
- Host.
- Username and password (optional).
- Live RTSP URL.
- Record RTSP URL.
- Enabled.

The form validates with Zod. The probe control calls `POST /cameras/probe`. The probe result shows reachability, error, codec, and resolution. It warns about H.265/HEVC streams.

The camera detail page shows a snapshot from `GET /cameras/{id}/snapshot`. It has edit and delete actions.

## The system alerts feed

The system alerts route lists alerts. It requests `limit: 100`. It refetches every 15 seconds. There is a filter for unacknowledged alerts.

Each alert row shows the severity, type, title, message, time, and camera link. You can acknowledge an unacknowledged alert.

## Settings

The settings page shows the system status, the site settings, and the Google Drive controls.

The `SystemStatusCard` shows health, version, camera counts, retention, and disk usage.

The `SettingsForm` edits the site name, retention period, recordings directory, and the recording enabled flag. Non-admins see a read-only form.

The Google Drive card shows the connection state. It provides connect, disconnect, and archive actions. The OAuth callback is processed in the page.

## User administration

The users route is admin-only. It lists the users and provides create and delete actions. You cannot delete the current user. You cannot delete the last admin.

The roles are `admin`, `operator`, and `viewer`.

## State management

The dashboard uses TanStack Query. The query client is in `src/lib/query.ts`. The default stale time is 30 seconds. The default retry count is one.

The query keys are scoped by feature. The key factories are in `src/lib/*/keys.ts`.

The dashboard does not use Svelte stores. Local state uses Svelte 5 runes.

## Streaming transport

The dashboard does not use a WebSocket. It uses:

- HTTP for the API.
- MP4 over HTTP for recorded playback.
- WebRTC over HTTP signaling for live view.
- Polling every 15 seconds for system alerts.

## Hosted-dashboard transport requirements

When the dashboard runs hosted (remote mode), the recorder must be reachable
from the browser over an HTTPS tunnel. Concretely:

- `NVR_PUBLIC_URL` on the recorder must be the HTTPS origin the browser reaches.
- The recorder's API, HLS, and WebRTC WHEP signaling URLs must all be HTTPS and
  reachable from the browser. Because the dashboard is served over HTTPS, the
  browser rejects plain-HTTP recorder addresses and mixed-content media URLs.
- The hosted dashboard and the recorder are cross-origin, so the recorder must
  allow the dashboard's origin under its CORS policy (it is added automatically
  from `NVR_HOSTED_DASHBOARD_URL`).

### Token transport and storage tradeoff

In hosted mode the dashboard authenticates with a bearer session token that the
recorder issues in the `X-Session-Token` response header. The token is stored in
the browser's `localStorage` and sent as `Authorization: Bearer` on remote
requests. This contrasts with the embedded mode, which uses an HttpOnly cookie
that JavaScript cannot read.

The tradeoff: a `localStorage` token is readable by any script running on the
dashboard origin, so it is only as safe as that origin. Keep the hosted
dashboard to trusted, HTTPS-served origins and a strict Content-Security-Policy
(`connect-src` must allow the recorder origins). The embedded UI retains the
stronger HttpOnly-cookie model and is the recommended default when the dashboard
and server are served together.

## Styling

The dashboard uses Tailwind CSS 4. The components use utility classes directly. There are no style blocks.

The visual theme is dark surfaces with emerald accents. Status colors are red, amber, and sky.
