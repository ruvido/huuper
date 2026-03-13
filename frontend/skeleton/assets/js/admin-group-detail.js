(() => {
  const statusNode = document.querySelector("#group-status");
  const summaryNode = document.querySelector("#group-summary");
  const membersNode = document.querySelector("#group-members");
  if (!statusNode || !summaryNode || !membersNode || !window.huuperAuth || !window.huuperListPage) {
    return;
  }

  const id = window.huuperListPage.queryParam("id");
  if (!id) {
    window.huuperListPage.setStatus(statusNode, "Missing group.");
    return;
  }

  window.huuperAuth.apiFetch(`/api/admin/groups/${encodeURIComponent(id)}`).then((group) => {
    window.huuperListPage.setStatus(statusNode, "");
    summaryNode.hidden = false;
    summaryNode.innerHTML = `<article><strong>${window.huuperListPage.escapeHTML(group.name || id)}</strong><p class="status">${[group.type, `${group.members_count} members`, group.requests_count > 0 ? `${group.requests_count} requests` : ""].filter(Boolean).join(" • ")}</p></article>`;

    const memberItems = Array.isArray(group.members) ? group.members : [];
    if (memberItems.length > 0) {
      membersNode.hidden = false;
      window.huuperListPage.renderList(membersNode, memberItems, (item) => {
        const meta = [];
        if (item.is_guardian) {
          if (item.proteges_count > 0) {
            meta.push(`guardian of ${item.proteges_count}`);
          } else {
            meta.push("guardian");
          }
        }
        if (!item.is_guardian && item.proteges_count > 0) {
          meta.push(`${item.proteges_count} requests`);
        }
        return `<a href="/admin/user/?id=${encodeURIComponent(item.id)}"><strong>${window.huuperListPage.escapeHTML(item.full_name || item.email || item.id)}</strong>${meta.length > 0 ? `<p class="status">${meta.join(" • ")}</p>` : ""}</a>`;
      });
    }
  }).catch(() => {
    window.huuperListPage.setStatus(statusNode, "Group unavailable.");
  });
})();
