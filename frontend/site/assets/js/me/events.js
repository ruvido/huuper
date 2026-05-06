// /me/events list orchestrator. Fetches /api/me/events?window=...,
// renders list items via appMeEventsList.render, and wires the
// future/past window switcher tabs.
//
// TODO: deeper edit flows are handled by the shared events wizard.
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
  const SPOTLIGHT_ID = "events-spotlight";

  // Mirrors battleplan arrangeSections: lifts the spotlight row out of the
  // main list into the flush spotlight container so it sits edge-to-edge.
  function arrangeSpotlight(listNode) {
    const container = document.getElementById(SPOTLIGHT_ID);
    if (!container || !listNode) return;
    container.innerHTML = "";
    const el = listNode.querySelector(".list-item-spotlight");
    if (el) {
      container.appendChild(el);
      container.hidden = false;
    } else {
      container.hidden = true;
    }
  }

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
    return document.getElementById("events-window-tabs");
  }

  let canCreateEvent = false;

  function syncCtaVisibility() {
    const cta = document.getElementById("events-new-cta");
    const section = document.getElementById("events-future-section");
    const visible = canCreateEvent && currentWindow !== "past";
    if (cta) cta.hidden = !visible;
    if (section) section.hidden = !visible;
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
        syncCtaVisibility();
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
          if (currentWindow !== "past" && items.length > 0) items[0].__spotlight = true;
          return { items };
        },
        renderItem: (item) => rndmod.renderItem(item, {
          esc: window.appListPage.escapeHTML,
          basePath: BASE_PATH,
          spotlight: item.__spotlight === true,
        }),
        afterRender: (node) => arrangeSpotlight(node),
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
      if (currentWindow !== "past" && items.length > 0) items[0].__spotlight = true;
      window.appListPage.renderList(listNode, items, (item) => rndmod.renderItem(item, {
        esc: window.appListPage.escapeHTML,
        basePath: BASE_PATH,
        spotlight: item.__spotlight === true,
      }));
      arrangeSpotlight(listNode);
      listNode.hidden = false;
      statusNode.hidden = true;
    } catch (_) {
      window.appListPage.setStatus(statusNode, lc.errorMessage || "Events unavailable.");
    }
  }

  async function gateNewEventCTA() {
    const cta = document.getElementById("events-new-cta");
    if (!cta) return;
    const lc = copy();
    if (lc.newLabel) cta.textContent = lc.newLabel;
    try {
      const payload = await window.appAuth.apiFetch("/api/me/access/events");
      canCreateEvent = !!(payload && payload.access === true);
    } catch (_) {
      canCreateEvent = false;
    }
    syncCtaVisibility();
  }

  gateNewEventCTA();
  renderTabs();
  syncCtaVisibility();
  loadList();
})();
