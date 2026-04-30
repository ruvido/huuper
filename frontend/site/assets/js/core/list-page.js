window.huuperListPage = (() => {
  function text(value) {
    const raw = String(value || "").trim();
    return raw;
  }

  function dateTime(value) {
    const raw = text(value);
    if (!raw) {
      return "";
    }

    const parsed = new Date(raw);
    if (Number.isNaN(parsed.getTime())) {
      return raw;
    }

    return new Intl.DateTimeFormat("it-IT", {
      day: "2-digit",
      month: "2-digit",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    }).format(parsed);
  }

  function escapeHTML(value) {
    return text(value)
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;")
      .replaceAll("'", "&#39;");
  }

  function queryParam(name) {
    return new URLSearchParams(window.location.search).get(name) || "";
  }

  function renderList(node, items, renderItem) {
    node.innerHTML = "";
    for (const item of items) {
      node.insertAdjacentHTML("beforeend", renderItem(item));
    }
  }

  function initials(value) {
    const raw = text(value);
    if (!raw) {
      return "?";
    }
    const parts = raw.split(/\s+/).filter(Boolean).slice(0, 2);
    if (parts.length === 0) {
      return raw.slice(0, 2).toUpperCase();
    }
    return parts.map((part) => part[0] || "").join("").toUpperCase();
  }

  function renderListItemLink(href, title, meta, options = {}) {
    const safeHref = escapeHTML(href);
    const safeTitle = escapeHTML(title);
    const safeMeta = text(meta);
    const sideHTML = options && options.sideHTML ? String(options.sideHTML) : "";
    return `
      <a class="list-item" href="${safeHref}">
        ${options.noMedia ? "" : `<span class="list-item-media" aria-hidden="true"><span class="list-item-media-face"><span class="list-item-media-text">${escapeHTML(initials(title))}</span></span></span>`}
        <span class="list-item-main">
          <span class="list-item-copy">
            <strong>${safeTitle}</strong>
            ${safeMeta ? `<span class="list-item-meta">${escapeHTML(safeMeta)}</span>` : ""}
          </span>
        </span>
        ${sideHTML ? `<span class="list-item-side list-item-side-badge">${sideHTML}</span>` : ""}
      </a>
    `;
  }

  function setStatus(node, message) {
    node.textContent = message;
    node.hidden = !message;
  }

  return {
    text,
    dateTime,
    escapeHTML,
    initials,
    queryParam,
    renderList,
    renderListItemLink,
    setStatus,
  };
})();
