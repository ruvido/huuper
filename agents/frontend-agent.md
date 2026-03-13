# Frontend Agent

## Scopo
Costruire e refactorare il frontend v2 in modo DRY, semantico, riusabile e coerente con l'approccio `frontend/skeleton -> frontend/site`.

## Boundary
- Modificare solo `frontend/skeleton`.
- `frontend/site` e' output generato.
- E' vietato editare manualmente `frontend/site`, salvo bootstrap esplicitamente richiesto e temporaneo.
- Il frontend legacy vive in `archive/frontend-svelte/` e non va toccato salvo richiesta esplicita.

## Regola principale
Prima i blocchi riusabili, poi le pagine.

Questo significa:
- non duplicare top bar, bottom bar, shell pagina, field, actions, list, card
- non creare nuove pagine se prima non esiste il componente/layout necessario
- non copiare HTML tra `public`, `me`, `admin`

## Struttura target
Il frontend v2 deve convergere verso una struttura come questa:

```text
frontend/skeleton/
  components/
  layouts/
  public/
  me/
  admin/
  assets/
```

Se il tree corrente non e' ancora li', ogni modifica deve andare in quella direzione.

## DRY
- Una shell condivisa si definisce una volta.
- Una top bar condivisa si definisce una volta.
- Le bottom bar si definiscono per dominio, non per pagina.
- I campi form si definiscono una volta.
- Le varianti `me` e `admin` riusano gli stessi blocchi di base, cambiando solo struttura e contenuto necessario.
- Una pagina dettaglio non deve fan-out su piu' endpoint se il backend puo' aggregare il payload in modo chiaro.

## Componenti obbligatori
Prima di continuare ad aggiungere pagine, devono esistere o emergere chiaramente questi blocchi:
- shell pagina
- top bar
- bottom nav membro
- bottom nav admin
- field
- action row
- list row
- section/panel

## Wireframe policy
- Wireframe vero, non mockup.
- Nessun testo meta che spiega la UI.
- Nessuna feature inventata.
- Nessun bottone senza destinazione reale o comportamento previsto.
- Nessuna navigazione finta.
- Nessuna sezione "placeholder" travestita da prodotto.
- Ogni pagina deve contenere solo cio' che e' strettamente necessario.
- Nessun fallback implicito o comportamento alternativo non richiesto.
- Nessuna logica "se fallisce allora prova anche..." senza richiesta esplicita.
- Evitare composizione client-side di una singola vista tramite molte chiamate correlate: preferire endpoint di dettaglio backend.

## Public / Me / Admin
- `public` contiene solo flussi pubblici reali.
- `me` e `admin` sono superfici separate.
- `me` e `admin` riusano i componenti, ma non si mescolano.
- `profile` non va forzato nella bottom bar se appartiene alla top bar.

## CSS
- CSS fortemente semantico.
- Ridurre le classi al minimo.
- Less is more.
- Mobile first sempre.
- In fase wireframe:
  - evitare design gratuito
  - evitare gerarchie estetiche arbitrarie
  - evitare decorazioni non necessarie

## Hot reload
- In sviluppo si usa il flusso backend con `--frontend-dev`.
- La sorgente osservata e' `frontend/skeleton`.
- L'output generato e' `frontend/site`.
- Vietato introdurre watcher custom opachi o pipeline alternative senza conferma.

## Quando implementi
Per ogni task frontend:
1. verificare se il blocco esiste gia'
2. se non esiste, estrarre prima il componente/layout
3. poi applicare la pagina
4. controllare che `frontend/site` sia solo output

## Red flags
Blocca e correggi se vedi:
- HTML duplicato tra pagine
- top bar copiata piu' volte
- bottom bar copiata piu' volte
- CTA non reali
- testo esplicativo invece di contenuto UI reale
- edit manuali in `frontend/site`
- fallback automatici non richiesti
- logica inferita dal contesto invece che confermata
