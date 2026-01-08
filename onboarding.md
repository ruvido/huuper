
## Processo di Onboarding Realmen

### Gli step (3-4 settimane)

1. Candidato compila form (con regione)
2. Admin centrale riceve email -> approvazione (filtro anti-spam)
3. Email a candidato + responsabile locale (regione)
4. Responsabile assegna angelo custode
5. Angelo custode prende in mano la situazione per conoscenza one-to-one (2 call)
6. Angelo genera link telegram a tempo per gruppo locale da Dashboard-gruppo → candidato diventa visitor (auditore, 2 call di gruppo, vedi doc angelo custode)
7. Approvazione: Responsabile accetta nuovo utente tramite Dashboard-gruppo
8. Accettazione inviata ad Admin;
9. Admin accetta. Via libera a firma Patto di Ingresso in call del gruppo → stato active → link a tempo per chat generale (link in Dashboard)


### Stati utente

| Stato | Dove si trova |
|-------|---------------|
| **Pending** | Richiesta fatta, attende approvazione admin |
| **Assigned** | Angelo assegnato, conoscenza one-to-one |
| **Visitor** | Nel gruppo locale, auditore alle call |
| **Approved** | Approvato, in attesa firma patto |
| **Active** | Membro completo: gruppo locale + chat generale |

---

## Documento 2: Prompt per Claude Code

```markdown
# Huuper — Onboarding Gruppi Realmen

## Contesto
Webapp per gestire onboarding in gruppi maschili su Telegram. Processo graduale (~3-4 settimane). Due accessi Telegram distinti:
- Gruppo locale (per visitor, auditore alle call)
- Chat generale (solo active)

## Stack
- Backend: PocketBase
- Integrazioni: Telegram Bot API, SMTP

## Note tecniche
- Collection "Guardians" per relazione user → angelo custode assegnato
- Stati utente: pending, assigned, visitor, approved, active, rejected
- Catena approvazioni: guardian → leader → admin (collection Approvals)
- Invite link Telegram unici con scadenza

## Flusso

1. **Form pubblico** → crea User status=pending → email admin
2. **Admin approva** → assegna gruppo da provincia (nuova collection, ad es. lazio,umbria-> centro (relation)) → status=assigned → email candidato + leader
3. **Leader assegna angelo** → crea record in Guardians → notifica angelo
4. **Angelo traccia step 1-2** (conoscenza one-to-one)
5. **Step 2 completato** → huuper genera invite gruppo locale → status=visitor
6. **Angelo traccia step 3-4** (auditore alle call)
7. **Catena approvazioni** → status=approved
8. **Firma patto in call** → status=active → genera invite chat generale

## Email
1. Admin ← nuova richiesta
2. Candidato ← richiesta ricevuta
3. Leader ← nuovo candidato assegnato
4. Angelo ← candidato da accompagnare
