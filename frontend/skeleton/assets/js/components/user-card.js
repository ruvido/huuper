window.huuperUserCard = (() => {
  function text(value) {
    return window.huuperListPage ? window.huuperListPage.text(value) : String(value || "").trim();
  }

  function escapeHTML(value) {
    return window.huuperListPage ? window.huuperListPage.escapeHTML(value) : text(value);
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

  function subtitle(item) {
    const parts = [];
    const age = Number.isFinite(item.age) ? item.age : null;
    const region = text(item.region);

    if (age !== null) {
      parts.push(`${age} years`);
    }
    if (region) {
      parts.push(region);
    }

    return parts.join(" • ");
  }

  function avatarURL(item) {
    const filename = text(item.avatar);
    const id = text(item.id);
    if (!filename || !id) {
      return "";
    }
    return `/api/files/users/${encodeURIComponent(id)}/${encodeURIComponent(filename)}`;
  }

  function trailing(item) {
    if (item.is_assistant) {
      return {
        tone: "active",
        title: "Assistant",
        detail: "",
      };
    }

    if (item.is_guardian) {
      const count = Number.isFinite(item.proteges_count) ? item.proteges_count : 0;
      return {
        tone: "active",
        title: "Guardian",
        detail: `${count > 0 ? count : 1} assigned`,
      };
    }

    return {
      tone: "",
      title: "",
      detail: "",
    };
  }

  function renderMember(item, href) {
    const name = item.full_name || item.email || item.id || "";
    const subline = subtitle(item);
    const meta = trailing(item);
    const safeHref = escapeHTML(href);

    return `
      <a href="${safeHref}" class="user-row">
        <span class="user-avatar${item.is_guardian ? " user-avatar-guardian" : ""}">
          ${avatarURL(item)
            ? `<img class="user-avatar-image" src="${escapeHTML(avatarURL(item))}" alt="${escapeHTML(name)}" />`
            : `<span class="user-avatar-text">${escapeHTML(initials(name))}</span>`}
          ${item.is_guardian ? '<span class="user-avatar-dot"></span>' : ""}
        </span>
        <span class="user-copy">
          <strong>${escapeHTML(name)}</strong>
          ${subline ? `<span class="user-subline">${escapeHTML(subline)}</span>` : ""}
        </span>
        ${(meta.title || meta.detail) ? `
          <span class="user-side${meta.tone ? ` user-side-${meta.tone}` : ""}">
            ${meta.title ? `<span class="user-side-title">${escapeHTML(meta.title)}</span>` : ""}
            ${meta.detail ? `<span class="user-side-detail">${escapeHTML(meta.detail)}</span>` : ""}
          </span>
        ` : ""}
      </a>
    `;
  }

  return {
    renderMember,
  };
})();
