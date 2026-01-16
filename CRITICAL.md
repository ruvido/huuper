# Critical Issues

## Critical: Settings Exposure

The `/api/settings/{name}` endpoint returns the full `data` payload to any
authenticated user. This leaks sensitive values such as
`settings.telegram.token`, even though the `settings` collection is admin-only.

We are **not** addressing this now, but it needs to be fixed before production.

## High: Telegram Tokens Never Expire

Telegram connect tokens are only cleaned up when a new token is generated. If no
one generates new tokens, old tokens remain valid indefinitely, so the intended
24-hour expiry is not enforced.
