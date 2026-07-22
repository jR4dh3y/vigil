#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
CFG="$ROOT/deploy/mediamtx.yml"

if [[ -x "$ROOT/.bin/mediamtx" ]]; then
	exec "$ROOT/.bin/mediamtx" "$CFG"
fi

if command -v mediamtx >/dev/null 2>&1; then
	exec mediamtx "$CFG"
fi

echo "MediaMTX not found. Install the binary at .bin/mediamtx or on PATH." >&2
echo "  See README.md (Prerequisites / Dev)." >&2
exit 1
