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

  function attendeeSubtitle(item) {
    const parts = [];
    const age = Number.isFinite(item.age) ? item.age : null;
    if (age !== null) {
      parts.push(`${age} years`);
    } else {
      const ageRange = text(item.age_range);
      if (ageRange) {
        parts.push(`${ageRange} years`);
      }
    }

    if (item.is_user) {
      const groupName = text(item.group_name);
      if (groupName) {
        parts.push(groupName);
      }
    } else {
      const region = text(item.region);
      if (region) {
        parts.push(region);
      }
    }

    return parts.join(" • ");
  }

  function avatarURL(item) {
    const filename = text(item.avatar);
    const id = text(item.user_id || item.id);
    if (!filename || !id) {
      return "";
    }
    return `/api/files/users/${encodeURIComponent(id)}/${encodeURIComponent(filename)}`;
  }

  function renderAvatar(item, name, options = {}) {
    const url = avatarURL(item);
    const classes = ["user-avatar"];
    if (options.guardian) {
      classes.push("user-avatar-guardian");
    }

    const fallback = `<span class="user-avatar-text"${url ? ' hidden' : ''}>${escapeHTML(initials(name))}</span>`;
    if (!url) {
      return `<span class="${classes.join(" ")}"><span class="user-avatar-face">${fallback}</span>${options.guardian ? '<span class="user-avatar-dot"></span>' : ""}</span>`;
    }

    return `
      <span class="${classes.join(" ")}">
        <span class="user-avatar-face">
          <img
            class="user-avatar-image"
            src="${escapeHTML(url)}"
            alt=""
            onerror="this.hidden=true; if(this.nextElementSibling){ this.nextElementSibling.hidden=false; }"
          />
          ${fallback}
        </span>
        ${options.guardian ? '<span class="user-avatar-dot"></span>' : ""}
      </span>
    `;
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

  function attendeeTrailing(item) {
    if (item && item.is_admin) {
      return {
        tone: "active",
        title: "Admin",
        detail: "",
      };
    }

    return {
      tone: "",
      title: "",
      detail: "",
    };
  }

  function renderBody(name, subline) {
    return `
      <span class="user-main">
        <span class="user-copy">
          <strong>${escapeHTML(name)}</strong>
          ${subline ? `<span class="user-subline">${escapeHTML(subline)}</span>` : ""}
        </span>
      </span>
    `;
  }

  function renderSide(meta) {
    if (!meta || (!meta.title && !meta.detail)) {
      return "";
    }

    return `
      <span class="user-side${meta.tone ? ` user-side-${meta.tone}` : ""}">
        ${meta.title ? `<span class="user-side-title">${escapeHTML(meta.title)}</span>` : ""}
        ${meta.detail ? `<span class="user-side-detail">${escapeHTML(meta.detail)}</span>` : ""}
      </span>
    `;
  }

  function renderMenu(menuData) {
    if (!menuData) {
      return "";
    }

    return `
      <button class="row-menu-trigger" type="button" aria-label="Open attendee menu" data-attendee="${escapeHTML(JSON.stringify(menuData))}">
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" aria-hidden="true">
          <path d="M9.5 13a1.5 1.5 0 1 1-3 0 1.5 1.5 0 0 1 3 0m0-5a1.5 1.5 0 1 1-3 0 1.5 1.5 0 0 1 3 0m0-5a1.5 1.5 0 1 1-3 0 1.5 1.5 0 0 1 3 0"/>
        </svg>
      </button>
    `;
  }

  function renderRow(config) {
    const tag = config.tag || "article";
    const rowClass = ["user-row"];
    if (config.menuData) {
      rowClass.push("user-row-has-menu");
    }
    if (config.linkable) {
      rowClass.push("user-row-linkable");
    }

    const attrs = [];
    if (tag === "a" && config.href) {
      attrs.push(`href="${escapeHTML(config.href)}"`);
    } else if (config.href) {
      attrs.push(`data-href="${escapeHTML(config.href)}"`);
    }

    return `
      <${tag} class="${rowClass.join(" ")}"${attrs.length > 0 ? ` ${attrs.join(" ")}` : ""}>
        ${config.avatar}
        ${renderBody(config.name, config.subline)}
        ${renderSide(config.meta)}
        ${renderMenu(config.menuData)}
      </${tag}>
    `;
  }

  function renderMember(item, href) {
    const name = item.full_name || item.email || item.id || "";
    const subline = subtitle(item);
    const meta = trailing(item);

    return renderRow({
      tag: "a",
      href,
      name,
      subline,
      meta,
      linkable: false,
      avatar: renderAvatar(item, name, { guardian: item.is_guardian }),
    });
  }

  function renderAttendee(item, href, options = {}) {
    const name = item.full_name || item.email || item.id || "";
    const subline = attendeeSubtitle(item);
    const meta = attendeeTrailing(item);
    const menuData = options.menu === false ? null : {
      id: item.id || "",
      user_id: item.user_id || "",
      full_name: name,
      status: item.status || "",
    };

    return renderRow({
      tag: "article",
      href: href || "",
      name,
      subline,
      meta,
      linkable: Boolean(href),
      menuData,
      avatar: renderAvatar(item, name),
    });
  }

  return {
    renderMember,
    renderAttendee,
  };
})();
