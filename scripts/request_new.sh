#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:9090}"
seed="$(date +%s)$RANDOM"

first_names=(
  "Luca" "Marco" "Andrea" "Matteo" "Davide" "Francesco"
  "Alessandro" "Giuseppe" "Antonio" "Stefano" "Simone" "Federico"
)
surnames=(
  "Rossi" "Bianchi" "Romano" "Colombo" "Ricci" "Marino"
  "Greco" "Bruno" "Gallo" "Conti" "DeLuca" "Mancini"
)
regions=(
  "Lombardia" "Lazio" "Piemonte" "Veneto" "Emilia-Romagna" "Toscana"
  "Campania" "Puglia" "Sicilia" "Liguria" "Marche" "Sardegna"
)
marital_statuses=("Single" "Married" "Divorced")
children_options=("No" "Yes")
email_domains=("gmail.com" "outlook.com" "hotmail.com" "icloud.com" "libero.it")
mobile_prefixes=("320" "327" "328" "329" "333" "334" "335" "339")
motivations=(
  "Mi interessa conoscere persone con interessi simili."
  "Cerco una community attiva e ben organizzata."
  "Vorrei partecipare agli eventi e contribuire al gruppo."
  "Mi piace l'idea del progetto e vorrei farne parte."
  "Sono motivato a partecipare in modo costante."
)

pick() {
  local arr_name="$1"
  local arr_len=0
  local idx=0

  eval "arr_len=\${#${arr_name}[@]}"
  if [ "${arr_len}" -le 0 ]; then
    return 1
  fi

  idx=$((RANDOM % arr_len))
  eval "printf '%s\n' \"\${${arr_name}[${idx}]}\""
}

first_name="$(pick first_names)"
surname="$(pick surnames)"
full_name="${first_name} ${surname}"

slug_first="$(echo "${first_name}" | tr '[:upper:]' '[:lower:]')"
slug_surname="$(echo "${surname}" | tr '[:upper:]' '[:lower:]')"
email_domain="$(pick email_domains)"
email="${slug_first}.${slug_surname}${seed: -4}@${email_domain}"

mobile="+39$(pick mobile_prefixes)$((1000000 + RANDOM % 9000000))"
region="$(pick regions)"
birth_year="$((1975 + RANDOM % 25))"
marital_status="$(pick marital_statuses)"
children="$(pick children_options)"
motivation="$(pick motivations)"

curl -sS -X POST "${BASE_URL}/api/requests/submit" \
  -H "Content-Type: application/json" \
  -d "{
    \"full_name\": \"${full_name}\",
    \"email\": \"${email}\",
    \"mobile\": \"${mobile}\",
    \"region\": \"${region}\",
    \"birth_year\": \"${birth_year}\",
    \"marital_status\": \"${marital_status}\",
    \"children\": \"${children}\",
    \"motivation\": \"${motivation}\"
  }"
