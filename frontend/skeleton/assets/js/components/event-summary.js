window.appEventSummary = (() => {
  function render(location, duration, description, url) {
    const esc = window.appListPage.escapeHTML;
    const locationText = esc(location || "");
    const durationText = esc(duration || "");
    const descriptionText = esc(description || "");
    const urlText = esc(url || "");
    return `
      <article class="detail-card">
        ${location ? `<strong>${locationText}</strong>` : ""}
        ${duration ? `<p class="event-summary-duration">${durationText}</p>` : ""}
        ${description ? `<p class="event-summary-description">${descriptionText}</p>` : ""}
        ${url ? `<p class="event-summary-url"><a href="${urlText}" target="_blank" rel="noopener noreferrer">${urlText}</a></p>` : ""}
      </article>
    `;
  }

  return { render };
})();
