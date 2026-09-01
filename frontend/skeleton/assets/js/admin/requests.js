(() => {
  if (!window.appRequestListPage) {
    return;
  }

  window.appRequestListPage.init({
    statusSelector: "#admin-requests-status",
    searchSelector: "#admin-requests-search",
    tabsSelector: "#admin-requests-tabs",
    allTabSelector: "#admin-requests-tab-all",
    allCountSelector: "#admin-requests-count-all",
    urgentTabSelector: "#admin-requests-tab-urgent",
    urgentCountSelector: "#admin-requests-count-urgent",
    archivedTabSelector: "#admin-requests-tab-archived",
    archivedCountSelector: "#admin-requests-count-archived",
    listSelector: "#admin-requests-list",
    loadURL: "/api/admin/requests",
    requestsURL: "/admin/requests/",
    itemHref: (id) => `/admin/request/?id=${encodeURIComponent(id)}`,
    detailURL: (id) => `/api/admin/requests/${encodeURIComponent(id)}`,
    actionURL: (id) => `/api/admin/requests/${encodeURIComponent(id)}/actions`,
    emailAlertBadge: true,
    errorMessage: "Requests unavailable.",
  });
})();
