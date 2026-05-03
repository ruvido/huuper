(() => {
  if (!window.appEventDetail) {
    return;
  }
  window.appEventDetail.init({
    detailURL: (id) => `/api/me/events/${encodeURIComponent(id)}`,
    includeRegistrationState: true,
    includeRegistrations: true,
    scope: "me",
    canManageRegistrations: false,
    showPending: false,
    showCancelled: false,
  });
})();
