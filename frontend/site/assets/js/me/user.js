(() => {
  if (!window.huuperUserDetail) {
    return;
  }
  window.huuperUserDetail.init({
    scope: "me",
    detailURL: (id) => `/api/me/users/${encodeURIComponent(id)}`,
    includeStatus: false,
    includeAdminFlag: false,
    includeGroups: false,
  });
})();
