#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
DASHBOARD_BUILD="$ROOT_DIR/apps/dashboard/build"
UI_PARENT="$ROOT_DIR/server/internal/ui"
UI_DIST="$UI_PARENT/dist"

for command in bun go make; do
  if ! command -v "$command" >/dev/null 2>&1; then
    printf 'error: required command not found: %s\n' "$command" >&2
    exit 1
  fi
done

cd "$ROOT_DIR"
printf 'Building dashboard...\n'
bun run build --filter=@nvr/dashboard

if [[ ! -f "$DASHBOARD_BUILD/index.html" ]]; then
  printf 'error: dashboard build did not produce %s/index.html\n' "$DASHBOARD_BUILD" >&2
  exit 1
fi

staging_dist="$(mktemp -d "$UI_PARENT/.dist.XXXXXX")"
cleanup() {
  rm -rf "$staging_dist"
}
trap cleanup EXIT

cp -a "$DASHBOARD_BUILD/." "$staging_dist/"
rm -rf "$UI_DIST"
mv "$staging_dist" "$UI_DIST"
trap - EXIT

printf 'Building Go binary...\n'
make -C "$ROOT_DIR/server" build-full
printf 'Built %s/server/bin/nvrd\n' "$ROOT_DIR"
