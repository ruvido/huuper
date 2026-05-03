(() => {
  if (!window.appGroupDetail) {
    return;
  }
  window.appGroupDetail.init({
    scope: "admin",
    detailURL: (id) => `/api/admin/groups/${encodeURIComponent(id)}`,
    manageAssistant: true,
    assistantURL: (id) => `/api/admin/groups/${encodeURIComponent(id)}/assistant`,
  });
})();
