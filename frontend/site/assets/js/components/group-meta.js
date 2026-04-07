window.huuperGroupMeta = (() => {
  function text(value) {
    return window.huuperListPage.text(value);
  }

  function escapeHTML(value) {
    return window.huuperListPage.escapeHTML(value);
  }

  function groupMeta(group, options = {}) {
    const parts = [];
    const type = text(group.type);
    const membersCount = Number.isFinite(group.members_count) ? group.members_count : null;
    const requestsCount = Number.isFinite(group.requests_count) ? group.requests_count : null;
    const pendingLabel = options.pendingLabel || "pending";
    const requestsVisible = options.requestsVisible !== false;

    if (type) parts.push(type);
    if (membersCount !== null) parts.push(`${membersCount} members`);
    if (requestsVisible && requestsCount !== null && requestsCount > 0) {
      parts.push(`${requestsCount} ${pendingLabel}`);
    }

    return parts.join(" • ");
  }

  function renderSummary(group, options = {}) {
    const type = text(group.type) || "unknown";
    const membersCount = Number.isFinite(group.members_count) ? String(group.members_count) : "0";
    const requestsCount = Number.isFinite(group.requests_count) && options.requestsVisible !== false && group.requests_count > 0
      ? String(group.requests_count)
      : "0";

    return `
      <article class="group-stats">
        <div class="group-stat">
          <span class="group-stat-label">Type</span>
          <strong class="group-stat-value">${escapeHTML(type)}</strong>
        </div>
        <div class="group-stat">
          <span class="group-stat-label">Members</span>
          <strong class="group-stat-value">${escapeHTML(membersCount)}</strong>
        </div>
        <div class="group-stat">
          <span class="group-stat-label">Pending</span>
          <strong class="group-stat-value">${escapeHTML(requestsCount)}</strong>
        </div>
      </article>
    `;
  }

  function memberMeta(item) {
    const meta = [];
    if (item.is_guardian) {
      if (item.proteges_count > 0) {
        meta.push(`guardian of ${item.proteges_count}`);
      } else {
        meta.push("guardian");
      }
    }
    return meta.join(" • ");
  }

  return {
    groupMeta,
    renderSummary,
    memberMeta,
  };
})();
