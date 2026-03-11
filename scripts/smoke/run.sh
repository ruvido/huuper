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

BASE_URL="${BASE:-http://127.0.0.1:8090}"
ADMIN_TOKEN="${ADMIN_TOKEN:-}"
MEMBER_TOKEN="${MEMBER_TOKEN:-}"
EVENT_SLUG="${EVENT_SLUG:-}"

tmp_body="$(mktemp)"
trap 'rm -f "${tmp_body}"' EXIT

require_bin() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "Missing required command: $1" >&2
    exit 1
  }
}

get_admin_token_if_needed() {
  if [[ -n "${ADMIN_TOKEN}" ]]; then
    return
  fi
  ADMIN_TOKEN="$("${SCRIPT_DIR}/auth_admin.sh")"
}

get_member_token_if_needed() {
  if [[ -n "${MEMBER_TOKEN}" ]]; then
    return
  fi
  MEMBER_TOKEN="$("${SCRIPT_DIR}/auth_member.sh")"
}

call() {
  local method="$1"
  local url="$2"
  local auth="${3:-none}"
  local payload="${4:-}"

  local -a args
  args=(-sS -X "${method}" "${url}" -o "${tmp_body}" -w "%{http_code}")
  case "${auth}" in
    admin)
      args+=(-H "Authorization: ${ADMIN_TOKEN}")
      ;;
    member)
      args+=(-H "Authorization: ${MEMBER_TOKEN}")
      ;;
  esac
  if [[ -n "${payload}" ]]; then
    args+=(-H "Content-Type: application/json" -d "${payload}")
  fi

  curl "${args[@]}"
}

assert_code() {
  local got="$1"
  local expected="$2"
  local label="$3"

  if [[ "${got}" != "${expected}" ]]; then
    echo "FAIL ${label}: expected ${expected}, got ${got}"
    echo "Body:"
    print_body
    exit 1
  fi
  echo "OK   ${label}: ${got}"
}

print_body() {
  if command -v jq >/dev/null 2>&1; then
    jq . "${tmp_body}" 2>/dev/null || cat "${tmp_body}"
  else
    cat "${tmp_body}"
  fi
}

require_bin curl
require_bin jq

echo "BASE_URL=${BASE_URL}"
echo

echo "== 1) Public smoke =="
code="$(call GET "${BASE_URL}/api/public/settings/signup" none)"
assert_code "${code}" "200" "public settings signup"

code="$(call GET "${BASE_URL}/api/public/settings/profile_schema" none)"
assert_code "${code}" "404" "public settings profile_schema hidden"

uniq="$(date +%s)"
payload="$(cat <<JSON
{
  "full_name": "Smoke Test",
  "email": "smoke.${uniq}@example.com",
  "mobile": "+393331112233",
  "region": "Lombardia",
  "birth_year": "1990",
  "marital_status": "Single",
  "children": "No",
  "motivation": "Smoke test del backend v2."
}
JSON
)"

code="$(call POST "${BASE_URL}/api/public/requests" none "${payload}")"
assert_code "${code}" "201" "public request submit"

echo
echo "Submit response:"
print_body

echo
echo "== 2) Auth smoke =="
get_admin_token_if_needed
get_member_token_if_needed

code="$(call GET "${BASE_URL}/api/me/settings/profile_schema" none)"
assert_code "${code}" "401" "me settings requires auth"

code="$(call GET "${BASE_URL}/api/me/settings/profile_schema" member)"
assert_code "${code}" "200" "me settings with member"

code="$(call GET "${BASE_URL}/api/admin/settings/profile_schema" none)"
assert_code "${code}" "401" "admin settings requires auth"

code="$(call GET "${BASE_URL}/api/admin/settings/profile_schema" member)"
assert_code "${code}" "403" "admin settings forbidden for member"

code="$(call GET "${BASE_URL}/api/admin/settings/profile_schema" admin)"
assert_code "${code}" "200" "admin settings with admin"

code="$(call GET "${BASE_URL}/api/admin/summary" member)"
assert_code "${code}" "403" "admin summary forbidden for member"

code="$(call GET "${BASE_URL}/api/admin/summary" admin)"
assert_code "${code}" "200" "admin summary with admin"

code="$(call POST "${BASE_URL}/api/me/telegram/token" member "{}")"
assert_code "${code}" "200" "me telegram token"

echo
echo "== 3) Optional event smoke =="
if [[ -n "${EVENT_SLUG}" ]]; then
  code="$(call GET "${BASE_URL}/api/me/events/${EVENT_SLUG}/status" member)"
  assert_code "${code}" "200" "me event status"

  code="$(call POST "${BASE_URL}/api/public/events/${EVENT_SLUG}/register" none '{"email":"smoke.event.'${uniq}'@example.com","data":{"full_name":"Smoke Event"}}')"
  assert_code "${code}" "201" "public event register"
else
  echo "SKIP event smoke: set EVENT_SLUG in scripts/.env"
fi

echo
echo "Smoke test backend v2 completed."
