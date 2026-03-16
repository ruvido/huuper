(() => {
  if (!window.huuperEntityList || !window.huuperAuth || !window.huuperListPage || !window.huuperRequestCard) {
    return;
  }

  window.huuperEntityList.init({
    statusSelector: "#requests-status",
    listSelector: "#requests-list",
    requiresAuth: true,
    requiresRequestCard: true,
    emptyMessage: "No requests.",
    errorMessage: "Requests unavailable.",
    load: () => window.huuperAuth.apiFetch("/api/me/requests"),
    renderItem: (item) => {
      const href = `/me/request/?id=${encodeURIComponent(item.id)}`;
      return window.huuperRequestCard.renderCompact(item, href);
    },
  });
})();
