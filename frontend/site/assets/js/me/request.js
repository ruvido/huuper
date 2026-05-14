(() => {
  if (!window.appRequestDetail) {
    return;
  }

  window.appRequestDetail.init({
    detailURL: (id) => `/api/me/requests/${encodeURIComponent(id)}`,
    actionURL: (id) => `/api/me/requests/${encodeURIComponent(id)}/actions`,
    requestsURL: "/me/requests/",
    assignGroupURL: (id) => `/me/request/assign-group/?id=${encodeURIComponent(id)}`,
    assignGuardianURL: (id) => `/me/request/assign-guardian/?id=${encodeURIComponent(id)}`,
    inlineGroupApprovalCancel: true,
  });
})();
