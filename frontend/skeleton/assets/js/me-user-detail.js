(() => {
  const statusNode = document.querySelector("#user-status");
  const summaryNode = document.querySelector("#user-summary");
  if (!statusNode || !summaryNode || !window.huuperAuth || !window.huuperListPage || !window.huuperRequestCard) {
    return;
  }

  const id = window.huuperListPage.queryParam("id");
  if (!id) {
    window.huuperListPage.setStatus(statusNode, "Missing user.");
    return;
  }

  window.huuperAuth.apiFetch(`/api/me/users/${encodeURIComponent(id)}`).then((payload) => {
    const telegram = payload.telegram || {};
    const telegramName = telegram.username || telegram.first_name || "";
    const rows = [`<p class="status">${window.huuperListPage.escapeHTML(payload.email || "")}</p>`];
    if (telegramName) {
      rows.push(`<p class="status">telegram: ${window.huuperListPage.escapeHTML(telegramName)}</p>`);
    }

    const guardianRequests = Array.isArray(payload.guardian_requests) ? payload.guardian_requests : [];
    if (guardianRequests.length > 0) {
      rows.push(`<p class="status">Guardian of:</p>`);
    }
    for (const request of guardianRequests) {
      rows.push(window.huuperRequestCard.renderCompact(request, `/me/request/?id=${encodeURIComponent(request.id)}`));
    }

    summaryNode.hidden = false;
    summaryNode.innerHTML = `<article><strong>${window.huuperListPage.escapeHTML(payload.full_name || payload.email || id)}</strong>${rows.join("")}</article>`;

    window.huuperListPage.setStatus(statusNode, "");
  }).catch(() => {
    window.huuperListPage.setStatus(statusNode, "User unavailable.");
  });
})();
