#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-https://branco.realmen.it}"
AUTO_LOGIN="${AUTO_LOGIN:-1}"

if [ $# -lt 2 ]; then
  echo "Usage: $0 <email> <password>" >&2
  echo "Example: $0 test.user@example.com 'Test1234!'" >&2
  exit 1
fi

EMAIL="$1"
PASSWORD="$2"

echo "create user: ${EMAIL} (status=active)"
CREATE_RESP="$(curl -sS -X POST "${BASE_URL}/api/collections/users/records" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"${EMAIL}\",
    \"password\": \"${PASSWORD}\",
    \"passwordConfirm\": \"${PASSWORD}\",
    \"status\": \"active\"
  }")"

if echo "$CREATE_RESP" | grep -q '"message"'; then
  # PocketBase errors include "message".
  echo "$CREATE_RESP"
  exit 1
fi

echo "user created"
echo "$CREATE_RESP"

if [ "$AUTO_LOGIN" = "1" ]; then
  echo "auth with password"
  AUTH_RESP="$(curl -sS -X POST "${BASE_URL}/api/collections/users/auth-with-password" \
    -H "Content-Type: application/json" \
    -d "{
      \"identity\": \"${EMAIL}\",
      \"password\": \"${PASSWORD}\"
    }")"
  echo "$AUTH_RESP"
fi
