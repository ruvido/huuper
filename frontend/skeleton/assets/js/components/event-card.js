window.huuperEventCard = (() => {
  function summaryLines(location, duration) {
    const locationText = window.huuperListPage.escapeHTML(location || "");
    const durationText = window.huuperListPage.escapeHTML(duration || "");
    return `
      <article class="event-summary-card">
        ${location ? `<strong>${locationText}</strong>` : ""}
        ${duration ? `<p class="event-summary-duration">${durationText}</p>` : ""}
      </article>
    `;
  }

  return { summaryLines };
})();
