#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
VPS_HOST="${VPS_HOST:-fiber}"
VPS_PATH="${VPS_PATH:-/home/ruvido/apps/huuper}"
SERVICE_NAME="${SERVICE_NAME:-huuper}"
DOCKER_COMPOSE_CMD="${DOCKER_COMPOSE_CMD:-docker compose}"
APP_HOST_PORT="${APP_HOST_PORT:-8090}"
RELEASE_ID="${1:-}"

if [ -z "$RELEASE_ID" ]; then
  echo "Usage: ./scripts/ops/rollback.sh <release_id>" >&2
  exit 1
fi

echo "remote: verify release exists"
ssh "$VPS_HOST" "test -d '$VPS_PATH/releases/$RELEASE_ID' || (echo 'Release not found: $RELEASE_ID' >&2; exit 1)"

echo "remote: switch current release -> $RELEASE_ID"
ssh "$VPS_HOST" "ln -sfn '$VPS_PATH/releases/$RELEASE_ID' '$VPS_PATH/current'"

echo "docker: up -d --build --force-recreate"
ssh "$VPS_HOST" "cd '$VPS_PATH/deploy' && $DOCKER_COMPOSE_CMD -f docker-compose.yml up -d --build --force-recreate $SERVICE_NAME"

echo "health: wait for container"
ssh "$VPS_HOST" "cd '$VPS_PATH/deploy' && $DOCKER_COMPOSE_CMD -f docker-compose.yml ps --status running --services | grep -qx '$SERVICE_NAME'"

echo "health: wait for http :$APP_HOST_PORT"
ssh "$VPS_HOST" "for i in \$(seq 1 30); do if curl -fsS 'http://127.0.0.1:$APP_HOST_PORT/api/health' >/dev/null; then exit 0; fi; sleep 1; done; exit 1"

echo "ok: rollback completed ($RELEASE_ID)"
