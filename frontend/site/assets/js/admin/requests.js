(() => {
  if (!window.huuperRequestListPage) {
    return;
  }

  window.huuperRequestListPage.init({
    statusSelector: "#admin-requests-status",
    tabsSelector: "#admin-requests-tabs",
    allTabSelector: "#admin-requests-tab-all",
    allCountSelector: "#admin-requests-count-all",
    urgentTabSelector: "#admin-requests-tab-urgent",
    urgentCountSelector: "#admin-requests-count-urgent",
    listSelector: "#admin-requests-list",
    loadURL: "/api/admin/requests",
    requestsURL: "/admin/requests/",
    roles: ["admin", "assistant", "guardian"],
    itemHref: (id) => `/admin/request/?id=${encodeURIComponent(id)}`,
    detailURL: (id) => `/api/admin/requests/${encodeURIComponent(id)}`,
    actionURL: (id) => `/api/admin/requests/${encodeURIComponent(id)}/actions`,
    errorMessage: "Requests unavailable.",
  });
})();
