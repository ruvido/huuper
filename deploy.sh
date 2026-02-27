#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
VPS_HOST="${VPS_HOST:-fiber}"
VPS_PATH="${VPS_PATH:-/home/ruvido/dev/huuper}"
SERVICE_NAME="${SERVICE_NAME:-huuper}"
BIN_NAME="${BIN_NAME:-huuper}"
BIN_DIR="${BIN_DIR:-bin}"
BIN_PATH="$ROOT_DIR/$BIN_DIR/$BIN_NAME"

echo "remote: prepare path"
ssh "$VPS_HOST" "mkdir -p '$VPS_PATH' '$VPS_PATH/$BIN_DIR'"

echo "frontend: build"
cd "$ROOT_DIR/frontend"
npm ci
npm run build
cd "$ROOT_DIR"

echo "backend: build linux binary"
mkdir -p "$ROOT_DIR/$BIN_DIR"
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o "$BIN_PATH" main.go

echo "docker: down"
ssh "$VPS_HOST" "cd '$VPS_PATH' && docker compose down || true"

echo "rsync: runtime files -> $VPS_HOST:$VPS_PATH"
rsync -avz --progress "$ROOT_DIR/docker-compose.yml" "$ROOT_DIR/Dockerfile" "$VPS_HOST:$VPS_PATH/"
rsync -avz --progress --delete "$ROOT_DIR/$BIN_DIR/" "$VPS_HOST:$VPS_PATH/$BIN_DIR/"
rsync -avz --progress --delete "$ROOT_DIR/migrations/" "$VPS_HOST:$VPS_PATH/migrations/"
rsync -avz --progress --delete "$ROOT_DIR/pb_public/" "$VPS_HOST:$VPS_PATH/pb_public/"

echo "docker: up -d --build --force-recreate"
ssh "$VPS_HOST" "cd '$VPS_PATH' && docker compose up -d --build --force-recreate $SERVICE_NAME"

echo "done"
