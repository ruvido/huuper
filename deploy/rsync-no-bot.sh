#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# Reuse the standard deploy flow with a compose file that disables the Telegram bot.
export DEPLOY_COMPOSE_FILE="${DEPLOY_COMPOSE_FILE:-docker-compose.no-bot.yml}"

exec "$SCRIPT_DIR/rsync.sh" "$@"
