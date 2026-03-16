window.huuperEventCard = (() => {
  function summary(event, meta) {
    return `<article><strong>${window.huuperListPage.escapeHTML(event.title || event.id || "")}</strong>${meta ? `<p class="meta-text">${window.huuperListPage.escapeHTML(meta)}</p>` : ""}</article>`;
  }

  return { summary };
})();
