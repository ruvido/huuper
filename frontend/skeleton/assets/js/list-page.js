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

  function setStatus(node, message) {
    node.textContent = message;
    node.hidden = !message;
  }

  return {
    text,
    dateTime,
    escapeHTML,
    queryParam,
    renderList,
    setStatus,
  };
})();
