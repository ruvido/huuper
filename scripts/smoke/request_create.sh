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

BASE_URL="${BASE_URL:-${BASE:-http://localhost:9090}}"

tmp_body="$(mktemp)"
trap 'rm -f "${tmp_body}"' EXIT

require_bin() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "Missing required command: $1" >&2
    exit 1
  }
}

call() {
  local method="$1"
  local url="$2"
  local payload="${3:-}"

  local -a args
  args=(-sS -X "${method}" "${url}" -o "${tmp_body}" -w "%{http_code}")
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

random_request_name() {
  local -a first_names=(
    "Luca"
    "Marco"
    "Andrea"
    "Matteo"
    "Simone"
    "Davide"
    "Federico"
    "Alessandro"
    "Gabriele"
    "Stefano"
  )
  local -a last_names=(
    "Rossi"
    "Ferrari"
    "Esposito"
    "Romano"
    "Conti"
    "Moretti"
    "Ricci"
    "Marini"
    "Greco"
    "Lombardi"
  )

  local first_index=$((RANDOM % ${#first_names[@]}))
  local last_index=$((RANDOM % ${#last_names[@]}))
  printf '%s %s' "${first_names[$first_index]}" "${last_names[$last_index]}"
}

require_bin curl
require_bin jq

uniq="$(date +%s)"
request_email="${REQUEST_SMOKE_EMAIL:-requests.smoke.${uniq}@realmen.it}"
request_name="${REQUEST_SMOKE_NAME:-$(random_request_name)}"

echo "BASE_URL=${BASE_URL}"
echo "Request email: ${request_email}"
echo "Request name: ${request_name}"

payload="$(jq -cn \
  --arg full_name "${request_name}" \
  --arg email "${request_email}" \
  --arg mobile "+393331112233" \
  --arg region "Lombardia" \
  --arg birth_year "1990" \
  --arg marital_status "Single" \
  --arg children "No" \
  --arg motivation "Local request create smoke." \
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

code="$(call POST "${BASE_URL}/api/public/requests" "${payload}")"
assert_code "${code}" "201" "public request submit"

echo
print_body
