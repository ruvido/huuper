# Smoke Scripts

Smoke test operativi per il backend v2.

## File
- `run.sh`: entrypoint unico per lo smoke test `public / me / admin`
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

## Uso
```bash
scripts/smoke/run.sh
```
