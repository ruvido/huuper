(() => {
  if (!window.appAuth || !window.appListPage) {
    return;
  }

  const statusNode = document.querySelector("#admin-users-status");
  const tabsNode = document.querySelector("#admin-users-tabs");
  const approvedTabNode = document.querySelector("#admin-users-tab-approved");
  const cancelledTabNode = document.querySelector("#admin-users-tab-cancelled");
  const approvedCountNode = document.querySelector("#admin-users-count-approved");
  const cancelledCountNode = document.querySelector("#admin-users-count-cancelled");
  const listNode = document.querySelector("#admin-users-list");
  const filterBarNode = document.querySelector("#admin-users-filter");

  if (!statusNode || !tabsNode || !approvedTabNode || !cancelledTabNode || !approvedCountNode || !cancelledCountNode || !listNode) {
    return;
  }

  let activeTab = "approved";
  let allItems = [];

  const params = new URLSearchParams(window.location.search);
  const filters = {
    region: params.get("region") || "",
    age: params.get("age") || "",
    marital: params.get("marital") || "",
    work: params.get("work") || "",
    sport: params.get("sport") || "",
    interest: params.get("interest") || "",
    skill: params.get("skill") || "",
  };

  const FILTER_LABELS = {
    region: "Regione",
    age: "Età",
    marital: "Stato",
    work: "Settore",
    sport: "Sport",
    interest: "Interesse",
    skill: "Skill",
  };

  function ageInBucket(birthYear, bucket) {
    const year = Number(birthYear);
    if (!year) return false;
    const age = new Date().getFullYear() - year;
    if (bucket === "nd") return false;
    if (bucket === "55+") return age >= 55;
    const [lo, hi] = bucket.split("-").map(Number);
    return age >= lo && age <= hi;
  }

  function arrayContains(raw, needle) {
    if (!Array.isArray(raw)) return false;
    const target = needle.trim().toLowerCase();
    for (const item of raw) {
      if (typeof item !== "string") continue;
      const parts = item.split(/[\n,;/]+/).map((p) => p.trim().toLowerCase());
      if (parts.includes(target)) return true;
    }
    return false;
  }

  function matchesFilters(user) {
    const data = user && user.data ? user.data : {};
    if (filters.region) {
      if (filters.region === "nd") {
        if ((data.region || "").trim()) return false;
      } else if ((data.region || "").trim() !== filters.region) {
        return false;
      }
    }
    if (filters.age) {
      if (filters.age === "nd") {
        if (data.birth_year) return false;
      } else if (!ageInBucket(data.birth_year, filters.age)) {
        return false;
      }
    }
    if (filters.marital) {
      if (filters.marital === "nd") {
        if ((data.marital_status || "").trim()) return false;
      } else if ((data.marital_status || "").trim() !== filters.marital) {
        return false;
      }
    }
    if (filters.work) {
      if (filters.work === "nd") {
        if ((data.work || "").trim()) return false;
      } else if ((data.work || "").trim() !== filters.work) {
        return false;
      }
    }
    if (filters.sport && !arrayContains(data.sports, filters.sport)) return false;
    if (filters.interest && !arrayContains(data.interests, filters.interest)) return false;
    if (filters.skill && !arrayContains(data.skills, filters.skill)) return false;
    return true;
  }

  function activeFilterChips() {
    return Object.entries(filters)
      .filter(([, v]) => v)
      .map(([k, v]) => ({ key: k, label: FILTER_LABELS[k] || k, value: v }));
  }

  function renderFilterBar() {
    if (!filterBarNode) return;
    const chips = activeFilterChips();
    if (chips.length === 0) {
      filterBarNode.hidden = true;
      filterBarNode.innerHTML = "";
      return;
    }
    filterBarNode.hidden = false;
    filterBarNode.innerHTML =
      chips
        .map(
          (c) => `<span class="filter-chip">
        <span class="filter-chip-label">${window.appListPage.escapeHTML(c.label)}</span>
        <span class="filter-chip-value">${window.appListPage.escapeHTML(c.value)}</span>
      </span>`
        )
        .join("") +
      `<a class="filter-chip-clear" href="/admin/users/">Reset</a>`;
  }

  function renderItem(item) {
    const href = `/admin/user/?id=${encodeURIComponent(item.id)}`;
    if (window.appListItem && window.appListItem.renderMember) {
      return window.appListItem.renderMember(item, href);
    }
    const status = window.appListPage.text(item.status);
    const role = item && item.admin === true ? "admin" : "";
    const meta = [status, role].filter(Boolean).join(" • ");
    const title = window.appListPage.text(item.full_name) || window.appListPage.text(item.email) || item.id;
    return window.appListPage.renderListItemLink(href, title, meta);
  }

  function render() {
    const filtered = allItems.filter(matchesFilters);
    const approved = filtered.filter((u) => window.appListPage.text(u.status) !== "cancelled");
    const cancelled = filtered.filter((u) => window.appListPage.text(u.status) === "cancelled");

    approvedCountNode.textContent = String(approved.length);
    cancelledCountNode.textContent = String(cancelled.length);
    approvedTabNode.classList.toggle("section-tab-current", activeTab === "approved");
    cancelledTabNode.classList.toggle("section-tab-current", activeTab === "cancelled");

    tabsNode.hidden = filtered.length === 0;

    const visibleItems = activeTab === "cancelled" ? cancelled : approved;

    if (visibleItems.length === 0) {
      listNode.hidden = true;
      const emptyMessage = activeTab === "cancelled" ? "No cancelled users." : "No users.";
      window.appListPage.setStatus(statusNode, emptyMessage);
      return;
    }

    window.appListPage.renderList(listNode, visibleItems, renderItem);
    listNode.hidden = false;
    window.appListPage.setStatus(statusNode, "");
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
    window.appAuth.apiFetch("/api/admin/users").then((payload) => {
      allItems = Array.isArray(payload.items) ? payload.items : [];
      renderFilterBar();
      render();
    }).catch(() => {
      window.appListPage.setStatus(statusNode, "Users unavailable.");
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
