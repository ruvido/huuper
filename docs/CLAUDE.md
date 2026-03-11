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
