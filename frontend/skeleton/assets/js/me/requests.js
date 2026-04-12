(() => {
  if (!window.huuperRequestListPage) {
    return;
  }

  window.huuperRequestListPage.init({
    statusSelector: "#requests-status",
    tabsSelector: "#requests-tabs",
    allTabSelector: "#requests-tab-all",
    allCountSelector: "#requests-count-all",
    urgentTabSelector: "#requests-tab-urgent",
    urgentCountSelector: "#requests-count-urgent",
    listSelector: "#requests-list",
    loadURL: "/api/me/requests",
    requestsURL: "/me/requests/",
    roles: ["assistant", "guardian"],
    itemHref: (id) => `/me/request/?id=${encodeURIComponent(id)}`,
    detailURL: (id) => `/api/me/requests/${encodeURIComponent(id)}`,
    actionURL: (id) => `/api/me/requests/${encodeURIComponent(id)}/actions`,
    errorMessage: "Requests unavailable.",
  });
})();
