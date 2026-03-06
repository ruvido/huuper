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

# Fill IDs before running.
BASE_URL="${BASE:-http://localhost:8090}"
ADMIN_TOKEN="${ADMIN_TOKEN:-}"
REQ_ID="k300zyr0bjxqoda"
GROUP_ID="brccrrq8457c1ys"
GUARDIAN_ID="h4874mk9ix3bybb"
WRONG_GUARDIAN_ID="qejvhfr1r3mywco"

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

if [[ -z "${REQ_ID}" || -z "${GROUP_ID}" || -z "${GUARDIAN_ID}" ]]; then
  echo "Set REQ_ID, GROUP_ID, GUARDIAN_ID at the top of this script."
  exit 1
fi

echo "== 2.a FAIL expected: missing group for 2-group_assigned =="
curl -sS -X POST "${BASE_URL}/api/requests/${REQ_ID}/action" \
  -H "Authorization: ${ADMIN_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"action":"transition","target_status":"2-group_assigned"}'
echo
echo

echo "== 2.b OK expected: 2-group_assigned with group =="
curl -sS -X POST "${BASE_URL}/api/requests/${REQ_ID}/action" \
  -H "Authorization: ${ADMIN_TOKEN}" \
  -H "Content-Type: application/json" \
  -d "{\"action\":\"transition\",\"target_status\":\"2-group_assigned\",\"group\":\"${GROUP_ID}\"}"
echo
echo

echo "== 2.c FAIL expected: missing guardian for 3-guardian_assigned =="
curl -sS -X POST "${BASE_URL}/api/requests/${REQ_ID}/action" \
  -H "Authorization: ${ADMIN_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"action":"transition","target_status":"3-guardian_assigned"}'
echo
echo

if [[ -n "${WRONG_GUARDIAN_ID}" ]]; then
  echo "== 2.d FAIL expected: wrong guardian not in group =="
  curl -sS -X POST "${BASE_URL}/api/requests/${REQ_ID}/action" \
    -H "Authorization: ${ADMIN_TOKEN}" \
    -H "Content-Type: application/json" \
    -d "{\"action\":\"transition\",\"target_status\":\"3-guardian_assigned\",\"guardian\":\"${WRONG_GUARDIAN_ID}\"}"
  echo
  echo
fi

echo "== 2.e OK expected: valid guardian in group =="
curl -sS -X POST "${BASE_URL}/api/requests/${REQ_ID}/action" \
  -H "Authorization: ${ADMIN_TOKEN}" \
  -H "Content-Type: application/json" \
  -d "{\"action\":\"transition\",\"target_status\":\"3-guardian_assigned\",\"guardian\":\"${GUARDIAN_ID}\"}"
echo
