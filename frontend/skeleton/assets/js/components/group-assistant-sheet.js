window.huuperGroupAssistantSheet = (() => {
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

  function memberMeta(item) {
    const parts = [];
    if (Number.isFinite(item.age)) {
      parts.push(`${item.age} years`);
    }
    if (text(item.region)) {
      parts.push(text(item.region));
    }
    return parts.join(" • ");
  }

  function renderChoice(item) {
    const label = text(item.full_name || item.email || item.id);
    const meta = memberMeta(item);
    const filename = text(item.avatar);
    const id = text(item.id);
    const url = filename && id ? `/api/files/users/${encodeURIComponent(id)}/${encodeURIComponent(filename)}` : "";
    return `
      <button class="sheet-option" type="button" data-choice-id="${escapeHTML(id)}" data-choice-label="${escapeHTML(label)}">
        <span class="list-item-media" aria-hidden="true">
          <span class="list-item-media-face">
            ${url ? `<img class="list-item-media-image" src="${escapeHTML(url)}" alt="" onerror="this.hidden=true; if(this.nextElementSibling){ this.nextElementSibling.hidden=false; }" />` : ""}
            <span class="list-item-media-text"${url ? " hidden" : ""}>${escapeHTML(initials(label))}</span>
          </span>
        </span>
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
    if (!window.huuperActionSheet || !window.huuperAuth || !window.huuperListPage) {
      return;
    }

    const assistantURL = text(config && config.assistantURL);
    const members = Array.isArray(config && config.members) ? config.members : [];
    if (!assistantURL || members.length === 0) {
      return;
    }

    const groupName = text(config && config.groupName) || text(config && config.groupID);
    let selectedID = "";
    let selectedLabel = "";

    window.huuperActionSheet.open({
      title: "Assign assistant",
      meta: groupName,
      contentHTML: `<div class="sheet-option-list">${members.map((item) => renderChoice(item)).join("")}</div>`,
      footerAction: {
        label: "Assign assistant",
        disabled: true,
        onSelect: async (button) => {
          if (!selectedID) {
            return;
          }

          button.disabled = true;
          try {
            await window.huuperAuth.apiFetch(assistantURL, {
              method: "POST",
              headers: { "Content-Type": "application/json" },
              body: JSON.stringify({ assistant: selectedID }),
            });
            window.location.replace(String(config.redirectURL || window.location.href).trim() || window.location.href);
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
              footerButton.textContent = "Assign assistant";
              footerButton.setAttribute("aria-label", `Assign assistant ${selectedLabel}`);
            }
          });
        });
      },
    });
  }

  return { open };
})();
