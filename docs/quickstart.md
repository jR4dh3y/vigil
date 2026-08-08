# Quickstart

This quickstart runs Vigil, creates the first administrator, and verifies the setup. It gives the commands for development and for production.

## Prerequisites

To run Vigil in development, you need:

- Bun (version 1.3.14).
- Go (version 1.26 or later).
- MediaMTX.
- FFmpeg.

The repository uses Bun for the JavaScript packages. The backend uses Go. MediaMTX is the media server. FFmpeg is a tool for snapshots and probing.

## Install The Dependencies

Install the JavaScript dependencies with Bun:

```bash
bun install
```

MediaMTX is available as a local binary in `.bin/mediamtx`. The launcher in `tools/mediamtx/run.sh` uses this binary first. It falls back to `mediamtx` from the PATH.

## 1. Run In Development

The root package has a `dev` script that runs all development tasks:

```bash
bun run dev
```

You can run one part at a time:

```bash
bun run dev:server      # Go API on :8080
bun run dev:mediamtx    # MediaMTX
bun run dev:dashboard   # Dashboard on :5173
bun run dev:landing     # Landing page
bun run dev:mobile      # Mobile app
```

The dashboard dev server proxies the API and the MediaMTX streams. The proxy is configured in `apps/dashboard/vite.config.ts`.

## 2. Bootstrap The First Admin

When you run the server for the first time, there are no users. The system asks you to create the first administrator.

Open the dashboard in a browser. You will see the setup form. Create an administrator account. Then log in.

See [setup-and-authentication](./setup-and-authentication.md) for the details of first-start setup.

## 3. Verify The Server

To verify that the server runs, open this URL in a browser:

```text
http://localhost:8080/api/v1/health
```

The response is a JSON object with a `status` field, for example
`{"status":"ok"}`.

To verify that MediaMTX runs, open its API:

```text
http://127.0.0.1:9997/v3/paths/list
```

## 4. Build The Production Binary

To build the production binary:

```bash
bun run build --filter=@nvr/dashboard
cp -a apps/dashboard/build/. server/internal/ui/dist/
cd server && go build -o bin/nvrd ./cmd/nvrd
./bin/nvrd
```

The dashboard is built first. Then the built files are copied into the Go binary. The single binary serves the API and the dashboard on `:8080`.

### Headless (hosted dashboard) binary

If you run the dashboard as a separate hosted service and point the recorder at
it, build the slim/headless binary. It omits the embedded dashboard and serves
a connection page that deep-links to the hosted dashboard:

```bash
cd server && make build-slim        # produces bin/nvrd-slim
```

The hosted dashboard, the recorder's public HTTPS URL, and the allowed CORS
origins are configured at first run with `nvrd setup`. See
[setup-and-authentication](./setup-and-authentication.md).

## 5. Bootstrap With `nvrd setup` (CLI)

`nvrd setup` creates the first administrator and persists the public and hosted
dashboard URLs. The password is never shown on the command line. Interactive:

```bash
./bin/nvrd setup \
  --public-url https://recorder.example.com \
  --hosted-url https://nvr.example.com/dashboard
```

It prompts for the admin username and password (hidden input). Non-interactive
(first-start automation):

```bash
cd server && go run ./cmd/nvrd setup \
  --username admin \
  --password-stdin \
  --public-url https://recorder.example.com \
  --hosted-url https://nvr.example.com/dashboard \
  --non-interactive <<< 'a-strong-password'
```

`--password-stdin` reads the password from standard input instead of a terminal
prompt, so it never appears in `argv`. Once an admin exists, `nvrd setup` rejects
credential-bearing reruns and only allows URL-only updates. See
[setup-and-authentication](./setup-and-authentication.md) for all flags, the env
first-start bootstrap, and the validation rules.

## 6. Run With Docker

To run Vigil with Docker:

```bash
cd deploy && docker compose up --build
```

For the headless/hosted-dashboard image (no embedded UI):

```bash
cd deploy && docker build --target runtime-slim -t nvr-slim ..
```

See [operations](./operations.md) for the details of the Docker setup.

## Code Generation

The API contract generates the Go server stubs and the TypeScript client. To regenerate the TypeScript client:

```bash
bun run gen:api
```

To regenerate the Go server:

```bash
bun run gen:server
```

See [development](./development.md) for the details of code generation.

## Verify The Setup

Run these checks:

```bash
bun run lint          # Biome format and lint
bun run check-types   # TypeScript type checks
```

The Go backend has its own build and test commands inside `server/`. See [development](./development.md) for the details.

## Next Steps

- Read [architecture](./architecture.md) for the system design.
- Read [backend](./backend.md) for the Go server.
- Read [api](./api.md) for the REST API.
- Read [operations](./operations.md) for the deployment options.