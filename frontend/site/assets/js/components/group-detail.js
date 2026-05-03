window.appGroupDetail = (() => {
  function tabLabel(label, count) {
    return `<span class="section-tab-text">${window.appListPage.escapeHTML(label)}</span><span class="section-tab-count">${window.appListPage.escapeHTML(String(count))}</span>`;
  }

  function currentAuth() {
    return window.appAuth && typeof window.appAuth.read === "function" ? window.appAuth.read() : null;
  }

  function canChangeAssistant(group) {
    const auth = currentAuth();
    const model = auth && auth.model ? auth.model : null;
    if (!model) {
      return false;
    }
    if (model.admin === true) {
      return true;
    }
    return window.appListPage.text(model.id) !== "" && window.appListPage.text(group && group.assistant) === window.appListPage.text(model.id);
  }

  function init(config) {
    const statusNode = document.querySelector("#group-status");
    const assistantSectionNode = document.querySelector("#group-assistant-section");
    const topbarAssistantActionNode = document.querySelector("[data-group-assistant-action]");
    const topbarNode = document.querySelector(".top-bar");
    const topbarTitleNode = document.querySelector(".top-bar-title");
    const tabsNode = document.querySelector("#group-tabs");
    const tabMembersNode = document.querySelector("#group-tab-members");
    const tabPendingNode = document.querySelector("#group-tab-pending");
    const pendingSectionNode = document.querySelector("#group-pending-section");
    const pendingNode = document.querySelector("#group-pending");
    const membersSectionNode = document.querySelector("#group-members-section");
    const membersNode = document.querySelector("#group-members");
    if (!statusNode || !assistantSectionNode || !tabsNode || !tabMembersNode || !tabPendingNode || !pendingSectionNode || !pendingNode || !membersSectionNode || !membersNode || !window.appAuth || !window.appListPage || !window.appRequestItem || !window.appListItem) {
      return;
    }

    function showTab(name, hasPending) {
      const showingPending = name === "pending" && hasPending;
      membersSectionNode.hidden = showingPending;
      pendingSectionNode.hidden = !showingPending;
      tabMembersNode.classList.toggle("section-tab-current", !showingPending);
      tabPendingNode.classList.toggle("section-tab-current", showingPending);
    }

    const id = window.appListPage.queryParam("id");
    if (!id) {
      window.appListPage.setStatus(statusNode, "Missing group.");
      return;
    }

    window.appAuth.apiFetch(config.detailURL(id)).then((group) => {
      window.appListPage.setStatus(statusNode, "");
      if (topbarNode) {
        topbarNode.classList.add("top-bar-flat");
      }
      const assistantAllowed = canChangeAssistant(group);
      if (topbarAssistantActionNode) {
        topbarAssistantActionNode.hidden = !assistantAllowed || !config.assistantURL;
        if (assistantAllowed && config.assistantURL && !topbarAssistantActionNode.dataset.bound) {
          topbarAssistantActionNode.dataset.bound = "1";
          topbarAssistantActionNode.addEventListener("click", () => {
            if (!window.appGroupAssistantSheet || typeof window.appGroupAssistantSheet.open !== "function") {
              return;
            }

            window.appGroupAssistantSheet.open({
              groupID: id,
              groupName: group.name || id,
              members: Array.isArray(group.members) ? group.members : [],
              assistantURL: typeof config.assistantURL === "function" ? config.assistantURL(id) : "",
              redirectURL: typeof config.assistantRedirectURL === "string" && config.assistantRedirectURL.trim()
                ? config.assistantRedirectURL.trim()
                : window.location.pathname + window.location.search,
            });
          });
        }
      }
      if (topbarTitleNode && group && typeof group.name === "string" && group.name.trim()) {
        topbarTitleNode.textContent = group.name.trim();
      }

      const pendingItems = Array.isArray(group.pending_requests) ? group.pending_requests : [];
      const hasPending = pendingItems.length > 0;
      tabsNode.hidden = !hasPending;
      tabPendingNode.hidden = !hasPending;
      tabMembersNode.innerHTML = tabLabel("Members", Array.isArray(group.members) ? group.members.length : 0);
      tabPendingNode.innerHTML = tabLabel("Pending", pendingItems.length);
      if (pendingItems.length > 0) {
        window.appListPage.renderList(pendingNode, pendingItems, (item) => {
          const href = `/${config.scope}/request/?id=${encodeURIComponent(item.id)}`;
          return window.appRequestItem.renderListItem(item, href);
        });
      }

      const memberItems = Array.isArray(group.members) ? group.members : [];
      const canManageAssistant = config.manageAssistant === true && window.appGroupMeta && window.appGroupMeta.assistantMissing(group);
      if (canManageAssistant && memberItems.length > 0) {
        assistantSectionNode.hidden = false;
        assistantSectionNode.innerHTML = `
          <button id="group-assistant-button" class="group-assistant-button" type="button">Assign assistant</button>
        `;
        const button = assistantSectionNode.querySelector("#group-assistant-button");
        if (button) {
          button.addEventListener("click", () => {
            if (!window.appGroupAssistantSheet || typeof window.appGroupAssistantSheet.open !== "function") {
              return;
            }
            window.appGroupAssistantSheet.open({
              groupID: id,
              groupName: group.name || id,
              members: memberItems,
              assistantURL: typeof config.assistantURL === "function" ? config.assistantURL(id) : "",
              redirectURL: typeof config.assistantRedirectURL === "string" && config.assistantRedirectURL.trim()
                ? config.assistantRedirectURL.trim()
                : window.location.pathname + window.location.search,
            });
          });
        }
      } else {
        assistantSectionNode.hidden = true;
        assistantSectionNode.innerHTML = "";
      }

      if (memberItems.length > 0) {
        window.appListPage.renderList(membersNode, memberItems, (item) => {
          return window.appListItem.renderMember(item, `/${config.scope}/user/?id=${encodeURIComponent(item.id)}`);
        });
      }

      if (memberItems.length > 0) {
        showTab("members", hasPending);
      } else if (hasPending) {
        showTab("pending", hasPending);
      }

      tabMembersNode.addEventListener("click", () => showTab("members", hasPending));
      tabPendingNode.addEventListener("click", () => showTab("pending", hasPending));
    }).catch(() => {
      window.appListPage.setStatus(statusNode, "Group unavailable.");
    });
  }

  return { init };
})();
