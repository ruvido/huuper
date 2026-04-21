window.huuperActionSheet = (() => {
  let root;
  let headerNode;
  let bodyNode;
  let actionsNode;
  let footerNode;

  function escapeHTML(value) {
    return String(value || "")
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;")
      .replaceAll("'", "&#39;");
  }

  function ensure() {
    if (root) {
      return;
    }

    root = document.createElement("div");
    root.className = "action-sheet";
    root.hidden = true;
    root.innerHTML = `
      <button class="action-sheet-backdrop" type="button" aria-label="Close menu"></button>
      <section class="action-sheet-panel" role="dialog" aria-modal="true" aria-label="Action menu">
        <div class="action-sheet-handle" aria-hidden="true"></div>
        <div class="action-sheet-header" hidden></div>
        <div class="action-sheet-body" hidden></div>
        <div class="action-sheet-actions"></div>
        <div class="action-sheet-footer" hidden></div>
      </section>
    `;

    document.body.appendChild(root);
    headerNode = root.querySelector(".action-sheet-header");
    bodyNode = root.querySelector(".action-sheet-body");
    actionsNode = root.querySelector(".action-sheet-actions");
    footerNode = root.querySelector(".action-sheet-footer");
    root.querySelector(".action-sheet-backdrop").addEventListener("click", close);

    document.addEventListener("keydown", (event) => {
      if (event.key === "Escape" && root && !root.hidden) {
        close();
      }
    });
  }

  function close() {
    if (!root) {
      return;
    }
    root.hidden = true;
    root.classList.remove("action-sheet-open");
    document.body.classList.remove("action-sheet-lock");
    headerNode.hidden = true;
    headerNode.innerHTML = "";
    bodyNode.hidden = true;
    bodyNode.innerHTML = "";
    actionsNode.innerHTML = "";
    footerNode.hidden = true;
    footerNode.innerHTML = "";
  }

  function open(config) {
    ensure();
    const actions = Array.isArray(config && config.actions) ? config.actions : [];
    const title = config && config.title ? String(config.title).trim() : "";
    const meta = config && config.meta ? String(config.meta).trim() : "";
    const contentHTML = config && config.contentHTML ? String(config.contentHTML) : "";
    headerNode.hidden = !(title || meta);
    headerNode.innerHTML = title || meta
      ? `${title ? `<strong class="action-sheet-title">${escapeHTML(title)}</strong>` : ""}${meta ? `<p class="action-sheet-meta">${escapeHTML(meta)}</p>` : ""}`
      : "";
    bodyNode.hidden = !contentHTML;
    bodyNode.innerHTML = contentHTML;
    actionsNode.innerHTML = "";
    footerNode.hidden = true;
    footerNode.innerHTML = "";

    actions.forEach((action) => {
      const button = document.createElement("button");
      button.type = "button";
      button.className = `action-sheet-action${action && action.tone ? ` action-sheet-action-${action.tone}` : ""}`;
      button.textContent = action && action.label ? action.label : "Action";
      button.addEventListener("click", async () => {
        const previousDisabled = button.disabled;
        button.disabled = true;
        button.classList.add("request-action-loading");
        try {
          const result = action && typeof action.onSelect === "function" ? action.onSelect(button) : null;
          if (result && typeof result.then === "function") {
            await result;
          }
          close();
        } catch (_) {
          button.disabled = previousDisabled;
          button.classList.remove("request-action-loading");
        }
      });
      actionsNode.appendChild(button);
    });

    if (config && config.footerAction) {
      const footerAction = config.footerAction;
      const button = document.createElement("button");
      button.type = "button";
      button.className = `action-sheet-footer-button${footerAction.tone ? ` action-sheet-footer-button-${footerAction.tone}` : ""}`;
      button.textContent = footerAction.label ? footerAction.label : "Confirm";
      button.disabled = footerAction.disabled === true;
      button.addEventListener("click", async () => {
        const previousDisabled = button.disabled;
        button.disabled = true;
        button.classList.add("request-action-loading");
        try {
          const result = typeof footerAction.onSelect === "function" ? footerAction.onSelect(button) : null;
          if (result && typeof result.then === "function") {
            await result;
          }
        } catch (_) {
          button.disabled = previousDisabled;
          button.classList.remove("request-action-loading");
        }
      });
      footerNode.appendChild(button);
      footerNode.hidden = false;
    }

    root.hidden = false;
    root.classList.add("action-sheet-open");
    document.body.classList.add("action-sheet-lock");
    if (config && typeof config.onOpen === "function") {
      config.onOpen(root);
    }
  }

  return {
    open,
    close,
  };
})();
