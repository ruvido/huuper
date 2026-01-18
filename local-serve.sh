#!/bin/bash

# Exit on error
set -euo pipefail

# Kill any process on port 8090
# lsof -ti:8090 | xargs kill -9 2>/dev/null || true

echo "🔨 Building Svelte frontend..."
pushd frontend >/dev/null

# Keep npm cache/logs inside the repo to avoid permission issues with ~/.npm.
export NPM_CONFIG_CACHE="${PWD}/.npm-cache"
export NPM_CONFIG_LOGS_DIR="${PWD}/.npm-logs"
mkdir -p "$NPM_CONFIG_CACHE" "$NPM_CONFIG_LOGS_DIR"

# Ensure frontend deps are installed (vite is a dev dependency).
if [ ! -d "node_modules" ] || ! command -v ./node_modules/.bin/vite >/dev/null 2>&1; then
  echo "📦 Installing frontend dependencies..."
  if ! npm install; then
    echo "❌ npm install failed. If you see EPERM for esbuild, the filesystem may be mounted noexec." >&2
    echo "   Try running from a writable exec-enabled path or ask to disable auto-install." >&2
    exit 1
  fi
fi

npm run build
popd >/dev/null

echo "🚀 Starting PocketBase server..."
go run . serve
