# Events refactor — TODO

Stato della migrazione del modello eventi (call/meetup, drop rally, drop series, single-record con cadence). Aggiornato dopo task #21–#24 completati.

## Decisioni del brainstorm

- **Type**: definiti da `settings.eventflow.types`. Il bootstrap attuale include `call` e `meetup`; nuovi tipi si aggiungono dai settings senza cambiare codice Go.
- **Group → visibility**: se `group` è settato → solo membri. Se vuoto → aperto a tutti. Convenzione già supportata da `ListForUser`.
- **Series**: niente token, niente record-chain. **1 record = 1 evento (anche ricorrente)**. La serie è descritta da `cadence + count` sullo stesso record.
- **Cadence**: capability applicativa hardcoded, salvata come stringa enum (SelectField in PB):
  - `once`
  - `weekly:mon`..`weekly:sun` (7)
  - `monthly:1st-mon`..`monthly:5th-sun`, `monthly:last-mon`..`monthly:last-sun` (35)
- **Count**: capability applicativa hardcoded, SelectField stretto, valori `1 / 3 / 6 / 12` (default 3 nel wizard, salvato come stringa in PB ma parsato a int via `eventCount(record)`).
- **End_date**: opzionale, datetime. Per meetup multi-giorno (weekend ritiro) o per definire la durata della prima occorrenza ricorrente (le successive ereditano lo stesso delta).
- **Location**: opzionale, free-text, principalmente per meetup.
- **Cancellazione**:
  - "Solo questa occorrenza" → `data.cancelled_dates: ["2026-05-11"]` aggiornato. Render skipa la data.
  - "Tutto l'evento" → delete del record (cascade su event_registrations).
  - **Niente override per-occorrenza** (cambio orario/giorno una settimana). Per spostare: cancella questa + crea evento `once` separato. Frizione intenzionale.

## Stato

### ✅ Completato

- **Migration** (`1767000000_events_cadence_location.go` + `1767200000_eventflow_settings_v2.go`): drop `series` + `idx_events_series`, add `cadence`/`count`/`end_date`/`location`, converte `events.type` a testo e sposta i tipi ammessi in `settings.eventflow.types`.
- **Migration cadence Select** (`1767000001_events_cadence_select.go`): converte `cadence` da TextField a SelectField con i 43 valori validi.
- **Migration count Select** (`1767000002_events_count_select.go`): converte `count` da NumberField a SelectField `["1","3","6","12"]`. Snap dei valori esistenti al più vicino.
- **Backend types** (`backend/internal/events/types.go`): drop `TypeRally`+`Series` field, add `StartDate`/`EndDate`/`Cadence`/`Count`/`Location` su `Item`+`MapItem`. Helper `eventCount(record)` per leggere il select-string come int.
- **Eventflow settings v2** (`docs/EVENTFLOW_SETTINGS.md` + migration `1767200000_eventflow_settings_v2.go`): `settings.eventflow.types` e' la source of truth dei tipi evento. `creators` e' array (`["admin", "assistant"]`), `required` e' mappa booleana dei field supportati.
- **Cadence helper** (`backend/internal/events/cadence.go`): `ComputeOccurrences`, `NextOccurrence`, `LastOccurrence`, parser weekday/Nth. Coverage tests OK.
- **Lifecycle** (`backend/internal/events/lifecycle.go`): nuovo `CreateInput` (single record + cadence/count/end_date/location), `Cancel` = delete record, **nuovo** `CancelOccurrence(record, dateStr)` → append a `data.cancelled_dates`. `Reschedule` shifta `start_date` (tutta la serie shifta).
- **Visibility** (`backend/internal/events/visibility.go`): drop `collapseSeries`. `ListForUser` filtra visibility in SQL + window in Go via `nextNonCancelledOccurrence`. Nuovo `OccurrencesFor(record)` ritorna lo schedule canonico con `cancelled`/`past` flags.
- **API handlers** (`api/me/events.go`, `api/admin/events.go`): detail handler ritorna `{event: MapItem, occurrences: [...]}`. Drop `eventsForRegistrationScope` e logica `RegistrationScope == "series"` (1 record = 1 registration). Nuovo handler `CancelOccurrenceHandler`.
- **Routes**: `POST /api/me/events/{id}/cancel-occurrence` + `POST /api/admin/events/{id}/cancel-occurrence` wired.
- **Tests** (`internal/events/lifecycle_test.go`): coverage config-driven type def, `parseDate`, `generateSlug`, `ComputeOccurrencesOnce/Weekly/Monthly/Last`, `NextOccurrenceSkipsCancelled`, `LastOccurrence`, validazioni cadence. `go test ./...` verde.
- **Frontend cadence helper** (`frontend/skeleton/assets/js/components/events-cadence.js`): port JS di `cadence.go`. Espone `computeOccurrences`, `nextOccurrence`, `metaText`, `cadenceLabel`, `shortDate` come `window.appEventsCadence`.
- **Frontend list render** (`me/events-render.js` + `admin/events-render.js`): chiamano `cad.metaText(item)`. Output:
  - meetup once + end_date diverso: `"15 MAY 2026 → 17 MAY 2026 · ROMA"`
  - meetup once: `"15 MAY 2026 · ROMA"`
  - call/meetup ricorrente: `"NEXT 4 MAY 2026 · WEEKLY × 3 · ROMA"`
- **CSS spotlight** (`app.css`): padding simmetrico `var(--space-4)` + `align-content: center` per centrare verticalmente il contenuto. La classe `.list-item-spotlight` è la stessa per active battleplan e first-upcoming event (drop di `.battleplan-list-item-active` come classe duplicata).
- **Pages templates**: `me/events/index.html` e `admin/events/index.html` includono `events-cadence.js` prima di `events-render.js`. La `status-list-content` template ha lo slot `SpotlightID` opzionale per il primo upcoming.
- **Wizard settings-driven** (`frontend/skeleton/assets/js/me/events-wizard*.js`): legge tipi/label/description/required da `settings.eventflow.types`, filtra via `creators`, invia `{type,start_date,cadence,count,end_date,location,url,title,group,data}` e non usa piu' `dates[]`/`series`.
- **Detail occurrences** (`frontend/skeleton/assets/js/components/event-detail.js` + detail templates): render `payload.occurrences`, mostra cancelled/past, espone "Cancel this one" solo con `can_edit`, chiama `POST /api/{scope}/events/{id}/cancel-occurrence`, e il cancel totale non invia piu' `scope`.

### ⏳ Da fare

#### #26 — End-to-end smoke + polish

**Scope**: verifica manuale dei flussi eventi ricorrenti e piccoli fix emersi dal giro browser.

**Task**:
- Eseguire la verifica end-to-end sotto.
- Sistemare eventuali dettagli di render/UX trovati nel detail/list/wizard.

## Note implementative

- **Niente `event_date` field**: rinominato `start_date` ovunque (backend `Item.StartDate`, frontend `item.start_date`). Frontend code che usa `item.event_date` va aggiornato.
- **Niente `series` field**: lookup esistenti (es. summary handlers per cross-event analytics) sono già stati ripuliti, ma se trovi referenze residue → drop.
- **Cancellazione**: `Cancel` (delete record) non accetta più `scope`. Il vecchio param `scope` (this/future/all) è stato rimosso dagli handler. Frontend che lo manda → 400 invalid_payload. Il wizard #26 deve smettere di inviarlo.
- **`RegistrationScope == "series"`** rimosso. Tutti gli eventi hanno scope per-record. La config `eventflow.types[].registration_scope` è ignorata dal backend (può essere ripulita dal seed se vuoi).
- **Spotlight CSS**: `.list-item-spotlight` ora è una classe **generica** in `app.css`, riutilizzabile altrove (request urgente, ecc.). Padding simmetrico, no rounded, full-width edge-to-edge dentro container `.list-card-flush`.
- **`.list-card-flush` / `.list-card-flush-first`**: rinominate da `.battleplan-list-card` / `.battleplan-list-card-flush-first` per essere generiche. Usate sia da battleplan-active che events-spotlight.

## Verifica end-to-end (dopo #25 + #26)

1. Crea call ricorrente: type=call, no group, start=lunedì prox 21:00, cadence=weekly:mon, count=3 → 1 record creato. Lista mostra "NEXT [date] · WEEKLY × 3".
2. Apri detail: 3 occorrenze elencate (4/11/18 maggio). Bottone "Cancel this one" su 11 maggio → cancellata, render rimossa dalla lista occorrenze upcoming.
3. Lista events: la "next" si aggiorna (ora 4 maggio se non passato, altrimenti 18 maggio).
4. Crea meetup multi-giorno: type=meetup, group=X, start=15 May 18:00, end=17 May 12:00, location=Roma, cadence=once → render lista "15 MAY → 17 MAY · ROMA".
5. Cancel intero evento (delete record) → registrations cascade.
6. Il primo upcoming dell'utente è in spotlight (yellow card invertita) — stesso aspetto del battleplan attivo.
