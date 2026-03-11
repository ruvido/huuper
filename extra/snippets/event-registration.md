---
title: "II Raduno Nazionale"
event_slug: "ii-raduno-nazionale"
api_base: "http://localhost:8090"
url: "/ii-raduno"
---

## Realmen conoscono il proprio corpo!

![II Raduno Nazionale ad Assisi](/images/ii-raduno-assisi.jpg)

Tre giorni per rimettere al centro il corpo, la disciplina e la fratellanza. Un raduno intenso in cui allenamento, confronto e momenti di verita si intrecciano per fare un salto in avanti insieme.

- **Quando:** 20-22 Marzo 2026
- **Dove:** Assisi (PG)

### Cosa faremo

- Sessioni di allenamento funzionale e mobilita.
- Incontri di gruppo e accountability dal vivo.
- Workshop pratici su abitudini, energia e resilienza.
- Tempo strutturato per confronto e crescita personale.

### Cosa portare

- Abbigliamento comodo per allenamento indoor/outdoor.
- Scarpe da ginnastica e un cambio completo.
- Un quaderno per appunti e riflessioni.
- Documento di identita.
- Contanti per la quota in struttura.

---

### Come iscriversi

Per partecipare, compila il form qui sotto.

I **posti sono limitati** e diamo priorita a chi sta gia partecipando ai [gruppi di accountability](). Per coprire tutti i costi dell'evento sara necessario versare una quota di **100 euro** in contanti presso la struttura.

Ricevuta la tua richiesta ti **contatteremo telefonicamente** nel giro di qualche giorno per confermare la tua partecipazione.

### Procedura di partecipazione

1. Compila il form con i tuoi dati reali.
2. Attendi la chiamata di conferma da parte del team.
3. Riceverai i dettagli logistici e le istruzioni finali.
4. Presentati in struttura con la quota di partecipazione in contanti.

<hr />

<form data-event-slug="{{< param "event_slug" >}}" data-api-base="{{< param "api_base" >}}">
  <fieldset>
    <legend><strong>I tuoi dati</strong></legend>
    <div class="grid">
      <input name="name" placeholder="Nome e Cognome" required />
      <input type="tel" name="mobile" placeholder="Numero di telefono (10 cifre)" inputmode="numeric" pattern="[0-9]{10}" minlength="10" maxlength="10" title="Inserisci esattamente 10 cifre" required />
      <input type="email" name="email" placeholder="Email" required />
    </div>
  </fieldset>
  <fieldset>
    <legend><strong>Quanti anni hai?</strong></legend>
    <div class="grid radio-grid">
      <label><input type="radio" name="age_range" value="18-24" required /> 18-24</label>
      <label><input type="radio" name="age_range" value="25-32" required /> 25-32</label>
      <label><input type="radio" name="age_range" value="33-39" required /> 33-39</label>
      <label><input type="radio" name="age_range" value="40+" required /> 40+</label>
    </div>
  </fieldset>
  <fieldset>
    <legend><strong>Da dove vieni?</strong></legend>
    <div class="grid radio-grid">
      <label><input type="radio" name="region" value="nord-est" required /> Nord-est</label>
      <label><input type="radio" name="region" value="nord-ovest" required /> Nord-ovest</label>
      <label><input type="radio" name="region" value="centro" required /> Centro</label>
      <label><input type="radio" name="region" value="sud-isole" required /> Sud e isole</label>
    </div>
  </fieldset>
  <fieldset>
    <legend><strong>Situazione sentimentale</strong></legend>
    <div class="grid radio-grid radio-grid--three">
      <label><input type="radio" name="marital_status" value="single" required /> Single</label>
      <label><input type="radio" name="marital_status" value="fidanzato" required /> Fidanzato</label>
      <label><input type="radio" name="marital_status" value="sposato" required /> Sposato</label>
    </div>
  </fieldset>
  <button type="submit" data-submit-button>
    <span data-submit-label><strong>Iscriviti</strong></span>
    <span data-submit-spinner aria-hidden="true"></span>
  </button>
</form>
<div data-event-message role="status" aria-live="polite">
  <div data-event-card tabindex="-1">
    <div data-event-icon>
      <svg data-icon-success viewBox="0 0 52 52" aria-hidden="true">
        <circle data-icon-ring cx="26" cy="26" r="24" fill="none" />
        <path data-icon-mark fill="none" d="M14.1 27.2l7.1 7.2 16.7-16.8" />
      </svg>
      <svg data-icon-error viewBox="0 0 52 52" aria-hidden="true">
        <circle data-icon-ring cx="26" cy="26" r="24" fill="none" />
        <path data-icon-mark fill="none" d="M16 16l20 20M36 16l-20 20" />
      </svg>
    </div>
    <div data-event-text></div>
  </div>
</div>

<style>
  [data-event-message] {
    margin-top: 0.75rem;
    display: none;
    justify-content: center;
  }
  [data-event-message][data-state] {
    display: flex;
  }
  [data-event-card] {
    margin-top: 2rem;
    width: min(320px, 92vw);
    //aspect-ratio: 1 / 1;
    border-radius: 1.25rem;
    border: 1px solid transparent;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    gap: 1rem;
    padding: 1.5rem;
    transform: translateY(6px);
    opacity: 0;
    transition: opacity 180ms ease, transform 180ms ease, border-color 180ms ease, background 180ms ease;
  }
  [data-event-message][data-state] [data-event-card] {
    opacity: 1;
    transform: translateY(0);
  }
  [data-event-message][data-state="success"] [data-event-card] {
    background: #ecfdf3;
    border-color: #a7f3d0;
    color: #065f46;
  }
  [data-event-message][data-state="error"] [data-event-card] {
    background: #fef2f2;
    border-color: #fecaca;
    color: #991b1b;
  }
  [data-event-icon] {
    width: 120px;
    height: 120px;
  }
  [data-event-icon] svg {
    width: 100%;
    height: 100%;
  }
  [data-icon-success],
  [data-icon-error] {
    display: none;
  }
  [data-event-message][data-state="success"] [data-icon-success],
  [data-event-message][data-state="error"] [data-icon-error] {
    display: block;
  }
  [data-icon-ring] {
    stroke-width: 2.5;
    stroke-dasharray: 166;
    stroke-dashoffset: 166;
    animation: event-stroke 0.6s cubic-bezier(0.65, 0, 0.45, 1) forwards;
  }
  [data-icon-mark] {
    stroke-width: 3.25;
    stroke-linecap: round;
    stroke-linejoin: round;
    stroke-dasharray: 48;
    stroke-dashoffset: 48;
    animation: event-stroke 0.3s cubic-bezier(0.65, 0, 0.45, 1) 0.6s forwards;
  }
  [data-event-message][data-state="success"] [data-icon-ring],
  [data-event-message][data-state="success"] [data-icon-mark] {
    stroke: #22c55e;
  }
  [data-event-message][data-state="error"] [data-icon-ring],
  [data-event-message][data-state="error"] [data-icon-mark] {
    stroke: #ef4444;
  }
  [data-event-text] {
    font-size: 1rem;
    line-height: 1.5;
    font-weight: 600;
  }
  form [data-submit-button] {
    position: relative;
    min-width: 9.5rem;
    background: var(--inverse);
    border-color: var(--inverse);
    color: #fff;
  }
  form [data-submit-button]:is(:hover, :focus) {
    background: var(--inverse);
    border-color: var(--inverse);
  }
  hr {
    width: 50%;
    margin: 1rem auto 1.25rem;
    border: 0;
    border-top: 1px solid #000000;
  }
  .radio-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
  .radio-grid--three {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
  @media (min-width: 900px) {
    .radio-grid {
      grid-template-columns: repeat(4, minmax(0, 1fr));
    }
    .radio-grid--three {
      grid-template-columns: repeat(3, minmax(0, 1fr));
    }
  }
  [data-submit-spinner] {
    width: 1.05rem;
    height: 1.05rem;
    border-radius: 999px;
    border: 2px solid currentColor;
    border-right-color: transparent;
    display: none;
    animation: event-spin 700ms linear infinite;
  }
  [data-submit-button][data-loading="true"] [data-submit-label] {
    opacity: 0.25;
  }
  [data-submit-button][data-loading="true"] [data-submit-spinner] {
    display: inline-block;
  }
  @keyframes event-spin {
    to {
      transform: rotate(360deg);
    }
  }
  @keyframes event-stroke {
    100% {
      stroke-dashoffset: 0;
    }
  }
</style>

<script>
(() => {
  const form = document.querySelector("form[data-event-slug][data-api-base]");
  const message = document.querySelector("[data-event-message]");
  if (!form || !message) return;
  const messageText = message.querySelector("[data-event-text]");

  const button = form.querySelector("[data-submit-button]");
  const setLoading = (value) => {
    if (!button) return;
    button.dataset.loading = value ? "true" : "false";
    button.disabled = !!value;
  };

  const setMessage = (state, value) => {
    message.dataset.state = state;
    if (messageText) {
      messageText.textContent = value;
    }
    const card = message.querySelector("[data-event-card]");
    if (card && typeof card.focus === "function") {
      card.focus();
    }
  };

  const messages = {
    invalidEvent: "Invalid event.",
    invalidEmail: "Email non valida.",
    errorGeneric: "Registration failed.",
    eventClosed: "Registrations are closed.",
    alreadySubmitted: "Registration already submitted.",
    successWithEmail: "Check your email for confirmation.",
    successNoEmail: "Thanks, we received your registration."
  };

  const slug = (form.dataset.eventSlug || "").trim();
  const apiBase = (form.dataset.apiBase || "").trim().replace(/\/$/, "");
  if (!slug || !apiBase) {
    setMessage("error", messages.invalidEvent);
    return;
  }

  const loadMessages = async () => {
    const escapedSlug = slug.replace(/"/g, '\\"');
    const url = `${apiBase}/api/collections/events/records?filter=slug="${escapedSlug}"&fields=id,data`;
    try {
      const res = await fetch(url);
      if (!res.ok) return;
      const payload = await res.json().catch(() => ({}));
      const record = payload && Array.isArray(payload.items) ? payload.items[0] : null;
      const eventMessages = record && record.data && record.data.messages ? record.data.messages : null;
      if (!eventMessages) return;

      if (typeof eventMessages.invalid_event === "string") messages.invalidEvent = eventMessages.invalid_event;
      if (typeof eventMessages.invalid_email === "string") messages.invalidEmail = eventMessages.invalid_email;
      if (typeof eventMessages.error_generic === "string") messages.errorGeneric = eventMessages.error_generic;
      if (typeof eventMessages.event_closed === "string") messages.eventClosed = eventMessages.event_closed;
      if (typeof eventMessages.already_submitted === "string") messages.alreadySubmitted = eventMessages.already_submitted;
      if (typeof eventMessages.success_with_email === "string") messages.successWithEmail = eventMessages.success_with_email;
      if (typeof eventMessages.success_no_email === "string") messages.successNoEmail = eventMessages.success_no_email;
    } catch {
      // keep defaults
    }
  };

  const messageForError = (payload) => {
    if (payload && payload.message === "invalid_email") return messages.invalidEmail;
    if (payload && payload.message === "event_closed") return messages.eventClosed;
    if (payload && payload.message === "already_submitted") return messages.alreadySubmitted;
    if (payload && payload.data && typeof payload.data === "object") {
      const uniqueError = Object.values(payload.data).some((item) => {
        return (
          item &&
          typeof item === "object" &&
          typeof item.code === "string" &&
          item.code.includes("unique")
        );
      });
      if (uniqueError) return messages.alreadySubmitted;
    }
    return messages.errorGeneric;
  };

  loadMessages();

  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    setLoading(true);

    try {
      const formData = new FormData(form);
      const entries = Object.fromEntries(formData.entries());
      const email = entries.email || "";
      delete entries.email;

      const res = await fetch(`${apiBase}/api/events/${encodeURIComponent(slug)}/register`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, data: entries })
      });

      const payload = await res.json().catch(() => ({}));
      if (!res.ok) {
        setMessage("error", messageForError(payload));
        form.style.display = "none";
        return;
      }

      const emailSent = payload && payload.email_sent === true;
      setMessage("success", emailSent ? messages.successWithEmail : messages.successNoEmail);
      form.reset();
      form.style.display = "none";
    } catch {
      setMessage("error", messages.errorGeneric);
    } finally {
      setLoading(false);
    }
  });
})();
</script>
