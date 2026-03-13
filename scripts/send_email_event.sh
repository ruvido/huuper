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

BASE_URL="${BASE_URL:-${BASE:-http://127.0.0.1:9090}}"
ADMIN_EMAIL="${ADMIN_EMAIL:-}"
EVENT_SLUG="${1:-}"
EMAIL_FILE="${2:-}"

usage() {
  echo "Use: $0 <event-slug> <file.md>" >&2
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "Comando richiesto non trovato: $1" >&2
    exit 1
  }
}

parse_subject() {
  awk '
    BEGIN {
      in_frontmatter = 0
      seen_frontmatter = 0
    }
    {
      line = $0
      sub(/\r$/, "", line)
      if (line ~ /^[[:space:]]*$/) {
        next
      }
      if (!seen_frontmatter) {
        if (line !~ /^[[:space:]]*---[[:space:]]*$/) {
          exit
        }
        in_frontmatter = 1
        seen_frontmatter = 1
        next
      }
      if (in_frontmatter && line ~ /^[[:space:]]*---[[:space:]]*$/) {
        exit
      }
      if (in_frontmatter && line ~ /^[[:space:]]*subject:[[:space:]]+/) {
        sub(/^[[:space:]]*subject:[[:space:]]+/, "", line)
        print line
      }
    }
  ' "$1"
}

parse_body() {
  awk '
    BEGIN {
      in_frontmatter = 0
      frontmatter_closed = 0
    }
    {
      raw = $0
      line = raw
      sub(/\r$/, "", line)
      if (!in_frontmatter && !frontmatter_closed) {
        if (line ~ /^[[:space:]]*$/) {
          next
        }
        if (line ~ /^[[:space:]]*---[[:space:]]*$/) {
          in_frontmatter = 1
          next
        }
        exit
      }
      if (in_frontmatter) {
        if (line ~ /^[[:space:]]*---[[:space:]]*$/) {
          in_frontmatter = 0
          frontmatter_closed = 1
          next
        }
        next
      }
      print raw
    }
  ' "$1"
}

json_value() {
  local json="$1"
  local query="$2"
  printf '%s' "$json" | jq -r "$query"
}

if [[ -z "${EVENT_SLUG}" || -z "${EMAIL_FILE}" ]]; then
  usage
  exit 1
fi

if [[ ! -f "${EMAIL_FILE}" ]]; then
  echo "File non trovato: ${EMAIL_FILE}" >&2
  exit 1
fi

require_cmd curl
require_cmd jq

: "${BASE_URL:?Missing base url (.env: BASE_URL or BASE)}"
: "${ADMIN_EMAIL:?Missing admin email (.env: ADMIN_EMAIL)}"

SUBJECT="$(parse_subject "${EMAIL_FILE}")"
BODY="$(parse_body "${EMAIL_FILE}")"

if [[ -z "${SUBJECT}" ]]; then
  echo "Subject mancante: usa YAML front matter con '---', 'subject: ...', '---'" >&2
  exit 1
fi

if [[ -z "$(printf '%s' "${BODY}" | tr -d '[:space:]')" ]]; then
  echo "Body email vuoto in ${EMAIL_FILE}" >&2
  exit 1
fi

BODY="md:
${BODY}"

TOKEN="$("${SCRIPT_DIR}/smoke/auth_admin.sh")"

event_response="$(
  curl -sS --get "${BASE_URL}/api/collections/events/records" \
    -H "Authorization: Bearer ${TOKEN}" \
    --data-urlencode "filter=slug='${EVENT_SLUG}'" \
    --data-urlencode "fields=id,slug,title" \
    --data-urlencode "perPage=1"
)"

EVENT_ID="$(json_value "${event_response}" '.items[0].id // empty')"
EVENT_TITLE="$(json_value "${event_response}" '.items[0].title // empty')"

if [[ -z "${EVENT_ID}" ]]; then
  echo "Evento non trovato per slug: ${EVENT_SLUG}" >&2
  exit 1
fi

echo "Evento: ${EVENT_TITLE:-"(senza titolo)"} [${EVENT_ID}]"
echo "Slug: ${EVENT_SLUG}"
echo "Test email su ADMIN_EMAIL: ${ADMIN_EMAIL}"
echo "Subject: ${SUBJECT}"

test_payload="$(
  jq -n \
    --arg subject "${SUBJECT}" \
    --arg body "${BODY}" \
    '{subject:$subject, body:$body, target:"all", dry_run:true}'
)"

test_response="$(
  curl -sS -X POST "${BASE_URL}/api/admin/events/${EVENT_ID}/email" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer ${TOKEN}" \
    -d "${test_payload}"
)"

echo "Test response:"
printf '%s\n' "${test_response}" | jq .

read -r -p "Send to all for event '${EVENT_SLUG}'? [yes/NO] " confirm
if [[ "${confirm}" != "yes" ]]; then
  echo "Invio annullato."
  exit 0
fi

live_payload="$(
  jq -n \
    --arg subject "${SUBJECT}" \
    --arg body "${BODY}" \
    '{subject:$subject, body:$body, target:"all", dry_run:false}'
)"

live_response="$(
  curl -sS -X POST "${BASE_URL}/api/admin/events/${EVENT_ID}/email" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer ${TOKEN}" \
    -d "${live_payload}"
)"

echo "Live response:"
printf '%s\n' "${live_response}" | jq .
