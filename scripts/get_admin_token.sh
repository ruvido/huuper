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

BASE="${BASE:-}"
ADMIN_EMAIL="${ADMIN_EMAIL:-}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-}"

: "${BASE:?Missing base url (.env: BASE)}"
: "${ADMIN_EMAIL:?Missing admin email (.env: ADMIN_EMAIL)}"
: "${ADMIN_PASSWORD:?Missing admin password (.env: ADMIN_PASSWORD)}"

payload="$(jq -n --arg identity "$ADMIN_EMAIL" --arg password "$ADMIN_PASSWORD" \
    '{identity:$identity, password:$password}')"

response="$(curl -sS -X POST "$BASE/api/collections/users/auth-with-password" \
    -H "Content-Type: application/json" \
    -d "$payload")"

token="$(printf '%s' "$response" | jq -r '.token')"

[ -n "$token" ] && [ "$token" != "null" ] || { echo "Token non trovato"; exit 1; }

echo "$token"
