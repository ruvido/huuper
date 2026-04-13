window.huuperUserDetail = (() => {
  function renderGroups(node, groups, scope) {
    if (!node || groups.length === 0) {
      return;
    }
    node.hidden = false;
    window.huuperListPage.renderList(node, groups, (group) => {
      const sideHTML = window.huuperGroupMeta && window.huuperGroupMeta.assistantMissing(group)
        ? window.huuperGroupMeta.assistantWarningBadge()
        : "";
      return window.huuperListPage.renderListItemLink(
        `/${scope}/group/?id=${encodeURIComponent(group.id)}`,
        group.name || group.id,
        window.huuperListPage.text(group.type),
        { sideHTML },
      );
    });
  }

  function init(config) {
    const statusNode = document.querySelector("#user-status");
    const summaryNode = document.querySelector("#user-summary");
    const guardianLabelNode = document.querySelector("#user-guardian-label");
    const guardianDividerNode = guardianLabelNode ? guardianLabelNode.previousElementSibling : null;
    const requestsNode = document.querySelector("#user-requests");
    const groupsLabelNode = document.querySelector("#user-groups-label");
    const groupsDividerNode = groupsLabelNode ? groupsLabelNode.previousElementSibling : null;
    const groupsNode = document.querySelector("#user-groups");
    if (!statusNode || !summaryNode || !requestsNode || !window.huuperAuth || !window.huuperListPage || !window.huuperRequestItem || !window.huuperUserSummary) {
      return;
    }

    const id = window.huuperListPage.queryParam("id");
    if (!id) {
      window.huuperListPage.setStatus(statusNode, "Missing user.");
      return;
    }

    window.huuperAuth.apiFetch(config.detailURL(id)).then((payload) => {
      summaryNode.hidden = false;
      summaryNode.innerHTML = window.huuperUserSummary.render(payload, {
        includeStatus: config.includeStatus === true,
        includeAdminFlag: config.includeAdminFlag === true,
      });
      const guardianRequests = Array.isArray(payload.guardian_requests) ? payload.guardian_requests : [];
      if (guardianRequests.length > 0) {
        if (guardianLabelNode) {
          guardianLabelNode.hidden = false;
        }
        if (guardianDividerNode && guardianDividerNode.classList && guardianDividerNode.classList.contains("section-divider")) {
          guardianDividerNode.hidden = false;
        }
        requestsNode.hidden = false;
        window.huuperListPage.renderList(requestsNode, guardianRequests, (request) => {
          return window.huuperRequestItem.renderListItem(request, `/${config.scope}/request/?id=${encodeURIComponent(request.id)}`);
        });
      } else {
        if (guardianLabelNode) {
          guardianLabelNode.hidden = true;
        }
        if (guardianDividerNode && guardianDividerNode.classList && guardianDividerNode.classList.contains("section-divider")) {
          guardianDividerNode.hidden = true;
        }
        requestsNode.hidden = true;
        requestsNode.innerHTML = "";
      }

      if (config.includeGroups) {
        const groups = Array.isArray(payload.groups) ? payload.groups : [];
        if (groupsLabelNode) {
          groupsLabelNode.hidden = groups.length === 0;
        }
        if (groupsDividerNode && groupsDividerNode.classList && groupsDividerNode.classList.contains("section-divider")) {
          groupsDividerNode.hidden = groups.length === 0;
        }
        renderGroups(groupsNode, groups, config.scope);
      }

      window.huuperListPage.setStatus(statusNode, "");
    }).catch(() => {
      window.huuperListPage.setStatus(statusNode, "User unavailable.");
    });
  }

  return { init };
})();
