window.huuperRequestAssignmentSheet = (() => {
  function text(value) {
    return window.huuperListPage.text(value);
  }

  function escapeHTML(value) {
    return window.huuperListPage.escapeHTML(value);
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

  function groupMeta(item) {
    return window.huuperGroupMeta.groupMeta(item, { requestsVisible: false });
  }

  function guardianMeta(item) {
    const parts = [];
    if (Number.isFinite(item.age)) {
      parts.push(`${item.age} years`);
    }
    if (text(item.region)) {
      parts.push(text(item.region));
    }
    return parts.join(" • ");
  }

  function optionRow(item, mode) {
    const label = mode === "group"
      ? text(item.name || item.id)
      : text(item.full_name || item.email || item.id);
    const meta = mode === "group" ? groupMeta(item) : guardianMeta(item);
    const avatar = mode === "group"
      ? `<span class="list-item-media" aria-hidden="true"><span class="list-item-media-face"><span class="list-item-media-text">${escapeHTML(initials(label))}</span></span></span>`
      : (() => {
          const filename = text(item.avatar);
          const id = text(item.id);
          const url = filename && id ? `/api/files/users/${encodeURIComponent(id)}/${encodeURIComponent(filename)}` : "";
          return `
            <span class="list-item-media" aria-hidden="true">
              <span class="list-item-media-face">
                ${url ? `<img class="list-item-media-image" src="${escapeHTML(url)}" alt="" onerror="this.hidden=true; if(this.nextElementSibling){ this.nextElementSibling.hidden=false; }" />` : ""}
                <span class="list-item-media-text"${url ? " hidden" : ""}>${escapeHTML(initials(label))}</span>
              </span>
            </span>
          `;
        })();

    return `
      <button class="sheet-option" type="button" data-choice-id="${escapeHTML(item.id)}" data-choice-label="${escapeHTML(label)}">
        ${avatar}
        <span class="sheet-option-copy">
          <strong>${escapeHTML(label)}</strong>
          ${meta ? `<span class="meta-text">${escapeHTML(meta)}</span>` : ""}
        </span>
        <svg class="sheet-option-check" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" aria-hidden="true">
          <path d="M16 2.5 6 12.5l-6-6 1.4-1.4L6 9.7 14.6 1.1z" fill="currentColor"/>
        </svg>
      </button>
    `;
  }

  function open(config) {
    if (!window.huuperActionSheet || !window.huuperAuth || !window.huuperListPage || !window.huuperGroupMeta || !window.huuperRequestActions) {
      return;
    }

    const payloadPromise = config.payload
      ? Promise.resolve(config.payload)
      : window.huuperAuth.apiFetch(config.detailURL(config.requestID));

    payloadPromise.then((payload) => {
      const workflow = payload && typeof payload.workflow === "object" ? payload.workflow : {};
      const action = text(workflow.pending_action);
      const mode = action === "set_guardian" ? "guardian" : "group";
      const options = workflow.options || {};
      const items = mode === "guardian"
        ? (Array.isArray(options.guardians) ? options.guardians : [])
        : (Array.isArray(options.groups) ? options.groups : []);

      if (items.length === 0) {
        return;
      }

      let selectedID = "";
      let selectedLabel = "";
      const actionLabel = text(workflow.pending_action_label || (mode === "guardian" ? "Assign guardian" : "Assign group"));

      window.huuperActionSheet.open({
        title: actionLabel,
        contentHTML: `<div class="sheet-option-list">${items.map((item) => optionRow(item, mode)).join("")}</div>`,
        footerAction: {
          label: actionLabel,
          disabled: true,
          onSelect: async (button) => {
            if (!selectedID) {
              return;
            }
            button.disabled = true;
            const body = { action };
            if (mode === "guardian") {
              body.guardian = selectedID;
            } else {
              body.group = selectedID;
            }
            try {
              await window.huuperRequestActions.submitAndRedirect({
                actionURL: config.actionURL(config.requestID),
                body,
                redirectURL: config.requestsURL,
              });
            } catch (_) {
              button.disabled = false;
            }
          },
        },
        onOpen: (root) => {
          const footerButton = root.querySelector(".action-sheet-footer-button");
          root.querySelectorAll(".sheet-option").forEach((node) => {
            node.addEventListener("click", () => {
              selectedID = text(node.getAttribute("data-choice-id"));
              selectedLabel = text(node.getAttribute("data-choice-label"));
              root.querySelectorAll(".sheet-option-current").forEach((current) => current.classList.remove("sheet-option-current"));
              node.classList.add("sheet-option-current");
              if (footerButton) {
                footerButton.disabled = false;
                footerButton.textContent = `${actionLabel}`;
                footerButton.setAttribute("aria-label", `${actionLabel} ${selectedLabel}`);
              }
            });
          });
        },
      });
    }).catch(() => {});
  }

  return { open };
})();
