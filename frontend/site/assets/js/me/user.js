(() => {
  if (!window.appUserDetail) {
    return;
  }
  window.appUserDetail.init({
    scope: "me",
    detailURL: (id) => `/api/me/users/${encodeURIComponent(id)}`,
    includeStatus: false,
    includeAdminFlag: false,
    includeGroups: false,
  });
})();
