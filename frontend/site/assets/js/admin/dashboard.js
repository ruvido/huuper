(() => {
  const listNode = document.querySelector("#admin-summary");
  const statusNode = document.querySelector("#admin-status");
  if (!listNode || !statusNode || !window.huuperAuth || !window.huuperListPage) {
    return;
  }

  async function load() {
    try {
      const payload = await window.huuperAuth.apiFetch("/api/admin/summary");
      const rows = [
        ["Users", payload.users && payload.users.total, "/admin/users/"],
        ["Groups", payload.groups && payload.groups.total, "/admin/groups/"],
        ["Events", payload.events && payload.events.total, "/admin/events/"],
      ].filter((row) => row[1] !== undefined && row[1] !== null);

      if (rows.length === 0) {
        window.huuperListPage.setStatus(statusNode, "No data.");
        return;
      }

      window.huuperListPage.renderList(listNode, rows, ([label, value, href]) => {
        const safeHref = window.huuperListPage.escapeHTML(href);
        const safeLabel = window.huuperListPage.escapeHTML(label);
        return `<a href="${safeHref}"><span class="data-label">${safeLabel}</span><strong class="data-value">${value}</strong></a>`;
      });
      listNode.hidden = false;
    } catch (_) {
      window.huuperListPage.setStatus(statusNode, "Summary unavailable.");
    }
  }

  load();
})();
