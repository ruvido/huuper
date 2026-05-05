// Side-effect actions for the events wizard: cancel-with-scope dialog and
// reschedule call. All functions take an explicit `deps` object so nothing
// is captured implicitly. The orchestrator wires these to buttons.
(() => {
  window.appEventsWizard = window.appEventsWizard || {};
  const ns = window.appEventsWizard;
  if (!ns.config) return;

  function eventsCopy() {
    return (window.appCopy && window.appCopy.events) || {};
  }

  function wizardCopy() {
    return eventsCopy().wizard || {};
  }

  function cancelCopy() {
    return eventsCopy().cancel || {};
  }

  function detailCopy() {
    return eventsCopy().detail || {};
  }

  // deps: { auth, scope, eventID, hasSeries, basePath, onDone }
  function openCancelDialog(deps) {
    const sheet = window.appActionSheet;
    if (!sheet || typeof sheet.open !== "function") return;
    const wc = wizardCopy();
    const cc = cancelCopy();
    const dc = detailCopy();
    const dialog = (dc.cancelDialog || {});
    const labels = cc.scopeLabels || {};

    if (!deps.hasSeries) {
      sheet.open({
        title: dialog.title || "Cancel this event?",
        actions: [],
        footerAction: {
          label: dialog.confirmLabel || "Cancel event",
          tone: "danger",
          onSelect: async () => {
            await postCancel(deps.auth, deps.scope, deps.eventID, "this");
            if (typeof deps.onDone === "function") deps.onDone();
          },
        },
      });
      return;
    }

    sheet.open({
      title: dialog.title || "Cancel event",
      meta: (wc.cancelMeta || ""),
      actions: [
        {
          label: labels.this || "Only this occurrence",
          onSelect: async () => {
            await postCancel(deps.auth, deps.scope, deps.eventID, "this");
            if (typeof deps.onDone === "function") deps.onDone();
          },
        },
        {
          label: labels.future || "This and future events",
          onSelect: async () => {
            await postCancel(deps.auth, deps.scope, deps.eventID, "future");
            if (typeof deps.onDone === "function") deps.onDone();
          },
        },
        {
          label: labels.all || "All events",
          tone: "danger",
          onSelect: async () => {
            await postCancel(deps.auth, deps.scope, deps.eventID, "all");
            if (typeof deps.onDone === "function") deps.onDone();
          },
        },
      ],
    });
  }

  async function postCancel(auth, scope, eventID, scopeValue) {
    return auth.apiFetch(`/api/${scope}/events/${encodeURIComponent(eventID)}/cancel`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ scope: scopeValue }),
    });
  }

  // deps: { auth, scope, eventID, newDateISO }
  async function postReschedule(deps) {
    return deps.auth.apiFetch(`/api/${deps.scope}/events/${encodeURIComponent(deps.eventID)}/reschedule`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ start_date: deps.newDateISO }),
    });
  }

  function openRescheduleDialog(deps) {
    const sheet = window.appActionSheet;
    if (!sheet || typeof sheet.open !== "function") return;
    const wc = wizardCopy();
    const labels = (wc.labels || {});
    const inputId = "events-wizard-reschedule-input";
    const initialDate = (deps.currentISO || "").slice(0, 10) || "";
    sheet.open({
      title: (wc.rescheduleTitle || "Reschedule event"),
      contentHTML: `
        <div class="events-wizard-field">
          <label class="field-label" for="${inputId}">${escapeHTML(labels.startDateField || "Date")}</label>
          <input id="${inputId}" class="field-input" type="date" value="${escapeHTML(initialDate)}" />
        </div>
      `,
      actions: [],
      footerAction: {
        label: (wc.rescheduleConfirm || "Save"),
        tone: "primary",
        onSelect: async () => {
          const dateEl = document.getElementById(inputId);
          const dateVal = dateEl ? dateEl.value : "";
          if (!dateVal) return;
          await postReschedule({ auth: deps.auth, scope: deps.scope, eventID: deps.eventID, newDateISO: dateVal });
          if (typeof deps.onDone === "function") deps.onDone();
        },
      },
    });
  }

  function escapeHTML(value) {
    return String(value || "")
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;")
      .replaceAll("'", "&#39;");
  }

  ns.actions = {
    openCancelDialog,
    openRescheduleDialog,
    postCancel,
    postReschedule,
  };
})();
