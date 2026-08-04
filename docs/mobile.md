# Mobile

This document describes the mobile app. The app is in `apps/mobile/`. It is an Expo application with Expo Router.

## Overview

The mobile app is the check-on-my-house client for Vigil. It provides live view, event playback, and event alerts. It is not an admin surface. It does not manage users or storage.

The app is built with Expo SDK 57. It uses Expo Router for navigation. It uses TanStack Query for server state. It uses Zustand for local state.

The app connects to a Vigil recorder over the network. The user enters the recorder URL. The app stores the URL and the session token on the device.

## Build and configuration

The app configuration is in `apps/mobile/app.json`.

- App name: `Vigil`.
- Slug: `vigil`.
- URI scheme: `vigil`.
- Bundle ID: `com.nvr.app`.

The app uses these Expo plugins:

- `expo-router`.
- `expo-secure-store`.
- `expo-notifications`.
- `expo-build-properties` with Android cleartext traffic enabled.

The cleartext setting allows HTTP connections to local recorders.

## The environment variables

| Variable | Purpose |
|---|---|
| `EXPO_PUBLIC_API_URL` | The initial recorder API URL. The default is `http://127.0.0.1:8080/api/v1`. |
| `EXPO_OS` | Platform checks for keyboard and notifications. |

## The routes

The app uses Expo Router file-based routing. The routes are in `apps/mobile/app/`.

| Route | Surface |
|---|---|
| `/` | Cold-start redirect. |
| `/login` | Login form. |
| `/setup` | First-time admin setup form. |
| `/server` | Recorder address form sheet. |
| `/(tabs)/(live)` | Live camera list and camera detail. |
| `/(tabs)/(events)` | Recent activity and event detail. |
| `/(tabs)/(settings)` | Settings. |

The root layout hydrates the recorder URL, the session token, and the local preferences before it mounts the app.

## Authentication

The app uses the browser-style session flow:

1. The root layout restores the recorder URL and the session token from storage.
2. The app queries `GET /auth/status` to decide the surface.
3. Login and setup call the auth endpoints and store the session token.
4. The backend issues the session in the `X-Session-Token` header. The app stores it in SecureStore.

The token storage is in `apps/mobile/lib/api/session.ts`. The storage key is `nvr_session`.

There is no refresh-token endpoint. The app updates the token from the `X-Session-Token` response header. A `401` response clears the token, except for the login, setup, and status requests.

## The API client

The app uses the shared `@nvr/api-client` package. The client is in `apps/mobile/lib/api/client.ts`.

The client caches one OpenAPI client per recorder URL. It uses the `expo/fetch` implementation. It sends the token as `Authorization: Bearer`.

The session middleware:

- Adds the bearer token on requests.
- Saves the session token from responses.
- Ignores responses from an old recorder URL.
- Clears the token on `401`.

## Recorder connection

The user enters the recorder URL in the server form. The app normalizes the URL:

- Adds `http://` if there is no scheme.
- Rejects non-HTTP URLs and embedded credentials.
- Ensures the path ends in `/api/v1`.

The app tests the recorder with `GET /health`. On success, it saves the URL, clears the session and the query cache, and returns to the start.

The URL storage key is `vigil_recorder_url`.

## Live view

The live camera list polls the cameras every 30 seconds. Each camera card requests a live stream only when it is focused and the camera is online.

The live session comes from `POST /cameras/{id}/live`. The app refreshes the session before the token expires.

The player chooses WHEP first. The WHEP hook creates a WebRTC peer connection with no STUN servers. It negotiates over HTTP. On failure, the player falls back to HLS.

The HLS player uses `expo-video`. It replaces the video URI and starts playback.

## Event playback

The event detail route loads the event, the camera, and the playback session. The playback session comes from `POST /cameras/{id}/playback`. The default duration is 60 seconds.

The app exposes the playback URL only while the route is focused.

## Notifications

The app uses local notifications. It does not register a remote push token.

The notification service:

- Creates an Android channel called `vigil-alerts`.
- Requests permission.
- Schedules local notifications for new events.

The events tab polls every 15 seconds. When monitoring is armed and notifications are enabled, the app schedules local notifications for unacknowledged warning and critical events.

A notification response routes to the event detail page.

The notification module is disabled on web and in Expo Go.

## Settings

The settings screen shows:

- The account.
- The armed monitoring switch.
- The notification permission switch.
- The recorder health summary.
- The recorder URL link.
- The app metadata.

The armed switch is a local preference. It does not call the recorder API.

## State management

The app uses three kinds of state:

- **Server state**: TanStack Query. The query client has a 30-second stale time and one retry.
- **Local state**: Zustand in `apps/mobile/lib/store.ts`. It stores the armed flag, the notification flag, and the event watermark.
- **Component state**: React state for forms and errors.

The Zustand store persists through SecureStore. The storage key is `vigil_preferences`.

## Theming

The app uses Expo Router theme providers. It selects the light or dark theme from the device color scheme.

The colors are in `apps/mobile/theme/colors.ts`. They use platform-specific values: iOS system colors, Android Material colors, and web defaults.