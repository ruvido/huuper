#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${SCRIPT_DIR}/.env"

if [[ -f "${ENV_FILE}" ]]; then
  set -a
  # shellcheck disable=SC1090
  source "${ENV_FILE}"
  set +a
fi

BASE="${BASE:-http://127.0.0.1:8090}"
ADMIN_EMAIL="${ADMIN_EMAIL:-}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-}"

: "${ADMIN_EMAIL:?Missing admin email (.env: ADMIN_EMAIL)}"
: "${ADMIN_PASSWORD:?Missing admin password (.env: ADMIN_PASSWORD)}"

curl -sS -X POST "$BASE/api/collections/users/auth-with-password" \
  -H "Content-Type: application/json" \
  -d "{\"identity\":\"${ADMIN_EMAIL}\",\"password\":\"${ADMIN_PASSWORD}\"}"
