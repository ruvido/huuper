(() => {
  if (!window.huuperRequestDetail) {
    return;
  }

  window.huuperRequestDetail.init({
    detailURL: (id) => `/api/admin/requests/${encodeURIComponent(id)}`,
    actionURL: (id) => `/api/admin/requests/${encodeURIComponent(id)}/actions`,
  });
})();
