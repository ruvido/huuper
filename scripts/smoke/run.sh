#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPTS_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
ENV_FILE="${SCRIPTS_DIR}/.env"

if [[ -f "${ENV_FILE}" ]]; then
  while IFS='=' read -r key raw; do
    [[ -z "${key}" ]] && continue
    [[ "${key}" =~ ^[[:space:]]*# ]] && continue

    key="$(printf '%s' "${key}" | xargs)"
    [[ -z "${key}" ]] && continue

    if [[ "${!key+x}" == "x" ]]; then
      continue
    fi

    value="$(printf '%s' "${raw:-}" | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')"
    value="${value%\"}"
    value="${value#\"}"
    export "${key}=${value}"
  done < "${ENV_FILE}"
fi

BASE_URL="${BASE_URL:-${BASE:-http://127.0.0.1:9090}}"
ADMIN_TOKEN="${ADMIN_TOKEN:-}"
MEMBER_TOKEN="${MEMBER_TOKEN:-}"
GUARDIAN_TOKEN="${GUARDIAN_TOKEN:-}"
GUARDIAN_EMAIL="${GUARDIAN_EMAIL:-}"
GUARDIAN_PASSWORD="${GUARDIAN_PASSWORD:-}"
EVENT_SLUG="${EVENT_SLUG:-}"
REQUEST_SMOKE_GROUP_ID="${REQUEST_SMOKE_GROUP_ID:-}"
REQUEST_SMOKE_GUARDIAN_ID="${REQUEST_SMOKE_GUARDIAN_ID:-}"

auth_with_password() {
  local email="$1"
  local password="$2"
  local payload
  payload="$(jq -n --arg identity "$email" --arg password "$password" '{identity:$identity, password:$password}')"
  curl -sS -X POST "${BASE_URL}/api/collections/users/auth-with-password" \
    -H "Content-Type: application/json" \
    -d "${payload}" | jq -r '.token // empty'
}

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
  ADMIN_TOKEN="$(bash "${SCRIPT_DIR}/auth_admin.sh")"
}

get_member_token_if_needed() {
  if [[ -n "${MEMBER_TOKEN}" ]]; then
    return
  fi
  MEMBER_TOKEN="$(bash "${SCRIPT_DIR}/auth_member.sh")"
}

get_guardian_token_if_needed() {
  if [[ -n "${GUARDIAN_TOKEN}" ]]; then
    return 0
  fi
  if [[ -z "${GUARDIAN_EMAIL}" || -z "${GUARDIAN_PASSWORD}" ]]; then
    return 1
  fi
  GUARDIAN_TOKEN="$(auth_with_password "${GUARDIAN_EMAIL}" "${GUARDIAN_PASSWORD}")"
  [[ -n "${GUARDIAN_TOKEN}" ]]
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
    guardian)
      args+=(-H "Authorization: ${GUARDIAN_TOKEN}")
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

json_value() {
  jq -r "$1 // empty" "${tmp_body}"
}

assert_json_value() {
  local jq_expr="$1"
  local expected="$2"
  local label="$3"
  local got
  got="$(json_value "${jq_expr}")"
  if [[ "${got}" != "${expected}" ]]; then
    echo "FAIL ${label}: expected ${expected}, got ${got}"
    echo "Body:"
    print_body
    exit 1
  fi
  echo "OK   ${label}: ${got}"
}

submit_request() {
  local uniq="$1"
  local payload
  payload="$(cat <<JSON
{
  "full_name": "Smoke Test ${uniq}",
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
  call POST "${BASE_URL}/api/public/requests" none "${payload}"
}

request_action_payload() {
  local action="$1"
  local group="${2:-}"
  local guardian="${3:-}"
  local reason="${4:-}"
  local mentoring_notes="${5:-}"
  jq -cn \
    --arg action "${action}" \
    --arg group "${group}" \
    --arg guardian "${guardian}" \
    --arg reason "${reason}" \
    --arg mentoring_notes "${mentoring_notes}" \
    '
      {action:$action}
      + (if $group != "" then {group:$group} else {} end)
      + (if $guardian != "" then {guardian:$guardian} else {} end)
      + (if $reason != "" then {reason:$reason} else {} end)
      + (if $mentoring_notes != "" then {mentoring_notes:$mentoring_notes} else {} end)
    '
}

run_minimal_request_flow_smoke() {
  local request_id="$1"

  echo "Minimal flow request_id=${request_id}"

  code="$(call POST "${BASE_URL}/api/admin/requests/${request_id}/actions" admin "$(request_action_payload "advance" "${REQUEST_SMOKE_GROUP_ID}")")"
  assert_code "${code}" "200" "request assign_group by admin"
  assert_json_value '.status' "2-group_assigned" "request status after assign_group"

  code="$(call POST "${BASE_URL}/api/admin/requests/${request_id}/actions" admin "$(request_action_payload "advance" "" "${REQUEST_SMOKE_GUARDIAN_ID}")")"
  assert_code "${code}" "200" "request assign_guardian by admin"
  assert_json_value '.status' "3-guardian_assigned" "request status after assign_guardian"

  code="$(call POST "${BASE_URL}/api/admin/requests/${request_id}/actions" admin "$(request_action_payload "advance" "" "" "" "Mentoring smoke admin.")")"
  assert_code "${code}" "200" "request mentoring by admin"
  assert_json_value '.status' "4-mentoring" "request status after mentoring"

  code="$(call POST "${BASE_URL}/api/admin/requests/${request_id}/actions" admin "$(request_action_payload "advance")")"
  assert_code "${code}" "200" "request group_approved by admin"
  assert_json_value '.status' "5-group_approved" "request status after group_approved"

  code="$(call POST "${BASE_URL}/api/admin/requests/${request_id}/actions" admin "$(request_action_payload "advance")")"
  assert_code "${code}" "200" "request admin_approved by admin"
  assert_json_value '.status' "6-admin_approved" "request status after admin_approved"

  code="$(call GET "${BASE_URL}/api/admin/requests/${request_id}" admin)"
  assert_code "${code}" "200" "request detail after minimal flow"
  assert_json_value '.group' "${REQUEST_SMOKE_GROUP_ID}" "request detail group"
  assert_json_value '.guardian' "${REQUEST_SMOKE_GUARDIAN_ID}" "request detail guardian"
  assert_json_value '.status' "6-admin_approved" "request detail final status"
}

run_role_request_flow_smoke() {
  local request_id="$1"

  echo "Role flow request_id=${request_id}"

  code="$(call POST "${BASE_URL}/api/admin/requests/${request_id}/actions" admin "$(request_action_payload "advance" "${REQUEST_SMOKE_GROUP_ID}")")"
  assert_code "${code}" "200" "request assign_group for role flow"
  assert_json_value '.status' "2-group_assigned" "role flow status after assign_group"

  code="$(call GET "${BASE_URL}/api/me/requests/${request_id}" member)"
  assert_code "${code}" "200" "assistant can view grouped request"
  assert_json_value '.workflow.next_action' "assign_guardian" "assistant sees assign_guardian step"

  code="$(call POST "${BASE_URL}/api/me/requests/${request_id}/actions" member "$(request_action_payload "advance" "" "${REQUEST_SMOKE_GUARDIAN_ID}")")"
  assert_code "${code}" "200" "assistant assign_guardian"
  assert_json_value '.status' "3-guardian_assigned" "role flow status after assign_guardian"

  code="$(call GET "${BASE_URL}/api/me/requests/${request_id}" guardian)"
  assert_code "${code}" "200" "guardian can view assigned request"
  assert_json_value '.workflow.next_action' "mentoring" "guardian sees mentoring step"

  code="$(call POST "${BASE_URL}/api/me/requests/${request_id}/actions" guardian "$(request_action_payload "advance" "" "" "" "Mentoring smoke guardian.")")"
  assert_code "${code}" "200" "guardian mentoring"
  assert_json_value '.status' "4-mentoring" "role flow status after mentoring"

  code="$(call POST "${BASE_URL}/api/me/requests/${request_id}/actions" member "$(request_action_payload "advance")")"
  assert_code "${code}" "200" "assistant group_approved"
  assert_json_value '.status' "5-group_approved" "role flow status after group_approved"

  code="$(call POST "${BASE_URL}/api/admin/requests/${request_id}/actions" admin "$(request_action_payload "advance")")"
  assert_code "${code}" "200" "admin admin_approved in role flow"
  assert_json_value '.status' "6-admin_approved" "role flow status after admin_approved"

  code="$(call GET "${BASE_URL}/api/admin/requests/${request_id}" admin)"
  assert_code "${code}" "200" "request detail after role flow"
  assert_json_value '.status' "6-admin_approved" "role flow final detail status"
}

require_bin curl
require_bin jq

echo "BASE_URL=${BASE_URL}"
echo

echo "== 1) Public smoke =="
code="$(call GET "${BASE_URL}/api/public/settings/signup" none)"
assert_code "${code}" "200" "public settings signup"

code="$(call GET "${BASE_URL}/api/public/settings/profile_schema" none)"
assert_code "${code}" "200" "public settings profile_schema"

uniq="$(date +%s)"
code="$(submit_request "${uniq}")"
assert_code "${code}" "201" "public request submit"
PUBLIC_REQUEST_ID="$(json_value '.id')"

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
echo "== 3) Request flow smoke =="
if [[ -n "${REQUEST_SMOKE_GROUP_ID}" && -n "${REQUEST_SMOKE_GUARDIAN_ID}" ]]; then
  run_minimal_request_flow_smoke "${PUBLIC_REQUEST_ID}"

  if get_guardian_token_if_needed; then
    role_uniq="$((uniq + 1))"
    code="$(submit_request "${role_uniq}")"
    assert_code "${code}" "201" "public request submit for role flow"
    ROLE_REQUEST_ID="$(json_value '.id')"
    run_role_request_flow_smoke "${ROLE_REQUEST_ID}"
  else
    echo "SKIP role-based request smoke: set GUARDIAN_EMAIL/GUARDIAN_PASSWORD or GUARDIAN_TOKEN"
  fi
else
  echo "SKIP request flow smoke: set REQUEST_SMOKE_GROUP_ID and REQUEST_SMOKE_GUARDIAN_ID"
fi

echo
echo "== 4) Optional event smoke =="
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
