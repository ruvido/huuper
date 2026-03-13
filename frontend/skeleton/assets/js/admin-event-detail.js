(() => {
  const statusNode = document.querySelector("#event-status");
  const summaryNode = document.querySelector("#event-summary");
  const registrationsNode = document.querySelector("#event-registrations");
  if (!statusNode || !summaryNode || !registrationsNode || !window.huuperAuth || !window.huuperListPage) {
    return;
  }

  const id = window.huuperListPage.queryParam("id");
  if (!id) {
    window.huuperListPage.setStatus(statusNode, "Missing event.");
    return;
  }

  window.huuperAuth.apiFetch(`/api/admin/events/${encodeURIComponent(id)}`).then((payload) => {
    const event = payload.event || {};
    summaryNode.hidden = false;
    summaryNode.innerHTML = `<article><strong>${window.huuperListPage.escapeHTML(event.title || id)}</strong><p class="status">${window.huuperListPage.dateTime(event.event_date)}</p></article>`;

    const items = Array.isArray(payload.registrations) ? payload.registrations : [];
    if (items.length > 0) {
      registrationsNode.hidden = false;
      window.huuperListPage.renderList(registrationsNode, items, (item) => `<article><strong>${window.huuperListPage.escapeHTML(item.email || item.id)}</strong>${item.status ? `<p class="status">${window.huuperListPage.escapeHTML(item.status)}</p>` : ""}</article>`);
    }
    window.huuperListPage.setStatus(statusNode, "");
  }).catch(() => {
    window.huuperListPage.setStatus(statusNode, "Event unavailable.");
  });
})();
