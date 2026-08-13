# Configuration

The backend reads its configuration from environment variables. There is no configuration file. The `Load` function in `server/internal/config` reads the variables and returns a `Config` value.

## Configuration Variables

| Variable | Purpose | Default |
|---|---|---|
| `NVR_HTTP_ADDR` | The HTTP listen address. | `:8080` |
| `NVR_DATA_DIR` | The data directory for the database. | `./data` |
| `NVR_RECORDINGS_DIR` | The recordings directory. | `./recordings` |
| `NVR_SECRETS_KEY` | The encryption key for secrets. | empty |
| `NVR_LOG_LEVEL` | The log level. | `info` |
| `NVR_RETENTION_DAYS` | The local fallback retention period in days. Drive-archived files are removed sooner after upload. | `7` |
| `NVR_MEDIAMTX_API_URL` | The MediaMTX control API URL. | `http://127.0.0.1:9997` |
| `NVR_MEDIAMTX_WEBRTC_URL` | The MediaMTX WebRTC URL. | `http://127.0.0.1:8889` |
| `NVR_MEDIAMTX_HLS_URL` | The MediaMTX HLS URL. | `http://127.0.0.1:8888` |
| `NVR_MEDIAMTX_PLAYBACK_URL` | The MediaMTX playback URL. | empty |
| `NVR_GOOGLE_CLIENT_ID` | The Google OAuth client ID. | empty |
| `NVR_GOOGLE_CLIENT_SECRET` | The Google OAuth client secret. | empty |
| `NVR_GOOGLE_REDIRECT_URL` | The Google OAuth redirect URL. | empty |
| `NVR_PUBLIC_URL` | The externally reachable HTTPS URL of this server. | empty |
| `NVR_HOSTED_DASHBOARD_URL` | The hosted dashboard URL used to reach this server. | empty |
| `NVR_CORS_ORIGINS` | Extra exact HTTPS origins allowed cross-origin (comma-separated). | empty |
| `NVR_ADMIN_USERNAME` | First-admin username for first-start env bootstrap. | empty |
| `NVR_ADMIN_PASSWORD` | First-admin password for first-start env bootstrap. | empty |

The recognized log levels are `debug`, `info`, `warn`, `warning`, and `error`. A malformed integer falls back to the default value.

## The secrets key

The `NVR_SECRETS_KEY` variable is the encryption key. It encrypts the camera credentials and the Google Drive tokens at rest. Set a long random value in production.

The key is required before you connect Google Drive. Without a key, the backend stores credentials as plaintext for development.

## Database settings

The backend stores some settings in the SQLite `settings` table. These settings override the environment variables. The settings are:

- The recordings directory.
- The recording enabled flag.
- The retention period. It controls non-Drive pruning and MediaMTX's
  `recordDeleteAfter` fallback; successful Drive uploads are cleaned up
  immediately.

When the recordings directory is empty, the backend disables recording.

## Public and hosted dashboard URLs

`NVR_PUBLIC_URL` is the externally reachable HTTPS URL of this recorder. It is
used by the slim/headless build to deep-link to the hosted dashboard, and it is
the URL the browser reaches for playback. `NVR_HOSTED_DASHBOARD_URL` points at
the separately hosted dashboard that manages this server.

The backend persists both values in the SQLite `settings` table (keys
`publicUrl` and `hostedDashboardUrl`), set through `nvrd setup` or the setup
wizard. The environment variables deliberately **override** the persisted
values when non-empty; the persisted database values apply when the environment
is unset. There is no third configuration layer.

Both values must be absolute `http` or `https` URLs. In production, set
`NVR_PUBLIC_URL` to an HTTPS origin so the browser, HLS, and WebRTC signaling
are reachable over the tunnel. See [operations](./operations.md) for the HTTPS
tunnel requirement.

## CORS origins

The backend allows cross-origin requests only from **exact** origins. The
effective allow-list is:

- Every origin in `NVR_CORS_ORIGINS` (comma-separated). Each must be HTTPS
  unless it is a localhost development origin (`http://localhost:*` or
  `http://127.0.0.1:*`).
- The origin derived from `NVR_HOSTED_DASHBOARD_URL` (the resolved hosted
  dashboard origin is always allowed).

Configured origins receive `Access-Control-Allow-Origin`,
`Access-Control-Allow-Headers` (including `Authorization` and
`X-Session-Token`), `Access-Control-Expose-Headers: X-Session-Token`, and
`Vary: Origin`, but not credential allowance. Localhost development origins
retain credential allowance for the embedded-UI dev flow.

## First-admin bootstrap

`NVR_ADMIN_USERNAME` and `NVR_ADMIN_PASSWORD` support idempotent first-start
automation. They are honored **only** when the database has no users; once any
user exists they are ignored. The username must be non-empty after trimming and
the password must contain at least eight runes — invalid partial or weak
bootstrap configuration is a startup error while setup is required. The password
is argon2id-hashed and never persisted. Remove `NVR_ADMIN_PASSWORD` from the
environment after the first admin is created.

## Google Drive configuration

You can set the Google OAuth details in two ways:

- As the `NVR_GOOGLE_*` environment variables.
- In the dashboard under Settings, Google Drive.

When you save the OAuth details in the dashboard, the backend encrypts them and stores them in the settings table. It prefers these persisted values when they are present.

To set up Google OAuth:

1. Enable the Google Drive API.
2. Configure the consent screen with the `drive.file` and `userinfo.email` scopes.
3. Create Web application credentials.
4. Make the authorized redirect URI exactly match `NVR_GOOGLE_REDIRECT_URL`.

The default redirect URL is:

```text
http://localhost:8080/api/v1/storage/gdrive/callback
```

## The example environment file

The example file is in `deploy/nvr.example.env`. Copy it and set your values. The file documents the active variables and the commented production options.

See [operations](./operations.md) for the Docker environment and the deployment model.
