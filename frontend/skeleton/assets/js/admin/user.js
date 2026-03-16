(() => {
  if (!window.huuperUserDetail) {
    return;
  }
  window.huuperUserDetail.init({
    scope: "admin",
    detailURL: (id) => `/api/admin/users/${encodeURIComponent(id)}`,
    includeStatus: true,
    includeAdminFlag: true,
    includeGroups: true,
  });
})();
