window.huuperRequestItem = (() => {
  function text(value) {
    return window.huuperListPage ? window.huuperListPage.text(value) : String(value || "").trim();
  }

  function escapeHTML(value) {
    return window.huuperListPage ? window.huuperListPage.escapeHTML(value) : text(value);
  }

  function dateOnly(value) {
    const raw = text(value);
    if (!raw) {
      return "";
    }

    const parsed = new Date(raw);
    if (Number.isNaN(parsed.getTime())) {
      return raw;
    }

    return new Intl.DateTimeFormat("it-IT", {
      day: "2-digit",
      month: "2-digit",
      year: "numeric",
    }).format(parsed);
  }

  function dateTime(value) {
    const raw = text(value);
    if (!raw) {
      return "";
    }

    const parsed = new Date(raw);
    if (Number.isNaN(parsed.getTime())) {
      return raw;
    }

    return new Intl.DateTimeFormat("it-IT", {
      day: "2-digit",
      month: "short",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    }).format(parsed);
  }

  function statusText(value) {
    const raw = text(value);
    if (!raw) {
      return "";
    }

    const normalized = raw.replace(/^\d+-/, "").trim();
    const labels = {
      submitted: "New request",
      assign_group: "Assign group",
      assign_guardian: "Assign guardian",
      guardian_assigned: "Mentoring",
      mentoring: "Mentoring",
      group_approved: "Group approval",
      admin_approved: "Final review",
      rejected: "Rejected",
      promoted: "Promoted",
    };

    if (labels[normalized]) {
      return labels[normalized];
    }

    const humanized = normalized.replaceAll("_", " ").replaceAll("-", " ").trim();
    if (!humanized) {
      return "";
    }

    return humanized.charAt(0).toUpperCase() + humanized.slice(1);
  }

  function actionText(value) {
    const raw = text(value);
    const labels = {
      set_group: "Assign group",
      set_guardian: "Assign guardian",
      set_mentoring_done: "Complete mentoring",
      set_group_approved: "Approve group",
      set_admin_approved: "Approve request",
      reject: "Reject request",
      promote: "Promote request",
    };

    return labels[raw] || raw.replaceAll("_", " ").replaceAll("-", " ").trim() || "Continue";
  }

  function title(item) {
    const data = item && typeof item.data === "object" ? item.data : {};
    return escapeHTML(item.full_name || data.full_name || data.name || item.email || item.id || "");
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

  function row(label, value) {
    const rendered = text(value);
    if (!rendered) {
      return "";
    }
    return `<p class="request-row"><span>${escapeHTML(label)}:</span> <strong>${escapeHTML(rendered)}</strong></p>`;
  }

  function notesText(value) {
    const raw = text(value);
    if (!raw) {
      return "";
    }
    return raw;
  }

  function detailField(label, value) {
    const rendered = text(value);
    if (!rendered) {
      return "";
    }

    const classes = ["request-detail-field"];
    return `
      <article class="${classes.join(" ")}">
        <span class="request-detail-term">${escapeHTML(label)}</span>
        <strong class="request-detail-description">${escapeHTML(rendered)}</strong>
      </article>
    `;
  }

  function ageText(birthYear) {
    const parsed = Number.parseInt(text(birthYear), 10);
    const currentYear = new Date().getFullYear();
    if (!Number.isFinite(parsed) || parsed < 1900 || parsed > currentYear) {
      return "";
    }
    return `${currentYear - parsed} years`;
  }

  function requestAgeDays(createdAt) {
    const raw = text(createdAt);
    if (!raw) {
      return "";
    }
    const parsed = new Date(raw);
    if (Number.isNaN(parsed.getTime())) {
      return "";
    }
    const diffMs = Date.now() - parsed.getTime();
    if (diffMs <= 0) {
      return "";
    }
    const diffDays = Math.floor(diffMs / 86400000);
    if (diffDays < 1) {
      const diffHours = Math.max(1, Math.floor(diffMs / 3600000));
      return diffHours === 1 ? "1 hour ago" : `${diffHours} hours ago`;
    }
    return diffDays === 1 ? "1 day ago" : `${diffDays} days ago`;
  }

  function renderDetail(item, options = {}) {
    const data = item && typeof item.data === "object" ? item.data : {};
    const fullName = text(data.full_name || item.full_name || item.email || item.id);
    const age = ageText(data.birth_year);
    const location = text(data.region);
    const ageLabel = requestAgeDays(item.created || data.created);
    const meta = age && location
      ? `${escapeHTML(age)} <span class="request-meta-dot" aria-hidden="true"></span> ${escapeHTML(location)}`
      : age
        ? escapeHTML(age)
        : location
          ? escapeHTML(location)
          : "";
    const details = [
      detailField("Phone", data.mobile),
      detailField("Email", item.email),
      detailField("Marital status", data.marital_status),
      detailField("Children", data.children),
    ].filter(Boolean).join("");
    const motivation = text(data.motivation)
      ? `
        <section class="request-motivation">
          <span class="request-detail-term">Motivation</span>
          <p class="request-motivation-quote">${escapeHTML(data.motivation)}</p>
        </section>
      `
      : "";

    return `
      <article class="request-sheet">
        <header class="request-sheet-header">
          <div class="request-identity">
            ${ageLabel ? `<span class="request-candidate-label">${escapeHTML(ageLabel)}</span>` : ""}
            <h1 class="request-title">${escapeHTML(fullName)}</h1>
            ${meta ? `<p class="request-subtitle">${meta}</p>` : ""}
          </div>
        </header>
        <section class="request-info-grid">${details}</section>
        ${motivation}
      </article>
    `;
  }

  function requestMeta(item) {
    const parts = [];
    const data = item && typeof item.data === "object" ? item.data : {};
    const region = text(item.region || data.region);
    const rawBirthYear = text(item.birth_year || data.birth_year);
    const birthYear = Number.parseInt(rawBirthYear, 10);
    const currentYear = new Date().getFullYear();

    if (region) {
      parts.push(region);
    }
    if (Number.isFinite(birthYear) && birthYear > 1900 && birthYear <= currentYear) {
      parts.push(`${currentYear - birthYear} years`);
    }

    return parts.join(" • ");
  }

  function renderListItem(item, href, options = {}) {
    const displayTitle = item.full_name || (item.data && (item.data.full_name || item.data.name)) || item.email || item.id;
    const safeTitle = escapeHTML(displayTitle);
    const meta = requestMeta(item);
    const workflow = item && typeof item.workflow === "object" ? item.workflow : {};
    const status = text(workflow.pending_action_label) || statusText(item.status_label || item.status);
    const side = status
      ? `<span class="list-item-side"><span class="list-item-side-title request-item-status">${escapeHTML(status)}</span></span>`
      : "";

    return `
      <a href="${escapeHTML(href)}" class="list-item request-item">
        <span class="list-item-copy request-item-copy">
          <strong>${safeTitle}</strong>
          ${meta ? `<span class="list-item-meta request-item-meta">${escapeHTML(meta)}</span>` : ""}
        </span>
        ${side}
      </a>
    `;
  }

  return {
    actionText,
    renderDetail,
    renderListItem,
    requestAgeDays,
    statusText,
  };
})();
