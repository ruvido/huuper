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

## Frontend
- Il frontend attivo usa l'approccio `frontend/skeleton` -> `frontend/site`.
- `frontend/skeleton` e' la source of truth editabile.
- `frontend/site` e' output generato e servito da PocketBase.
- Non modificare manualmente `frontend/site`.
- L'archivio del vecchio frontend Svelte vive in `archive/frontend-svelte/` e non fa parte del path attivo.
- Usare CSS semantico e componenti/template chiari.
- Il primo obiettivo frontend v2 e' wireframing funzionale, non design finale.
- Il dev mode deve restare semplice e trasparente, con watcher/build standard e senza automazioni opache.

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
- Fallback impliciti o comportamento automatico non richiesto.
- Drift rispetto alle convenzioni ufficiali di Go e PocketBase.
- Reintrodurre il frontend legacy nel path attivo.

## Collaboration Constraint
- Se una modifica e' ambigua, fermarsi e chiarire.
- Se una modifica tocca struttura, policy o naming cross-cutting, proporre prima le opzioni.

## Decision policy
- In caso di dubbio scegliere l'opzione piu' semplice, piu' chiara, con meno moving parts.
- Prima ufficializzare una convenzione, verificare che migliori davvero discoverability, robustezza e manutenzione.

## Riferimenti
- https://pocketbase.io/docs/
- https://pocketbase.io/docs/use-as-framework/
- https://go.dev/doc/
- https://core.telegram.org/bots/api
