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

    const normalized = raw.replace(/^\d+-/, "").replaceAll("_", " ").replaceAll("-", " ").trim();
    if (!normalized) {
      return "";
    }

    return normalized.charAt(0).toUpperCase() + normalized.slice(1);
  }

  function title(item) {
    const data = item && typeof item.data === "object" ? item.data : {};
    return escapeHTML(item.full_name || data.full_name || data.name || item.email || item.id || "");
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

    if (raw.startsWith("md:")) {
      return raw.slice(3).trim();
    }

    return raw;
  }

  function notesHTML(item) {
    const mentoringHTML = text(item.mentoring_notes_html);
    if (mentoringHTML) {
      return mentoringHTML;
    }

    const mentoringText = notesText(item.mentoring_notes);
    if (mentoringText) {
      return `<p>${escapeHTML(mentoringText)}</p>`;
    }

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
    const renderedNotes = showNotes ? notesHTML(item) : "";
    const hasMentoringNotes = text(item.mentoring_notes_html) || text(item.mentoring_notes);
    const notesLabel = hasMentoringNotes ? "Mentoring notes" : "Notes";
    const notesBlock = renderedNotes && !inlineMentoring ? `<div class="request-notes"><p class="request-row"><span>${escapeHTML(notesLabel)}:</span></p><div>${renderedNotes}</div></div>` : "";

    return `<article class="request-card"><strong>${title(item)}</strong>${rows.join("")}${notesBlock}</article>`;
  }

  function renderCompact(item, href) {
    const rows = [];
    if (item.assigned_at) {
      rows.push(row("Since", dateOnly(item.assigned_at)));
    } else {
      const workflow = item && typeof item.workflow === "object" ? item.workflow : {};
      rows.push(row("Status", workflow.current_action_label || item.status_label || statusText(item.status)));
    }

    return `<a class="request-card request-card-link" href="${escapeHTML(href)}"><strong>${title(item)}</strong>${rows.join("")}</a>`;
  }

  return {
    renderDetail,
    renderCompact,
    statusText,
  };
})();
