(() => {
  const listNode = document.querySelector("#admin-events-list");
  const statusNode = document.querySelector("#admin-events-status");
  if (!listNode || !statusNode || !window.huuperAuth || !window.huuperListPage) {
    return;
  }

  async function load() {
    try {
      const payload = await window.huuperAuth.apiFetch("/api/admin/events");
      const items = Array.isArray(payload.items) ? payload.items : [];
      if (items.length === 0) {
        window.huuperListPage.setStatus(statusNode, "No events.");
        return;
      }

      window.huuperListPage.renderList(listNode, items, (item) => {
        const title = window.huuperListPage.escapeHTML(item.title || item.id);
        const date = window.huuperListPage.dateTime(item.event_date);
        const href = `/admin/event/?id=${encodeURIComponent(item.id)}`;
        return `<a href="${href}"><strong>${title}</strong>${date ? `<p class="status">${date}</p>` : ""}</a>`;
      });
      listNode.hidden = false;
    } catch (_) {
      window.huuperListPage.setStatus(statusNode, "Events unavailable.");
    }
  }

  load();
})();
