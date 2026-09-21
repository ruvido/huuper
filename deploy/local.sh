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

if [[ ! -d "${REPO_ROOT}/frontend/site" ]]; then
  echo "❌ frontend/site non trovato." >&2
  exit 1
fi

echo "🚀 Starting PocketBase server..."
pushd "${REPO_ROOT}" >/dev/null
go run ./backend serve --http=127.0.0.1:9090 --dir=./data --frontend-dev --disable-telegram-bot
#go run ./backend serve --http=127.0.0.1:9090 --dir=./data --frontend-dev
