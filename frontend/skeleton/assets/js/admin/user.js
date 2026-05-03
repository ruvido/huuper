(() => {
  if (!window.appUserDetail) {
    return;
  }
  window.appUserDetail.init({
    scope: "admin",
    detailURL: (id) => `/api/admin/users/${encodeURIComponent(id)}`,
    cancelURL: (id) => `/api/admin/users/${encodeURIComponent(id)}/cancel`,
    includeStatus: true,
    includeAdminFlag: true,
    includeGroups: true,
  });
})();
