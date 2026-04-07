(() => {
  if (!window.huuperRequestDetail) {
    return;
  }

  window.huuperRequestDetail.init({
    detailURL: (id) => `/api/me/requests/${encodeURIComponent(id)}`,
    actionURL: (id) => `/api/me/requests/${encodeURIComponent(id)}/actions`,
    requestsURL: "/me/requests/",
    assignGroupURL: (id) => `/me/request/assign-group/?id=${encodeURIComponent(id)}`,
    assignGuardianURL: (id) => `/me/request/assign-guardian/?id=${encodeURIComponent(id)}`,
  });
})();
