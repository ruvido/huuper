(() => {
  const statusNode = document.querySelector("#user-status");
  const summaryNode = document.querySelector("#user-summary");
  const groupsNode = document.querySelector("#user-groups");
  if (!statusNode || !summaryNode || !groupsNode || !window.huuperAuth || !window.huuperListPage || !window.huuperRequestCard) {
    return;
  }

  const id = window.huuperListPage.queryParam("id");
  if (!id) {
    window.huuperListPage.setStatus(statusNode, "Missing user.");
    return;
  }

  window.huuperAuth.apiFetch(`/api/admin/users/${encodeURIComponent(id)}`).then((payload) => {
    const telegram = payload.telegram || {};
    const telegramName = telegram.username || telegram.first_name || "";
    const rows = [
      `<p class="status">${[window.huuperListPage.escapeHTML(payload.email || ""), window.huuperListPage.escapeHTML(payload.status || ""), payload.admin ? "admin" : ""].filter(Boolean).join(" • ")}</p>`,
    ];
    if (telegramName) {
      rows.push(`<p class="status">telegram: ${window.huuperListPage.escapeHTML(telegramName)}</p>`);
    }

    const guardianRequests = Array.isArray(payload.guardian_requests) ? payload.guardian_requests : [];
    if (guardianRequests.length > 0) {
      rows.push(`<p class="status">Guardian of:</p>`);
    }
    for (const request of guardianRequests) {
      rows.push(window.huuperRequestCard.renderCompact(request, `/admin/request/?id=${encodeURIComponent(request.id)}`));
    }

    summaryNode.hidden = false;
    summaryNode.innerHTML = `<article><strong>${window.huuperListPage.escapeHTML(payload.full_name || payload.email || id)}</strong>${rows.join("")}</article>`;

    const groups = Array.isArray(payload.groups) ? payload.groups : [];
    if (groups.length > 0) {
      groupsNode.hidden = false;
      window.huuperListPage.renderList(groupsNode, groups, (group) => `<article><strong>${window.huuperListPage.escapeHTML(group.name || group.id)}</strong>${group.type ? `<p class="status">${window.huuperListPage.escapeHTML(group.type)}</p>` : ""}</article>`);
    }
    window.huuperListPage.setStatus(statusNode, "");
  }).catch(() => {
    window.huuperListPage.setStatus(statusNode, "User unavailable.");
  });
})();
