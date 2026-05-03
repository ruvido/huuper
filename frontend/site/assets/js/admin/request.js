(() => {
  if (!window.appRequestDetail) {
    return;
  }

  window.appRequestDetail.init({
    detailURL: (id) => `/api/admin/requests/${encodeURIComponent(id)}`,
    actionURL: (id) => `/api/admin/requests/${encodeURIComponent(id)}/actions`,
    requestsURL: "/admin/requests/",
    assignGroupURL: (id) => `/admin/request/assign-group/?id=${encodeURIComponent(id)}`,
    assignGuardianURL: (id) => `/admin/request/assign-guardian/?id=${encodeURIComponent(id)}`,
  });
})();
