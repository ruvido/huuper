(() => {
  if (!window.appEntityList || !window.appAuth || !window.appListPage) {
    return;
  }

  window.appEntityList.init({
    statusSelector: "#groups-status",
    listSelector: "#groups-list",
    requiresAuth: true,
    emptyMessage: "No groups.",
    errorMessage: "Groups unavailable.",
    load: () => window.appAuth.apiFetch("/api/me/groups"),
    renderItem: (item) => {
      const type = window.appListPage.text(item.type);
      const membersCount = Number.isFinite(item.members_count) ? item.members_count : null;
      const requestsCount = Number.isFinite(item.requests_count) ? item.requests_count : null;
      const meta = [];
      if (type) meta.push(type);
      if (membersCount !== null) meta.push(`${membersCount} members`);
      if (requestsCount !== null && requestsCount > 0) meta.push(`${requestsCount} pending`);
      const inviteLink = window.appListPage.text(item.invite_link);
      const isMember = item.is_member === true;
      const assistantBadge = window.appGroupMeta && window.appGroupMeta.assistantMissing(item)
        ? window.appGroupMeta.assistantWarningBadge()
        : "";

      const mediaHTML = `
        <span class="list-item-media" aria-hidden="true">
          <span class="list-item-media-face">
            <span class="list-item-media-text">${window.appListPage.escapeHTML(window.appListPage.initials(item.name || item.id))}</span>
          </span>
        </span>
      `;
      const mainHTML = `
        <span class="list-item-main">
          <span class="list-item-copy">
            <strong>${window.appListPage.escapeHTML(item.name || item.id)}</strong>
            ${meta.length ? `<span class="list-item-meta">${window.appListPage.escapeHTML(meta.join(" • "))}</span>` : ""}
          </span>
        </span>
      `;

      if (isMember) {
        const detailHref = `/me/group/?id=${encodeURIComponent(item.id)}`;
        return `
          <a class="list-item" href="${window.appListPage.escapeHTML(detailHref)}">
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
            ${inviteLink ? `<span class="action-row"><a class="group-join-button" href="${window.appListPage.escapeHTML(inviteLink)}" target="_blank" rel="noopener noreferrer">Join</a></span>` : ""}
            ${assistantBadge}
          </span>
        </article>
      `;
    },
  });
})();
