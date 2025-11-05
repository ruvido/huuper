# members

Self-hosted webapp to manage private telegram/discord groups, including:
- user login 
- user profile page
- group list
- admin page for admins only

## Pilastri
- **Backend**: PocketBase (Go framework)
- **Frontend**: Svelte + hash routing
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

## svelte
- simple hash routing

## Docs Ufficiali
- PocketBase: https://pocketbase.io/docs/
- Go Framework: https://pocketbase.io/docs/use-as-framework/
- Svelte: https://svelte.dev/docs/svelte/
- Hash Routing: https://svelte.dev/docs/kit/configuration#router

## Comandi
```bash
./pocketbase serve                    # Start dev
cd frontend && npm run build         # Build frontend
go build -o pocketbase main.go       # Build binary
```