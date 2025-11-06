#!/bin/bash

# Exit on error
set -e

echo "🔨 Building Svelte frontend..."
cd frontend && npm run build && cd ..

echo "🚀 Starting PocketBase server..."
go run . serve
