// /me/events list orchestrator. Fetches /api/me/events?window=...,
// renders list items via appMeEventsList.render, and wires the
// future/past window switcher tabs.
//
// TODO: wizard for create/edit/cancel-with-scope is OUT OF SCOPE for this
// round — see the events backend (api/me/events.go) for the existing
// endpoints we'll wire next round.
(() => {
  if (!window.appEntityList || !window.appAuth || !window.appListPage) {
    return;
  }
  window.appMeEventsList = window.appMeEventsList || {};
  const ns = window.appMeEventsList;
  if (!ns.render) return;
  const rndmod = ns.render;

  const SCOPE = "me";
  const BASE_PATH = "/me";

  function copy() {
    return (window.appCopy && window.appCopy.events && window.appCopy.events.list) || {};
  }

  let currentWindow = readWindowFromURL() || "future";

  function readWindowFromURL() {
    try {
      const value = new URLSearchParams(window.location.search).get("window");
      return value === "past" ? "past" : value === "future" ? "future" : "";
    } catch (_) {
      return "";
    }
  }

  function syncURLWindow(value) {
    try {
      const url = new URL(window.location.href);
      url.searchParams.set("window", value);
      window.history.replaceState(null, "", url.pathname + url.search + url.hash);
    } catch (_) {
      /* non-fatal */
    }
  }

  function ensureTabsHost() {
    let host = document.getElementById("events-window-tabs");
    if (host) return host;
    const list = document.getElementById("events-list");
    const status = document.getElementById("events-status");
    const anchor = status || list;
    if (!anchor || !anchor.parentNode) return null;
    host = document.createElement("section");
    host.id = "events-window-tabs";
    host.className = "section-tabs events-window-tabs";
    anchor.parentNode.insertBefore(host, anchor);
    return host;
  }

  function renderTabs() {
    const host = ensureTabsHost();
    if (!host) return;
    host.innerHTML = rndmod.renderWindowTabs(currentWindow);
    host.querySelectorAll("button[data-window]").forEach((btn) => {
      btn.addEventListener("click", () => {
        const next = btn.getAttribute("data-window") || "future";
        if (next === currentWindow) return;
        currentWindow = next;
        syncURLWindow(currentWindow);
        renderTabs();
        loadList();
      });
    });
  }

  let listInited = false;

  function loadList() {
    const lc = copy();
    if (!listInited) {
      listInited = true;
      window.appEntityList.init({
        statusSelector: "#events-status",
        listSelector: "#events-list",
        requiresAuth: true,
        errorMessage: lc.errorMessage || "Events unavailable.",
        renderEmpty: () => rndmod.renderEmpty(),
        load: async () => {
          const url = `/api/${SCOPE}/events?window=${encodeURIComponent(currentWindow)}`;
          const payload = await window.appAuth.apiFetch(url);
          const items = Array.isArray(payload && payload.items) ? payload.items : [];
          return { items };
        },
        renderItem: (item) => rndmod.renderItem(item, {
          esc: window.appListPage.escapeHTML,
          basePath: BASE_PATH,
        }),
      });
      return;
    }
    // Re-trigger entity-list by simulating a pageshow back/forward refresh:
    // simplest path is to dispatch a synthetic load by re-fetching directly
    // and rendering inline.
    refreshList();
  }

  async function refreshList() {
    const listNode = document.getElementById("events-list");
    const statusNode = document.getElementById("events-status");
    if (!listNode || !statusNode) return;
    const lc = copy();
    try {
      const url = `/api/${SCOPE}/events?window=${encodeURIComponent(currentWindow)}`;
      const payload = await window.appAuth.apiFetch(url);
      const items = Array.isArray(payload && payload.items) ? payload.items : [];
      if (items.length === 0) {
        listNode.innerHTML = rndmod.renderEmpty();
        listNode.hidden = false;
        statusNode.hidden = true;
        return;
      }
      window.appListPage.renderList(listNode, items, (item) => rndmod.renderItem(item, {
        esc: window.appListPage.escapeHTML,
        basePath: BASE_PATH,
      }));
      listNode.hidden = false;
      statusNode.hidden = true;
    } catch (_) {
      window.appListPage.setStatus(statusNode, lc.errorMessage || "Events unavailable.");
    }
  }

  renderTabs();
  loadList();
})();
