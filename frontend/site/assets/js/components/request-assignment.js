window.huuperRequestAssignment = (() => {
  function redirectToRequests(requestsURL) {
    window.location.replace(requestsURL);
  }

  function text(value) {
    return window.huuperListPage.text(value);
  }

  function actionText(value) {
    const raw = text(value);
    const labels = {
      group: "Assign group",
      guardian: "Assign guardian",
    };
    return labels[raw] || "Assign";
  }

  function metaText(item, mode) {
    if (mode === "group") {
      return window.huuperGroupMeta.groupMeta(item, {
        requestsVisible: false,
      });
    }
    const parts = [];
    if (Number.isFinite(item.age)) {
      parts.push(`${item.age} years`);
    }
    if (text(item.region)) {
      parts.push(text(item.region));
    }
    return parts.join(" • ");
  }

  function choiceLabel(item, mode) {
    if (mode === "group") {
      return text(item.name || item.id);
    }
    return text(item.full_name || item.email || item.id);
  }

  function renderChoice(item, mode) {
    const name = choiceLabel(item, mode);
    const meta = metaText(item, mode);
    const initials = window.huuperListPage.escapeHTML(name.split(/\s+/).filter(Boolean).slice(0, 2).map((part) => part[0] || "").join("").toUpperCase() || "?");
    const avatar = mode === "group"
      ? `<span class="list-item-media" aria-hidden="true"><span class="list-item-media-face"><span class="list-item-media-text">${initials}</span></span></span>`
      : (() => {
          const filename = text(item.avatar);
          const id = text(item.id);
          const url = filename && id ? `/api/files/users/${encodeURIComponent(id)}/${encodeURIComponent(filename)}` : "";
          return `
            <span class="list-item-media" aria-hidden="true">
              <span class="list-item-media-face">
                ${url ? `<img class="list-item-media-image" src="${window.huuperListPage.escapeHTML(url)}" alt="" onerror="this.hidden=true; if(this.nextElementSibling){ this.nextElementSibling.hidden=false; }" />` : ""}
                <span class="list-item-media-text"${url ? " hidden" : ""}>${initials}</span>
              </span>
            </span>
          `;
        })();
    return `
      <button class="picker-row picker-row-with-avatar" type="button" data-choice="${window.huuperListPage.escapeHTML(item.id)}" data-choice-label="${window.huuperListPage.escapeHTML(name)}">
        ${avatar}
        <span class="picker-row-copy">
          <strong>${window.huuperListPage.escapeHTML(name)}</strong>
          ${meta ? `<p class="meta-text">${window.huuperListPage.escapeHTML(meta)}</p>` : ""}
        </span>
      </button>
    `;
  }

  function bindChoices(node, items, payload, config, mode, statusNode) {
    node.querySelectorAll("[data-choice]").forEach((button) => {
      button.addEventListener("click", () => {
        const choiceID = text(button.getAttribute("data-choice"));
        const choiceName = text(button.getAttribute("data-choice-label"));
        window.huuperActionSheet.open({
          title: text(payload.full_name || payload.email || payload.id),
          meta: mode === "group" ? `Assign to ${choiceName}` : `Assign guardian ${choiceName}`,
          actions: [
            {
              label: actionText(mode),
              onSelect: async () => {
                try {
                  window.huuperListPage.setStatus(statusNode, "");
                  const body = { action: "advance" };
                  if (mode === "group") {
                    body.group = choiceID;
                  } else {
                    body.guardian = choiceID;
                  }
                  await window.huuperAuth.apiFetch(config.actionURL(config.requestID), {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify(body),
                  });
                  redirectToRequests(config.requestsURL);
                } catch (_) {
                  window.huuperListPage.setStatus(statusNode, "Action unavailable.");
                }
              },
            },
          ],
        });
      });
    });
  }

  function init(config) {
    const statusNode = document.querySelector(config.statusSelector);
    const listNode = document.querySelector(config.listSelector);
    const titleNode = document.querySelector(".top-bar-title");
    if (!statusNode || !listNode || !window.huuperAuth || !window.huuperListPage || !window.huuperActionSheet || !window.huuperListItem || !window.huuperGroupMeta) {
      return;
    }

    const requestID = window.huuperListPage.queryParam("id");
    if (!requestID) {
      window.huuperListPage.setStatus(statusNode, "Missing request.");
      return;
    }
    config.requestID = requestID;

    window.huuperAuth.apiFetch(config.detailURL(requestID)).then((payload) => {
      const workflow = payload && typeof payload.workflow === "object" ? payload.workflow : {};
      const requiredField = text(workflow.required_field);
      if (!workflow.can_advance || requiredField !== config.mode) {
        window.huuperListPage.setStatus(statusNode, "Action unavailable.");
        return;
      }

      if (titleNode) {
        titleNode.textContent = actionText(config.mode);
      }

      const options = workflow.options || {};
      const items = config.mode === "group"
        ? (Array.isArray(options.groups) ? options.groups : [])
        : (Array.isArray(options.guardians) ? options.guardians : []);

      if (items.length === 0) {
        window.huuperListPage.setStatus(statusNode, "No options.");
        return;
      }

      listNode.hidden = false;
      window.huuperListPage.renderList(listNode, items, (item) => renderChoice(item, config.mode));
      bindChoices(listNode, items, payload, config, config.mode, statusNode);
      window.huuperListPage.setStatus(statusNode, "");
    }).catch(() => {
      window.huuperListPage.setStatus(statusNode, "Request unavailable.");
    });
  }

  return { init };
})();
