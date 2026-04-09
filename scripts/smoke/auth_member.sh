#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPTS_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
ENV_FILE="${SCRIPTS_DIR}/.env"

if [[ -f "${ENV_FILE}" ]]; then
  set -a
  # shellcheck disable=SC1090
  source "${ENV_FILE}"
  set +a
fi

BASE_URL="${BASE_URL:-${BASE:-http://127.0.0.1:9090}}"
USER_EMAIL="${USER_EMAIL:-${MEMBER_EMAIL:-}}"
USER_PASSWORD="${USER_PASSWORD:-${MEMBER_PASSWORD:-}}"

: "${BASE_URL:?Missing base url (.env: BASE_URL or BASE)}"
: "${USER_EMAIL:?Missing user email (.env: USER_EMAIL or MEMBER_EMAIL)}"
: "${USER_PASSWORD:?Missing user password (.env: USER_PASSWORD or MEMBER_PASSWORD)}"

command -v jq >/dev/null 2>&1 || { echo "jq non installato" >&2; exit 1; }

payload="$(jq -n --arg identity "$USER_EMAIL" --arg password "$USER_PASSWORD" \
  '{identity:$identity, password:$password}')"

response="$(curl -sS -X POST "$BASE_URL/api/collections/users/auth-with-password" \
  -H "Content-Type: application/json" \
  -d "$payload")"

token="$(printf '%s' "$response" | jq -r '.token // empty')"

[ -n "$token" ] || { echo "Token utente non trovato"; exit 1; }

echo "$token"
