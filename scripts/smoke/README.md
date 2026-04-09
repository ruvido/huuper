# Smoke Scripts

Smoke test operativi per il backend v2.

## File
- `request_notifications.sh`: smoke interattivo canonico per verificare il request flow passo per passo, con ruoli reali e email attese
- `auth_admin.sh`: ottiene un token admin da `scripts/.env`
- `auth_member.sh`: ottiene un token member da `scripts/.env`

## Variabili richieste in `scripts/.env`
- `BASE`
- `ADMIN_EMAIL`
- `ADMIN_PASSWORD`
- `MEMBER_EMAIL` oppure `USER_EMAIL`
- `MEMBER_PASSWORD` oppure `USER_PASSWORD`

## Variabili opzionali
- `ADMIN_TOKEN`: bypass del login admin
- `MEMBER_TOKEN`: bypass del login member
- `GUARDIAN_TOKEN`: bypass del login guardian per lo smoke request
- `GUARDIAN_EMAIL`
- `GUARDIAN_PASSWORD`
- `REQUEST_SMOKE_GROUP_ID`: gruppo usato per lo smoke request
- `REQUEST_SMOKE_GUARDIAN_ID`: guardian usato per lo smoke request

## Uso
```bash
scripts/smoke/request_notifications.sh
```

Lo script crea una request con email `requests.smoke.<timestamp>@realmen.it`, usa i ruoli reali del flow e chiede conferma prima di ogni step:

- `set_group`
- `set_guardian`
- `set_mentoring_done`
- `set_group_approved`
- `set_admin_approved`

Per ogni step mostra quale email ti aspetti di vedere.

Non esiste piu' un entrypoint smoke generico: per il dominio `request` questo script e' la source of truth operativa.
