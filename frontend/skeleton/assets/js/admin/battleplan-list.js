(() => {
  if (!window.huuperEntityList || !window.huuperAuth || !window.huuperListPage) {
    return;
  }

  (async () => {
    try {
      const settings = await window.huuperAuth.apiFetch("/api/admin/settings/battleplan");
      const title = settings && settings.data && settings.data.title;
      if (!title) return;
      const hero = document.getElementById("battleplan-hero");
      const target = document.getElementById("battleplan-hero-title");
      if (!hero || !target) return;
      target.innerHTML = title;
      hero.hidden = false;
    } catch (_) {
      /* hero is optional */
    }
  })();

  const STATUS_LABELS = {
    active: "Active",
    completed: "Completed",
    archived: "Archived",
  };

  function durationDays(item) {
    if (!item.start_date || !item.end_date) return 0;
    const start = new Date(item.start_date);
    const end = new Date(item.end_date);
    if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime())) return 0;
    return Math.round((end - start) / (1000 * 60 * 60 * 24));
  }

  function shortDate(str) {
    if (!str) return "";
    const d = new Date(str);
    if (isNaN(d)) return "";
    const months = ["Jan","Feb","Mar","Apr","May","Jun","Jul","Aug","Sep","Oct","Nov","Dec"];
    return `${d.getDate()} ${months[d.getMonth()]} ${d.getFullYear()}`;
  }

  function metaHTML(item, esc) {
    const days = durationDays(item);
    const vis = item.visibility || "";
    const start = shortDate(item.start_date);
    const closeDate = item.status === "archived" ? shortDate(item.updated) : shortDate(item.end_date);
    const range = (start && closeDate) ? `${esc(start)} → ${esc(closeDate)}` : esc(start || closeDate);
    const tags = [
      days ? `<label class="field-label wizard-summary-tag wizard-summary-tag-outline">${days}D</label>` : "",
      vis ? `<label class="field-label wizard-summary-tag wizard-summary-tag-outline">${esc(vis.toUpperCase())}</label>` : "",
    ].filter(Boolean).join("");
    return `<span class="battleplan-list-meta">${range ? `<span class="battleplan-list-date">${range}</span>` : ""}${tags ? `<span class="battleplan-list-tags">${tags}</span>` : ""}</span>`;
  }

  function itemTime(item) {
    const raw = item.start_date || item.created || "";
    const parsed = new Date(raw);
    return Number.isNaN(parsed.getTime()) ? 0 : parsed.getTime();
  }

  function orderedItems(items) {
    const statusRank = { active: 0, completed: 1, archived: 2 };
    return [...items].sort((a, b) => {
      const ar = Object.prototype.hasOwnProperty.call(statusRank, a.status) ? statusRank[a.status] : 3;
      const br = Object.prototype.hasOwnProperty.call(statusRank, b.status) ? statusRank[b.status] : 3;
      if (ar !== br) return ar - br;
      return itemTime(b) - itemTime(a);
    });
  }

  function activeBattleplan(items) {
    return items.find((item) => item && item.status === "active") || null;
  }

  function wireNewButton(items) {
    const trigger = document.querySelector('[data-battleplan-new]');
    if (!trigger) return;
    const active = activeBattleplan(items);
    trigger.onclick = (event) => {
      if (!active) return;
      event.preventDefault();
      const sheet = window.huuperActionSheet;
      if (!sheet || typeof sheet.open !== "function") return;
      sheet.open({
        title: "Vuoi abbandonare il tuo attuale battleplan?",
        actions: [],
        footerAction: {
          label: "YES",
          onSelect: async () => {
            await window.huuperAuth.apiFetch(`/api/me/battleplans/${encodeURIComponent(active.id)}/status`, {
              method: "POST",
              headers: { "Content-Type": "application/json" },
              body: JSON.stringify({ status: "archived" }),
            });
            window.location.href = "/admin/battleplan/new/";
          },
        },
      });
    };
  }

  function arrangeSections(listNode, items) {
    if (!listNode) return;
    const copy = (window.huuperCopy && window.huuperCopy.battleplan && window.huuperCopy.battleplan.list) || {};
    const activeSection = document.getElementById("battleplan-active-section");
    const archiveSection = document.getElementById("battleplan-archive-section");
    const activeContainer = document.getElementById("battleplan-active");
    const activeLabel = document.getElementById("battleplan-active-label");
    const archiveLabel = document.getElementById("battleplan-archive-label");

    const activeEl = listNode.querySelector(".battleplan-list-item-active");
    if (activeContainer) {
      activeContainer.innerHTML = "";
      if (activeEl) {
        activeContainer.appendChild(activeEl);
        activeContainer.hidden = false;
      } else {
        activeContainer.hidden = true;
      }
    }

    const hasActive = items.some((it) => it.status === "active");
    const hasOther = items.some((it) => it.status !== "active");

    if (activeLabel) {
      activeLabel.textContent = copy.activeSectionLabel || "";
      activeLabel.hidden = !(hasActive && copy.activeSectionLabel);
    }
    if (archiveLabel) {
      archiveLabel.textContent = copy.archiveSectionLabel || "";
      archiveLabel.hidden = !(hasOther && copy.archiveSectionLabel);
    }
    if (activeSection) activeSection.hidden = !hasActive;
    if (archiveSection) archiveSection.hidden = !hasOther;
  }

  function renderEmpty() {
    return `
      <section class="empty-state empty-state-icon-only" aria-label="Battleplan">
        <svg class="empty-state-icon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor" aria-hidden="true"><path d="M8 15A7 7 0 1 1 8 1a7 7 0 0 1 0 14m0 1A8 8 0 1 0 8 0a8 8 0 0 0 0 16"/><path d="M8 13A5 5 0 1 1 8 3a5 5 0 0 1 0 10m0 1A6 6 0 1 0 8 2a6 6 0 0 0 0 12"/><path d="M8 11a3 3 0 1 1 0-6 3 3 0 0 1 0 6m0 1a4 4 0 1 0 0-8 4 4 0 0 0 0 8"/><path d="M9.5 8a1.5 1.5 0 1 1-3 0 1.5 1.5 0 0 1 3 0"/></svg>
        <p class="empty-state-label">nessun piano di battaglia</p>
      </section>
    `;
  }

  window.huuperEntityList.init({
    statusSelector: "#battleplan-status",
    listSelector: "#battleplan-list",
    requiresAuth: true,
    errorMessage: "Battleplan unavailable.",
    renderEmpty,
    load: async () => {
      const payload = await window.huuperAuth.apiFetch("/api/me/battleplans?per_page=200");
      const items = Array.isArray(payload.items) ? payload.items : [];
      return { ...payload, items: orderedItems(items) };
    },
    renderItem: (item) => {
      const lp = window.huuperListPage;
      const title = (item.data && item.data.priority && item.data.priority.title) || "Battleplan";
      const href = `/admin/battleplan/view/${encodeURIComponent(item.id)}/`;
      const status = STATUS_LABELS[item.status] || item.status || "";
      const meta = metaHTML(item, lp.escapeHTML);
      const activeArrow = `<svg class="battleplan-list-arrow" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32" aria-hidden="true"><rect x="0" y="0" width="32" height="32" rx="6" fill="#0c0c0c"/><g fill="none" stroke="#f2cc0d" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><line x1="9" y1="16" x2="22" y2="16"/><polyline points="17,11 22,16 17,21"/></g></svg>`;
      const sideContent = item.status === "active"
        ? activeArrow
        : (status ? `<span class="list-item-side-title request-item-status">${lp.escapeHTML(status)}</span>` : "");
      const side = sideContent ? `<span class="list-item-side">${sideContent}</span>` : "";
      const activeClass = item.status === "active" ? " battleplan-list-item-active" : "";
      const archivedClass = item.status === "archived" ? " battleplan-list-item-archived" : "";
      return `
        <a href="${lp.escapeHTML(href)}" class="list-item request-item battleplan-list-item${activeClass}${archivedClass}">
          <span class="list-item-copy request-item-copy">
            <strong>${lp.escapeHTML(title)}</strong>
            <span class="list-item-meta request-item-meta">${meta}</span>
          </span>
          ${side}
        </a>
      `;
    },
    afterRender: (node, items) => {
      wireNewButton(items);
      arrangeSections(node, items);
    },
  });
})();
