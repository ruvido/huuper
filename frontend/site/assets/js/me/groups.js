(() => {
  if (!window.huuperEntityList || !window.huuperAuth || !window.huuperListPage) {
    return;
  }

  window.huuperEntityList.init({
    statusSelector: "#groups-status",
    listSelector: "#groups-list",
    requiresAuth: true,
    emptyMessage: "No groups.",
    errorMessage: "Groups unavailable.",
    load: () => window.huuperAuth.apiFetch("/api/me/groups"),
    renderItem: (item) => {
      const type = window.huuperListPage.text(item.type);
      const membersCount = Number.isFinite(item.members_count) ? item.members_count : null;
      const requestsCount = Number.isFinite(item.requests_count) ? item.requests_count : null;
      const meta = [];
      if (type) meta.push(type);
      if (membersCount !== null) meta.push(`${membersCount} members`);
      if (requestsCount !== null && requestsCount > 0) meta.push(`${requestsCount} pending`);
      const href = `/me/group/?id=${encodeURIComponent(item.id)}`;
      return window.huuperListPage.renderListItemLink(href, item.name || item.id, meta.join(" • "));
    },
  });
})();
