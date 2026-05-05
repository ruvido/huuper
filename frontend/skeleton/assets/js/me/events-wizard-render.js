// Render functions for the events wizard. These functions take a `ctx`
// object containing { state, copy, dom, PREFIX, esc } and read from it
// explicitly. No closure over outer state.
(() => {
  window.appEventsWizard = window.appEventsWizard || {};
  const ns = window.appEventsWizard;
  if (!ns.config) return;

  const cfgmod = ns.config;

  function esc(value) {
    return String(value || "")
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;")
      .replaceAll("'", "&#39;");
  }

  function setNextIcon(kind, PREFIX) {
    const nextBtn = document.getElementById(`${PREFIX}-next`);
    if (!nextBtn) return;
    const def = nextBtn.querySelector('[data-wizard-next-icon="default"]');
    const ed = nextBtn.querySelector('[data-wizard-next-icon="edit"]');
    if (def) def.hidden = kind === "edit";
    if (ed) ed.hidden = kind !== "edit";
  }

  function typeOptionLabel(typeValue, copy) {
    void copy;
    return typeValue ? typeValue.charAt(0).toUpperCase() + typeValue.slice(1) : "";
  }

  function scheduleLabel(value, copy) {
    const labels = (copy && copy.scheduleLabels) || {};
    return labels[value] || value;
  }

  function renderTypeStep(ctx) {
    const { state, dom, copy } = ctx;
    const types = cfgmod.allowedTypes(state.cfg, state.isAdmin);
    const stepCopy = (copy.steps && copy.steps.type) || {};
    const cards = types.map((t) => {
      const selected = state.input.type === t.value;
      const label = t.label || typeOptionLabel(t.value, copy);
      const description = t.description || "";
      return `
        <label class="events-wizard-type-option${selected ? " is-selected" : ""}">
          <input type="radio" name="event_type" value="${esc(t.value)}" ${selected ? "checked" : ""} />
          <span class="events-wizard-type-option-body">
            <strong>${esc(label)}</strong>
            ${description ? `<span class="events-wizard-type-option-meta">${esc(description)}</span>` : ""}
          </span>
        </label>
      `;
    }).join("");
    const empty = !types.length
      ? `<p class="meta-text">${esc(copy.errors.noAllowedTypes || "")}</p>`
      : "";
    dom.stage.innerHTML = `
      <section class="wizard-step wizard-step-priority events-wizard-step">
        <header class="wizard-step-header">
          <h2 class="display-hero">${esc(stepCopy.title || "")}</h2>
          ${stepCopy.subtitle ? `<cite class="intro-quote">${esc(stepCopy.subtitle)}</cite>` : ""}
        </header>
        <div class="events-wizard-type-grid">
          ${cards || empty}
        </div>
      </section>
    `;
  }

  function renderGroupStep(ctx) {
    const { state, dom, copy } = ctx;
    const stepCopy = (copy.steps && copy.steps.group) || {};
    const groups = state.availableGroups || [];
    const empty = !groups.length
      ? `<p class="meta-text">${esc(copy.errors.noEligibleGroups || "")}</p>`
      : "";
    const opts = groups.map((g) => {
      const selected = state.input.group === g.id;
      return `
        <label class="events-wizard-type-option${selected ? " is-selected" : ""}">
          <input type="radio" name="event_group" value="${esc(g.id)}" ${selected ? "checked" : ""} />
          <span class="events-wizard-type-option-body">
            <strong>${esc(g.name || g.id)}</strong>
          </span>
        </label>
      `;
    }).join("");
    dom.stage.innerHTML = `
      <section class="wizard-step wizard-step-priority events-wizard-step">
        <header class="wizard-step-header">
          <h2 class="display-hero">${esc(stepCopy.title || "")}</h2>
          ${stepCopy.subtitle ? `<cite class="intro-quote">${esc(stepCopy.subtitle)}</cite>` : ""}
        </header>
        <div class="events-wizard-type-grid">
          ${opts || empty}
        </div>
      </section>
    `;
  }

  function renderDatesStep(ctx) {
    const { state, dom, copy } = ctx;
    const stepCopy = (copy.steps && copy.steps.dates) || {};
    const typeDef = cfgmod.typeDef(state.cfg, state.input.type);
    const endDateRequired = cfgmod.required(typeDef, "end_date");
    const schedule = state.input.schedule || "once";
    const cadence = cfgmod.cadenceFromInput(state.input);
    const previewDates = cfgmod.generateScheduleDates(state.input.startDate, cadence, state.input.count);
    const previewLines = previewDates.map((d) => `<li>${esc(d)}</li>`).join("");
    const countOptions = cfgmod.COUNT_OPTIONS.map((n) => `
      <label class="segmented-option">
        <input type="radio" name="event_count" value="${n}" ${Number(state.input.count) === n ? "checked" : ""} />
        <span>${n}</span>
      </label>
    `).join("");
    const weekdayOptions = cfgmod.WEEKDAYS.map((day) => `
      <label class="segmented-option">
        <input type="radio" name="event_weekday" value="${esc(day)}" ${state.input.weekday === day ? "checked" : ""} />
        <span>${esc(scheduleLabel(day, copy))}</span>
      </label>
    `).join("");
    const monthDayOptions = cfgmod.WEEKDAYS.map((day) => `
      <label class="segmented-option">
        <input type="radio" name="event_month_day" value="${esc(day)}" ${state.input.monthDay === day ? "checked" : ""} />
        <span>${esc(scheduleLabel(day, copy))}</span>
      </label>
    `).join("");
    const monthlyPositionOptions = cfgmod.MONTH_POSITIONS.map((pos) => `
      <label class="segmented-option">
        <input type="radio" name="event_month_position" value="${esc(pos)}" ${state.input.monthPosition === pos ? "checked" : ""} />
        <span>${esc(scheduleLabel(`position_${pos}`, copy))}</span>
      </label>
    `).join("");
    const repeatedSection = schedule !== "once" ? `
      <div class="events-wizard-field">
        <label class="field-label">${esc((copy.labels && copy.labels.countField) || "")}</label>
        <div class="segmented">${countOptions}</div>
      </div>
    ` : "";
    const weeklySection = "";
    const monthlySection = "";
    void weekdayOptions;
    void monthDayOptions;
    void monthlyPositionOptions;

    dom.stage.innerHTML = `
      <section class="wizard-step wizard-step-priority events-wizard-step">
        <header class="wizard-step-header">
          <h2 class="display-hero">${esc(stepCopy.title || "")}</h2>
          ${stepCopy.subtitle ? `<cite class="intro-quote">${esc(stepCopy.subtitle)}</cite>` : ""}
        </header>

        <div class="events-wizard-field">
          <label class="field-label">${esc((copy.labels && copy.labels.scheduleField) || "")}</label>
          <div class="segmented">
            <label class="segmented-option">
              <input type="radio" name="event_schedule" value="once" ${schedule === "once" ? "checked" : ""} />
              <span>${esc((copy.labels && copy.labels.onceOption) || "Once")}</span>
            </label>
            <label class="segmented-option">
              <input type="radio" name="event_schedule" value="weekly" ${schedule === "weekly" ? "checked" : ""} />
              <span>${esc((copy.labels && copy.labels.weeklyOption) || "Weekly")}</span>
            </label>
            <label class="segmented-option">
              <input type="radio" name="event_schedule" value="monthly" ${schedule === "monthly" ? "checked" : ""} />
              <span>${esc((copy.labels && copy.labels.monthlyOption) || "Monthly")}</span>
            </label>
          </div>
        </div>

        <div class="events-wizard-field">
          <label class="field-label" for="${ctx.PREFIX}-startdate">${esc((copy.labels && copy.labels.startDateField) || "")}</label>
          <input class="field-input" id="${ctx.PREFIX}-startdate" type="date" name="event_start_date" value="${esc(state.input.startDate)}" />
        </div>

        <div class="events-wizard-field">
          <label class="field-label" for="${ctx.PREFIX}-starttime">${esc((copy.labels && copy.labels.startTimeField) || "")} ${esc((copy.labels && copy.labels.optionalTag) || "")}</label>
          <input class="field-input" id="${ctx.PREFIX}-starttime" type="time" name="event_start_time" value="${esc(state.input.startTime || "")}" />
        </div>

        <div class="events-wizard-field">
          <label class="field-label" for="${ctx.PREFIX}-enddate">${esc((copy.labels && copy.labels.endDateField) || "")}${endDateRequired ? "" : ` ${esc((copy.labels && copy.labels.optionalTag) || "")}`}</label>
          <input class="field-input${state.fieldErrors.end_date ? " field-input-error" : ""}" id="${ctx.PREFIX}-enddate" type="date" name="event_end_date" value="${esc(state.input.endDate)}" />
        </div>

        ${weeklySection}
        ${monthlySection}
        ${repeatedSection}

        ${previewLines ? `
          <div class="events-wizard-field">
            <label class="field-label">${esc((copy.labels && copy.labels.previewField) || "")}</label>
            <ul class="events-wizard-preview-list wizard-summary-muted-card">${previewLines}</ul>
          </div>
        ` : ""}
      </section>
    `;
  }

  function renderDetailsStep(ctx) {
    const { state, dom, copy } = ctx;
    const stepCopy = (copy.steps && copy.steps.details) || {};
    const typeDef = cfgmod.typeDef(state.cfg, state.input.type);
    const titleRequired = cfgmod.required(typeDef, "title");
    const urlRequired = cfgmod.required(typeDef, "url");
    const locationRequired = cfgmod.required(typeDef, "location");
    const descRaw = (state.input.data && state.input.data.description) || "";
    dom.stage.innerHTML = `
      <section class="wizard-step wizard-step-priority events-wizard-step">
        <header class="wizard-step-header">
          <h2 class="display-hero">${esc(stepCopy.title || "")}</h2>
          ${stepCopy.subtitle ? `<cite class="intro-quote">${esc(stepCopy.subtitle)}</cite>` : ""}
        </header>

        <div class="events-wizard-field">
          <label class="field-label" for="${ctx.PREFIX}-title">${esc((copy.labels && copy.labels.titleField) || "")}${titleRequired ? "" : ` ${esc((copy.labels && copy.labels.optionalTag) || "")}`}</label>
          <input class="field-input${state.fieldErrors.title ? " field-input-error" : ""}" id="${ctx.PREFIX}-title" type="text" name="event_title" value="${esc(state.input.title)}" placeholder="${esc((copy.placeholders && copy.placeholders.title) || "")}" />
        </div>

        <div class="events-wizard-field">
          <label class="field-label" for="${ctx.PREFIX}-location">${esc((copy.labels && copy.labels.locationField) || "")}${locationRequired ? "" : ` ${esc((copy.labels && copy.labels.optionalTag) || "")}`}</label>
          <input class="field-input${state.fieldErrors.location ? " field-input-error" : ""}" id="${ctx.PREFIX}-location" type="text" name="event_location" value="${esc(state.input.location)}" placeholder="${esc((copy.placeholders && copy.placeholders.location) || "")}" />
        </div>

        <div class="events-wizard-field">
          <label class="field-label" for="${ctx.PREFIX}-url">${esc((copy.labels && copy.labels.urlField) || "")}${urlRequired ? "" : ` ${esc((copy.labels && copy.labels.optionalTag) || "")}`}</label>
          <input class="field-input${state.fieldErrors.url ? " field-input-error" : ""}" id="${ctx.PREFIX}-url" type="url" name="event_url" value="${esc(state.input.url)}" placeholder="${esc((copy.placeholders && copy.placeholders.url) || "")}" />
        </div>

        <div class="events-wizard-field">
          <label class="field-label" for="${ctx.PREFIX}-desc">${esc((copy.labels && copy.labels.descriptionField) || "")} ${esc((copy.labels && copy.labels.optionalTag) || "")}</label>
          <textarea class="field-input" id="${ctx.PREFIX}-desc" name="event_description" rows="4" placeholder="${esc((copy.placeholders && copy.placeholders.description) || "")}">${esc(descRaw)}</textarea>
        </div>
      </section>
    `;
  }

  function renderConfirm(ctx) {
    const { state, dom, copy, PREFIX } = ctx;
    const stepCopy = (copy.steps && copy.steps.confirm) || {};
    const td = cfgmod.typeDef(state.cfg, state.input.type) || {};
    const typeLabel = td.label || typeOptionLabel(state.input.type, copy);
    const previewDates = cfgmod.generateScheduleDates(state.input.startDate, cfgmod.cadenceFromInput(state.input), state.input.count);
    const dateLines = previewDates.map((d) => `<li>${esc(d)}</li>`).join("");
    const groupName = state.input.groupName || "";
    const desc = (state.input.data && state.input.data.description) || "";
    const title = (state.input.title || "").trim();
    const location = (state.input.location || "").trim();
    const url = (state.input.url || "").trim();

    dom.stage.innerHTML = `
      <section class="wizard-step wizard-step-confirm events-wizard-step">
        <header class="wizard-step-header">
          <h2 class="display-hero">${esc(stepCopy.title || "")}</h2>
        </header>
        <div class="wizard-summary">
          <div class="events-wizard-summary-row">
            <label class="field-label wizard-summary-tag">${esc((copy.labels && copy.labels.summaryType) || "Type")}</label>
            <p class="wizard-summary-value">${esc(typeLabel)}</p>
          </div>
          ${groupName ? `
            <div class="events-wizard-summary-row">
              <label class="field-label wizard-summary-tag">${esc((copy.labels && copy.labels.summaryGroup) || "Group")}</label>
              <p class="wizard-summary-value">${esc(groupName)}</p>
            </div>
          ` : ""}
          <div class="events-wizard-summary-row">
            <label class="field-label wizard-summary-tag">${esc((copy.labels && copy.labels.summaryDates) || "Dates")}</label>
            ${dateLines ? `<ul class="wizard-summary-muted-card events-wizard-preview-list">${dateLines}</ul>` : `<p class="wizard-summary-value">—</p>`}
          </div>
          ${title ? `
            <div class="events-wizard-summary-row">
              <label class="field-label wizard-summary-tag">${esc((copy.labels && copy.labels.summaryTitle) || "Title")}</label>
              <p class="wizard-summary-value">${esc(title)}</p>
            </div>
          ` : ""}
          ${location ? `
            <div class="events-wizard-summary-row">
              <label class="field-label wizard-summary-tag">${esc((copy.labels && copy.labels.summaryLocation) || "Location")}</label>
              <p class="wizard-summary-value">${esc(location)}</p>
            </div>
          ` : ""}
          ${url ? `
            <div class="events-wizard-summary-row">
              <label class="field-label wizard-summary-tag">${esc((copy.labels && copy.labels.summaryURL) || "URL")}</label>
              <p class="wizard-summary-value"><a href="${esc(url)}" target="_blank" rel="noopener noreferrer">${esc(url)}</a></p>
            </div>
          ` : ""}
          ${desc ? `
            <div class="events-wizard-summary-row">
              <label class="field-label wizard-summary-tag">${esc((copy.labels && copy.labels.summaryDescription) || "Description")}</label>
              <p class="wizard-summary-muted-card">${esc(desc).replaceAll("\n", "<br>")}</p>
            </div>
          ` : ""}
        </div>
      </section>
    `;
    void PREFIX;
  }

  ns.render = {
    esc,
    setNextIcon,
    typeOptionLabel,
    scheduleLabel,
    renderTypeStep,
    renderGroupStep,
    renderDatesStep,
    renderDetailsStep,
    renderConfirm,
  };
})();
