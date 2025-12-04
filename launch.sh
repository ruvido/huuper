#!/bin/bash

# Exit on error
set -e

# Kill any process on port 8090
lsof -ti:8090 | xargs kill -9 2>/dev/null || true

echo "🔨 Building Svelte frontend..."
cd frontend && npm run build && cd ..

echo "🚀 Starting PocketBase server..."
go run . serve
