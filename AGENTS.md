# AGENTS

## Scopo
Linee guida minime per contribuire al progetto in modo coerente, semplice e mantenibile.

## Principi
- Less is more.
- DRY sempre: evitare duplicazioni di logica, dati e processi.
- Preferire soluzioni state of the art ma semplici da operare in self-hosting.
- Quando c'e' dubbio, scegliere l'opzione piu' semplice, piu' chiara e con meno moving parts.

## Regole operative
- No custom code quando esiste una soluzione standard o affidabile gia' disponibile.
- Non prendere iniziativa architetturale o di scope senza discuterne prima.
- Prima di implementare cambi strutturali, proporre opzioni e attendere conferma.
- Tenere i cambi piccoli, iterativi e verificabili.
- Non introdurre automazioni opache difficili da mantenere.
- Non aggiungere dipendenze nuove se non strettamente necessario.
- Non introdurre fallback impliciti, percorsi alternativi automatici o comportamento inferito non richiesto esplicitamente.

## Scope attuale
- Backend e frontend web sono in scope.
- Mobile e' fuori scope, salvo richiesta esplicita.

## Architettura dati
- Source of truth dei dati applicativi: PocketBase per metadata, relazioni, auth e permessi.
- Il database non deve sostituire confini di dominio chiari nell'applicazione.
- PocketBase e' la scelta predefinita per metadata + auth.

## Backend
- Backend in Go con PocketBase framework.
- La logica server-side deve stare in Go.
- Usare `.env` per configurazioni locali e credenziali di bootstrap dove previsto.
- Preferire refactor iterativi piccoli con build verde a ogni step.
- Centralizzare autorizzazione, validazione e access rules; evitare duplicazione negli handler.
- Mantenere il tree strutturato e sensato, senza refactor strutturali non discussi.
- I file backend devono seguire naming di dominio chiaro, non naming tecnico rumoroso.
- Evitare file come `_get`, `_list`, `_create`, `_handler` quando il dominio puo' vivere in `groups.go`, `users.go`, `events.go`, `requests.go`, `settings.go`.
- Se una capability appartiene chiaramente a un dominio esistente, va nello stesso file o package del dominio, non in un file-frammento aggiunto per comodita'.

## Frontend
- Il frontend attivo usa l'approccio `frontend/skeleton` -> `frontend/site`.
- `frontend/skeleton` e' la source of truth editabile.
- `frontend/site` e' output generato e servito da PocketBase.
- Non modificare manualmente `frontend/site`.
- Il frontend v2 ha watch e live reload attivi: dopo le modifiche a `frontend/skeleton`, l'output si rigenera automaticamente. Non ripetere istruzioni su rebuild manuale salvo blocco esplicito.
- L'archivio del vecchio frontend Svelte vive in `archive/frontend-svelte/` e non fa parte del path attivo.
- Usare CSS semantico e componenti/template chiari.
- Non hardcodare mai testo user-facing dentro componenti, template, JS o handler: il copy frontend deve stare in `frontend/skeleton/copy/*`; il copy backend deve stare in `backend/internal/copywriting/` (per request: `backend/internal/copywriting/requests/`) o arrivare da settings.
- Il primo obiettivo frontend v2 e' wireframing funzionale, non design finale.
- Il dev mode deve restare semplice e trasparente, con watcher/build standard e senza automazioni opache.
- Per il frontend v2 valgono anche `agents/frontend-agent.md` e `agents/review-agent.md`.
- Evitare fan-out di chiamate dal frontend quando il backend puo' aggregare i dati in un endpoint chiaro.
- Per le pagine dettaglio preferire un solo endpoint backend per pagina, soprattutto se il frontend altrimenti farebbe 2 o piu' richieste correlate.

## API Design
- Le API devono essere chiare, robuste, prevedibili e adatte a interfacce web, CLI, mobile o desktop.
- Il namespacing deve essere esplicito e stabile.
- Gli endpoint admin hanno potere globale sulle risorse applicative.
- Gli endpoint non admin devono esporre solo cio' che riguarda il profilo dell'utente autenticato e le sue relazioni autorizzate.
- Naming, payload, errori e regole di accesso devono essere coerenti tra domini diversi.

## Cosa evitare
- Feature creep.
- Astrazioni furbe o premature.
- Duplicazione di logica tra handler.
- Duplicazione di fetch lato frontend per comporre una singola vista quando il backend puo' esporre un endpoint di dettaglio aggregato.
- Reintrodurre naming sporco o frammentato nel backend, soprattutto file suffix-based come `_get`, `_list`, `_handler`.
- Fallback impliciti o comportamento automatico non richiesto.
- Drift rispetto alle convenzioni ufficiali di Go e PocketBase.
- Reintrodurre il frontend legacy nel path attivo.

## Collaboration Constraint
- Se una modifica e' ambigua, fermarsi e chiarire.
- Se una modifica tocca struttura, policy o naming cross-cutting, proporre prima le opzioni.
- Se un flusso ha piu' interpretazioni possibili, non implementare scorciatoie o fallback: fermarsi e chiedere conferma.

## Decision policy
- In caso di dubbio scegliere l'opzione piu' semplice, piu' chiara, con meno moving parts.
- Prima ufficializzare una convenzione, verificare che migliori davvero discoverability, robustezza e manutenzione.

## Riferimenti
- https://pocketbase.io/docs/
- https://pocketbase.io/docs/use-as-framework/
- https://go.dev/doc/
- https://core.telegram.org/bots/api
