#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPTS_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
REPO_ROOT="$(cd "${SCRIPTS_DIR}/.." && pwd)"
DATA_DB="${REPO_ROOT}/data/data.db"
ENV_FILE="${SCRIPTS_DIR}/.env"

if [[ -f "${ENV_FILE}" ]]; then
  set -a
  # shellcheck disable=SC1090
  source "${ENV_FILE}"
  set +a
fi

BASE_URL="${BASE_URL:-${BASE:-http://localhost:9090}}"
TOKEN_VALUE="${ONBOARDING_TOKEN:-123}"
TOKEN_USER_ID="${ONBOARDING_USER_ID:-h4874mk9ix3bybb}"
TOKEN_GROUP_ID="${ONBOARDING_GROUP_ID:-zwlo6tfz2xf4tr5}"
TOKEN_SERVICE="${ONBOARDING_SERVICE:-onboarding}"
AVATAR_PATH="${ONBOARDING_AVATAR_PATH:-${REPO_ROOT}/avatar.jpg}"

require_bin() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "Missing required command: $1" >&2
    exit 1
  }
}

log() {
  printf '[%s] %s\n' "$(date +'%H:%M:%S')" "$*"
}

urlencode() {
  jq -rn --arg v "$1" '$v|@uri'
}

json_pretty() {
  if command -v jq >/dev/null 2>&1; then
    jq . "$1" 2>/dev/null || cat "$1"
  else
    cat "$1"
  fi
}

call() {
  local method="$1"
  local url="$2"
  local token="${3:-}"
  local payload="${4:-}"
  local body_file="${5:-}"

  local -a args
  args=(-sS -X "${method}" "${url}" -o "${body_file}" -w "%{http_code}")
  if [[ -n "${token}" ]]; then
    args+=(-H "Authorization: Bearer ${token}")
  fi
  if [[ -n "${payload}" ]]; then
    args+=(-H "Content-Type: application/json" -d "${payload}")
  fi

  curl "${args[@]}"
}

require_bin curl
require_bin jq
require_bin sqlite3
tmp_body="$(mktemp)"
trap 'rm -f "${tmp_body}"' EXIT

log "BASE_URL=${BASE_URL}"
log "token=${TOKEN_VALUE} user=${TOKEN_USER_ID} group=${TOKEN_GROUP_ID} service=${TOKEN_SERVICE}"
log "avatar_path=${AVATAR_PATH}"

if [[ ! -f "${AVATAR_PATH}" ]]; then
  log "Avatar file not found: ${AVATAR_PATH}"
  exit 1
fi

tmp_avatar="${AVATAR_PATH}"
log "Using avatar file: ${tmp_avatar}"

log "Recreating onboarding token directly in PocketBase SQLite"
if [[ ! -f "${DATA_DB}" ]]; then
  log "PocketBase database not found: ${DATA_DB}"
  exit 1
fi

created_at="$(date -u +"%Y-%m-%d %H:%M:%S").000Z"
sqlite3 "${DATA_DB}" <<SQL
BEGIN;
DELETE FROM tokens WHERE token = '${TOKEN_VALUE}' AND service = '${TOKEN_SERVICE}';
INSERT INTO tokens (created, id, service, token, user, "group")
VALUES ('${created_at}', 'smoke_onboarding_${TOKEN_VALUE}', '${TOKEN_SERVICE}', '${TOKEN_VALUE}', '${TOKEN_USER_ID}', '${TOKEN_GROUP_ID}');
COMMIT;
SQL

created_token_row="$(sqlite3 "${DATA_DB}" "SELECT id, created, service, token, user, \"group\" FROM tokens WHERE token = '${TOKEN_VALUE}' AND service = '${TOKEN_SERVICE}' LIMIT 1;")"
log "Created token row:"
printf '%s\n' "${created_token_row}"

log "Submitting onboarding finalize with multipart avatar"
profile_payload="$(jq -cn \
  --arg work "Sanita Medicina" \
  --argjson skills '["Muratura"]' \
  --argjson interests '["Bushcraft"]' \
  --argjson sports '["Palestra"]' \
  '{work:$work, skills:$skills, interests:$interests, sports:$sports}')"

http_code="$(curl -sS -X POST \
  "${BASE_URL}/api/public/onboarding/${TOKEN_VALUE}/finalize" \
  -F "data=${profile_payload}" \
  -F "avatar=@${tmp_avatar};type=image/jpeg;filename=smoke-avatar.jpg" \
  -o "${tmp_body}" \
  -w "%{http_code}")"

log "Finalize HTTP ${http_code}"
json_pretty "${tmp_body}"

if [[ "${http_code}" != "200" ]]; then
  log "Finalize did not return 200"
  exit 1
fi

log "Authenticating as admin for verification"
admin_token="$("${SCRIPTS_DIR}/smoke/auth_admin.sh")"
log "Admin token acquired"

log "Verifying user record"
user_response="$(curl -sS -H "Authorization: Bearer ${admin_token}" "${BASE_URL}/api/admin/users/${TOKEN_USER_ID}")"
user_avatar="$(printf '%s' "${user_response}" | jq -r '.avatar // empty')"
user_data="$(printf '%s' "${user_response}" | jq -c '.data // {}')"
log "User avatar: ${user_avatar:-<empty>}"
log "User data: ${user_data}"

if [[ -z "${user_avatar}" ]]; then
  log "FAIL avatar is still empty"
  printf '%s\n' "${user_response}" | jq . 2>/dev/null || printf '%s\n' "${user_response}"
  exit 1
fi

log "Smoke OK: avatar saved in PocketBase"
