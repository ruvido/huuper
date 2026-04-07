window.huuperUserDetail = (() => {
  function renderStatusLine(parts) {
    const clean = parts.filter(Boolean);
    return clean.length ? `<p class="meta-text">${clean.join(" • ")}</p>` : "";
  }

  function renderGroups(node, groups, scope) {
    if (!node || groups.length === 0) {
      return;
    }
    node.hidden = false;
    window.huuperListPage.renderList(node, groups, (group) => {
      return window.huuperListPage.renderListItemLink(`/${scope}/group/?id=${encodeURIComponent(group.id)}`, group.name || group.id, window.huuperListPage.text(group.type));
    });
  }

  function init(config) {
    const statusNode = document.querySelector("#user-status");
    const summaryNode = document.querySelector("#user-summary");
    const groupsNode = document.querySelector("#user-groups");
    if (!statusNode || !summaryNode || !window.huuperAuth || !window.huuperListPage || !window.huuperRequestItem) {
      return;
    }

    const id = window.huuperListPage.queryParam("id");
    if (!id) {
      window.huuperListPage.setStatus(statusNode, "Missing user.");
      return;
    }

    window.huuperAuth.apiFetch(config.detailURL(id)).then((payload) => {
      const telegram = payload.telegram || {};
      const telegramName = telegram.username || telegram.first_name || "";
      const rows = [];
      const meta = [];

      if (payload.email) meta.push(window.huuperListPage.escapeHTML(payload.email));
      if (config.includeStatus && payload.status) meta.push(window.huuperListPage.escapeHTML(payload.status));
      if (config.includeAdminFlag && payload.admin) meta.push("admin");
      const statusLine = renderStatusLine(meta);
      if (statusLine) rows.push(statusLine);

      if (telegramName) {
        rows.push(`<p class="meta-text">Telegram: ${window.huuperListPage.escapeHTML(telegramName)}</p>`);
      }

      const guardianRequests = Array.isArray(payload.guardian_requests) ? payload.guardian_requests : [];
      if (guardianRequests.length > 0) {
        rows.push(`<p class="meta-text">Guardian of:</p>`);
      }
      for (const request of guardianRequests) {
        rows.push(window.huuperRequestItem.renderListItem(request, `/${config.scope}/request/?id=${encodeURIComponent(request.id)}`));
      }

      summaryNode.hidden = false;
      summaryNode.innerHTML = `<article class="detail-card"><strong>${window.huuperListPage.escapeHTML(payload.full_name || payload.email || id)}</strong>${rows.join("")}</article>`;

      if (config.includeGroups) {
        renderGroups(groupsNode, Array.isArray(payload.groups) ? payload.groups : [], config.scope);
      }

      window.huuperListPage.setStatus(statusNode, "");
    }).catch(() => {
      window.huuperListPage.setStatus(statusNode, "User unavailable.");
    });
  }

  return { init };
})();
