window.huuperRequestListPage = (() => {
  function text(value) {
    return window.huuperListPage.text(value);
  }

  function bindActions(node, config) {
    if (!window.huuperRequestAssignmentSheet) {
      return;
    }

    node.querySelectorAll(".request-item-action").forEach((button) => {
      if (button.dataset.bound === "true") {
        return;
      }
      button.dataset.bound = "true";
      button.addEventListener("click", () => {
        const requestID = text(button.getAttribute("data-request-id"));
        if (!requestID) {
          return;
        }
        window.huuperRequestAssignmentSheet.open({
          requestID,
          requestsURL: config.requestsURL,
          detailURL: config.detailURL,
          actionURL: config.actionURL,
        });
      });
    });
  }

  function renderEmptyState(label) {
    return `
      <section class="empty-state empty-state-icon-only" aria-label="${label}">
        <svg class="empty-state-icon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor" aria-hidden="true">
          <path d="M4.98 4a.5.5 0 0 0-.39.188L1.54 8H6a.5.5 0 0 1 .5.5 1.5 1.5 0 1 0 3 0A.5.5 0 0 1 10 8h4.46l-3.05-3.812A.5.5 0 0 0 11.02 4zm-1.17-.437A1.5 1.5 0 0 1 4.98 3h6.04a1.5 1.5 0 0 1 1.17.563l3.7 4.625a.5.5 0 0 1 .106.374l-.39 3.124A1.5 1.5 0 0 1 14.117 13H1.883a1.5 1.5 0 0 1-1.489-1.314l-.39-3.124a.5.5 0 0 1 .106-.374z"/>
        </svg>
      </section>
    `;
  }

  function init(config) {
    const statusNode = document.querySelector(config.statusSelector);
    const tabsNode = document.querySelector(config.tabsSelector);
    const allTabNode = document.querySelector(config.allTabSelector);
    const urgentTabNode = document.querySelector(config.urgentTabSelector);
    const allCountNode = document.querySelector(config.allCountSelector);
    const urgentCountNode = document.querySelector(config.urgentCountSelector);
    const listNode = document.querySelector(config.listSelector);

    if (
      !statusNode ||
      !tabsNode ||
      !allTabNode ||
      !urgentTabNode ||
      !allCountNode ||
      !urgentCountNode ||
      !listNode ||
      !window.huuperAuth ||
      !window.huuperListPage ||
      !window.huuperRequestItem
    ) {
      return;
    }

    let activeTab = "all";
    let allItems = [];

    function urgentItems(items) {
      return items.filter((item) => {
        const workflow = item && typeof item.workflow === "object" ? item.workflow : {};
        const nextRole = text(workflow.next_role);
        const allowedRoles = Array.isArray(config.roles) ? config.roles : [];
        return workflow.can_take_action === true && allowedRoles.includes(nextRole);
      });
    }

    function render() {
      const urgent = urgentItems(allItems);
      const visibleItems = activeTab === "urgent" ? urgent : allItems;

      allCountNode.textContent = String(allItems.length);
      urgentCountNode.textContent = String(urgent.length);
      allTabNode.classList.toggle("section-tab-current", activeTab === "all");
      urgentTabNode.classList.toggle("section-tab-current", activeTab === "urgent");
      tabsNode.hidden = allItems.length === 0;

      if (allItems.length === 0) {
        listNode.innerHTML = renderEmptyState("No requests");
        listNode.hidden = false;
        statusNode.hidden = true;
        return;
      }

      if (visibleItems.length === 0) {
        listNode.innerHTML = renderEmptyState(activeTab === "urgent" ? "No urgent requests" : "No requests");
        listNode.hidden = false;
        statusNode.hidden = true;
        return;
      }

      window.huuperListPage.renderList(listNode, visibleItems, (item) => {
        return window.huuperRequestItem.renderListItem(item, config.itemHref(item.id), {
          assignGroupURL: config.assignGroupHref,
          assignGuardianURL: config.assignGuardianHref,
        });
      });
      listNode.hidden = false;
      statusNode.hidden = true;
      bindActions(listNode, config);
    }

    allTabNode.addEventListener("click", () => {
      activeTab = "all";
      render();
    });

    urgentTabNode.addEventListener("click", () => {
      activeTab = "urgent";
      render();
    });

    window.huuperAuth.apiFetch(config.loadURL).then((payload) => {
      allItems = Array.isArray(payload.items) ? payload.items : [];
      render();
    }).catch(() => {
      window.huuperListPage.setStatus(statusNode, config.errorMessage || "Requests unavailable.");
    });
  }

  return { init };
})();
