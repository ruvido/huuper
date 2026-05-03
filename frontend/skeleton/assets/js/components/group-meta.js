window.appGroupMeta = (() => {
  function text(value) {
    return window.appListPage.text(value);
  }

  function escapeHTML(value) {
    return window.appListPage.escapeHTML(value);
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

  function assistantMissing(item) {
    return !text(item && item.assistant);
  }

  function assistantWarningBadge() {
    return `
      <span class="group-warning-badge" aria-label="Assistant missing" title="Assistant missing">
        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" viewBox="0 0 16 16" aria-hidden="true">
          <path d="M11 5a3 3 0 1 1-6 0 3 3 0 0 1 6 0m-9 8c0 1 1 1 1 1h5.256A4.5 4.5 0 0 1 8 12.5a4.5 4.5 0 0 1 1.544-3.393Q8.844 9.002 8 9c-5 0-6 3-6 4"/>
          <path d="M16 12.5a3.5 3.5 0 1 1-7 0 3.5 3.5 0 0 1 7 0m-3.5-2a.5.5 0 0 0-.5.5v1.5a.5.5 0 0 0 1 0V11a.5.5 0 0 0-.5-.5m0 4a.5.5 0 1 0 0-1 .5.5 0 0 0 0 1"/>
        </svg>
      </span>
    `;
  }

  return {
    groupMeta,
    renderSummary,
    memberMeta,
    assistantMissing,
    assistantWarningBadge,
  };
})();
