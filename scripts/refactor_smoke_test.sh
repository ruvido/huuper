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

tmp_body="$(mktemp)"
trap 'rm -f "${tmp_body}"' EXIT

get_token_if_needed() {
  if [[ -n "${ADMIN_TOKEN}" ]]; then
    return
  fi

  local auth_json
  auth_json="$("${SCRIPT_DIR}/get_admin_token.sh")"
  if command -v jq >/dev/null 2>&1; then
    ADMIN_TOKEN="$(printf '%s' "${auth_json}" | jq -r '.token // empty')"
  else
    ADMIN_TOKEN="$(printf '%s' "${auth_json}" | sed -n 's/.*"token"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p')"
  fi
}

call() {
  local method="$1"
  local url="$2"
  local auth="${3:-no}"
  local payload="${4:-}"

  local -a args
  args=(-sS -X "${method}" "${url}" -o "${tmp_body}" -w "%{http_code}")
  if [[ "${auth}" == "yes" ]]; then
    args+=(-H "Authorization: ${ADMIN_TOKEN}")
  fi
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
    cat "${tmp_body}"
    exit 1
  fi
  echo "OK   ${label}: ${got}"
}

echo "BASE_URL=${BASE_URL}"
echo

echo "== 1) Settings access checks =="
code="$(call GET "${BASE_URL}/api/settings/signup" no)"
assert_code "${code}" "200" "public signup"

code="$(call GET "${BASE_URL}/api/settings/profile_schema" no)"
assert_code "${code}" "401" "profile_schema without auth"

code="$(call GET "${BASE_URL}/api/settings/onboarding" no)"
assert_code "${code}" "401" "onboarding without auth"

get_token_if_needed
if [[ -z "${ADMIN_TOKEN}" ]]; then
  echo "FAIL cannot obtain ADMIN_TOKEN"
  exit 1
fi

code="$(call GET "${BASE_URL}/api/settings/profile_schema" yes)"
assert_code "${code}" "200" "profile_schema with auth"

code="$(call GET "${BASE_URL}/api/settings/onboarding" yes)"
assert_code "${code}" "200" "onboarding with auth"

echo
echo "== 2) Submit check (new canonical payload) =="
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
  "motivation": "Smoke test del refactor."
}
JSON
)"

code="$(call POST "${BASE_URL}/api/requests/submit" no "${payload}")"
assert_code "${code}" "201" "submit request"

echo "Response:"
cat "${tmp_body}"
echo
echo
echo "Smoke test completed."
