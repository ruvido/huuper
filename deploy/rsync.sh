#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
VPS_HOST="${VPS_HOST:-fiber}"
VPS_PATH="${VPS_PATH:-/home/ruvido/apps/huuper}"
SERVICE_NAME="${SERVICE_NAME:-huuper}"
DOCKER_COMPOSE_CMD="${DOCKER_COMPOSE_CMD:-docker compose}"
DEPLOY_COMPOSE_FILE="${DEPLOY_COMPOSE_FILE:-docker-compose.yml}"
BIN_NAME="${BIN_NAME:-huuper}"
APP_HOST_PORT="${APP_HOST_PORT:-8090}"
TARGET_GOOS="${TARGET_GOOS:-linux}"
TARGET_GOARCH="${TARGET_GOARCH:-amd64}"
FRONTEND_ARCHIVE_DIR="${FRONTEND_ARCHIVE_DIR:-$VPS_PATH/shared/frontend-history}"
DEPLOY_URL="${DEPLOY_URL:-https://branco.realmen.it}"

DEPLOY_START_EPOCH="$(date +%s)"
RELEASE_ID="${RELEASE_ID:-$(date +%Y%m%d-%H%M%S)-$(git -C "$ROOT_DIR" rev-parse --short HEAD)}"
TMP_RELEASE_DIR="/tmp/huuper-release-$RELEASE_ID"
REMOTE_ENV_FILE="$TMP_RELEASE_DIR/shared.env"

echo "release: $RELEASE_ID"
echo "remote: prepare release layout"
ssh "$VPS_HOST" "mkdir -p '$VPS_PATH/releases/$RELEASE_ID' '$VPS_PATH/deploy' '$VPS_PATH/shared/data' '$FRONTEND_ARCHIVE_DIR'"

if [ ! -f "$ROOT_DIR/.env" ]; then
  echo "missing .env in project root" >&2
  exit 1
fi

if [ ! -f "$ROOT_DIR/deploy/$DEPLOY_COMPOSE_FILE" ]; then
  echo "missing deploy/$DEPLOY_COMPOSE_FILE in project root" >&2
  exit 1
fi

echo "remote: check docker permissions"
ssh "$VPS_HOST" "docker info >/dev/null 2>&1 || (echo 'Docker non accessibile senza sudo.' >&2; exit 1)"

cd "$ROOT_DIR"

# frontend/site is a build artifact and is gitignored, so whatever sits on disk
# is not necessarily what frontend/skeleton says. Rebuilding it here is what
# makes this script a complete deploy: without it a frontend change silently
# never ships.
echo "frontend: build skeleton -> site"
go run ./backend build-frontend

if [[ ! -d "$ROOT_DIR/frontend/site" ]]; then
  echo "frontend build produced no frontend/site" >&2
  exit 1
fi

echo "backend: build linux binary"
mkdir -p "$TMP_RELEASE_DIR/bin" "$TMP_RELEASE_DIR/backend/migrations" "$TMP_RELEASE_DIR/frontend/site"
CGO_ENABLED=0 GOOS="$TARGET_GOOS" GOARCH="$TARGET_GOARCH" go build -installsuffix cgo -o "$TMP_RELEASE_DIR/bin/$BIN_NAME" ./backend

echo "prepare: copy runtime artifacts"
rsync -a --delete "$ROOT_DIR/backend/migrations/" "$TMP_RELEASE_DIR/backend/migrations/"
rsync -a --delete "$ROOT_DIR/frontend/site/" "$TMP_RELEASE_DIR/frontend/site/"

echo "prepare: remote env URL=$DEPLOY_URL"
awk -v url="$DEPLOY_URL" '
  BEGIN { done = 0 }
  /^URL=/ {
    print "URL=" url
    done = 1
    next
  }
  { print }
  END {
    if (!done) {
      print "URL=" url
    }
  }
' "$ROOT_DIR/.env" > "$REMOTE_ENV_FILE"

echo "rsync: release -> $VPS_HOST:$VPS_PATH/releases/$RELEASE_ID/"
rsync -avz --progress --delete \
  "$TMP_RELEASE_DIR/" \
  "$VPS_HOST:$VPS_PATH/releases/$RELEASE_ID/"

echo "rsync: deploy config -> $VPS_HOST:$VPS_PATH/deploy/"
rsync -avz --progress --delete \
  "$ROOT_DIR/deploy/$DEPLOY_COMPOSE_FILE" \
  "$ROOT_DIR/deploy/Dockerfile" \
  "$VPS_HOST:$VPS_PATH/deploy/"

echo "rsync: shared .env -> $VPS_HOST:$VPS_PATH/shared/.env"
rsync -avz --progress \
  "$REMOTE_ENV_FILE" \
  "$VPS_HOST:$VPS_PATH/shared/.env"

echo "remote: validate shared env"
ssh "$VPS_HOST" "test -f '$VPS_PATH/shared/.env' || (echo 'Missing $VPS_PATH/shared/.env' >&2; exit 1)"

echo "remote: snapshot current frontend"
ssh "$VPS_HOST" "if [ -d '$VPS_PATH/current/frontend/site' ]; then mkdir -p '$FRONTEND_ARCHIVE_DIR/$RELEASE_ID'; rsync -a --delete '$VPS_PATH/current/frontend/site/' '$FRONTEND_ARCHIVE_DIR/$RELEASE_ID/'; fi"

echo "remote: switch current release"
ssh "$VPS_HOST" "ln -sfn '$VPS_PATH/releases/$RELEASE_ID' '$VPS_PATH/current'"

for attempt in 1 2 3; do
  echo "docker: remove stale container name if present (attempt $attempt)"
  ssh "$VPS_HOST" "docker rm -f '$SERVICE_NAME' >/dev/null 2>&1 || true"

  echo "docker: up -d --build --force-recreate (attempt $attempt)"
  ssh "$VPS_HOST" "cd '$VPS_PATH/deploy' && $DOCKER_COMPOSE_CMD -f '$DEPLOY_COMPOSE_FILE' up -d --build --force-recreate $SERVICE_NAME"

  echo "verify: container was actually recreated with this release"
  # Parsed on the VPS: `date -d` is GNU-only and is not available on macOS,
  # where it aborted the deploy right before the healthcheck.
  CONTAINER_CREATED_EPOCH="$(ssh "$VPS_HOST" "date -d \"\$(docker inspect '$SERVICE_NAME' --format '{{.Created}}')\" +%s")"
  if [ "$CONTAINER_CREATED_EPOCH" -ge "$DEPLOY_START_EPOCH" ]; then
    break
  fi

  echo "container was not recreated (still older than this deploy) - retrying" >&2
  if [ "$attempt" = 3 ]; then
    echo "container still not recreated after 3 attempts - giving up" >&2
    exit 1
  fi
  sleep 3
done

echo "health: wait for container"
ssh "$VPS_HOST" "cd '$VPS_PATH/deploy' && $DOCKER_COMPOSE_CMD -f '$DEPLOY_COMPOSE_FILE' ps --status running --services | grep -qx '$SERVICE_NAME'"

echo "health: wait for http :$APP_HOST_PORT"
if ! ssh "$VPS_HOST" "for i in \$(seq 1 30); do if curl -fsS 'http://127.0.0.1:$APP_HOST_PORT/api/health' >/dev/null; then exit 0; fi; sleep 1; done; exit 1"; then
  echo "healthcheck failed, dumping remote status/logs"
  ssh "$VPS_HOST" "cd '$VPS_PATH/deploy' && $DOCKER_COMPOSE_CMD -f '$DEPLOY_COMPOSE_FILE' ps"
  ssh "$VPS_HOST" "cd '$VPS_PATH/deploy' && $DOCKER_COMPOSE_CMD -f '$DEPLOY_COMPOSE_FILE' logs --tail=200 '$SERVICE_NAME'"
  exit 1
fi

echo "cleanup: remove local temp release"
rm -rf "$TMP_RELEASE_DIR"

echo "ok: deploy completed ($RELEASE_ID)"
