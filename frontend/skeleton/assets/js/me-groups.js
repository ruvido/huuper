(() => {
  const listNode = document.querySelector("#groups-list");
  const statusNode = document.querySelector("#groups-status");
  if (!listNode || !statusNode || !window.huuperAuth || !window.huuperListPage) {
    return;
  }

  async function load() {
    try {
      const payload = await window.huuperAuth.apiFetch("/api/me/groups");
      const items = Array.isArray(payload.items) ? payload.items : [];
      if (items.length === 0) {
        window.huuperListPage.setStatus(statusNode, "No groups.");
        return;
      }

      window.huuperListPage.renderList(listNode, items, (item) => {
        const name = window.huuperListPage.escapeHTML(item.name || item.id);
        const type = window.huuperListPage.text(item.type);
        const membersCount = Number.isFinite(item.members_count) ? item.members_count : null;
        const requestsCount = Number.isFinite(item.requests_count) ? item.requests_count : null;
        const meta = [];
        if (type) meta.push(type);
        if (membersCount !== null) meta.push(`${membersCount} members`);
        if (requestsCount !== null && requestsCount > 0) meta.push(`${requestsCount} requests`);
        const href = `/me/group/?id=${encodeURIComponent(item.id)}`;
        return `<a href="${href}"><strong>${name}</strong>${meta.length ? `<p class="status">${meta.join(" • ")}</p>` : ""}</a>`;
      });
      listNode.hidden = false;
    } catch (_) {
      window.huuperListPage.setStatus(statusNode, "Groups unavailable.");
    }
  }

  load();
})();
