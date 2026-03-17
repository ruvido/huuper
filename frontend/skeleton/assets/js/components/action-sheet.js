window.huuperActionSheet = (() => {
  let root;
  let actionsNode;

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
        <div class="action-sheet-actions"></div>
      </section>
    `;

    document.body.appendChild(root);
    actionsNode = root.querySelector(".action-sheet-actions");
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
    actionsNode.innerHTML = "";
  }

  function open(config) {
    ensure();
    const actions = Array.isArray(config && config.actions) ? config.actions : [];
    actionsNode.innerHTML = "";

    actions.forEach((action) => {
      const button = document.createElement("button");
      button.type = "button";
      button.className = `action-sheet-action${action && action.tone ? ` action-sheet-action-${action.tone}` : ""}`;
      button.textContent = action && action.label ? action.label : "Action";
      button.addEventListener("click", () => {
        if (action && typeof action.onSelect === "function") {
          action.onSelect();
        }
        close();
      });
      actionsNode.appendChild(button);
    });

    root.hidden = false;
    root.classList.add("action-sheet-open");
  }

  return {
    open,
    close,
  };
})();
