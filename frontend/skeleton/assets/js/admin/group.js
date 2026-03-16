(() => {
  if (!window.huuperGroupDetail) {
    return;
  }
  window.huuperGroupDetail.init({
    scope: "admin",
    detailURL: (id) => `/api/admin/groups/${encodeURIComponent(id)}`,
  });
})();
