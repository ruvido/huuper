(() => {
  if (!window.huuperEntityList || !window.huuperAuth || !window.huuperListPage) {
    return;
  }

  window.huuperEntityList.init({
    statusSelector: "#admin-groups-status",
    listSelector: "#admin-groups-list",
    requiresAuth: true,
    emptyMessage: "No groups.",
    errorMessage: "Groups unavailable.",
    load: () => window.huuperAuth.apiFetch("/api/admin/groups"),
    renderItem: (item) => {
      const meta = window.huuperListPage.text(item.type);
      const href = `/admin/group/?id=${encodeURIComponent(item.id)}`;
      return window.huuperListPage.renderCompactLink(href, item.name || item.id, meta);
    },
  });
})();
