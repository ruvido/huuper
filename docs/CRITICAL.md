# Critical Issues

## Resolved: Settings Exposure

The `/api/settings/{name}` endpoint now whitelists setting names and only returns
safe fields. For `telegram`, it exposes only `data.name` (not the bot token).
The `settings` collection is admin-only via migration `1763300000_settings_auth_only.go`.

## High: Telegram Tokens Never Expire

Telegram connect tokens are only cleaned up when a new token is generated. If no
one generates new tokens, old tokens remain valid indefinitely, so the intended
24-hour expiry is not enforced. Current cleanup runs on token generation and
deletes records older than 7 days in `api/telegram_token.go`.
