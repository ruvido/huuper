(() => {
  if (!window.appGroupDetail) {
    return;
  }
  window.appGroupDetail.init({
    scope: "me",
    detailURL: (id) => `/api/me/groups/${encodeURIComponent(id)}`,
    assistantURL: (id) => `/api/me/groups/${encodeURIComponent(id)}/assistant`,
  });
})();
