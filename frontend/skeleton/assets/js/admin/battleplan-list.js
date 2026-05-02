(() => {
  if (!window.huuperEntityList || !window.huuperAuth || !window.huuperListPage) {
    return;
  }

  (() => {
    const fab = document.querySelector('[data-battleplan-new]');
    if (!fab) return;
    const aria = window.huuperCopy
      && window.huuperCopy.battleplan
      && window.huuperCopy.battleplan.fab
      && window.huuperCopy.battleplan.fab.newAria;
    if (aria) fab.setAttribute("aria-label", aria);
  })();

  function showPlaceholder() {
    const placeholder = document.getElementById("battleplan-placeholder");
    if (!placeholder) return;
    const labelEl = document.getElementById("battleplan-placeholder-label");
    const labelText = window.huuperCopy
      && window.huuperCopy.battleplan
      && window.huuperCopy.battleplan.placeholder
      && window.huuperCopy.battleplan.placeholder.label;
    if (labelEl && labelText) labelEl.textContent = labelText;
    placeholder.hidden = false;
    const fab = document.querySelector('[data-battleplan-new]');
    if (fab) fab.hidden = true;
  }

  function isAdminUser() {
    try {
      const raw = localStorage.getItem("huuper.auth");
      const auth = raw ? JSON.parse(raw) : null;
      return !!(auth && auth.model && auth.model.admin === true);
    } catch (_) {
      return false;
    }
  }

  (async () => {
    try {
      const settings = await window.huuperAuth.apiFetch("/api/me/settings/battleplan");
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
    draft: "Draft",
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
    const statusRank = { active: 0, draft: 1, completed: 2, archived: 3 };
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

  function basePath() {
    return window.location.pathname.startsWith("/me/") ? "/me/battleplan" : "/admin/battleplan";
  }

  function findDraft(items) {
    return (items || []).find((it) => it && it.status === "draft") || null;
  }

  function confirmOverwriteDraft(draft, onConfirm) {
    const sheet = window.huuperActionSheet;
    const dlg = ((window.huuperCopy || {}).battleplan || {}).list || {};
    const cfg = dlg.overwriteDraftDialog || {};
    if (!sheet || typeof sheet.open !== "function") return;
    sheet.open({
      title: cfg.title || "Overwrite existing draft?",
      actions: [],
      footerAction: {
        label: cfg.confirmLabel || "Yes",
        tone: "danger",
        onSelect: async () => {
          await window.huuperAuth.apiFetch(`/api/me/battleplans/${encodeURIComponent(draft.id)}`, { method: "DELETE" });
          await onConfirm();
        },
      },
    });
  }

  function bindNewPlanTrigger(trigger, draft) {
    if (!trigger) return;
    trigger.onclick = (event) => {
      const go = () => { window.location.href = `${basePath()}/new/`; };
      if (draft) {
        event.preventDefault();
        confirmOverwriteDraft(draft, go);
        return;
      }
      if (trigger.tagName !== "A") {
        event.preventDefault();
        go();
      }
    };
  }

  function wireNewButton(items) {
    const active = activeBattleplan(items);
    const draft = findDraft(items);
    bindNewPlanTrigger(document.querySelector('[data-battleplan-new]'), draft);
    bindNewPlanTrigger(document.getElementById("battleplan-active-new"), draft);

    const dupBtn = document.getElementById("battleplan-active-duplicate");
    if (dupBtn) {
      dupBtn.onclick = async (event) => {
        event.preventDefault();
        if (!active) return;
        const createDraft = async () => {
          const payload = {
            start_date: (active.start_date || "").slice(0, 10),
            duration_days: durationDays(active),
            visibility: active.visibility,
            status: "draft",
            data: active.data || {},
          };
          const created = await window.huuperAuth.apiFetch("/api/me/battleplans", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(payload),
          });
          if (created && created.id) {
            window.location.href = `${basePath()}/edit/${encodeURIComponent(created.id)}/`;
          }
        };
        if (draft) {
          confirmOverwriteDraft(draft, createDraft);
        } else {
          await createDraft();
        }
      };
    }
  }

  function arrangeSections(listNode, items) {
    if (!listNode) return;
    const copy = (window.huuperCopy && window.huuperCopy.battleplan && window.huuperCopy.battleplan.list) || {};
    const activeSection = document.getElementById("battleplan-active-section");
    const draftSection = document.getElementById("battleplan-draft-section");
    const archiveSection = document.getElementById("battleplan-archive-section");
    const activeContainer = document.getElementById("battleplan-active");
    const draftContainer = document.getElementById("battleplan-draft");
    const activeLabel = document.getElementById("battleplan-active-label");
    const draftLabel = document.getElementById("battleplan-draft-label");
    const archiveLabel = document.getElementById("battleplan-archive-label");
    const activeActions = document.getElementById("battleplan-active-actions");
    const dupBtn = document.getElementById("battleplan-active-duplicate");
    const newBtn = document.getElementById("battleplan-active-new");

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

    const draftEls = Array.from(listNode.querySelectorAll(".battleplan-list-item-draft"));
    if (draftContainer) {
      draftContainer.innerHTML = "";
      draftEls.forEach((el) => draftContainer.appendChild(el));
      draftContainer.hidden = draftEls.length === 0;
    }

    const hasActive = items.some((it) => it.status === "active");
    const hasDraft = items.some((it) => it.status === "draft");
    const hasOther = items.some((it) => it.status !== "active" && it.status !== "draft");

    if (activeLabel) {
      activeLabel.textContent = copy.activeSectionLabel || "";
      activeLabel.hidden = !(hasActive && copy.activeSectionLabel);
    }
    if (draftLabel) {
      draftLabel.textContent = copy.draftSectionLabel || "";
      draftLabel.hidden = !(hasDraft && copy.draftSectionLabel);
    }
    if (archiveLabel) {
      archiveLabel.textContent = copy.archiveSectionLabel || "";
      archiveLabel.hidden = !(hasOther && copy.archiveSectionLabel);
    }
    if (dupBtn) dupBtn.textContent = copy.duplicateLabel || "";
    if (newBtn) newBtn.textContent = copy.newPlanLabel || "";
    if (activeActions) activeActions.hidden = !hasActive;
    if (activeSection) activeSection.hidden = !hasActive;
    if (draftSection) draftSection.hidden = !hasDraft;
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

  (async () => {
    let access = false;
    try {
      const payload = await window.huuperAuth.apiFetch("/api/me/access/battleplan");
      access = !!(payload && payload.access === true);
    } catch (_) {
      access = false;
    }
    if (!access) {
      showPlaceholder();
      return;
    }
    initList();
  })();

  function initList() {
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
      const href = `${basePath()}/view/${encodeURIComponent(item.id)}/`;
      const status = STATUS_LABELS[item.status] || item.status || "";
      const meta = metaHTML(item, lp.escapeHTML);
      const activeArrow = `<svg class="battleplan-list-arrow" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32" aria-hidden="true"><rect x="0" y="0" width="32" height="32" rx="6" fill="#0c0c0c"/><g fill="none" stroke="#f2cc0d" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><line x1="9" y1="16" x2="22" y2="16"/><polyline points="17,11 22,16 17,21"/></g></svg>`;
      let sideContent = "";
      if (item.status === "active") {
        sideContent = activeArrow;
      } else if (item.status !== "draft" && status) {
        sideContent = `<span class="list-item-side-title request-item-status">${lp.escapeHTML(status)}</span>`;
      }
      const side = sideContent ? `<span class="list-item-side">${sideContent}</span>` : "";
      const activeClass = item.status === "active" ? " battleplan-list-item-active" : "";
      const draftClass = item.status === "draft" ? " battleplan-list-item-draft" : "";
      const archivedClass = item.status === "archived" ? " battleplan-list-item-archived" : "";
      const inner = `
          <span class="list-item-copy request-item-copy">
            <strong>${lp.escapeHTML(title)}</strong>
            <span class="list-item-meta request-item-meta">${meta}</span>
          </span>
          ${side}`;
      const className = `list-item request-item battleplan-list-item${activeClass}${draftClass}${archivedClass}`;
      return `
        <a href="${lp.escapeHTML(href)}" class="${className}">${inner}
        </a>
      `;
    },
    afterRender: (node, items) => {
      wireNewButton(items);
      arrangeSections(node, items);
    },
    });
  }
})();
