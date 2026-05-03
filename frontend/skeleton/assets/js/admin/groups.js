(() => {
  if (!window.appEntityList || !window.appAuth || !window.appListPage) {
    return;
  }

  window.appEntityList.init({
    statusSelector: "#admin-groups-status",
    listSelector: "#admin-groups-list",
    requiresAuth: true,
    emptyMessage: "No groups.",
    errorMessage: "Groups unavailable.",
    load: () => window.appAuth.apiFetch("/api/admin/groups"),
    renderItem: (item) => {
      const meta = window.appListPage.text(item.type);
      const href = `/admin/group/?id=${encodeURIComponent(item.id)}`;
      const sideHTML = window.appGroupMeta && window.appGroupMeta.assistantMissing(item)
        ? window.appGroupMeta.assistantWarningBadge()
        : "";
      return window.appListPage.renderListItemLink(href, item.name || item.id, meta, { sideHTML });
    },
  });
})();
