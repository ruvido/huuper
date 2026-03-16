(() => {
  if (!window.huuperEntityList || !window.huuperAuth || !window.huuperListPage) {
    return;
  }

  window.huuperEntityList.init({
    statusSelector: "#admin-events-status",
    listSelector: "#admin-events-list",
    requiresAuth: true,
    emptyMessage: "No events.",
    errorMessage: "Events unavailable.",
    load: () => window.huuperAuth.apiFetch("/api/admin/events"),
    renderItem: (item) => {
      const href = `/admin/event/?id=${encodeURIComponent(item.id)}`;
      return window.huuperListPage.renderCompactLink(href, item.title || item.id, window.huuperListPage.dateTime(item.event_date));
    },
  });
})();
