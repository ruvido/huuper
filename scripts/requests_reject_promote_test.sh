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

BASE_URL="${BASE:-http://127.0.0.1:8090}"
ADMIN_TOKEN="${ADMIN_TOKEN:-}"

# Use two different requests:
# - REQ_ID_REJECT: request used for reject test
# - REQ_ID_PROMOTE: request used for full approve+promote test
REQ_ID_REJECT="ye537izdtlcoota"
REQ_ID_PROMOTE="5e085yu7dyux9cj"

# Optional: needed only if REQ_ID_PROMOTE starts before 3-guardian_assigned.
GROUP_ID=""
GUARDIAN_ID=""

REJECT_REASON="${REJECT_REASON:-Non idoneo dopo valutazione}"

if [[ -z "${ADMIN_TOKEN}" ]]; then
  AUTH_JSON="$("${SCRIPT_DIR}/get_admin_token.sh")"
  if command -v jq >/dev/null 2>&1; then
    ADMIN_TOKEN="$(printf '%s' "${AUTH_JSON}" | jq -r '.token // empty')"
  else
    ADMIN_TOKEN="$(printf '%s' "${AUTH_JSON}" | sed -n 's/.*"token"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p')"
  fi
fi

if [[ -z "${ADMIN_TOKEN}" ]]; then
  echo "Failed to obtain ADMIN_TOKEN from get_admin_token.sh."
  exit 1
fi

call_action() {
  local req_id="$1"
  local payload="$2"
  curl -sS -X POST "${BASE_URL}/api/requests/${req_id}/action" \
    -H "Authorization: ${ADMIN_TOKEN}" \
    -H "Content-Type: application/json" \
    -d "${payload}"
}

echo "BASE_URL=${BASE_URL}"
echo

if [[ -n "${REQ_ID_REJECT}" ]]; then
  echo "== 3) REJECT (admin-only) on REQ_ID_REJECT =="
  call_action "${REQ_ID_REJECT}" "{\"action\":\"reject\",\"reason\":\"${REJECT_REASON}\"}"
  echo
  echo
else
  echo "Skip reject flow: REQ_ID_REJECT empty."
  echo
fi

if [[ -n "${REQ_ID_PROMOTE}" ]]; then
  if [[ -n "${GROUP_ID}" && -n "${GUARDIAN_ID}" ]]; then
    echo "== 4.a) transition -> 2-group_assigned =="
    call_action "${REQ_ID_PROMOTE}" "{\"action\":\"transition\",\"target_status\":\"2-group_assigned\",\"group\":\"${GROUP_ID}\"}"
    echo
    echo

    echo "== 4.b) transition -> 3-guardian_assigned =="
    call_action "${REQ_ID_PROMOTE}" "{\"action\":\"transition\",\"target_status\":\"3-guardian_assigned\",\"guardian\":\"${GUARDIAN_ID}\"}"
    echo
    echo
  else
    echo "Skip 2-group_assigned and 3-guardian_assigned (GROUP_ID/GUARDIAN_ID not set)."
    echo "Assuming REQ_ID_PROMOTE is already at 3-guardian_assigned (or later)."
    echo
  fi

  echo "== 4.c) transition -> 4-mentoring =="
  call_action "${REQ_ID_PROMOTE}" '{"action":"transition","target_status":"4-mentoring"}'
  echo
  echo

  echo "== 4.d) transition -> 5-group_approved =="
  call_action "${REQ_ID_PROMOTE}" '{"action":"transition","target_status":"5-group_approved"}'
  echo
  echo

  echo "== 4.e) transition -> 6-admin_approved =="
  call_action "${REQ_ID_PROMOTE}" '{"action":"transition","target_status":"6-admin_approved"}'
  echo
  echo

  echo "== 4.f) promote (creates users.status=active and deletes request) =="
  call_action "${REQ_ID_PROMOTE}" '{"action":"promote"}'
  echo
else
  echo "Skip promote flow: REQ_ID_PROMOTE empty."
fi
