#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FRONTEND_DIR="$ROOT_DIR/frontend"
CACHE_DIR="$ROOT_DIR/.cache"
STAMP_FILE="$CACHE_DIR/frontend-build.sha256"
MODE="${1:-prod}"

mkdir -p "$CACHE_DIR"

if ! command -v rg >/dev/null 2>&1; then
	echo "Error: rg not found" >&2
	exit 1
fi

build_hash() {
	(
		cd "$FRONTEND_DIR"
		{
			rg --files src public
			echo "package.json"
			echo "bun.lock"
			echo "vite.config.js"
			echo "svelte.config.js"
			echo "index.html"
			echo "../VERSION"
		} | sort | while IFS= read -r file; do
			if [[ -f "$file" ]]; then
				shasum "$file"
			fi
		done
	) | shasum | awk '{print $1}'
}

CURRENT_HASH="$(build_hash)"
LAST_HASH=""
if [[ -f "$STAMP_FILE" ]]; then
	LAST_HASH="$(cat "$STAMP_FILE")"
fi

if [[ "$CURRENT_HASH" == "$LAST_HASH" ]]; then
	echo "Frontend unchanged: build skipped."
	exit 0
fi

echo "Frontend changed: running build ($MODE)..."
cd "$FRONTEND_DIR"
if [[ "$MODE" == "fast" ]]; then
	bun run build:fast
else
	bun run build
fi

echo "$CURRENT_HASH" > "$STAMP_FILE"
echo "Frontend build completed."
