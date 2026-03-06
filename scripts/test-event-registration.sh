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
EVENT_ID="${1:-i48zu71lysxrhi9}"
# EMAIL="${2:-test+$(date +%s)@realmen.it}"
EMAIL="test123@realmen.it"
FULL_NAME="${3:-Ruvido}"
MOBILE="${4:-3400000000}"
REGION="${5:-nord-est}"
MARITAL_STATUS="${6:-single}"
AGE_RANGE="${7:-25-32}"

: "${BASE:?Missing base url (.env: BASE)}"

command -v jq >/dev/null 2>&1 || { echo "jq non installato" >&2; exit 1; }

EVENT_JSON="$(curl -sS "${BASE}/api/collections/events/records/${EVENT_ID}")"
SLUG="$(printf '%s' "$EVENT_JSON" | jq -er '.slug')" || {
  echo "Evento non trovato o slug mancante: ${EVENT_ID}" >&2
  exit 1
}

PAYLOAD="$(jq -n \
  --arg email "$EMAIL" \
  --arg full_name "$FULL_NAME" \
  --arg mobile "$MOBILE" \
  --arg region "$REGION" \
  --arg marital_status "$MARITAL_STATUS" \
  --arg age_range "$AGE_RANGE" \
  '{
    email:$email,
    data:{
      full_name:$full_name,
      mobile:$mobile,
      region:$region,
      marital_status:$marital_status,
      age_range:$age_range
    }
  }')"

curl -sS -X POST "${BASE}/api/events/${SLUG}/register" \
  -H "Content-Type: application/json" \
  -d "$PAYLOAD"
