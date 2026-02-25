EVENT_ID="i48zu71lysxrhi9"
curl -X POST "http://127.0.0.1:8090/api/admin/events/$EVENT_ID/email" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $1" \
    -d '{
      "subject": "Test invio evento",
      "body": "Questa è una prova identica del messaggio finale.",
      "target": "all",
      "dry_run": true
    }'
