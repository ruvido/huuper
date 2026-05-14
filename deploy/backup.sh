#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
VPS_HOST="${VPS_HOST:-fiber}"
VPS_PATH="${VPS_PATH:-/home/ruvido/apps/huuper}"
SERVICE_NAME="${SERVICE_NAME:-huuper}"
DOCKER_COMPOSE_CMD="${DOCKER_COMPOSE_CMD:-docker compose}"
DEPLOY_COMPOSE_FILE="${DEPLOY_COMPOSE_FILE:-docker-compose.yml}"
APP_HOST_PORT="${APP_HOST_PORT:-8090}"

BACKUP_ID="${BACKUP_ID:-$(date +%Y%m%d-%H%M%S)}"
ROLLBACK_DIR="$ROOT_DIR/rollback/data-$BACKUP_ID"
BACKUPS_DIR="$ROOT_DIR/backups"

echo "backup: $BACKUP_ID"
echo "source:   $VPS_HOST:$VPS_PATH/shared/data"
echo "rollback: $ROLLBACK_DIR  (live db + storage)"
echo "backups:  $BACKUPS_DIR   (pocketbase zip archives)"

echo "remote: check docker permissions"
ssh "$VPS_HOST" "docker info >/dev/null 2>&1 || (echo 'Docker non accessibile senza sudo.' >&2; exit 1)"

echo "remote: verify data dir exists"
ssh "$VPS_HOST" "test -d '$VPS_PATH/shared/data' || (echo 'Missing $VPS_PATH/shared/data' >&2; exit 1)"

mkdir -p "$ROLLBACK_DIR" "$BACKUPS_DIR"

cleanup_started=0
restart_service() {
  if [ "$cleanup_started" -eq 1 ]; then
    return
  fi
  cleanup_started=1
  echo "docker: start $SERVICE_NAME"
  ssh "$VPS_HOST" "cd '$VPS_PATH/deploy' && $DOCKER_COMPOSE_CMD -f '$DEPLOY_COMPOSE_FILE' start $SERVICE_NAME" || true

  echo "health: wait for http :$APP_HOST_PORT"
  if ! ssh "$VPS_HOST" "for i in \$(seq 1 30); do if curl -fsS 'http://127.0.0.1:$APP_HOST_PORT/api/health' >/dev/null; then exit 0; fi; sleep 1; done; exit 1"; then
    echo "healthcheck failed, dumping remote status/logs" >&2
    ssh "$VPS_HOST" "cd '$VPS_PATH/deploy' && $DOCKER_COMPOSE_CMD -f '$DEPLOY_COMPOSE_FILE' ps" >&2 || true
    ssh "$VPS_HOST" "cd '$VPS_PATH/deploy' && $DOCKER_COMPOSE_CMD -f '$DEPLOY_COMPOSE_FILE' logs --tail=200 '$SERVICE_NAME'" >&2 || true
  fi
}
trap restart_service EXIT

echo "docker: stop $SERVICE_NAME"
ssh "$VPS_HOST" "cd '$VPS_PATH/deploy' && $DOCKER_COMPOSE_CMD -f '$DEPLOY_COMPOSE_FILE' stop $SERVICE_NAME"

echo "rsync: live db -> $ROLLBACK_DIR/"
rsync -avz --progress --delete \
  --exclude='backups/' --exclude='backups/**' \
  "$VPS_HOST:$VPS_PATH/shared/data/" \
  "$ROLLBACK_DIR/"

echo "rsync: pocketbase archives -> $BACKUPS_DIR/"
rsync -avz --progress \
  "$VPS_HOST:$VPS_PATH/shared/data/backups/" \
  "$BACKUPS_DIR/"

echo "ok: backup completed"
echo "  rollback: $ROLLBACK_DIR"
echo "  backups:  $BACKUPS_DIR"
