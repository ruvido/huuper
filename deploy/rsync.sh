#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
VPS_HOST="${VPS_HOST:-fiber}"
VPS_PATH="${VPS_PATH:-/home/ruvido/apps/huuper}"
SERVICE_NAME="${SERVICE_NAME:-huuper}"
DOCKER_COMPOSE_CMD="${DOCKER_COMPOSE_CMD:-docker compose}"
BIN_NAME="${BIN_NAME:-huuper}"
APP_HOST_PORT="${APP_HOST_PORT:-8090}"
TARGET_GOOS="${TARGET_GOOS:-linux}"
TARGET_GOARCH="${TARGET_GOARCH:-amd64}"

RELEASE_ID="${RELEASE_ID:-$(date +%Y%m%d-%H%M%S)-$(git -C "$ROOT_DIR" rev-parse --short HEAD)}"
TMP_RELEASE_DIR="/tmp/huuper-release-$RELEASE_ID"

echo "release: $RELEASE_ID"
echo "remote: prepare release layout"
ssh "$VPS_HOST" "mkdir -p '$VPS_PATH/releases/$RELEASE_ID' '$VPS_PATH/deploy' '$VPS_PATH/shared/data'"

if [ ! -f "$ROOT_DIR/.env" ]; then
  echo "missing .env in project root" >&2
  exit 1
fi

echo "remote: check docker permissions"
ssh "$VPS_HOST" "docker info >/dev/null 2>&1 || (echo 'Docker non accessibile senza sudo.' >&2; exit 1)"

if [[ ! -d "$ROOT_DIR/frontend/site" ]]; then
  echo "missing frontend/site in project root" >&2
  exit 1
fi

echo "backend: build linux binary"
mkdir -p "$TMP_RELEASE_DIR/bin" "$TMP_RELEASE_DIR/backend/migrations" "$TMP_RELEASE_DIR/frontend/site"
cd "$ROOT_DIR"
CGO_ENABLED=0 GOOS="$TARGET_GOOS" GOARCH="$TARGET_GOARCH" go build -a -installsuffix cgo -o "$TMP_RELEASE_DIR/bin/$BIN_NAME" ./backend

echo "prepare: copy runtime artifacts"
rsync -a --delete "$ROOT_DIR/backend/migrations/" "$TMP_RELEASE_DIR/backend/migrations/"
rsync -a --delete "$ROOT_DIR/frontend/site/" "$TMP_RELEASE_DIR/frontend/site/"

echo "rsync: release -> $VPS_HOST:$VPS_PATH/releases/$RELEASE_ID/"
rsync -avz --progress --delete \
  "$TMP_RELEASE_DIR/" \
  "$VPS_HOST:$VPS_PATH/releases/$RELEASE_ID/"

echo "rsync: deploy config -> $VPS_HOST:$VPS_PATH/deploy/"
rsync -avz --progress --delete \
  "$ROOT_DIR/deploy/docker-compose.yml" \
  "$ROOT_DIR/deploy/Dockerfile" \
  "$VPS_HOST:$VPS_PATH/deploy/"

echo "rsync: shared .env -> $VPS_HOST:$VPS_PATH/shared/.env"
rsync -avz --progress \
  "$ROOT_DIR/.env" \
  "$VPS_HOST:$VPS_PATH/shared/.env"

echo "remote: validate shared env"
ssh "$VPS_HOST" "test -f '$VPS_PATH/shared/.env' || (echo 'Missing $VPS_PATH/shared/.env' >&2; exit 1)"

echo "remote: switch current release"
ssh "$VPS_HOST" "ln -sfn '$VPS_PATH/releases/$RELEASE_ID' '$VPS_PATH/current'"

echo "docker: remove stale container name if present"
ssh "$VPS_HOST" "docker rm -f '$SERVICE_NAME' >/dev/null 2>&1 || true"

echo "docker: up -d --build --force-recreate"
ssh "$VPS_HOST" "cd '$VPS_PATH/deploy' && $DOCKER_COMPOSE_CMD -f docker-compose.yml up -d --build --force-recreate $SERVICE_NAME"

echo "health: wait for container"
ssh "$VPS_HOST" "cd '$VPS_PATH/deploy' && $DOCKER_COMPOSE_CMD -f docker-compose.yml ps --status running --services | grep -qx '$SERVICE_NAME'"

echo "health: wait for http :$APP_HOST_PORT"
if ! ssh "$VPS_HOST" "for i in \$(seq 1 30); do if curl -fsS 'http://127.0.0.1:$APP_HOST_PORT/api/health' >/dev/null; then exit 0; fi; sleep 1; done; exit 1"; then
  echo "healthcheck failed, dumping remote status/logs"
  ssh "$VPS_HOST" "cd '$VPS_PATH/deploy' && $DOCKER_COMPOSE_CMD -f docker-compose.yml ps"
  ssh "$VPS_HOST" "cd '$VPS_PATH/deploy' && $DOCKER_COMPOSE_CMD -f docker-compose.yml logs --tail=200 '$SERVICE_NAME'"
  exit 1
fi

echo "cleanup: remove local temp release"
rm -rf "$TMP_RELEASE_DIR"

echo "ok: deploy completed ($RELEASE_ID)"
