# settings.battleplan — Schema Reference

The `settings` collection row with `name = "battleplan"` holds runtime
configuration for the battleplan feature. Its `data` JSON is consumed by:

- The Go backend (`backend/internal/battleplans/config.go`, `Config` struct)
  for input validation in `Create`/`Update`.
- The frontend wizard (`frontend/skeleton/assets/js/admin/battleplan-wizard*.js`)
  for rendering steps, defaults, copy strings.

The runtime contract is enforced by `validateSettings()` and `validateCopy()`
in `frontend/skeleton/assets/js/admin/battleplan-wizard-config.js`. Anything
they require is mandatory for the wizard to mount.

## Top-level fields

| Field        | Type                  | Required | Consumer                                    |
|--------------|-----------------------|----------|---------------------------------------------|
| `priority`   | object                | yes      | wizard step 1 copy (new vs edit)            |
| `wizard`     | object                | yes      | intro + confirmation step copy              |
| `durations`  | array of `{value, default?}` | yes | duration picker on intro/priority step    |
| `visibility` | array of `{value, label, default?}` | yes | visibility picker, backend validation |
| `pillars`    | array of `{key, label, description}` | yes | wizard pillar steps + backend validation |
| `cadences`   | array of `{type, label, default?}` | yes | routine cadence selector                  |

### `priority`

```
{
  "new":  { "title": string, "text": string },
  "edit": { "title": string, "text": string }
}
```

`priority.new` shown when creating a new plan; `priority.edit` when editing
an existing one or viewing read-only. Both required (validateSettings throws
otherwise). Legacy `priority.label` / `priority.description` (single copy
block) have been migrated to the `new` / `edit` split — see migration
`1766400000_battleplan_settings.go` for the seed shape (which still uses
the legacy keys; an in-place data migration is expected before relying on
the wizard).

### `wizard`

```
{
  "intro": {
    "show": boolean (optional, defaults true),
    "title": string (required when show=true),
    "text":  string,
    "button": string (required when show=true)
  },
  "confirmation": {
    "title": string (required),
    "text":  string,
    "button": string
  }
}
```

When `wizard.intro.show === false`, the intro step is suppressed and
`title`/`button` are not required.

### `durations`

Array of `{ "value": number, "default"?: bool }`. At least one entry must
be marked `default: true`. `value` is days (e.g. 30, 60, 90).

### `visibility`

Array of `{ "value": string, "label": string, "default"?: bool }`.
Exactly one default expected. Backend uses `value` for record validation;
frontend renders `label`.

### `pillars`

Array of `{ "key": string, "label": string, "description": string }`. The
wizard renders one step per pillar (in array order). Backend
`Config.IsValidPillarKey` rejects unknown keys at write time.

### `cadences`

Array of `{ "type": string, "label": string, "default"?: bool }`. Valid
types: `paused`, `daily`, `specific_days`, `times_per_week`. The default
type seeds new routines.

## Example document

```json
{
  "priority": {
    "new":  { "title": "Nuovo Piano",     "text": "Scegli una priorità." },
    "edit": { "title": "Modifica Piano",  "text": "Aggiorna la priorità." }
  },
  "durations":  [{ "value": 30 }, { "value": 60, "default": true }, { "value": 90 }],
  "visibility": [{ "value": "group", "label": "Gruppo", "default": true },
                 { "value": "public", "label": "Pubblico" }],
  "pillars": [
    { "key": "interiorita", "label": "Interiorità", "description": "" },
    { "key": "relazioni",   "label": "Relazioni",   "description": "" },
    { "key": "risorse",     "label": "Risorse",     "description": "" },
    { "key": "salute",      "label": "Salute",      "description": "" }
  ],
  "cadences": [
    { "type": "paused",         "label": "In pausa",     "default": true },
    { "type": "daily",          "label": "Ogni giorno" },
    { "type": "specific_days",  "label": "Giorni specifici" },
    { "type": "times_per_week", "label": "Volte a settimana" }
  ],
  "wizard": {
    "intro":        { "show": true, "title": "Piano di Battaglia", "text": "...", "button": "INIZIA" },
    "confirmation": { "title": "Pronto a partire", "text": "...", "button": "Conferma" }
  }
}
```
