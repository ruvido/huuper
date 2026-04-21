(() => {
  if (!window.huuperEntityList || !window.huuperAuth || !window.huuperListPage) {
    return;
  }

  window.huuperEntityList.init({
    statusSelector: "#groups-status",
    listSelector: "#groups-list",
    requiresAuth: true,
    emptyMessage: "No groups.",
    errorMessage: "Groups unavailable.",
    load: () => window.huuperAuth.apiFetch("/api/me/groups"),
    renderItem: (item) => {
      const type = window.huuperListPage.text(item.type);
      const membersCount = Number.isFinite(item.members_count) ? item.members_count : null;
      const requestsCount = Number.isFinite(item.requests_count) ? item.requests_count : null;
      const meta = [];
      if (type) meta.push(type);
      if (membersCount !== null) meta.push(`${membersCount} members`);
      if (requestsCount !== null && requestsCount > 0) meta.push(`${requestsCount} pending`);
      const inviteLink = window.huuperListPage.text(item.invite_link);
      const isMember = item.is_member === true;
      const assistantBadge = window.huuperGroupMeta && window.huuperGroupMeta.assistantMissing(item)
        ? window.huuperGroupMeta.assistantWarningBadge()
        : "";

      const mediaHTML = `
        <span class="list-item-media" aria-hidden="true">
          <span class="list-item-media-face">
            <span class="list-item-media-text">${window.huuperListPage.escapeHTML(window.huuperListPage.initials(item.name || item.id))}</span>
          </span>
        </span>
      `;
      const mainHTML = `
        <span class="list-item-main">
          <span class="list-item-copy">
            <strong>${window.huuperListPage.escapeHTML(item.name || item.id)}</strong>
            ${meta.length ? `<span class="list-item-meta">${window.huuperListPage.escapeHTML(meta.join(" • "))}</span>` : ""}
          </span>
        </span>
      `;

      if (isMember) {
        const detailHref = `/me/group/?id=${encodeURIComponent(item.id)}`;
        return `
          <a class="list-item" href="${window.huuperListPage.escapeHTML(detailHref)}">
            ${mediaHTML}
            ${mainHTML}
            ${assistantBadge ? `<span class="list-item-side">${assistantBadge}</span>` : ""}
          </a>
        `;
      }

      return `
        <article class="list-item">
          ${mediaHTML}
          ${mainHTML}
          <span class="list-item-side">
            ${inviteLink ? `<span class="action-row"><a class="group-join-button" href="${window.huuperListPage.escapeHTML(inviteLink)}" target="_blank" rel="noopener noreferrer">Join</a></span>` : ""}
            ${assistantBadge}
          </span>
        </article>
      `;
    },
  });
})();
