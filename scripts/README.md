# Scripts

Script operativi minimali del repository.

## Tree
- `smoke/`: smoke test backend v2
- `send_email_event.sh`: invio email per evento da `slug`, con test su `ADMIN_EMAIL` e conferma prima del live

## `send_email_event.sh`
Accetta file markdown con formato unico YAML front matter:
`---`
`subject: Oggetto`
`---`

Il body email e' tutto il contenuto successivo, interpretato come markdown.

## Note
- Il build frontend e il deploy sono gestiti sotto `deploy/`
- Non mantenere qui script storici ridondanti o one-off
