// Pure render helpers for the /admin/events list page. Mirrors me/events-render
// — kept separate so the two scopes can diverge (admin may grow extra columns
// later) without coupling. Cadence math lives in window.appEventsCadence.
(() => {
  if (!window.appEventsCadence) return;
  window.appAdminEventsList = window.appAdminEventsList || {};
  const ns = window.appAdminEventsList;
  const cad = window.appEventsCadence;

  function eventsCopy() {
    return (window.appCopy && window.appCopy.events) || {};
  }

  function typeLabel(type) {
    const labels = (eventsCopy().types) || {};
    return labels[type] || (type ? type.charAt(0).toUpperCase() + type.slice(1) : "");
  }

  function deriveTitle(item) {
    const fallback = (eventsCopy().list && eventsCopy().list.autoTitleFallback) || "Event";
    const explicit = (item.title || "").trim();
    if (explicit) return explicit;
    const tlabel = item.type_label || typeLabel(item.type);
    const group = (item.group_name || "").trim();
    if (tlabel && group) return `${tlabel} · ${group}`;
    if (tlabel) return tlabel;
    return fallback;
  }

  // ctx: { esc, basePath, spotlight?: boolean }
  function renderItem(item, ctx) {
    const { esc, basePath, spotlight } = ctx;
    const title = deriveTitle(item);
    const href = `${basePath}/event/?id=${encodeURIComponent(item.id)}`;
    const meta = cad.metaText(item);
    const tlabel = item.type_label || typeLabel(item.type);
    const side = tlabel
      ? `<span class="list-item-side"><span class="list-item-side-title">${esc(tlabel)}</span></span>`
      : "";
    const cls = `list-item events-list-item${spotlight ? " list-item-spotlight" : ""}`;
    return `
      <a href="${esc(href)}" class="${cls}">
        <span class="list-item-copy">
          <strong>${esc(title)}</strong>
          ${meta ? `<span class="list-item-meta">${esc(meta)}</span>` : ""}
        </span>
        ${side}
      </a>
    `;
  }

  function renderEmpty() {
    const c = eventsCopy().list || {};
    return `
      <section class="empty-state empty-state-icon-only" aria-label="${c.sectionAriaLabel || "Events"}">
        <p class="empty-state-label">${c.emptyLabel || ""}</p>
      </section>
    `;
  }

  function renderWindowTabs(currentWindow) {
    const c = eventsCopy().list || {};
    const upcoming = c.futureLabel || "Upcoming";
    const past = c.pastLabel || "Past";
    const isFuture = currentWindow !== "past";
    return `
      <button id="events-tab-future" class="section-tab${isFuture ? " section-tab-current" : ""}" type="button" data-window="future">
        <span class="section-tab-text">${upcoming}</span>
      </button>
      <button id="events-tab-past" class="section-tab${!isFuture ? " section-tab-current" : ""}" type="button" data-window="past">
        <span class="section-tab-text">${past}</span>
      </button>
    `;
  }

  ns.render = {
    typeLabel,
    deriveTitle,
    renderItem,
    renderEmpty,
    renderWindowTabs,
  };
})();
