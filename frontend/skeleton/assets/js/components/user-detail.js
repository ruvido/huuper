window.appUserDetail = (() => {
  function renderGroups(node, groups, scope) {
    if (!node || groups.length === 0) {
      return;
    }
    node.hidden = false;
    window.appListPage.renderList(node, groups, (group) => {
      const sideHTML = window.appGroupMeta && window.appGroupMeta.assistantMissing(group)
        ? window.appGroupMeta.assistantWarningBadge()
        : "";
      return window.appListPage.renderListItemLink(
        `/${scope}/group/?id=${encodeURIComponent(group.id)}`,
        group.name || group.id,
        window.appListPage.text(group.type),
        { sideHTML },
      );
    });
  }

  function init(config) {
    const statusNode = document.querySelector("#user-status");
    const summaryNode = document.querySelector("#user-summary");
    const deleteButtonNode = document.querySelector("[data-request-reject]");
    const guardianLabelNode = document.querySelector("#user-guardian-label");
    const guardianDividerNode = guardianLabelNode ? guardianLabelNode.previousElementSibling : null;
    const requestsNode = document.querySelector("#user-requests");
    const groupsLabelNode = document.querySelector("#user-groups-label");
    const groupsDividerNode = groupsLabelNode ? groupsLabelNode.previousElementSibling : null;
    const groupsNode = document.querySelector("#user-groups");
    if (!statusNode || !summaryNode || !requestsNode || !window.appAuth || !window.appListPage || !window.appRequestItem || !window.appUserSummary) {
      return;
    }

    const id = window.appListPage.queryParam("id");
    if (!id) {
      window.appListPage.setStatus(statusNode, "Missing user.");
      return;
    }

    const cancelCopy = (((window.appCopy || {}).ui || {}).admin || {}).user || {};
    const cancelDialogCopy = cancelCopy.cancelDialog || {};
    const cancelDialogTitle = window.appListPage.text(cancelDialogCopy.title) || "Set user as cancelled?";
    const cancelDialogDescription = window.appListPage.text(cancelDialogCopy.description) || "The user will not be visible anymore in groups.";
    const cancelDialogConfirmLabel = window.appListPage.text(cancelDialogCopy.confirmLabel) || "Cancel user";
    const cancelDialogFallbackError = window.appListPage.text(cancelDialogCopy.fallbackError) || "Cancel unavailable.";

    if (deleteButtonNode && typeof config.cancelURL === "function" && window.appActionSheet) {
      deleteButtonNode.addEventListener("click", () => {
        window.appActionSheet.open({
          title: cancelDialogTitle,
          contentHTML: `<p class="meta-text">${window.appListPage.escapeHTML(cancelDialogDescription)}</p>`,
          footerAction: {
            label: cancelDialogConfirmLabel,
            tone: "danger",
            onSelect: async () => {
              try {
                window.appListPage.setStatus(statusNode, "");
                await Promise.all([
                  window.appAuth.apiFetch(config.cancelURL(id), { method: "POST" }),
                  new Promise((resolve) => setTimeout(resolve, 800)),
                ]);
                window.location.href = `/${config.scope}/users/`;
                await new Promise(() => {});
              } catch (error) {
                const message = window.appRequestActions && typeof window.appRequestActions.errorMessage === "function"
                  ? window.appRequestActions.errorMessage(error, cancelDialogFallbackError)
                  : cancelDialogFallbackError;
                window.appListPage.setStatus(statusNode, message);
                throw error;
              }
            },
          },
        });
      });
    }

    window.appAuth.apiFetch(config.detailURL(id)).then((payload) => {
      summaryNode.hidden = false;
      summaryNode.innerHTML = window.appUserSummary.render(payload, {
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
        window.appListPage.renderList(requestsNode, guardianRequests, (request) => {
          return window.appRequestItem.renderListItem(request, `/${config.scope}/request/?id=${encodeURIComponent(request.id)}`);
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

      window.appListPage.setStatus(statusNode, "");
    }).catch(() => {
      window.appListPage.setStatus(statusNode, "User unavailable.");
    });
  }

  return { init };
})();
