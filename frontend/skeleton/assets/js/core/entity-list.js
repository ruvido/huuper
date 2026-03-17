window.huuperEntityList = (() => {
  function init(config) {
    const listNode = document.querySelector(config.listSelector);
    const statusNode = document.querySelector(config.statusSelector);
    if (!listNode || !statusNode || !window.huuperListPage) {
      return;
    }
    if (config.requiresAuth && !window.huuperAuth) {
      return;
    }
    if (config.requiresRequestCard && !window.huuperRequestCard) {
      return;
    }

    async function load() {
      try {
        const payload = await config.load();
        const items = Array.isArray(payload.items) ? payload.items : [];
        if (items.length === 0) {
          if (typeof config.renderEmpty === "function") {
            listNode.innerHTML = config.renderEmpty();
            listNode.hidden = false;
            statusNode.hidden = true;
          } else {
            window.huuperListPage.setStatus(statusNode, config.emptyMessage);
          }
          return;
        }

        window.huuperListPage.renderList(listNode, items, config.renderItem);
        listNode.hidden = false;
      } catch (_) {
        window.huuperListPage.setStatus(statusNode, config.errorMessage);
      }
    }

    load();
  }

  return { init };
})();
