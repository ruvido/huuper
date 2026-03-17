(() => {
  if (!window.huuperEventDetail) {
    return;
  }
  window.huuperEventDetail.init({
    detailURL: (id) => `/api/me/events/${encodeURIComponent(id)}`,
    includeRegistrationState: true,
    includeRegistrations: true,
    scope: "me",
    canManageRegistrations: false,
    showPending: false,
    showCancelled: false,
  });
})();
