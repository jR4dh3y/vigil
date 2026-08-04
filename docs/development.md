# Development

This document describes the repository layout, the build, the code generation, and the quality checks. The repository is a Bun monorepo with Turborepo.

## Repository Layout

```text
server/                Go backend (nvrd)
  cmd/nvrd/            server entrypoint
  internal/            backend packages
  migrations/          SQL migrations
  api/                 OpenAPI contract
apps/
  dashboard/           SvelteKit web dashboard
  mobile/              Expo mobile app
  landing/             Astro marketing website
packages/
  api-client/          TypeScript API client
  typescript-config/   shared TypeScript configs
deploy/                Docker and MediaMTX configs
tools/                 local tools and launchers
recordings/            recorded media files (runtime data)
docs/                  this documentation
```

## The Package Manager

The repository uses Bun as the package manager. The package manager version is pinned in the root `package.json`:

```text
packageManager: bun@1.3.14
```

Install the packages with:

```bash
bun install
```

Install new packages with the Bun CLI. Open `package.json` and the workspace manifests to see the package structure.

## The Workspaces

The root `package.json` defines the workspaces:

```json
"workspaces": [
  "apps/*",
  "packages/*",
  "server",
  "tools/*"
]
```

The workspaces are:

- `apps/dashboard`: the SvelteKit dashboard.
- `apps/mobile`: the Expo mobile app.
- `apps/landing`: the Astro landing page.
- `packages/api-client`: the TypeScript API client.
- `packages/typescript-config`: the shared TypeScript configs.
- `server`: the Go backend.
- `tools/mediamtx`: the MediaMTX launcher.

## The Root Scripts

The root `package.json` provides these scripts:

| Script | Purpose |
|---|---|
| `build` | Build all workspaces. |
| `dev` | Run all development tasks. |
| `lint` | Run Biome format and lint. |
| `check-types` | Run the workspace type checks. |
| `format` | Rewrite files with Biome formatting. |
| `dev:dashboard` | Run the dashboard only. |
| `dev:landing` | Run the landing page only. |
| `dev:mobile` | Run the mobile app only. |
| `dev:server` | Run the Go server only. |
| `dev:mediamtx` | Run the MediaMTX launcher only. |
| `gen:api` | Regenerate the TypeScript API client. |
| `gen:server` | Regenerate the Go server. |
| `lint:fix` | Apply Biome fixes. |
| `check` | Biome check alias. |

## Turborepo

The Turborepo configuration is in `turbo.json`. It defines the shared tasks:

- `build`: depends on upstream builds.
- `lint`: depends on upstream lint.
- `check-types`: depends on upstream type checks.
- `dev`: no cache and persistent.
- `gen`: no cache and outputs to `src/generated/**`.

## Code Generation

The repository generates code from the OpenAPI contract.

### The TypeScript API Client

The `packages/api-client` package generates its types from the OpenAPI contract. The command is:

```bash
bun run gen:api
```

The generation script runs `openapi-typescript` on `server/api/openapi.yaml`. The output is `packages/api-client/src/generated/schema.ts`.

The build runs the generation script. The package exports the generated types and a client factory.

### The Go Server

The Go server generates its stubs from the OpenAPI contract and the sqlc queries. The commands are in `server/Makefile`:

```bash
# Regenerate the API server stubs and the sqlc queries
cd server && make generate

# Regenerate only the API server stubs
make oapi

# Regenerate only the sqlc queries
make sqlc
```

The `oapi-codegen` and `sqlc` command-line tools must be installed. The root script `gen:server` runs `make generate`.

## Linting And Formatting

The repository uses Biome for linting and formatting. The configuration is in `biome.json`.

Biome uses:

- Tabs for indentation.
- A line width of 100.
- Double quotes for JavaScript.
- Required semicolons.
- The recommended linter preset.

Biome excludes:

- `node_modules`.
- Build output.
- The Go `server` tree.
- The generated API schema.
- Svelte files.

The Astro override disables two unused-code rules for Astro files.

Run the checks with:

```bash
bun run lint
bun run format
```

## TypeScript Checks

The root `check-types` script runs all workspace type checks with Turborepo. Each app has its own type check:

- The dashboard uses `svelte-check`.
- The mobile app uses `tsc --noEmit`.
- The landing page uses `astro check`.
- The API client uses `tsc --noEmit`.

The shared TypeScript configs are in `packages/typescript-config`. The configs are `base.json`, `nextjs.json`, and `react-library.json`. The apps currently extend their framework configs instead.

## The Go Checks

The Go backend is outside the Biome tooling. Build and test it inside `server/`:

```bash
cd server
make build   # build the binary
go test ./...  # run the tests
```

The `server/package.json` has a `dev` script that runs the server with the development environment.

## The Local Tools

The `tools/mediamtx` package has a launcher. The `dev` script runs `bash ./run.sh`. The launcher starts MediaMTX with the deployment configuration.

The `.bin/` directory holds the local MediaMTX binary. It is ignored by Git.

## Documentation Style

Write all documentation in Simplified Technical English (STE). The rules are in `ste100-style-guide.md`. Refer to the repository `AGENTS.md` for the project conventions.

## The Quality Workflow

The recommended workflow before you finish a change:

```bash
bun run lint
bun run format
bun run check-types
cd server && go build ./... && go test ./...
```