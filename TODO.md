## Battleplan 
### critical
X save draft
X duplicate
X icona freccia su next (fulmine solo alla fine)
X invece che "+" "crea nuovo piano" in evidenza (se c'e' draft => modifica draft)
X note per il piano
x set end date
x centrare textarea anche rispetto al fatto che si apre la tastiera in basso
x bug: titolo mancante alla fine del processo
X bug: spazio in alto sopra al titolo
x bug: non appare il trattino per mettere la data
- Uniforma la disposizione dei bottoni in battle plan view sia se è active sia se è draft. Quando il battle plan è un draft, metti tre bottoni: 1. edit largo tratteggiato poi 2 archivo e 3 activate. 
- margine sx sputtanato tra topbar e resto della pagina!
- timeline : quadratino deve avere prima un padding in alto dove si vede un pezzo di trattino verticale poi dopo si vede il quadratino. ora inizia da quadratino
- cosmetic fixes!!


### settings
- descrizione pilastri mancante

### nice to have
- icona "libro" per link alla pagina pdb
- reminder in request a 7gg e 1gg dalla scadenza (anche un celebrate!)

---

Le routine dovrebbero essere salvate in una collection a parte. Nei campi ci deve essere anche relation to user. Un altro campo deve contenere l'array di quanti user stanno usando quella routine. Perché le routine possono essere condivisibili cioè leggibili da tutti. 

## Onboarding
- gruppo locale "greyed out" se non sei entrato nel gruppo generale (breve spiegazione in copy) prima entri in default poi dpuoi entrare in other groups. come si definisce questo comportamento in maniera data-driven?


## request
il json e' ancora un fritto misto di admin_approved= e name=

 1. Verifica end-to-end reale
     Provare questi 3 casi su dev/staging:
      - user completo senza onboarding_completed_at → migration lo marca complete e login va a /me/;
      - user incompleto con password valida → login redirige a /onboarding/?token=...;
      - user in limbo → migration resetta password/token e logga nuovo URL.
  2. Controllare i log migration
     Dopo startup/deploy cerca:
      - [migration limbo-onboarding] reset user
      - [migration onboarding-backfill] marked complete
      - [migration onboarding-backfill] user requires onboarding