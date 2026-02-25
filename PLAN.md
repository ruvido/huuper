# PLAN

Operational continuation plan for the data-driven Realmen onboarding flow.

## Current State

- `requests` collection exists (admin-only access) with:
  - `guardian` relation -> `users`
  - `data` JSON
- `groups` has `assistant` relation -> `users` (`MaxSelect: 1`)
- Settings upsert migrations added for:
  - `requests_flow`
  - `users` (`pact_required`)
  - `signup` (minimal public form fields)
- New API endpoints exist:
  - `POST /api/requests/submit` (public)
  - `POST /api/requests/{id}/action` (auth)
- `action` supports:
  - `transition`
  - `reject`

## Critical Decisions (Locked)

- Flow is data-driven from `settings.requests_flow`.
- Public intake fields are data-driven from `settings.signup`.
- `rejected` must be a top-level bool field on `requests` (NOT inside `data`).
- `assistant` is unique per group via `groups.assistant` single relation.
- Keep architecture minimal: no extra log collections for now.

## Immediate Priority

Implement and verify the `requests.rejected` migration and runtime consistency.

### Why

- API logic already moved to top-level `rejected`.
- DB must include the field for all existing environments.

### Action

1. Run migration `1764500005_requests_rejected_field.go`.
2. Verify backfill from `data.rejected` to top-level `rejected`.
3. Verify `data.rejected` is no longer used in new records.

### Verify (DB)

Use sqlite checks on active DB:

```sql
PRAGMA table_info(requests);
SELECT id, rejected, json_extract(data, '$.status') AS status FROM requests LIMIT 20;
```

## Next Implementation Steps

1. Complete `action` API with `promote`.
   - Input: `{ "action": "promote" }`
   - Allowed: admin only
   - Allowed status: `6-admin_approved`
   - Behavior:
     - create/update `users` record
     - apply `settings.users.pact_required`
     - remove request (or mark completed if deletion policy changes)

2. Add list API for dashboard.
   - `GET /api/requests`
   - Filters (query params):
     - `status`
     - `rejected`
     - `group_id`
     - `guardian`
   - Sorting:
     - default `-updated`

3. Frontend dashboard minimum.
   - Build columns from `requests_flow.statuses`.
   - Card actions:
     - next step (`transition`)
     - reject (`reject`)
     - promote (`promote`, when eligible)

## API Contracts (Current)

### Submit

`POST /api/requests/submit`

Body:

```json
{
  "name": "Mario Rossi",
  "email": "mario.rossi@example.com",
  "mobile": "+393331112233",
  "region": "Lombardia",
  "birth_year": "1990",
  "marital_status": "Single",
  "children": "No",
  "motivation": "..."
}
```

### Action: transition

`POST /api/requests/{id}/action`

```json
{
  "action": "transition",
  "target_status": "2-group_assigned"
}
```

### Action: reject

`POST /api/requests/{id}/action`

```json
{
  "action": "reject",
  "reason": "..."
}
```

## Known Risks / Notes

- Ensure the active runtime DB path is correct (`pb_data` nesting issue happened before).
- Do not rely on `settings.onboarding` for request flow; use `settings.requests_flow`.
- Keep `settings` keys stable; avoid renaming once frontend starts consuming them.

## Definition of Done (Near Term)

1. `requests.rejected` field exists in DB and is populated correctly.
2. `action=transition` and `action=reject` fully operate against top-level `rejected`.
3. `action=promote` implemented and validated end-to-end.
4. Dashboard can load and act on requests without hardcoded statuses.
