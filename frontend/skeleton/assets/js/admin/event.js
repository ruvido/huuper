(() => {
  if (!window.huuperEventDetail) {
    return;
  }
  window.huuperEventDetail.init({
    detailURL: (id) => `/api/admin/events/${encodeURIComponent(id)}`,
    includeRegistrationState: false,
    includeRegistrations: true,
    scope: "admin",
    canManageRegistrations: true,
    showPending: true,
    showCancelled: true,
  });
})();
