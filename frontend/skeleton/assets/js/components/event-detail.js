window.huuperEventDetail = (() => {
  function init(config) {
    const statusNode = document.querySelector("#event-status");
    const summaryNode = document.querySelector("#event-summary");
    const registrationsNode = document.querySelector("#event-registrations");
    if (!statusNode || !summaryNode || !window.huuperAuth || !window.huuperListPage || !window.huuperEventCard) {
      return;
    }

    const id = window.huuperListPage.queryParam("id");
    if (!id) {
      window.huuperListPage.setStatus(statusNode, "Missing event.");
      return;
    }

    window.huuperAuth.apiFetch(config.detailURL(id)).then((payload) => {
      const event = payload.event || {};
      const metaParts = [];
      const eventDate = window.huuperListPage.dateTime(event.event_date);
      if (eventDate) metaParts.push(eventDate);
      if (config.includeRegistrationState) {
        metaParts.push(payload.registered ? "Registered" : "Not registered");
      }
      summaryNode.hidden = false;
      summaryNode.innerHTML = window.huuperEventCard.summary(event, metaParts.filter(Boolean).join(" • "));

      if (config.includeRegistrations && registrationsNode) {
        const items = Array.isArray(payload.registrations) ? payload.registrations : [];
        if (items.length > 0) {
          registrationsNode.hidden = false;
          window.huuperListPage.renderList(registrationsNode, items, (item) => {
            return window.huuperListPage.renderCompactLink("#", item.email || item.id, item.status || "");
          });
        }
      }

      window.huuperListPage.setStatus(statusNode, "");
    }).catch(() => {
      window.huuperListPage.setStatus(statusNode, "Event unavailable.");
    });
  }

  return { init };
})();
