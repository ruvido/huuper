// Side-effect actions for the events wizard: event cancel dialog and
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

  function detailCopy() {
    return eventsCopy().detail || {};
  }

  // deps: { auth, scope, eventID, basePath, onDone }
  function openCancelDialog(deps) {
    const sheet = window.appActionSheet;
    if (!sheet || typeof sheet.open !== "function") return;
    const dc = detailCopy();
    const dialog = (dc.cancelDialog || {});

    sheet.open({
      title: dialog.title || "Cancel this event?",
      actions: [],
      footerAction: {
        label: dialog.confirmLabel || "Cancel event",
        tone: "danger",
        onSelect: async () => {
          await postCancel(deps.auth, deps.scope, deps.eventID);
          if (typeof deps.onDone === "function") deps.onDone();
        },
      },
    });
  }

  async function postCancel(auth, scope, eventID) {
    return auth.apiFetch(`/api/${scope}/events/${encodeURIComponent(eventID)}/cancel`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({}),
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
        <div>
          <label class="field-label" for="${inputId}">${escapeHTML(labels.startDateField || "Date")}</label>
          <input id="${inputId}" class="field-input" type="text" value="${escapeHTML(initialDate)}" placeholder="YYYY-MM-DD" inputmode="numeric" autocomplete="off" autocapitalize="off" spellcheck="false" maxlength="10" />
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
          await postReschedule({ auth: deps.auth, scope: deps.scope, eventID: deps.eventID, newDateISO: `${dateVal}T00:00:00Z` });
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
