// Pure render helpers for the /me/events list page. Builds list-item HTML
// with type-badge + group-name + formatted date. Published onto
// window.appMeEventsList so the orchestrator can consume the same helpers.
(() => {
  window.appMeEventsList = window.appMeEventsList || {};
  const ns = window.appMeEventsList;

  const TYPE_TAG_TONE = {
    rally: "events-tag-rally",
    call: "events-tag-call",
    meetup: "events-tag-meetup",
  };

  function eventsCopy() {
    return (window.appCopy && window.appCopy.events) || {};
  }

  function typeLabel(type) {
    const c = eventsCopy();
    const labels = (c.types) || {};
    return labels[type] || (type ? type.charAt(0).toUpperCase() + type.slice(1) : "");
  }

  function shortDate(value) {
    if (!value) return "";
    const d = new Date(value);
    if (Number.isNaN(d.getTime())) return "";
    const months = ["Jan","Feb","Mar","Apr","May","Jun","Jul","Aug","Sep","Oct","Nov","Dec"];
    const hh = String(d.getHours()).padStart(2, "0");
    const mm = String(d.getMinutes()).padStart(2, "0");
    return `${d.getDate()} ${months[d.getMonth()]} ${d.getFullYear()} · ${hh}:${mm}`;
  }

  // Falls back to "<TypeLabel> · <GroupName>" when the event has no title
  // (call/meetup created without a title still need a meaningful headline).
  function deriveTitle(item) {
    const fallback = (eventsCopy().list && eventsCopy().list.autoTitleFallback) || "Event";
    const explicit = (item.title || "").trim();
    if (explicit) return explicit;
    const tlabel = typeLabel(item.type);
    const group = (item.group_name || "").trim();
    if (tlabel && group) return `${tlabel} · ${group}`;
    if (tlabel) return tlabel;
    return fallback;
  }

  function badgeHTML(type, esc) {
    const label = typeLabel(type);
    if (!label) return "";
    const tone = TYPE_TAG_TONE[type] || "events-tag-default";
    return `<label class="field-label wizard-summary-tag wizard-summary-tag-outline events-tag ${tone}">${esc(label)}</label>`;
  }

  function metaHTML(item, esc) {
    const date = shortDate(item.event_date);
    const group = (item.group_name || "").trim();
    const tags = badgeHTML(item.type, esc);
    const parts = [];
    if (date) parts.push(`<span class="events-list-date">${esc(date)}</span>`);
    if (group) parts.push(`<span class="events-list-group">${esc(group)}</span>`);
    return `<span class="events-list-meta">${parts.join(" · ")}${tags ? `<span class="events-list-tags">${tags}</span>` : ""}</span>`;
  }

  // ctx: { esc, basePath ("/me" or "/admin") }
  function renderItem(item, ctx) {
    const { esc, basePath } = ctx;
    const title = deriveTitle(item);
    const href = `${basePath}/event/?id=${encodeURIComponent(item.id)}`;
    const meta = metaHTML(item, esc);
    return `
      <a href="${esc(href)}" class="list-item events-list-item">
        <span class="list-item-copy events-list-copy">
          <strong>${esc(title)}</strong>
          <span class="list-item-meta events-list-meta-row">${meta}</span>
        </span>
      </a>
    `;
  }

  function renderEmpty() {
    const c = eventsCopy().list || {};
    const ariaLabel = c.sectionAriaLabel || "Events";
    const emptyLabel = c.emptyLabel || "";
    return `
      <section class="empty-state empty-state-icon-only" aria-label="${ariaLabel}">
        <p class="empty-state-label">${emptyLabel}</p>
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
    shortDate,
    deriveTitle,
    badgeHTML,
    metaHTML,
    renderItem,
    renderEmpty,
    renderWindowTabs,
  };
})();
