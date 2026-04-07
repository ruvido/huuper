window.huuperEventSummary = (() => {
  function render(location, duration) {
    const locationText = window.huuperListPage.escapeHTML(location || "");
    const durationText = window.huuperListPage.escapeHTML(duration || "");
    return `
      <article class="detail-card">
        ${location ? `<strong>${locationText}</strong>` : ""}
        ${duration ? `<p class="event-summary-duration">${durationText}</p>` : ""}
      </article>
    `;
  }

  return { render };
})();
