(() => {
  if (!window.huuperEntityList || !window.huuperAuth || !window.huuperListPage) {
    return;
  }

  const STATUS_LABELS = {
    active: "Attivo",
    completed: "Completato",
    abandoned: "Abbandonato",
  };

  function durationDays(item) {
    if (!item.start_date || !item.end_date) return 0;
    const start = new Date(item.start_date);
    const end = new Date(item.end_date);
    if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime())) return 0;
    return Math.round((end - start) / (1000 * 60 * 60 * 24));
  }

  function metaFor(item) {
    const status = STATUS_LABELS[item.status] || item.status || "";
    const days = durationDays(item);
    return [status, days ? `${days} giorni` : ""].filter(Boolean).join(" · ");
  }

  function renderEmpty() {
    return `
      <section class="empty-state empty-state-icon-only" aria-label="Battleplan">
        <svg class="empty-state-icon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor" aria-hidden="true"><path d="M8 15A7 7 0 1 1 8 1a7 7 0 0 1 0 14m0 1A8 8 0 1 0 8 0a8 8 0 0 0 0 16"/><path d="M8 13A5 5 0 1 1 8 3a5 5 0 0 1 0 10m0 1A6 6 0 1 0 8 2a6 6 0 0 0 0 12"/><path d="M8 11a3 3 0 1 1 0-6 3 3 0 0 1 0 6m0 1a4 4 0 1 0 0-8 4 4 0 0 0 0 8"/><path d="M9.5 8a1.5 1.5 0 1 1-3 0 1.5 1.5 0 0 1 3 0"/></svg>
        <p class="empty-state-label">nessun piano di battaglia</p>
      </section>
    `;
  }

  window.huuperEntityList.init({
    statusSelector: "#battleplan-status",
    listSelector: "#battleplan-list",
    requiresAuth: true,
    errorMessage: "Battleplan unavailable.",
    renderEmpty,
    load: () => window.huuperAuth.apiFetch("/api/me/battleplans"),
    renderItem: (item) => {
      const title = (item.data && item.data.priority && item.data.priority.title) || "Piano di Battaglia";
      const href = `/admin/battleplan/?id=${encodeURIComponent(item.id)}`;
      return window.huuperListPage.renderListItemLink(href, title, metaFor(item));
    },
  });
})();
