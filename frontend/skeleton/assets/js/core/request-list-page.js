window.appRequestListPage = (() => {
  function renderEmptyState(label) {
    return `
      <section class="empty-state empty-state-icon-only" aria-label="${label}">
        <svg class="empty-state-icon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor" aria-hidden="true">
          <path d="M4.98 4a.5.5 0 0 0-.39.188L1.54 8H6a.5.5 0 0 1 .5.5 1.5 1.5 0 1 0 3 0A.5.5 0 0 1 10 8h4.46l-3.05-3.812A.5.5 0 0 0 11.02 4zm-1.17-.437A1.5 1.5 0 0 1 4.98 3h6.04a1.5 1.5 0 0 1 1.17.563l3.7 4.625a.5.5 0 0 1 .106.374l-.39 3.124A1.5 1.5 0 0 1 14.117 13H1.883a1.5 1.5 0 0 1-1.489-1.314l-.39-3.124a.5.5 0 0 1 .106-.374z"/>
        </svg>
      </section>
    `;
  }

  function itemSearchText(item) {
    const data = item && typeof item.data === "object" ? item.data : {};
    return [item && item.full_name, data.full_name, data.name, item && item.email]
      .filter(Boolean)
      .join(" ")
      .toLowerCase();
  }

  function debounce(fn, delayMs) {
    let timer = null;
    return (...args) => {
      if (timer) {
        window.clearTimeout(timer);
      }
      timer = window.setTimeout(() => fn(...args), delayMs);
    };
  }

  function init(config) {
    const statusNode = document.querySelector(config.statusSelector);
    const searchNode = config.searchSelector ? document.querySelector(config.searchSelector) : null;
    const tabsNode = document.querySelector(config.tabsSelector);
    const allTabNode = document.querySelector(config.allTabSelector);
    const urgentTabNode = document.querySelector(config.urgentTabSelector);
    const archivedTabNode = config.archivedTabSelector ? document.querySelector(config.archivedTabSelector) : null;
    const allCountNode = document.querySelector(config.allCountSelector);
    const urgentCountNode = document.querySelector(config.urgentCountSelector);
    const archivedCountNode = config.archivedCountSelector ? document.querySelector(config.archivedCountSelector) : null;
    const listNode = document.querySelector(config.listSelector);

    if (
      !statusNode ||
      !tabsNode ||
      !allTabNode ||
      !urgentTabNode ||
      !allCountNode ||
      !urgentCountNode ||
      !listNode ||
      !window.appAuth ||
      !window.appListPage ||
      !window.appRequestItem
    ) {
      return;
    }

    let activeTab = "urgent";
    let allItems = [];
    let archivedItems = [];
    let archivedLoaded = false;
    let archivedLoading = false;
    let searchTerm = "";

    function urgentItems(items) {
      return items.filter((item) => {
        const workflow = item && typeof item.workflow === "object" ? item.workflow : {};
        return workflow.actor_is_assigned === true;
      });
    }

    function filterBySearch(items) {
      if (!searchTerm) {
        return items;
      }
      return items.filter((item) => itemSearchText(item).includes(searchTerm));
    }

    function loadArchivedItems() {
      if (!archivedTabNode || archivedLoading) {
        return;
      }
      archivedLoading = true;
      const url = new URL(config.loadURL, window.location.origin);
      url.searchParams.set("status", "archived");
      window.appAuth.apiFetch(`${url.pathname}${url.search}`).then((payload) => {
        archivedItems = Array.isArray(payload.items) ? payload.items : [];
        archivedLoaded = true;
        archivedLoading = false;
        render();
      }).catch(() => {
        archivedLoading = false;
        window.appListPage.setStatus(statusNode, config.errorMessage || "Requests unavailable.");
      });
    }

    function render() {
      const urgent = filterBySearch(urgentItems(allItems));
      const all = filterBySearch(allItems);
      const archived = filterBySearch(archivedItems);
      const visibleItems = activeTab === "urgent" ? urgent : activeTab === "archived" ? archived : all;

      allCountNode.textContent = String(all.length);
      urgentCountNode.textContent = String(urgent.length);
      if (archivedCountNode) {
        archivedCountNode.textContent = String(archived.length);
      }
      allTabNode.classList.toggle("section-tab-current", activeTab === "all");
      urgentTabNode.classList.toggle("section-tab-current", activeTab === "urgent");
      if (archivedTabNode) {
        archivedTabNode.classList.toggle("section-tab-current", activeTab === "archived");
      }
      tabsNode.hidden = allItems.length === 0 && archivedItems.length === 0;

      if (activeTab === "archived" && archivedLoading) {
        listNode.hidden = true;
        window.appListPage.setStatus(statusNode, "Loading…");
        return;
      }

      if (visibleItems.length === 0) {
        const emptyLabel = activeTab === "urgent"
          ? "No urgent requests"
          : activeTab === "archived"
            ? "No archived requests"
            : "No requests";
        listNode.innerHTML = renderEmptyState(emptyLabel);
        listNode.hidden = false;
        statusNode.hidden = true;
        return;
      }

      window.appListPage.renderList(listNode, visibleItems, (item) => {
        return window.appRequestItem.renderListItem(item, config.itemHref(item.id), {
          emailAlertBadge: config.emailAlertBadge === true,
        });
      });
      listNode.hidden = false;
      statusNode.hidden = true;
    }

    allTabNode.addEventListener("click", () => {
      activeTab = "all";
      render();
    });

    urgentTabNode.addEventListener("click", () => {
      activeTab = "urgent";
      render();
    });

    if (archivedTabNode) {
      archivedTabNode.addEventListener("click", () => {
        activeTab = "archived";
        if (!archivedLoaded) {
          loadArchivedItems();
        }
        render();
      });
    }

    if (searchNode) {
      const onSearchInput = debounce(() => {
        searchTerm = searchNode.value.trim().toLowerCase();
        render();
      }, 300);
      searchNode.addEventListener("input", onSearchInput);
    }

    window.appAuth.apiFetch(config.loadURL).then((payload) => {
      allItems = Array.isArray(payload.items) ? payload.items : [];
      render();
    }).catch(() => {
      window.appListPage.setStatus(statusNode, config.errorMessage || "Requests unavailable.");
    });
  }

  return { init };
})();
