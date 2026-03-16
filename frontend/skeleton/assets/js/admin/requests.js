(() => {
  if (!window.huuperEntityList || !window.huuperAuth || !window.huuperListPage || !window.huuperRequestCard) {
    return;
  }

  window.huuperEntityList.init({
    statusSelector: "#admin-requests-status",
    listSelector: "#admin-requests-list",
    requiresAuth: true,
    requiresRequestCard: true,
    emptyMessage: "No requests.",
    errorMessage: "Requests unavailable.",
    load: () => window.huuperAuth.apiFetch("/api/admin/requests"),
    renderItem: (item) => {
      const href = `/admin/request/?id=${encodeURIComponent(item.id)}`;
      return window.huuperRequestCard.renderCompact(item, href);
    },
  });
})();
