// Battleplan list page orchestrator. Owns the access check + appEntityList
// wiring; delegates render fns to ns.render and DOM-mutating wiring to
// ns.actions. Each split file publishes onto window.appBattleplanList; we
// read from it here.
(() => {
  if (!window.appEntityList || !window.appAuth || !window.appListPage) {
    return;
  }
  window.appBattleplanList = window.appBattleplanList || {};
  const ns = window.appBattleplanList;
  if (!ns.render || !ns.actions) return; // split modules required

  const rndmod = ns.render;
  const actmod = ns.actions;

  // Initial wiring that must happen before items are loaded.
  actmod.initFabAriaLabel();
  actmod.initHeroTitle();

  // Reads the active status labels straight from copy each call so any copy
  // hot-reload (theoretically) takes effect on the next render. Returns an
  // object keyed by status enum → human label.
  function statusLabels() {
    const c = (window.appCopy && window.appCopy.battleplan && window.appCopy.battleplan.list) || {};
    return c.statusLabels || {};
  }

  function listCopy() {
    return (window.appCopy && window.appCopy.battleplan && window.appCopy.battleplan.list) || {};
  }

  function isAdminUser() {
    try {
      const raw = localStorage.getItem("app.auth");
      const auth = raw ? JSON.parse(raw) : null;
      return !!(auth && auth.model && auth.model.admin === true);
    } catch (_) {
      return false;
    }
  }
  // Currently unused in this orchestrator but kept as a public hook for
  // admin-only future toggles (kept simple, no need to expose globally).
  void isAdminUser;

  function orderedItems(items) {
    const statusRank = { active: 0, draft: 1, completed: 2, archived: 3 };
    return [...items].sort((a, b) => {
      const ar = Object.prototype.hasOwnProperty.call(statusRank, a.status) ? statusRank[a.status] : 3;
      const br = Object.prototype.hasOwnProperty.call(statusRank, b.status) ? statusRank[b.status] : 3;
      if (ar !== br) return ar - br;
      return rndmod.itemTime(b) - rndmod.itemTime(a);
    });
  }

  (async () => {
    let access = false;
    try {
      const payload = await window.appAuth.apiFetch("/api/me/access/battleplan");
      access = !!(payload && payload.access === true);
    } catch (_) {
      access = false;
    }
    if (!access) {
      actmod.showPlaceholder();
      return;
    }
    // Show the standalone "New Plan" CTA right away. arrangeSections will hide
    // it once items load if an active plan exists.
    const emptyNew = document.getElementById("battleplan-empty-new");
    if (emptyNew) {
      const copy = listCopy();
      emptyNew.textContent = copy.newPlanLabel || "";
      emptyNew.hidden = false;
    }
    initList();
  })();

  function initList() {
    const bpCopy = (window.appCopy && window.appCopy.battleplan) || {};
    const lc = listCopy();
    window.appEntityList.init({
      statusSelector: "#battleplan-status",
      listSelector: "#battleplan-list",
      requiresAuth: true,
      errorMessage: lc.errorMessage || "",
      renderEmpty: () => rndmod.renderEmpty(lc),
      load: async () => {
        const payload = await window.appAuth.apiFetch("/api/me/battleplans?per_page=200");
        const items = Array.isArray(payload.items) ? payload.items : [];
        return { ...payload, items: orderedItems(items) };
      },
      renderItem: (item) => rndmod.renderItem(item, {
        copy: lc,
        statusLabels: statusLabels(),
        esc: window.appListPage.escapeHTML,
        basePath: window.appBattleplan.basePath(),
        titleFallback: bpCopy.titleBattleplan || "",
      }),
      afterRender: (node, items) => {
        actmod.wireNewButton(items);
        actmod.arrangeSections(node, items);
      },
    });
  }
})();
