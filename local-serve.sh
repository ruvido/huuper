#!/bin/bash

# Exit on error
set -euo pipefail

# Kill any process on port 8090
# lsof -ti:8090 | xargs kill -9 2>/dev/null || true

echo "🔨 Building Svelte frontend..."
pushd frontend >/dev/null

if ! command -v bun >/dev/null 2>&1; then
  echo "❌ bun non trovato. Installa bun e riprova." >&2
  exit 1
fi

echo "📦 Installing frontend dependencies..."
bun install

bun run build
popd >/dev/null

echo "🚀 Starting PocketBase server..."
go run . serve
