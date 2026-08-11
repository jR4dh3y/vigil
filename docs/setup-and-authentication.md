# Setup And Authentication

This document describes the first-start setup and the authentication model. The authentication code is in `server/internal/auth`. The API handlers are in `server/internal/api`.

## First Start

On a new database, there are no users. The system requires a first administrator.

The flow is:

1. `GET /auth/status` reports that setup is required.
2. The client shows the setup form.
3. `POST /auth/setup` creates the first user as `admin`.
4. The setup response issues a session.

The setup endpoint rejects a request with `409` when a user already exists. The password must be at least eight characters.

## `nvrd setup` (CLI)

The server binary also exposes an explicit `nvrd setup` command. It creates the
first admin and persists the public and hosted dashboard URLs; it never starts
the HTTP server and never waits for a terminal. URLs are stored in the database
`settings` table, and the password is never persisted or shown on the command
line.

```
Usage: nvrd setup [flags]

Flags:
  --username <name>       Admin username (default: $NVR_ADMIN_USERNAME)
  --password-stdin        Read the admin password from standard input instead
                          of prompting on the terminal
  --public-url <url>      Public URL of this server (default: $NVR_PUBLIC_URL)
  --hosted-url <url>      Hosted dashboard URL used to reach this server
                          (default: $NVR_HOSTED_DASHBOARD_URL)
  --non-interactive       Fail instead of prompting when required values are
                          missing
  -h, --help              Show this help
```

### Interactive

Run `nvrd setup` with the URL flags. It prompts for the admin username and the
password on a hidden terminal line:

```bash
./server/bin/nvrd setup \
  --public-url https://recorder.example.com \
  --hosted-url https://nvr.example.com/dashboard
```

### Non-interactive

For first-start automation, pass the password over standard input so it never
appears in `argv`:

```bash
./server/bin/nvrd setup \
  --username admin \
  --password-stdin \
  --public-url https://recorder.example.com \
  --hosted-url https://nvr.example.com/dashboard \
  --non-interactive <<< 'a-strong-password'
```

The password source is resolved in this order: `--password-stdin`, then the
`NVR_ADMIN_PASSWORD` environment variable, then a hidden terminal prompt. There
is no `--password` flag and the password is never read from a file. In
non-interactive mode a missing password is an error.

### URL-only reruns

Once an admin exists, `nvrd setup` rejects any rerun that carries credential
input (a username, a password source, or `--password-stdin`). It allows URL-only
updates, which persist the public and hosted dashboard URLs without touching
users:

```bash
./server/bin/nvrd setup --public-url https://recorder.example.com
```

### Env first-start bootstrap

The server also supports first-start automation without a CLI: when the pair
`NVR_ADMIN_USERNAME` and `NVR_ADMIN_PASSWORD` is set and the database has no
users, startup creates the first admin. This is idempotent first-start config
only — once any user exists, these values are ignored. Partial or invalid
bootstrap configuration (an empty username, or a password under eight runes) is
a startup error while setup is required.

The password from env bootstrap is argon2id-hashed and never persisted or
returned in plaintext. After the first admin is created, remove
`NVR_ADMIN_PASSWORD` from the environment so it is not resident in the process
or the deployment config.

## Passwords

The backend hashes passwords with Argon2id. The functions are in `server/internal/auth/password.go`. They use the `argon2id` defaults.

## Sessions

A session uses an opaque token. The session functions are in `server/internal/auth/session.go`.

A session token is 32 random bytes. The backend stores only the SHA-256 hash of the token. The session TTL is 30 days.

The session is stored in the `sessions` table. Refer to [database](./database.md) for the schema.

## Session Transport

The backend accepts the session in three forms:

- The `nvr_session` cookie.
- The `Authorization: Bearer` header.
- The `X-Session-Token` header.

The cookie takes precedence when more than one form is present.

The cookie is `HttpOnly`, `Path=/`, `SameSite=Lax`, and `Secure=false`. The source comments say a reverse proxy terminates TLS. For production, shield the service with a reverse proxy for HTTPS.

The backend issues the session through both the cookie and the `X-Session-Token` header. This supports the mobile and API clients.

## The Request Principal

The `SessionMiddleware` resolves the session token and attaches the user to the request context. The user type is `auth.User`, with the ID, the username, and the role.

An expired session is deleted lazily during lookup.

## Roles

The backend has three roles:

- `admin`: full access.
- `operator`: camera management and probing.
- `viewer`: read-only access.

The role constants are in `server/internal/auth/session.go`.

## Authorization Rules

The authorization helpers are in `server/internal/api/authz.go`. The helpers are `requireUser`, `requireOperator`, and `requireAdmin`.

The rules are:

| Action | Required role |
|---|---|
| User administration | `admin` |
| Settings writes | `admin` |
| Google Drive administration | `admin` |
| Camera create, update, delete, probe | `operator` |
| Camera reads, live, snapshots | any authenticated user |
| Recordings, timeline, playback | any authenticated user |
| Events list and acknowledge | any authenticated user |
| System reads | any authenticated user |

## Logout

`POST /auth/logout` deletes the session and clears the cookie. The client also clears its local session token.