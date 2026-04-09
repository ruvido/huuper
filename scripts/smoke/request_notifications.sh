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
REQUEST_SMOKE_GROUP_ID="${REQUEST_SMOKE_GROUP_ID:-}"
REQUEST_SMOKE_GUARDIAN_ID="${REQUEST_SMOKE_GUARDIAN_ID:-}"
MEMBER_EMAIL="${MEMBER_EMAIL:-${USER_EMAIL:-}}"
ADMIN_EMAIL="${ADMIN_EMAIL:-}"
ADMIN_NOTIFICATION_EMAIL="${ADMIN_NOTIFICATION_EMAIL:-}"

tmp_body="$(mktemp)"
trap 'rm -f "${tmp_body}"' EXIT

require_bin() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "Missing required command: $1" >&2
    exit 1
  }
}

auth_with_password() {
  local email="$1"
  local password="$2"
  local payload
  payload="$(jq -n --arg identity "$email" --arg password "$password" '{identity:$identity, password:$password}')"
  curl -sS -X POST "${BASE_URL}/api/collections/users/auth-with-password" \
    -H "Content-Type: application/json" \
    -d "${payload}" | jq -r '.token // empty'
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
    return
  fi
  : "${GUARDIAN_EMAIL:?Missing GUARDIAN_EMAIL}"
  : "${GUARDIAN_PASSWORD:?Missing GUARDIAN_PASSWORD}"
  GUARDIAN_TOKEN="$(auth_with_password "${GUARDIAN_EMAIL}" "${GUARDIAN_PASSWORD}")"
  [[ -n "${GUARDIAN_TOKEN}" ]] || {
    echo "Guardian token not found" >&2
    exit 1
  }
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

load_admin_notification_email() {
  if [[ -n "${ADMIN_NOTIFICATION_EMAIL}" ]]; then
    return
  fi
  code="$(call GET "${BASE_URL}/api/admin/settings/email" admin)"
  assert_code "${code}" "200" "load admin email settings"
  ADMIN_NOTIFICATION_EMAIL="$(json_value '.data.admin')"
  : "${ADMIN_NOTIFICATION_EMAIL:?Missing settings.email.data.admin}"
}

request_action_payload() {
  local action="$1"
  local group="${2:-}"
  local guardian="${3:-}"
  local mentoring_notes="${4:-}"
  jq -cn \
    --arg action "${action}" \
    --arg group "${group}" \
    --arg guardian "${guardian}" \
    --arg mentoring_notes "${mentoring_notes}" \
    '
      {action:$action}
      + (if $group != "" then {group:$group} else {} end)
      + (if $guardian != "" then {guardian:$guardian} else {} end)
      + (if $mentoring_notes != "" then {mentoring_notes:$mentoring_notes} else {} end)
    '
}

confirm_step() {
  local label="$1"
  local expected="$2"
  local reply
  echo
  echo "${label}"
  echo "Expected email: ${expected}"
  read -r -p "Run this step now? (Y/n) " reply
  case "${reply:-Y}" in
    n|N)
      echo "Skipped."
      return 1
      ;;
    *)
      return 0
      ;;
  esac
}

require_bin curl
require_bin jq

: "${REQUEST_SMOKE_GROUP_ID:?Missing REQUEST_SMOKE_GROUP_ID}"
: "${REQUEST_SMOKE_GUARDIAN_ID:?Missing REQUEST_SMOKE_GUARDIAN_ID}"
: "${MEMBER_EMAIL:?Missing MEMBER_EMAIL}"
: "${GUARDIAN_EMAIL:?Missing GUARDIAN_EMAIL}"

get_admin_token_if_needed
get_member_token_if_needed
get_guardian_token_if_needed
load_admin_notification_email

uniq="$(date +%s)"
request_email="requests.smoke.${uniq}@realmen.it"
request_name="Request Smoke ${uniq}"
request_id=""

echo "BASE_URL=${BASE_URL}"
echo "Request email: ${request_email}"
echo "Admin email: ${ADMIN_NOTIFICATION_EMAIL}"
echo "Assistant email: ${MEMBER_EMAIL}"
echo "Guardian email: ${GUARDIAN_EMAIL}"

if confirm_step "Submit request" "${ADMIN_NOTIFICATION_EMAIL}"; then
  payload="$(jq -cn \
    --arg full_name "${request_name}" \
    --arg email "${request_email}" \
    --arg mobile "+393331112233" \
    --arg region "Lombardia" \
    --arg birth_year "1990" \
    --arg marital_status "Single" \
    --arg children "No" \
    --arg motivation "Interactive request email smoke." \
    '{
      full_name:$full_name,
      email:$email,
      mobile:$mobile,
      region:$region,
      birth_year:$birth_year,
      marital_status:$marital_status,
      children:$children,
      motivation:$motivation
    }')"
  code="$(call POST "${BASE_URL}/api/public/requests" none "${payload}")"
  assert_code "${code}" "201" "public request submit"
  request_id="$(json_value '.id')"
  echo "Request ID: ${request_id}"
else
  echo "A request is required to continue." >&2
  exit 1
fi

if confirm_step "Assign group" "${MEMBER_EMAIL}"; then
  code="$(call POST "${BASE_URL}/api/admin/requests/${request_id}/actions" admin "$(request_action_payload "advance" "${REQUEST_SMOKE_GROUP_ID}")")"
  assert_code "${code}" "200" "assign_group"
fi

if confirm_step "Assign guardian" "${GUARDIAN_EMAIL}"; then
  code="$(call POST "${BASE_URL}/api/me/requests/${request_id}/actions" member "$(request_action_payload "advance" "" "${REQUEST_SMOKE_GUARDIAN_ID}")")"
  assert_code "${code}" "200" "assign_guardian"
fi

if confirm_step "Complete mentoring" "${MEMBER_EMAIL}"; then
  code="$(call POST "${BASE_URL}/api/me/requests/${request_id}/actions" guardian "$(request_action_payload "advance" "" "" "Interactive request email smoke.")")"
  assert_code "${code}" "200" "mentoring"
fi

if confirm_step "Group approval" "${ADMIN_NOTIFICATION_EMAIL}"; then
  code="$(call POST "${BASE_URL}/api/me/requests/${request_id}/actions" member "$(request_action_payload "advance")")"
  assert_code "${code}" "200" "group_approved"
fi

if confirm_step "Admin approval" "${request_email}"; then
  code="$(call POST "${BASE_URL}/api/admin/requests/${request_id}/actions" admin "$(request_action_payload "advance")")"
  assert_code "${code}" "200" "admin_approved"
fi

echo
echo "Interactive request notification smoke completed."
echo "Request ID: ${request_id}"
echo "Candidate email: ${request_email}"
