(() => {
  if (!window.huuperUserDetail) {
    return;
  }
  window.huuperUserDetail.init({
    scope: "admin",
    detailURL: (id) => `/api/admin/users/${encodeURIComponent(id)}`,
    cancelURL: (id) => `/api/admin/users/${encodeURIComponent(id)}/cancel`,
    includeStatus: true,
    includeAdminFlag: true,
    includeGroups: true,
  });
})();
