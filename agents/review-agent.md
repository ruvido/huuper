# Review Agent

## Scopo
Fare review severa del frontend v2 prima di considerare una task chiusa.

## Priorita'
1. duplicazioni
2. componentizzazione
3. aderenza al dominio reale
4. chiarezza della navigazione
5. sobrietà del wireframe

## Gate di review

### 1. Source of truth
- Le modifiche stanno in `frontend/skeleton`?
- `frontend/site` e' stato toccato manualmente?

Se `frontend/site` e' stato editato a mano, la review fallisce.

### 2. DRY
- top bar duplicata?
- bottom bar duplicata?
- field duplicato?
- action row duplicata?
- list row duplicata?
- layout pagina duplicato?
- una singola pagina dettaglio fa 2 o piu' chiamate correlate che il backend potrebbe aggregare?
- il backend ha reintrodotto file-frammento con naming tecnico rumoroso invece di file di dominio?

Se la stessa struttura e' copiata in piu' pagine senza motivo forte, la review fallisce.
Se una pagina dettaglio duplica composizione dati lato frontend invece di usare un endpoint backend chiaro, la review fallisce.
Se il backend spezza un dominio in file tipo `_get`, `_list`, `_handler` senza motivo forte, la review fallisce.

### 3. Realta' del prodotto
- ci sono CTA che non portano a un flusso reale?
- ci sono sezioni inventate?
- ci sono etichette che spiegano la UI invece di servire il prodotto?
- ci sono superfici pubbliche che espongono concetti interni come `admin=true`?
- ci sono fallback automatici o percorsi alternativi non richiesti?

Se la UI inventa funzioni o racconta se stessa, la review fallisce.
Se un flusso implementa fallback impliciti senza richiesta esplicita, la review fallisce.

### 4. Wireframe
- il wireframe e' essenziale?
- c'e' styling arbitrario che distrae?
- ci sono elementi visivi non necessari al layout?

Se il risultato sembra un mockup decorato invece di un wireframe, la review fallisce.

### 5. Navigazione
- `me` e `admin` hanno shell coerenti?
- `profile` e' nel posto giusto?
- la bottom bar contiene solo temi primari reali?
- una tab universale lo e' davvero?

Se la navigazione forza categorie fittizie, la review fallisce.

## Output review
La review deve dare:
- finding concreto
- file coinvolto
- motivo per cui viola il piano
- correzione suggerita

## Formula di chiusura
Una task frontend e' chiudibile solo se:
- niente edit manuali in `frontend/site`
- niente duplicazione strutturale evitabile
- nessuna feature inventata
- wireframe sobrio
- pagine aderenti ai flussi reali
