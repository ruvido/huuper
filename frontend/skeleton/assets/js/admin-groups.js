(() => {
  const listNode = document.querySelector("#admin-groups-list");
  const statusNode = document.querySelector("#admin-groups-status");
  if (!listNode || !statusNode || !window.huuperAuth || !window.huuperListPage) {
    return;
  }

  async function load() {
    try {
      const payload = await window.huuperAuth.apiFetch("/api/admin/groups");
      const items = Array.isArray(payload.items) ? payload.items : [];
      if (items.length === 0) {
        window.huuperListPage.setStatus(statusNode, "No groups.");
        return;
      }

      window.huuperListPage.renderList(listNode, items, (item) => {
        const name = window.huuperListPage.escapeHTML(item.name || item.id);
        const type = window.huuperListPage.text(item.type);
        const href = `/admin/group/?id=${encodeURIComponent(item.id)}`;
        return `<a href="${href}"><strong>${name}</strong>${type ? `<p class="status">${type}</p>` : ""}</a>`;
      });
      listNode.hidden = false;
    } catch (_) {
      window.huuperListPage.setStatus(statusNode, "Groups unavailable.");
    }
  }

  load();
})();
