(() => {
  if (!window.appRequestAssignment) {
    return;
  }

  const mode = window.location.pathname.includes("/assign-guardian/") ? "guardian" : "group";
  const scope = window.location.pathname.startsWith("/admin/") ? "admin" : "me";
  const base = scope === "admin" ? "/api/admin/requests" : "/api/me/requests";
  const requestsURL = scope === "admin" ? "/admin/requests/" : "/me/requests/";

  window.appRequestAssignment.init({
    mode,
    statusSelector: "#request-assignment-status",
    listSelector: "#request-assignment-list",
    detailURL: (id) => `${base}/${encodeURIComponent(id)}`,
    actionURL: (id) => `${base}/${encodeURIComponent(id)}/actions`,
    requestsURL,
  });
})();
