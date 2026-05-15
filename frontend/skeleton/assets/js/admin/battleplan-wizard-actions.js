// View-page action wiring (the buttons that appear under a viewed/draft
// battleplan). All functions take a `deps` object so nothing is captured
// implicitly — see configureViewActions(deps) for the full shape.
(() => {
  window.appBattleplanWizard = window.appBattleplanWizard || {};
  const ns = window.appBattleplanWizard;
  if (!ns.render) return; // render module must load first (we use setNextIcon)

  const setNextIcon = ns.render.setNextIcon;

  // deps: { auth, state, dom, basePath, PREFIX, buildPayload }
  async function archiveCurrentView(deps) {
    const { auth, state, dom, basePath } = deps;
    if (!dom.back || dom.back.disabled) return;
    dom.back.disabled = true;
    try {
      await auth.apiFetch(`/api/me/battleplans/${encodeURIComponent(state.viewId)}/status`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ status: "archived" }),
      });
      window.location.href = `${basePath()}/`;
    } catch { dom.back.disabled = false; }
  }

  function openDeleteDialog(deps) {
    const { auth, state, basePath } = deps;
    const sheet = window.appActionSheet;
    const dlg = ((window.appCopy || {}).battleplan || {}).view || {};
    const dialog = dlg.deleteDialog || {};
    if (!sheet || typeof sheet.open !== "function") return;
    sheet.open({
      title: dialog.title || "",
      meta: dialog.meta,
      actions: [],
      footerAction: {
        label: dialog.confirmLabel || "",
        tone: "danger",
        onSelect: async () => {
          await auth.apiFetch(`/api/me/battleplans/${encodeURIComponent(state.viewId)}`, { method: "DELETE" });
          window.location.href = `${basePath()}/`;
        },
      },
    });
  }

  async function duplicateCurrentView(deps) {
    const { auth, state, basePath, buildPayload } = deps;
    let items = [];
    try {
      const list = await auth.apiFetch("/api/me/battleplans?per_page=200");
      items = (list && Array.isArray(list.items)) ? list.items : [];
    } catch (_) {
      /* if list fetch fails, fall through with empty items so create proceeds */
    }
    await window.appBattleplan.createDraft({
      payload: buildPayload(),
      items,
      excludeId: state.viewId,
      basePath: basePath(),
    });
  }

  async function activateCurrentView(deps) {
    const { auth, state, basePath, PREFIX } = deps;
    const nextBtn = document.getElementById(`${PREFIX}-next`);
    if (!nextBtn || nextBtn.disabled) return;
    nextBtn.disabled = true;
    try {
      await auth.apiFetch(`/api/me/battleplans/${encodeURIComponent(state.viewId)}/activate`, {
        method: "POST",
      });
      window.location.href = `${basePath()}/`;
    } catch { nextBtn.disabled = false; }
  }

  function ensureAddNoteButton(deps) {
    const { state, dom, auth, PREFIX } = deps;
    if (!dom.sticky || state.loadedStatus !== "active") return null;
    let button = document.getElementById(`${PREFIX}-add-note`);
    if (!button) {
      button = document.createElement("button");
      button.id = `${PREFIX}-add-note`;
      button.className = "battleplan-view-note-button";
      button.type = "button";
      dom.sticky.insertBefore(button, dom.sticky.firstChild);
    }
    const viewCopy = (((window.appCopy || {}).battleplan || {}).view) || {};
    button.textContent = (viewCopy.addNoteLabel || "").toUpperCase();
    button.onclick = () => {
      if (!window.appRequestNoteSheet) return;
      window.appRequestNoteSheet.open({
        title: viewCopy.addNoteTitle || "",
        submitLabel: viewCopy.addNoteSubmitLabel || "",
        emptyStatus: viewCopy.addNoteEmptyStatus || "",
        statusNode: dom.status,
        onSubmit: async (note, sheetButton) => {
          sheetButton.disabled = true;
          await auth.apiFetch(`/api/me/battleplans/${encodeURIComponent(state.viewId)}/notes`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ note }),
          });
          window.location.reload();
        },
      });
    };
    return button;
  }

  function configureViewActions(deps) {
    const { state, dom, basePath, PREFIX } = deps;
    const bpCopy = (window.appCopy && window.appCopy.battleplan) || {};
    const viewCopy = bpCopy.view || {};
    const listCopy = bpCopy.list || {};
    const status = state.loadedStatus || "";
    ensureAddNoteButton(deps);

    if (dom.back) {
      if (status === "archived") {
        dom.back.textContent = (viewCopy.deleteLabel || "").toUpperCase();
        dom.back.onclick = (event) => { event.preventDefault(); openDeleteDialog(deps); };
      } else {
        dom.back.textContent = (bpCopy.archiveLabel || "").toUpperCase();
        dom.back.onclick = (event) => { event.preventDefault(); archiveCurrentView(deps); };
      }
    }

    const nextBtn = document.getElementById(`${PREFIX}-next`);
    const nextLabel = document.getElementById(`${PREFIX}-next-label`);
    if (!nextBtn || !nextLabel) return;
    nextBtn.setAttribute("type", "button");
    nextBtn.removeAttribute("form");
    if (status === "draft") {
      nextLabel.textContent = (viewCopy.activateLabel || "").toUpperCase();
      setNextIcon("edit", PREFIX);
      nextBtn.onclick = () => activateCurrentView(deps);
    } else if (status === "active" || status === "draft") {
      nextLabel.textContent = (viewCopy.editLabel || "").toUpperCase();
      setNextIcon("edit", PREFIX);
      nextBtn.onclick = () => {
        window.location.href = `${basePath()}/edit/${encodeURIComponent(state.viewId)}/`;
      };
    } else {
      nextLabel.textContent = (listCopy.duplicateLabel || "").toUpperCase();
      setNextIcon("default", PREFIX);
      nextBtn.onclick = () => duplicateCurrentView(deps);
    }
  }

  ns.actions = {
    archiveCurrentView,
    openDeleteDialog,
    duplicateCurrentView,
    activateCurrentView,
    configureViewActions,
  };
})();
