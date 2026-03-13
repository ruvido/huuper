(() => {
  if (!window.huuperRequestDetail) {
    return;
  }

  window.huuperRequestDetail.init({
    detailURL: (id) => `/api/me/requests/${encodeURIComponent(id)}`,
    actionURL: (id) => `/api/me/requests/${encodeURIComponent(id)}/actions`,
  });
})();
