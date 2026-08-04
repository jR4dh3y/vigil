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

## Passwords

The backend hashes passwords with Argon2id. The functions are in `server/internal/auth/password.go`. They use the `argon2id` defaults.

## Sessions

A session uses an opaque token. The session functions are in `server/internal/auth/session.go`.

A session token is 32 random bytes. The backend stores only the SHA-256 hash of the token. The session TTL is 30 days.

The session is stored in the `sessions` table. Refer to `database.md` for the schema.

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