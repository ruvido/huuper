(() => {
  if (!window.huuperRequestAssignment) {
    return;
  }

  window.huuperRequestAssignment.init({
    mode: "guardian",
    statusSelector: "#request-assignment-status",
    listSelector: "#request-assignment-list",
    detailURL: (id) => `/api/admin/requests/${encodeURIComponent(id)}`,
    actionURL: (id) => `/api/admin/requests/${encodeURIComponent(id)}/actions`,
    requestsURL: "/admin/requests/",
  });
})();
