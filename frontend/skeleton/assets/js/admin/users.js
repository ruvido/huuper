(() => {
  if (!window.huuperAuth || !window.huuperListPage) {
    return;
  }

  const statusNode = document.querySelector("#admin-users-status");
  const tabsNode = document.querySelector("#admin-users-tabs");
  const approvedTabNode = document.querySelector("#admin-users-tab-approved");
  const cancelledTabNode = document.querySelector("#admin-users-tab-cancelled");
  const approvedCountNode = document.querySelector("#admin-users-count-approved");
  const cancelledCountNode = document.querySelector("#admin-users-count-cancelled");
  const listNode = document.querySelector("#admin-users-list");

  if (!statusNode || !tabsNode || !approvedTabNode || !cancelledTabNode || !approvedCountNode || !cancelledCountNode || !listNode) {
    return;
  }

  let activeTab = "approved";
  let allItems = [];

  function renderItem(item) {
    const status = window.huuperListPage.text(item.status);
    const role = item && item.admin === true ? "admin" : "";
    const meta = [status, role].filter(Boolean).join(" • ");
    const href = `/admin/user/?id=${encodeURIComponent(item.id)}`;
    const title = window.huuperListPage.text(item.full_name) || window.huuperListPage.text(item.email) || item.id;
    return window.huuperListPage.renderListItemLink(href, title, meta);
  }

  function render() {
    const approved = allItems.filter((u) => window.huuperListPage.text(u.status) !== "cancelled");
    const cancelled = allItems.filter((u) => window.huuperListPage.text(u.status) === "cancelled");

    approvedCountNode.textContent = String(approved.length);
    cancelledCountNode.textContent = String(cancelled.length);
    approvedTabNode.classList.toggle("section-tab-current", activeTab === "approved");
    cancelledTabNode.classList.toggle("section-tab-current", activeTab === "cancelled");

    tabsNode.hidden = allItems.length === 0;

    const visibleItems = activeTab === "cancelled" ? cancelled : approved;

    if (visibleItems.length === 0) {
      listNode.hidden = true;
      const emptyMessage = activeTab === "cancelled" ? "No cancelled users." : "No users.";
      window.huuperListPage.setStatus(statusNode, emptyMessage);
      return;
    }

    window.huuperListPage.renderList(listNode, visibleItems, renderItem);
    listNode.hidden = false;
    window.huuperListPage.setStatus(statusNode, "");
  }

  approvedTabNode.addEventListener("click", () => {
    activeTab = "approved";
    render();
  });

  cancelledTabNode.addEventListener("click", () => {
    activeTab = "cancelled";
    render();
  });

  function load() {
    window.huuperAuth.apiFetch("/api/admin/users").then((payload) => {
      allItems = Array.isArray(payload.items) ? payload.items : [];
      render();
    }).catch(() => {
      window.huuperListPage.setStatus(statusNode, "Users unavailable.");
    });
  }

  window.addEventListener("pageshow", (event) => {
    const nav = performance.getEntriesByType && performance.getEntriesByType("navigation")[0];
    if (event.persisted || (nav && nav.type === "back_forward")) {
      load();
    }
  });

  load();
})();
