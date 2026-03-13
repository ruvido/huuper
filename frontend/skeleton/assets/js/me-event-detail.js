(() => {
  const statusNode = document.querySelector("#event-status");
  const summaryNode = document.querySelector("#event-summary");
  if (!statusNode || !summaryNode || !window.huuperAuth || !window.huuperListPage) {
    return;
  }

  const id = window.huuperListPage.queryParam("id");
  if (!id) {
    window.huuperListPage.setStatus(statusNode, "Missing event.");
    return;
  }

  fetch(`/api/collections/events/records/${encodeURIComponent(id)}`).then(async (response) => {
    if (!response.ok) throw new Error("event_failed");
    const event = await response.json();
    let registrationText = "";
    if (event.slug) {
      try {
        const registration = await window.huuperAuth.apiFetch(`/api/me/events/${encodeURIComponent(event.slug)}/status`);
        registrationText = registration && registration.registered ? "registered" : "not registered";
      } catch (_) {}
    }
    summaryNode.hidden = false;
    summaryNode.innerHTML = `<article><strong>${window.huuperListPage.escapeHTML(event.title || id)}</strong><p class="status">${[window.huuperListPage.dateTime(event.event_date), registrationText].filter(Boolean).join(" • ")}</p></article>`;
    window.huuperListPage.setStatus(statusNode, "");
  }).catch(() => {
    window.huuperListPage.setStatus(statusNode, "Event unavailable.");
  });
})();
