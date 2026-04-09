# Smoke Scripts

Smoke test operativi per il backend v2.

## File
- `run.sh`: entrypoint unico per lo smoke test `public / me / admin`
- `request_notifications.sh`: smoke interattivo per verificare il request flow passo per passo e le email attese
- `auth_admin.sh`: ottiene un token admin da `scripts/.env`
- `auth_member.sh`: ottiene un token member da `scripts/.env`

## Variabili richieste in `scripts/.env`
- `BASE`
- `ADMIN_EMAIL`
- `ADMIN_PASSWORD`
- `MEMBER_EMAIL` oppure `USER_EMAIL`
- `MEMBER_PASSWORD` oppure `USER_PASSWORD`

## Variabili opzionali
- `EVENT_SLUG`: abilita il blocco smoke per gli eventi
- `ADMIN_TOKEN`: bypass del login admin
- `MEMBER_TOKEN`: bypass del login member
- `GUARDIAN_TOKEN`: bypass del login guardian per lo smoke request a ruoli reali
- `GUARDIAN_EMAIL`
- `GUARDIAN_PASSWORD`
- `REQUEST_SMOKE_GROUP_ID`: gruppo usato per il pipeline smoke delle request
- `REQUEST_SMOKE_GUARDIAN_ID`: guardian usato per il pipeline smoke delle request

## Uso
```bash
scripts/smoke/run.sh
```

Per verificare le email del request flow in modo interattivo:

```bash
scripts/smoke/request_notifications.sh
```

Lo script crea una request con email `requests.smoke.<timestamp>@realmen.it` e chiede conferma prima di ogni step del flow:

- `set_group`
- `set_guardian`
- `set_mentoring_done`
- `set_group_approved`
- `set_admin_approved`

Per ogni step mostra quale email ti aspetti di vedere.
