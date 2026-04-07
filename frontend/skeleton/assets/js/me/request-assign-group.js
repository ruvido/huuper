(() => {
  if (!window.huuperRequestAssignment) {
    return;
  }

  window.huuperRequestAssignment.init({
    mode: "group",
    statusSelector: "#request-assignment-status",
    listSelector: "#request-assignment-list",
    detailURL: (id) => `/api/me/requests/${encodeURIComponent(id)}`,
    actionURL: (id) => `/api/me/requests/${encodeURIComponent(id)}/actions`,
    requestsURL: "/me/requests/",
  });
})();
