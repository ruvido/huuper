# Critical Issue: Settings Exposure

The `/api/settings/{name}` endpoint used to allow public access and return full
`settings` records, including sensitive values like the Telegram bot token. This
is now mitigated by requiring auth on the route and locking the `settings`
collection to admin-only, but the token still lives in the `settings` data.

We are **not** addressing this now, but it needs to be fixed before production.
