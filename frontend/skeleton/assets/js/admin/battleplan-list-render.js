// Render helpers for the battleplan list page. Pure functions: take an item
// and a copy/escape context, return HTML strings. No DOM mutation, no state.
// Published onto window.appBattleplanList so the orchestrator and actions
// modules can read from the same shared namespace.
(() => {
  window.appBattleplanList = window.appBattleplanList || {};
  const ns = window.appBattleplanList;

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

  function itemTime(item) {
    const raw = item.start_date || item.created || "";
    const parsed = new Date(raw);
    return Number.isNaN(parsed.getTime()) ? 0 : parsed.getTime();
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

  // The yellow→black arrow badge shown only for the active plan. Inlined SVG
  // (no shared icon partial) to avoid a templating round-trip per render.
  function renderActiveArrowSVG() {
    return `<svg class="battleplan-list-arrow" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32" aria-hidden="true"><rect x="0" y="0" width="32" height="32" rx="6" fill="#0c0c0c"/><g fill="none" stroke="#f2cc0d" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><line x1="9" y1="16" x2="22" y2="16"/><polyline points="17,11 22,16 17,21"/></g></svg>`;
  }

  // ctx: { copy, statusLabels, esc, escapeHTML, basePath, titleFallback }
  function renderItem(item, ctx) {
    const { statusLabels, esc, basePath, titleFallback } = ctx;
    const title = (item.data && item.data.priority && item.data.priority.title) || titleFallback;
    const href = `${basePath}/view/${encodeURIComponent(item.id)}/`;
    const status = statusLabels[item.status] || item.status || "";
    const meta = metaHTML(item, esc);
    let sideContent = "";
    if (item.status === "active") {
      sideContent = renderActiveArrowSVG();
    } else if (item.status !== "draft" && status) {
      sideContent = `<span class="list-item-side-title request-item-status">${esc(status)}</span>`;
    }
    const side = sideContent ? `<span class="list-item-side">${sideContent}</span>` : "";
    const activeClass = item.status === "active" ? " battleplan-list-item-active" : "";
    const draftClass = item.status === "draft" ? " battleplan-list-item-draft" : "";
    const archivedClass = item.status === "archived" ? " battleplan-list-item-archived" : "";
    const inner = `
        <span class="list-item-copy request-item-copy">
          <strong>${esc(title)}</strong>
          <span class="list-item-meta request-item-meta">${meta}</span>
        </span>
        ${side}`;
    const className = `list-item request-item battleplan-list-item${activeClass}${draftClass}${archivedClass}`;
    return `
      <a href="${esc(href)}" class="${className}">${inner}
      </a>
    `;
  }

  // copy: full appCopy.battleplan.list object — reads .emptyLabel and
  // .sectionAriaLabel for the empty-state markup.
  function renderEmpty(copy) {
    const ariaLabel = (copy && copy.sectionAriaLabel) || "";
    const emptyLabel = (copy && copy.emptyLabel) || "";
    return `
      <section class="empty-state empty-state-icon-only" aria-label="${ariaLabel}">
        <svg class="empty-state-icon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor" aria-hidden="true"><path d="M8 15A7 7 0 1 1 8 1a7 7 0 0 1 0 14m0 1A8 8 0 1 0 8 0a8 8 0 0 0 0 16"/><path d="M8 13A5 5 0 1 1 8 3a5 5 0 0 1 0 10m0 1A6 6 0 1 0 8 2a6 6 0 0 0 0 12"/><path d="M8 11a3 3 0 1 1 0-6 3 3 0 0 1 0 6m0 1a4 4 0 1 0 0-8 4 4 0 0 0 0 8"/><path d="M9.5 8a1.5 1.5 0 1 1-3 0 1.5 1.5 0 0 1 3 0"/></svg>
        <p class="empty-state-label">${emptyLabel}</p>
      </section>
    `;
  }

  ns.render = {
    durationDays,
    shortDate,
    itemTime,
    metaHTML,
    renderActiveArrowSVG,
    renderItem,
    renderEmpty,
  };
})();
