window.huuperRequestCard = (() => {
  function text(value) {
    return window.huuperListPage ? window.huuperListPage.text(value) : String(value || "").trim();
  }

  function escapeHTML(value) {
    return window.huuperListPage ? window.huuperListPage.escapeHTML(value) : text(value);
  }

  function dateTime(value) {
    return window.huuperListPage ? window.huuperListPage.dateTime(value) : text(value);
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

  function statusText(value) {
    const raw = text(value);
    if (!raw) {
      return "";
    }

    const normalized = raw.replace(/^\d+-/, "").trim();
    const labels = {
      submitted: "Submitted",
      assign_group: "Group assignment",
      assign_guardian: "Guardian assignment",
      guardian_assigned: "Mentoring",
      mentoring: "Mentoring",
      group_approved: "Group approval",
      admin_approved: "Final approval",
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

  function mentoringNotesHTML(item) {
    const mentoringHTML = text(item.mentoring_notes_html);
    if (mentoringHTML) {
      return mentoringHTML;
    }

    const mentoringText = notesText(item.mentoring_notes);
    if (mentoringText) {
      return `<p>${escapeHTML(mentoringText)}</p>`;
    }
    return "";
  }

  function workflowNotesHTML(item) {
    const workflow = item && typeof item.workflow === "object" ? item.workflow : {};
    const rawHTML = text(workflow.next_action_notes_html);
    if (rawHTML) {
      return rawHTML;
    }

    const rawText = notesText(workflow.next_action_notes);
    if (!rawText) {
      return "";
    }
    return `<p>${escapeHTML(rawText)}</p>`;
  }

  function inlineMentoringNotes(item) {
    const raw = notesText(item.mentoring_notes);
    if (!raw) {
      return "";
    }
    if (raw.includes("\n")) {
      return "";
    }
    return raw;
  }

  function renderDetail(item, options = {}) {
    const showNotes = options.showNotes !== false;
    const rows = [
      row("Status", statusText(item.status_label || item.status)),
      row("Guardian", item.guardian_name),
      row("Since", item.assigned_at ? dateOnly(item.assigned_at) : ""),
      row("Created", item.created ? dateOnly(item.created) : ""),
    ].filter(Boolean);
    const inlineMentoring = inlineMentoringNotes(item);
    if (inlineMentoring) {
      rows.push(row("Mentoring notes", inlineMentoring));
    }
    const renderedMentoringNotes = showNotes && !inlineMentoring ? mentoringNotesHTML(item) : "";
    const renderedWorkflowNotes = showNotes ? workflowNotesHTML(item) : "";
    const mentoringNotesBlock = renderedMentoringNotes ? `<div class="request-notes"><p class="request-row"><span>Mentoring notes:</span></p><div>${renderedMentoringNotes}</div></div>` : "";
    const workflowNotesBlock = renderedWorkflowNotes ? `<div class="request-notes"><p class="request-row"><span>Notes:</span></p><div>${renderedWorkflowNotes}</div></div>` : "";

    return `<article class="request-card"><strong>${title(item)}</strong>${rows.join("")}${mentoringNotesBlock}${workflowNotesBlock}</article>`;
  }

  function renderCompact(item, href) {
    const displayTitle = item.full_name || (item.data && (item.data.full_name || item.data.name)) || item.email || item.id;
    const title = window.huuperListPage.escapeHTML(displayTitle);
    const since = item.assigned_at ? `Since: ${dateOnly(item.assigned_at)}` : "";
    const workflow = item && typeof item.workflow === "object" ? item.workflow : {};
    const status = text(workflow.current_action_label) || statusText(item.status_label || item.status);
    return `
      <a href="${window.huuperListPage.escapeHTML(href)}" class="user-row request-compact-row">
        <span class="user-avatar" aria-hidden="true">
          <span class="user-avatar-text">${window.huuperListPage.escapeHTML(initials(displayTitle))}</span>
        </span>
        <span class="user-copy request-compact-copy">
          <strong>${title}</strong>
          ${since ? `<span class="user-subline request-compact-since">${window.huuperListPage.escapeHTML(since)}</span>` : ""}
        </span>
        ${status ? `<span class="user-side"><span class="user-side-title request-compact-status">${window.huuperListPage.escapeHTML(status)}</span></span>` : ""}
      </a>
    `;
  }

  return {
    renderDetail,
    renderCompact,
    statusText,
  };
})();
