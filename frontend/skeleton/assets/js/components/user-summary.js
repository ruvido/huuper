window.huuperUserSummary = (() => {
  function text(value) {
    return window.huuperListPage.text(value);
  }

  function escapeHTML(value) {
    return window.huuperListPage.escapeHTML(value);
  }

  function row(label, value) {
    const rendered = text(value);
    if (!rendered) {
      return "";
    }
    return `
      <article class="request-detail-field">
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

  function splitLines(value) {
    return text(value)
      .split(/\r?\n/)
      .map((part) => part.trim())
      .filter(Boolean);
  }

  function normalizeList(value) {
    return Array.isArray(value) ? value.map((item) => text(item)).filter(Boolean) : [];
  }

  function section(title, bodyHTML) {
    if (!bodyHTML) {
      return "";
    }
    return `
      <article class="request-detail-field user-summary-rich-field">
        <span class="request-detail-term">${escapeHTML(title)}</span>
        <div class="request-detail-description">${bodyHTML}</div>
      </article>
    `;
  }

  function listHTML(items) {
    if (!items.length) {
      return "";
    }
    return `<ul>${items.map((item) => `<li>${escapeHTML(item)}</li>`).join("")}</ul>`;
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

  function render(payload, options = {}) {
    const data = payload && typeof payload.data === "object" ? payload.data : {};
    const fullName = text(payload.full_name || data.full_name || payload.email || payload.id);
    const rows = [];
    const avatar = text(payload.avatar);
    const email = text(payload.email);
    const status = options.includeStatus === true ? text(payload.status) : "";
    const telegram = payload.telegram || {};
    const telegramName = text(telegram.username || telegram.first_name);
    const admin = options.includeAdminFlag === true && payload.admin === true;
    const age = ageText(data.birth_year);
    const birthYear = text(data.birth_year);
    const region = text(data.region);
    const maritalStatus = text(data.marital_status);
    const children = text(data.children);
    const work = text(data.work);
    const why = splitLines(data.why).join(" ");
    const skills = normalizeList(data.skills);
    const interests = normalizeList(data.interests);
    const sports = normalizeList(data.sports);

    if (age) rows.push(row("Age", age));
    if (birthYear) rows.push(row("Birth year", birthYear));
    if (region) rows.push(row("Region", region));
    if (maritalStatus) rows.push(row("Marital status", maritalStatus));
    if (children) rows.push(row("Children", children));
    if (work) rows.push(row("Work", work.replaceAll("\n", " • ")));
    if (email) rows.push(row("Email", email));
    if (status) rows.push(row("Status", status));
    if (telegramName) rows.push(row("Telegram", telegramName));
    if (admin) rows.push(row("Role", "admin"));

    const copyBlocks = [
      section("Why", why ? `<p>${escapeHTML(why)}</p>` : ""),
      section("Skills", listHTML(skills)),
      section("Interests", listHTML(interests)),
      section("Sports", listHTML(sports)),
    ].filter(Boolean).join("");
    const avatarHTML = avatar
      ? `<img class="user-summary-avatar" src="/api/files/users/${encodeURIComponent(payload.id)}/${encodeURIComponent(avatar)}" alt="${escapeHTML(fullName)}" />`
      : `<span class="user-summary-avatar user-summary-avatar-fallback">${escapeHTML(initials(fullName))}</span>`;

    return `
      <article class="request-sheet user-summary-sheet">
        <header class="request-sheet-header">
          <div class="request-identity user-summary-identity">
            ${avatarHTML}
            <h1 class="request-title">${escapeHTML(fullName)}</h1>
          </div>
        </header>
        <section class="request-info-grid">
          ${rows.join("")}
        </section>
        ${copyBlocks}
      </article>
    `;
  }

  return { render };
})();
