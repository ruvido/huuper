(() => {
  const publicFields = window.appPublicFields;
  const auth = window.appAuth;

  const loadingNode = document.querySelector("#event-loading");
  const bodyNode = document.querySelector("#event-body");
  const errorNode = document.querySelector("#event-error");
  const titleNode = document.querySelector("#event-title");
  const metaNode = document.querySelector("#event-meta");
  const depositNode = document.querySelector("#event-deposit");
  const emailFieldWrap = document.querySelector("#event-email-field");
  const emailField = document.querySelector("#event-email");
  const emailErrorNode = document.querySelector("#event-email-error");
  const form = document.querySelector("#event-register-form");
  const statusNode = document.querySelector("#event-status");
  const submitButton = document.querySelector("#event-submit");

  if (!loadingNode || !bodyNode || !errorNode || !form || !submitButton || !auth) {
    return;
  }

  function slugFromURL() {
    const segments = window.location.pathname.split("/").filter(Boolean);
    return segments.length >= 2 ? segments[1] : "";
  }

  function formatDate(raw) {
    if (!raw) return "";
    const date = new Date(raw);
    if (Number.isNaN(date.getTime())) return "";
    return date.toLocaleDateString("it-IT", { day: "numeric", month: "long", year: "numeric" });
  }

  function showError(message) {
    loadingNode.hidden = true;
    errorNode.textContent = message;
    errorNode.hidden = false;
  }

  const slug = slugFromURL();
  if (!slug) {
    showError("Evento non trovato.");
    return;
  }

  const session = auth.read();
  const isMember = !!(session && session.token);
  let depositCents = 0;

  fetch("/api/public/events/" + encodeURIComponent(slug))
    .then((response) => {
      if (!response.ok) {
        throw new Error("not_found");
      }
      return response.json();
    })
    .then((payload) => {
      const item = payload.event || {};
      depositCents = Number(item.deposit_cents || 0);

      titleNode.textContent = item.title || "Raduno";
      const metaParts = [formatDate(item.start_date), item.location].filter(Boolean);
      metaNode.textContent = metaParts.join(" · ");

      if (depositCents > 0) {
        depositNode.textContent = "Caparra richiesta: " + (depositCents / 100).toFixed(2) + " €";
        depositNode.hidden = false;
      }

      if (isMember) {
        emailFieldWrap.hidden = true;
        submitButton.textContent = depositCents > 0 ? "Iscriviti e paga la caparra" : "Iscriviti";
      } else {
        submitButton.textContent = "Richiedi di partecipare";
      }

      loadingNode.hidden = true;
      bodyNode.hidden = false;
    })
    .catch(() => {
      showError("Evento non trovato o iscrizioni chiuse.");
    });

  if (emailField) {
    emailField.addEventListener("input", () => {
      emailErrorNode.hidden = true;
      emailErrorNode.textContent = "";
    });
  }

  form.addEventListener("submit", async (event) => {
    event.preventDefault();
    statusNode.hidden = true;
    statusNode.textContent = "";

    let email = "";
    if (isMember) {
      email = session.model && session.model.email ? session.model.email : "";
    } else if (publicFields && emailField) {
      const fieldAdvance = publicFields.validateFieldOnAdvance({ key: "email", type: "email" }, emailField.value);
      if (fieldAdvance.error) {
        emailErrorNode.textContent = fieldAdvance.error;
        emailErrorNode.hidden = false;
        return;
      }
      email = fieldAdvance.normalized || emailField.value.trim().toLowerCase();
    }

    submitButton.disabled = true;
    try {
      const payload = await auth.apiFetch("/api/public/events/" + encodeURIComponent(slug) + "/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, data: { source: "event_page" } }),
      });

      if (payload.checkout_url) {
        window.location.href = payload.checkout_url;
        return;
      }

      form.hidden = true;
      statusNode.textContent = isMember
        ? "Iscrizione ricevuta, controlla la tua email."
        : "Richiesta ricevuta, ti contatteremo a breve.";
      statusNode.hidden = false;
    } catch (error) {
      const code = (error && error.payload && error.payload.message) || "";
      const messages = {
        already_submitted: "Hai già inviato una registrazione per questo evento.",
        event_closed: "Le iscrizioni per questo evento sono chiuse.",
        invalid_email: "Email non valida.",
      };
      statusNode.textContent = messages[code] || "Registrazione non riuscita, riprova.";
      statusNode.hidden = false;
    } finally {
      submitButton.disabled = false;
    }
  });
})();
