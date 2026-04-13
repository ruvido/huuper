(() => {
  if (!window.huuperGroupDetail) {
    return;
  }
  window.huuperGroupDetail.init({
    scope: "me",
    detailURL: (id) => `/api/me/groups/${encodeURIComponent(id)}`,
    assistantURL: (id) => `/api/me/groups/${encodeURIComponent(id)}/assistant`,
  });
})();
