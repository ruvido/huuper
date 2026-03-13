(() => {
  const listNode = document.querySelector("#requests-list");
  const statusNode = document.querySelector("#requests-status");
  if (!listNode || !statusNode || !window.huuperAuth || !window.huuperListPage || !window.huuperRequestCard) {
    return;
  }

  async function load() {
    try {
      const payload = await window.huuperAuth.apiFetch("/api/me/requests");
      const items = Array.isArray(payload.items) ? payload.items : [];
      if (items.length === 0) {
        window.huuperListPage.setStatus(statusNode, "No requests.");
        return;
      }

      window.huuperListPage.renderList(listNode, items, (item) => {
        const href = `/me/request/?id=${encodeURIComponent(item.id)}`;
        return window.huuperRequestCard.renderCompact(item, href);
      });
      listNode.hidden = false;
    } catch (_) {
      window.huuperListPage.setStatus(statusNode, "Requests unavailable.");
    }
  }

  load();
})();
