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
| `NVR_RETENTION_DAYS` | The recording retention period. | `7` |
| `NVR_MEDIAMTX_API_URL` | The MediaMTX control API URL. | `http://127.0.0.1:9997` |
| `NVR_MEDIAMTX_WEBRTC_URL` | The MediaMTX WebRTC URL. | `http://127.0.0.1:8889` |
| `NVR_MEDIAMTX_HLS_URL` | The MediaMTX HLS URL. | `http://127.0.0.1:8888` |
| `NVR_MEDIAMTX_PLAYBACK_URL` | The MediaMTX playback URL. | empty |
| `NVR_GOOGLE_CLIENT_ID` | The Google OAuth client ID. | empty |
| `NVR_GOOGLE_CLIENT_SECRET` | The Google OAuth client secret. | empty |
| `NVR_GOOGLE_REDIRECT_URL` | The Google OAuth redirect URL. | empty |

The recognized log levels are `debug`, `info`, `warn`, `warning`, and `error`. A malformed integer falls back to the default value.

## The secrets key

The `NVR_SECRETS_KEY` variable is the encryption key. It encrypts the camera credentials and the Google Drive tokens at rest. Set a long random value in production.

The key is required before you connect Google Drive. Without a key, the backend stores credentials as plaintext for development.

## Database settings

The backend stores some settings in the SQLite `settings` table. These settings override the environment variables. The settings are:

- The recordings directory.
- The recording enabled flag.
- The retention period.

When the recordings directory is empty, the backend disables recording.

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