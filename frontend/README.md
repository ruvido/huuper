# Frontend v2

Frontend attivo del progetto.

## Struttura
- `skeleton/`: source of truth editabile
- `site/`: output generato e servito da PocketBase

## Regole
- Non modificare `site/` manualmente, salvo bootstrap iniziale esplicito
- La prima fase e' wireframing funzionale
- Il frontend legacy Svelte vive in `archive/frontend-svelte/`

## Dev
- sorgente: `frontend/skeleton`
- output: `frontend/site`
- sviluppo locale: `go run ./backend serve --http=127.0.0.1:9090 --dir=./data --frontend-dev`
- in alternativa: `./deploy/local.sh`
