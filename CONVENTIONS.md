# CONVENTIONS

## Scopo
Convenzioni operative e di design per il backend del repository.

## Priorita'
1. Chiarezza.
2. Semplicita'.
3. DRY.
4. Robustezza.
5. Discoverability.

## Namespacing API
- Usare namespace espliciti e stabili.
- Convenzione target:
  - `/api/public/...`
  - `/api/me/...`
  - `/api/admin/...`
- Evitare namespace ambigui come `private` o troppo vaghi rispetto al dominio.

## Modello di accesso
- `public`: risorse anonime o bootstrap pubblico.
- `me`: risorse dell'utente autenticato e delle sue relazioni autorizzate.
- `admin`: operazioni globali sulle risorse applicative.

## Regola di esposizione dati
- Un endpoint `me` non espone mai dati globali non legati all'utente autenticato.
- Un endpoint `admin` puo' operare su qualsiasi record applicativo.
- Le policy non devono essere ricostruite in ogni handler: vanno centralizzate in helper condivisi.

## Tree backend
- Organizzare il backend per domini chiari, non per accumulo storico.
- Privilegiare file piccoli e responsabilita' chiare.
- Evitare package o helper "misc" senza semantica precisa.
- Prima di spostare file o cambiare struttura, proporre il target tree.

## Handler
- Un handler deve fare poche cose:
  - validare input
  - delegare authz
  - delegare logica condivisa
  - costruire la response
- Evitare query, policy e trasformazioni duplicate in piu' handler.

## Authz
- Centralizzare controlli come:
  - require auth
  - require admin
  - require self
  - require membership
  - require visibility on request/group/resource
- I nomi degli helper devono descrivere chiaramente la capability verificata.

## Input e output
- Input chiari, stretti e prevedibili.
- Errori coerenti e stabili.
- Paginazione coerente dove serve.
- Filtri e sort espliciti e documentabili.
- Le response devono essere adatte a client diversi senza dipendere dalla UI web.

## PocketBase
- Usare PocketBase come base standard per auth, metadata, relazioni e persistence.
- Preferire feature ufficiali del framework rispetto a logica custom evitabile.
- Le settings JSON sono ammesse, ma parsing e validazione devono stare in punti unici e tipizzati quanto basta.

## Refactoring
- Refactor piccoli e incrementali.
- Prima consolidare policy e naming, poi struttura e route.
- Nessun refactor strutturale senza una proposta chiara del delta.

## Review checklist
- Riduce duplicazione?
- Migliora chiarezza e discoverability?
- Rafforza i confini `public / me / admin`?
- Riduce comportamento implicito?
- Mantiene il backend compilabile?
