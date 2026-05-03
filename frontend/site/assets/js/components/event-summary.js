window.appEventSummary = (() => {
  function render(location, duration) {
    const locationText = window.appListPage.escapeHTML(location || "");
    const durationText = window.appListPage.escapeHTML(duration || "");
    return `
      <article class="detail-card">
        ${location ? `<strong>${locationText}</strong>` : ""}
        ${duration ? `<p class="event-summary-duration">${durationText}</p>` : ""}
      </article>
    `;
  }

  return { render };
})();
