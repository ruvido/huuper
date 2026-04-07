(() => {
  if (!window.huuperEntityList || !window.huuperListPage) {
    return;
  }

  window.huuperEntityList.init({
    statusSelector: "#events-status",
    listSelector: "#events-list",
    emptyMessage: "No events.",
    errorMessage: "Events unavailable.",
    load: async () => {
      const response = await fetch("/api/collections/events/records?sort=event_date&perPage=200");
      if (!response.ok) {
        throw new Error("events_failed");
      }
      return response.json();
    },
    renderItem: (item) => {
      const href = `/me/event/?id=${encodeURIComponent(item.id)}`;
      return window.huuperListPage.renderListItemLink(href, item.title || item.id, window.huuperListPage.dateTime(item.event_date));
    },
  });
})();
