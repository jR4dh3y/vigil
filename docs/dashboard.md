# Dashboard

This document describes the web dashboard. The dashboard is in `apps/dashboard/`. It is a SvelteKit single-page application.

## Overview

The dashboard is the main admin surface for Vigil. It provides live view, timeline playback, camera setup, events, settings, and user administration.

The dashboard is a static client. It renders the API state. It has no business logic. The Go backend embeds the built dashboard files.

The dashboard is built with Svelte 5 runes mode. It uses Tailwind CSS 4. It uses TanStack Query for server state.

## Build and configuration

The dashboard uses the `adapter-static` adapter. The configuration is in `apps/dashboard/vite.config.ts`. The build output goes to `build/`. The SPA fallback is `index.html`.

The dashboard has no `svelte.config.js`. The adapter and compiler configuration are in `vite.config.ts`.

The dev server runs on port `5173`. It proxies:

- `/api` to the Go backend on `:8080`.
- `/mtx-hls` to MediaMTX HLS on `:8888`.
- `/mtx-webrtc` to MediaMTX WebRTC on `:8889`.

The API base URL is `VITE_API_BASE`. The default is `/api/v1`.

## The routes

The dashboard uses SvelteKit file-based routing. The routes are in `apps/dashboard/src/routes/`.

| URL | Purpose |
|---|---|
| `/` | The live camera grid. |
| `/cameras` | The camera list. |
| `/cameras/new` | Create a camera. |
| `/cameras/[id]` | Camera detail and edit. |
| `/cameras/[id]/timeline` | Timeline playback. |
| `/events` | The events feed. |
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

The shell has a sidebar with four sections: Live, Cameras, Events, and Settings. The Settings section has a Users tab for admins.

## The live grid

The home page shows a three-by-three grid of live camera tiles. Each tile requests a live stream with `POST /cameras/{id}/live`. The tile plays the stream.

The tile prefers HLS. The `HlsPlayer` component uses `hls.js`. It falls back to native HLS where available. On a fatal HLS error, the tile switches to WHEP.

The `WhepPlayer` component creates a WebRTC peer connection and negotiates over HTTP. It sends the SDP offer to the WHEP URL and applies the SDP answer.

The live query refreshes before the token expires. The refresh interval is about ten seconds before expiry.

## Timeline playback

The timeline route shows the recording coverage for a camera. The presets are 1 hour, 24 hours, and 7 days. The default range is the last 24 hours.

The route queries `GET /cameras/{id}/recordings`. The response has the segments and the coverage bars.

The `CoverageTimeline` component renders the coverage bars. You can click or drag to scrub. You can use the keyboard arrows and the Home and End keys.

A seek calls `POST /cameras/{id}/playback`. The `PlaybackPlayer` plays the MediaMTX HLS URL with the playback token.

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

## The events feed

The events route lists the events. It requests `limit: 100`. It refetches every 15 seconds. There is a filter for unacknowledged events.

Each event row shows the severity, type, title, message, time, and camera link. You can acknowledge an unacknowledged event.

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
- HLS over HTTP for playback.
- WebRTC over HTTP signaling for live view.
- Polling every 15 seconds for events.

## Styling

The dashboard uses Tailwind CSS 4. The components use utility classes directly. There are no style blocks.

The visual theme is dark surfaces with emerald accents. Status colors are red, amber, and sky.