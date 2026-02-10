curl -X POST "http://localhost:8090/api/requests/submit" \
    -H "Content-Type: application/json" \
    -d '{
      "full_name": "Antonello",
      "email": "antonello@venditti.it",
      "mobile": "+393331112233",
      "region": "Lombardia",
      "birth_year": "1990",
      "marital_status": "Single",
      "children": "No",
      "motivation": "Vorrei partecipare alla community."
    }'
