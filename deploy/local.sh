#!/bin/bash

# Exit on error
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="${SCRIPT_DIR}"
while [[ "${REPO_ROOT}" != "/" ]]; do
  if [[ -d "${REPO_ROOT}/frontend" && -f "${REPO_ROOT}/go.mod" ]]; then
    break
  fi
  REPO_ROOT="$(dirname "${REPO_ROOT}")"
done

if [[ ! -d "${REPO_ROOT}/frontend" || ! -f "${REPO_ROOT}/go.mod" ]]; then
  echo "❌ Impossibile trovare la root del progetto (frontend + go.mod)." >&2
  exit 1
fi

# Kill any process on port 9090
# lsof -ti:9090 | xargs kill -9 2>/dev/null || true

echo "🔨 Building Svelte frontend..."
pushd "${REPO_ROOT}/frontend" >/dev/null

if ! command -v bun >/dev/null 2>&1; then
  echo "❌ bun non trovato. Installa bun e riprova." >&2
  exit 1
fi

echo "📦 Installing frontend dependencies..."
bun install
# Ensure Vite binaries are executable (some installs can lose +x bit on macOS/networked setups)
chmod +x node_modules/.bin/vite node_modules/vite/bin/vite.js

bun run build
popd >/dev/null

echo "🚀 Starting PocketBase server..."
pushd "${REPO_ROOT}" >/dev/null
go run . serve --http=127.0.0.1:9090
