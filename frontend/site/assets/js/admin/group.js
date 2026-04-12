(() => {
  if (!window.huuperGroupDetail) {
    return;
  }
  window.huuperGroupDetail.init({
    scope: "admin",
    detailURL: (id) => `/api/admin/groups/${encodeURIComponent(id)}`,
    manageAssistant: true,
    assistantURL: (id) => `/api/admin/groups/${encodeURIComponent(id)}/assistant`,
    assistantRedirectURL: "/admin/groups/",
  });
})();
