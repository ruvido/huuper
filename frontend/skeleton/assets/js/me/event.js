(() => {
  if (!window.huuperEventDetail) {
    return;
  }
  window.huuperEventDetail.init({
    detailURL: (id) => `/api/me/events/${encodeURIComponent(id)}`,
    includeRegistrationState: true,
    includeRegistrations: false,
  });
})();
