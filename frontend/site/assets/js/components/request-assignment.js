window.appRequestAssignment = (() => {
  function text(value) {
    return window.appListPage.text(value);
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
      return window.appGroupMeta.groupMeta(item, {
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
    const initials = window.appListPage.escapeHTML(name.split(/\s+/).filter(Boolean).slice(0, 2).map((part) => part[0] || "").join("").toUpperCase() || "?");
    const avatar = mode === "group"
      ? `<span class="list-item-media" aria-hidden="true"><span class="list-item-media-face"><span class="list-item-media-text">${initials}</span></span></span>`
      : (() => {
          const filename = text(item.avatar);
          const id = text(item.id);
          const url = filename && id ? `/api/files/users/${encodeURIComponent(id)}/${encodeURIComponent(filename)}` : "";
          return `
            <span class="list-item-media" aria-hidden="true">
              <span class="list-item-media-face">
                ${url ? `<img class="list-item-media-image" src="${window.appListPage.escapeHTML(url)}" alt="" onerror="this.hidden=true; if(this.nextElementSibling){ this.nextElementSibling.hidden=false; }" />` : ""}
                <span class="list-item-media-text"${url ? " hidden" : ""}>${initials}</span>
              </span>
            </span>
          `;
        })();
    return `
      <button class="picker-row picker-row-with-avatar" type="button" data-choice="${window.appListPage.escapeHTML(item.id)}" data-choice-label="${window.appListPage.escapeHTML(name)}">
        ${avatar}
        <span class="picker-row-copy">
          <strong>${window.appListPage.escapeHTML(name)}</strong>
          ${meta ? `<p class="meta-text">${window.appListPage.escapeHTML(meta)}</p>` : ""}
        </span>
      </button>
    `;
  }

  function bindChoices(node, items, payload, config, mode, statusNode) {
    node.querySelectorAll("[data-choice]").forEach((button) => {
      button.addEventListener("click", () => {
        const choiceID = text(button.getAttribute("data-choice"));
        const choiceName = text(button.getAttribute("data-choice-label"));
        const workflow = payload && typeof payload.workflow === "object" ? payload.workflow : {};
        const action = text(workflow.pending_action);
        window.appActionSheet.open({
          title: text(payload.full_name || payload.email || payload.id),
          meta: mode === "group" ? `Assign to ${choiceName}` : `Assign guardian ${choiceName}`,
          actions: [
            {
              label: actionText(mode),
              onSelect: async () => {
                try {
                  window.appListPage.setStatus(statusNode, "");
                  const body = { action };
                  if (mode === "group") {
                    body.group = choiceID;
                  } else {
                    body.guardian = choiceID;
                  }
                  await window.appRequestActions.submitAndRedirect({
                    actionURL: config.actionURL(config.requestID),
                    body,
                    redirectURL: config.requestsURL,
                  });
                } catch (error) {
                  window.appListPage.setStatus(
                    statusNode,
                    window.appRequestActions.errorMessage(error, "Action unavailable."),
                  );
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
    if (!statusNode || !listNode || !window.appAuth || !window.appListPage || !window.appActionSheet || !window.appListItem || !window.appGroupMeta || !window.appRequestActions) {
      return;
    }

    const requestID = window.appListPage.queryParam("id");
    if (!requestID) {
      window.appListPage.setStatus(statusNode, "Missing request.");
      return;
    }
    config.requestID = requestID;

    window.appAuth.apiFetch(config.detailURL(requestID)).then((payload) => {
      const workflow = payload && typeof payload.workflow === "object" ? payload.workflow : {};
      const requiredField = text(workflow.required_field);
      const expectedAction = config.mode === "group" ? "set_group" : "set_guardian";
      if (workflow.can_take_pending_action !== true || requiredField !== config.mode || text(workflow.pending_action) !== expectedAction) {
        window.appListPage.setStatus(statusNode, "Action unavailable.");
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
        window.appListPage.setStatus(statusNode, "No options.");
        return;
      }

      listNode.hidden = false;
      window.appListPage.renderList(listNode, items, (item) => renderChoice(item, config.mode));
      bindChoices(listNode, items, payload, config, config.mode, statusNode);
      window.appListPage.setStatus(statusNode, "");
    }).catch(() => {
      window.appListPage.setStatus(statusNode, "Request unavailable.");
    });
  }

  return { init };
})();
