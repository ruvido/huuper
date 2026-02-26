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
EVENT_ID="${1:-}"     # $1 = event id
EMAIL_FILE="${2:-}"   # $2 = file yaml
SEND_FLAG="${3:-}"
DRY_RUN=true

if [[ -z "$EVENT_ID" || -z "$EMAIL_FILE" ]]; then
  echo "Use: $0 <event_id> <email.yml> [--send]" >&2
  exit 1
fi

if [[ ! -f "$EMAIL_FILE" ]]; then
  echo "File non trovato: $EMAIL_FILE" >&2
  exit 1
fi

if [[ -n "$SEND_FLAG" && "$SEND_FLAG" != "--send" ]]; then
  echo "Use: $0 <event_id> <email.yml> [--send]" >&2
  exit 1
fi

if [[ "$SEND_FLAG" == "--send" ]]; then
  DRY_RUN=false
fi

command -v yq >/dev/null 2>&1 || { echo "yq non installato" >&2; exit 1; }

# -e/-r: fallisce se campo manca o è null
SUBJECT="$(yq -er '.subject' "$EMAIL_FILE")" || { echo "Campo 'subject' mancante" >&2; exit 1; }
BODY="$(yq -er '.body' "$EMAIL_FILE")" || { echo "Campo 'body' mancante" >&2; exit 1; }

# opzionale: evita stringhe vuote
[[ -n "$SUBJECT" ]] || { echo "Campo 'subject' vuoto" >&2; exit 1; }
[[ -n "$BODY" ]] || { echo "Campo 'body' vuoto" >&2; exit 1; }

TOKEN="$($SCRIPT_DIR/get_admin_token.sh)"

PAYLOAD="$(jq -n \
  --arg subject "$SUBJECT" \
  --arg body "$BODY" \
  --argjson dry_run "$DRY_RUN" \
  '{subject:$subject, body:$body, target:"active", dry_run:$dry_run}')"

curl -sS -X POST "${BASE}/api/admin/events/${EVENT_ID}/email" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d "$PAYLOAD"

# curl -sS -X POST "${BASE}/api/admin/events/${EVENT_ID}/email" \
#   -H "Content-Type: application/json" \
#   -H "Authorization: Bearer ${TOKEN}" \
#   -d '{"subject":"Test 2 invio evento","body":"Test","target":"all","dry_run":true}'
#
