# PLAN

## Obiettivo
Refactoring iterativo del backend per renderlo piu' DRY, robusto, sicuro, discoverable e coerente nel namespacing, senza introdurre scope creep o astrazioni inutili.

## Vincoli
- Seguire `AGENTS.md`.
- Nessun big-bang refactor.
- Nessun cambio strutturale senza proposta e conferma.
- Nessuna dipendenza nuova se non strettamente necessaria.
- Ogni iterazione deve lasciare il backend compilabile.

## Target architetturale
- Backend pronto per web, CLI, mobile o desktop.
- API chiare e stabili.
- Due entrypoint di accesso principali:
  - `admin=true`: controllo completo sulle risorse applicative.
  - `admin!=true`: accesso limitato al proprio profilo e alle proprie relazioni autorizzate.
- Policy di accesso centralizzata, non dispersa negli handler.
- Tree backend organizzato per domini e responsabilita' chiare.

## Workflow multiagent

### 1. Architecture Guardian
Scopo:
- verificare che ogni proposta rispetti semplicita', DRY, no custom code e no scope drift.

Output:
- via libera o blocco motivato
- tradeoff espliciti

### 2. Repo Mapper
Scopo:
- mappare route, handler, helper condivisi, collection PocketBase, policy esistenti e duplicazioni.

Output:
- inventory del backend
- mappa dei confini di dominio
- elenco duplicazioni e naming incoerenti

### 3. Access Model Agent
Scopo:
- definire la matrice di accesso per `public`, `me`, `admin` e per i ruoli derivati.

Output:
- tabella endpoint -> ruolo/capability -> risorsa visibile
- proposta di helper centralizzati per authz

### 4. API Contract Agent
Scopo:
- definire namespacing, convenzioni di naming, shape delle risposte, error model, filtri e paginazione.

Output:
- contratto API target
- delta tra stato attuale e stato desiderato

### 5. Refactor Agent
Scopo:
- applicare refactor piccoli, lineari e verificabili.

Output:
- patch limitate
- riduzione della duplicazione
- miglioramento di discoverability e robustezza

### 6. Review Agent
Scopo:
- verificare regressioni, rischi, incompletezze e allineamento con le convenzioni.

Output:
- findings ordinati per severita'
- rischi residui
- decisione go/no-go per la prossima iterazione

## Ordine di lavoro
1. Mappare lo stato corrente.
2. Definire la matrice di accesso.
3. Definire il namespacing target.
4. Consolidare helper di authz e validazione.
5. Riorganizzare route e handler per dominio.
6. Allineare naming, payload ed errori.
7. Verificare build e regressioni.

## Iterazioni consigliate

### Iterazione 1
- Inventory backend.
- Matrice accessi attuale.
- Elenco duplicazioni e incoerenze.
- Nessun refactor strutturale.

### Iterazione 2
- Consolidamento authz.
- Estrarre helper condivisi per visibilita', ownership e ruoli.
- Ridurre duplicazione tra handler.

### Iterazione 3
- Proposta di namespacing target.
- Proposta di riorganizzazione del tree.
- Conferma prima di toccare struttura e path.

### Iterazione 4
- Refactor dei route group e naming handler.
- Allineamento dei path al modello approvato.

### Iterazione 5
- Pulizia convenzioni API.
- Uniformare errori, paginazione, filtri e response shape.

## Definition of Done
- Le policy `admin` vs `me` sono esplicite e centralizzate.
- Il namespacing delle API e' coerente e discoverable.
- Il tree backend e' organizzato per domini comprensibili.
- Gli handler delegano policy e validazione a helper condivisi.
- Le API restano robuste e adatte a piu' client.
- Ogni step e' stato verificato con build e review.
