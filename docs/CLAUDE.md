# members

Self-hosted webapp to manage private telegram/discord groups, including:
- user login 
- user profile page
- group list
- admin page for admins only

## Pilastri
- **Backend**: PocketBase (Go framework)
- **Frontend**: `frontend/skeleton` -> `frontend/site`
- **Deploy**: Single binary

## Strategia di Sviluppo
- **Less is more**: codice semplice e chiaro
- **I docs ufficiali sono dio**: consulta SEMPRE docs ufficiali prima di implementare
- **Best practices only**: seguire convenzioni ufficiali
- **Zero ridondanza**: evitare codice complesso e duplicato

## Code language
Tutto il codice è in inglese, sempre: nomi di collection, campi DB, package/funzioni/variabili Go, route/path API, chiavi JSON, nomi di file/variabili/classi CSS nel frontend, e le label in `frontend/skeleton/copy/*.js`. Nessuna eccezione.
Solo i **VALORI** dei record (contenuto reale digitato dagli admin: descrizioni, FAQ, testi email, ecc.) possono essere in italiano — mai la struttura/schema/codice.
- App interna (admin/member, `frontend/skeleton/copy/*.js`): tutte le stringhe SOLO in inglese, definite in `copy/*.js` e referenziate via `appCopy`. Mai stringhe hardcoded inline nei componenti.
- Pagine pubbliche (`assets/js/public/*.js`, es. registrazione/pagamento evento/retreat): chrome/label di UI in inglese come il resto del codice; il contenuto che viene renderizzato da `data` (testi scritti dall'admin, tipicamente IT) resta quello che l'admin ha digitato, verbatim.

## pocketbase
- extend in go
- custom api e hooks in go (no js sdk server side)
- usa .env per settare default admin

## Docs Ufficiali
- PocketBase: https://pocketbase.io/docs/
- Go Framework: https://pocketbase.io/docs/use-as-framework/

## Comandi
```bash
./pocketbase serve                    # Start dev
go run ./backend serve --dir=./data  # Start dev backend
go build -o pocketbase ./backend     # Build binary
```
