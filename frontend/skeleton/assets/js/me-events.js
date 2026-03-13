(() => {
  const listNode = document.querySelector("#events-list");
  const statusNode = document.querySelector("#events-status");
  if (!listNode || !statusNode || !window.huuperListPage) {
    return;
  }

  async function load() {
    try {
      const response = await fetch("/api/collections/events/records?sort=event_date&perPage=200");
      if (!response.ok) {
        throw new Error("events_failed");
      }
      const payload = await response.json();
      const items = Array.isArray(payload.items) ? payload.items : [];
      if (items.length === 0) {
        window.huuperListPage.setStatus(statusNode, "No events.");
        return;
      }

      window.huuperListPage.renderList(listNode, items, (item) => {
        const title = window.huuperListPage.escapeHTML(item.title || item.id);
        const date = window.huuperListPage.dateTime(item.event_date);
        const href = `/me/event/?id=${encodeURIComponent(item.id)}`;
        return `<a href="${href}"><strong>${title}</strong>${date ? `<p class="status">${date}</p>` : ""}</a>`;
      });
      listNode.hidden = false;
    } catch (_) {
      window.huuperListPage.setStatus(statusNode, "Events unavailable.");
    }
  }

  load();
})();
