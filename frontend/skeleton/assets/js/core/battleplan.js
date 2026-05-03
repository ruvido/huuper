window.appBattleplan = (() => {
  function basePath() {
    return window.location.pathname.startsWith("/me/") ? "/me/battleplan" : "/admin/battleplan";
  }

  function findDraft(items) {
    return (items || []).find((it) => it && it.status === "draft") || null;
  }

  function findActive(items) {
    return (items || []).find((it) => it && it.status === "active") || null;
  }

  function confirmOverwriteDraft(draft, onConfirm) {
    const sheet = window.appActionSheet;
    const cfg = ((window.appCopy || {}).battleplan || {}).list || {};
    const dlg = cfg.overwriteDraftDialog || {};
    if (!sheet || typeof sheet.open !== "function") return;
    sheet.open({
      title: dlg.title || "",
      actions: [],
      footerAction: {
        label: dlg.confirmLabel || "",
        tone: "danger",
        onSelect: async () => {
          await window.appAuth.apiFetch(`/api/me/battleplans/${encodeURIComponent(draft.id)}`, { method: "DELETE" });
          await onConfirm();
        },
      },
    });
  }

  // Phase E unified createDraft helper. Centralizes the
  // "POST /api/me/battleplans → maybe overwrite existing draft → redirect to
  // edit/<id>/" flow so list-actions and wizard-actions both call this
  // instead of duplicating the same 25-line dance.
  //
  // opts: {
  //   payload:     full body to POST (status: "draft" is forced).
  //   items:       battleplan list to scan for an existing draft.
  //   excludeId:   optional id to skip when matching existingDraft (when
  //                duplicating from a draft itself, you don't want to treat
  //                the source as a collision).
  //   basePath:    base url string (caller passes the resolved value).
  // }
  async function createDraft(opts) {
    const { payload, items, excludeId, basePath: bp } = opts || {};
    const targetBase = typeof bp === "string" ? bp : basePath();
    const post = async () => {
      const body = Object.assign({}, payload || {}, { status: "draft" });
      const created = await window.appAuth.apiFetch("/api/me/battleplans", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });
      if (created && created.id) {
        window.location.href = `${targetBase}/edit/${encodeURIComponent(created.id)}/`;
      }
    };
    const draft = findDraft(items || []);
    const collidingDraft = draft && (!excludeId || draft.id !== excludeId) ? draft : null;
    if (collidingDraft) {
      confirmOverwriteDraft(collidingDraft, post);
      return;
    }
    await post();
  }

  return { basePath, findDraft, findActive, confirmOverwriteDraft, createDraft };
})();
