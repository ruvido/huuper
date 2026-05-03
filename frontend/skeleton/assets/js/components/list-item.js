window.appListItem = (() => {
  function text(value) {
    return window.appListPage ? window.appListPage.text(value) : String(value || "").trim();
  }

  function escapeHTML(value) {
    return window.appListPage ? window.appListPage.escapeHTML(value) : text(value);
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

  function renderMedia(item, label, options = {}) {
    const url = avatarURL(item);
    const classes = ["list-item-media"];

    const fallback = `<span class="list-item-media-text"${url ? ' hidden' : ""}>${escapeHTML(initials(label))}</span>`;
    if (!url) {
      return `<span class="${classes.join(" ")}" aria-hidden="true"><span class="list-item-media-face">${fallback}</span></span>`;
    }

    return `
      <span class="${classes.join(" ")}" aria-hidden="true">
        <span class="list-item-media-face">
          <img
            class="list-item-media-image"
            src="${escapeHTML(url)}"
            alt=""
            onerror="this.hidden=true; if(this.nextElementSibling){ this.nextElementSibling.hidden=false; }"
          />
          ${fallback}
        </span>
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

  function renderBody(name, meta) {
    return `
      <span class="list-item-main">
        <span class="list-item-copy">
          <strong>${escapeHTML(name)}</strong>
          ${meta ? `<span class="list-item-meta">${escapeHTML(meta)}</span>` : ""}
        </span>
      </span>
    `;
  }

  function renderSide(side) {
    if (!side || (!side.title && !side.detail)) {
      return "";
    }

    return `
      <span class="list-item-side${side.tone ? ` list-item-side-${side.tone}` : ""}">
        ${side.title ? `<span class="list-item-side-title">${escapeHTML(side.title)}</span>` : ""}
        ${side.detail ? `<span class="list-item-side-detail">${escapeHTML(side.detail)}</span>` : ""}
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
    const classes = ["list-item"];
    if (config.menuData) {
      classes.push("list-item-has-menu");
    }
    if (config.linkable) {
      classes.push("list-item-linkable");
    }

    const attrs = [];
    if (tag === "a" && config.href) {
      attrs.push(`href="${escapeHTML(config.href)}"`);
    } else if (tag === "button") {
      attrs.push(`type="button"`);
    } else if (config.href) {
      attrs.push(`data-href="${escapeHTML(config.href)}"`);
    }

    return `
      <${tag} class="${classes.join(" ")}"${attrs.length > 0 ? ` ${attrs.join(" ")}` : ""}>
        ${config.media}
        ${renderBody(config.name, config.metaText)}
        ${renderSide(config.side)}
        ${renderMenu(config.menuData)}
      </${tag}>
    `;
  }

  function renderMember(item, href) {
    const name = item.full_name || item.email || item.id || "";
    return renderRow({
      tag: "a",
      href,
      name,
      metaText: subtitle(item),
      side: trailing(item),
      linkable: false,
      media: renderMedia(item, name),
    });
  }

  function renderAttendee(item, href, options = {}) {
    const name = item.full_name || item.email || item.id || "";
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
      metaText: attendeeSubtitle(item),
      side: attendeeTrailing(item),
      linkable: Boolean(href),
      menuData,
      media: renderMedia(item, name),
    });
  }

  function renderSelectableMember(item) {
    const name = item.full_name || item.email || item.id || "";
    return renderRow({
      tag: "button",
      name,
      metaText: subtitle(item),
      side: trailing(item),
      linkable: false,
      media: renderMedia(item, name),
    });
  }

  return {
    renderMember,
    renderAttendee,
    renderSelectableMember,
  };
})();
