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
      return `
        <label class="segmented-option">
          <input type="radio" name="event_type" value="${esc(t.value)}" ${selected ? "checked" : ""} />
          <span>${esc(label)}</span>
        </label>
      `;
    }).join("");
    const empty = !types.length
      ? `<p class="meta-text">${esc(copy.errors.noAllowedTypes || "")}</p>`
      : "";
    dom.stage.innerHTML = `
      <section class="wizard-step">
        <header class="wizard-step-header">
          <h2 class="display-hero">${esc(stepCopy.title || "")}</h2>
        </header>
        <div class="segmented">
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
        <label class="segmented-option">
          <input type="radio" name="event_group" value="${esc(g.id)}" ${selected ? "checked" : ""} />
          <span>${esc(g.name || g.id)}</span>
        </label>
      `;
    }).join("");
    dom.stage.innerHTML = `
      <section class="wizard-step">
        <header class="wizard-step-header">
          <h2 class="display-hero">${esc(stepCopy.title || "")}</h2>
        </header>
        <div class="segmented">
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
    const timeRequired = cfgmod.required(typeDef, "time");
    const schedule = state.input.schedule || "once";
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
      <div>
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
      <section class="wizard-step">
        <header class="wizard-step-header">
          <h2 class="display-hero">${esc(stepCopy.title || "")}</h2>
        </header>

        <div>
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

        <div>
          <label class="field-label" for="${ctx.PREFIX}-startdate">${esc((copy.labels && copy.labels.startDateField) || "")}</label>
          <input class="field-input${state.fieldErrors.start_date ? " field-input-error" : ""}" id="${ctx.PREFIX}-startdate" type="text" name="event_start_date" value="${esc(state.input.startDate)}" placeholder="YYYY-MM-DD" inputmode="numeric" autocomplete="off" autocapitalize="off" spellcheck="false" maxlength="10" />
        </div>

        ${timeRequired ? `
          <div>
            <label class="field-label" for="${ctx.PREFIX}-starttime">${esc((copy.labels && copy.labels.startTimeField) || "")}</label>
            <input class="field-input${state.fieldErrors.time ? " field-input-error" : ""}" id="${ctx.PREFIX}-starttime" type="text" name="event_start_time" value="${esc(state.input.startTime || "")}" placeholder="HH:MM" inputmode="numeric" autocomplete="off" autocapitalize="off" spellcheck="false" maxlength="5" />
          </div>
        ` : ""}

        ${endDateRequired ? `
          <div>
            <label class="field-label" for="${ctx.PREFIX}-enddate">${esc((copy.labels && copy.labels.endDateField) || "")}</label>
            <input class="field-input${state.fieldErrors.end_date ? " field-input-error" : ""}" id="${ctx.PREFIX}-enddate" type="text" name="event_end_date" value="${esc(state.input.endDate)}" placeholder="YYYY-MM-DD" inputmode="numeric" autocomplete="off" autocapitalize="off" spellcheck="false" maxlength="10" />
          </div>
        ` : ""}

        ${weeklySection}
        ${monthlySection}
        ${repeatedSection}
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
    const descriptionRequired = cfgmod.required(typeDef, "description");
    const descRaw = (state.input.data && state.input.data.description) || "";
    dom.stage.innerHTML = `
      <section class="wizard-step">
        <header class="wizard-step-header">
          <h2 class="display-hero">${esc(stepCopy.title || "")}</h2>
        </header>

        ${titleRequired ? `
          <div>
            <label class="field-label" for="${ctx.PREFIX}-title">${esc((copy.labels && copy.labels.titleField) || "")}</label>
            <input class="field-input${state.fieldErrors.title ? " field-input-error" : ""}" id="${ctx.PREFIX}-title" type="text" name="event_title" value="${esc(state.input.title)}" placeholder="${esc((copy.placeholders && copy.placeholders.title) || "")}" />
          </div>
        ` : ""}

        ${locationRequired ? `
          <div>
            <label class="field-label" for="${ctx.PREFIX}-location">${esc((copy.labels && copy.labels.locationField) || "")}</label>
            <input class="field-input${state.fieldErrors.location ? " field-input-error" : ""}" id="${ctx.PREFIX}-location" type="text" name="event_location" value="${esc(state.input.location)}" placeholder="${esc((copy.placeholders && copy.placeholders.location) || "")}" />
          </div>
        ` : ""}

        ${urlRequired ? `
          <div>
            <label class="field-label" for="${ctx.PREFIX}-url">${esc((copy.labels && copy.labels.urlField) || "")}</label>
            <input class="field-input${state.fieldErrors.url ? " field-input-error" : ""}" id="${ctx.PREFIX}-url" type="url" name="event_url" value="${esc(state.input.url)}" placeholder="${esc((copy.placeholders && copy.placeholders.url) || "")}" />
          </div>
        ` : ""}

        ${descriptionRequired ? `
          <div>
            <label class="field-label" for="${ctx.PREFIX}-desc">${esc((copy.labels && copy.labels.descriptionField) || "")}</label>
            <textarea class="field-input${state.fieldErrors.description ? " field-input-error" : ""}" id="${ctx.PREFIX}-desc" name="event_description" rows="4" placeholder="${esc((copy.placeholders && copy.placeholders.description) || "")}">${esc(descRaw)}</textarea>
          </div>
        ` : ""}
      </section>
    `;
  }

  function renderConfirm(ctx) {
    const { state, dom, copy, PREFIX } = ctx;
    const stepCopy = (copy.steps && copy.steps.confirm) || {};
    const td = cfgmod.typeDef(state.cfg, state.input.type) || {};
    const typeLabel = td.label || typeOptionLabel(state.input.type, copy);
    const previewDates = cfgmod.generateScheduleDates(state.input.startDate, cfgmod.cadenceFromInput(state.input), state.input.count);
    const dateTags = previewDates.map((d) => `<span class="wizard-summary-tag wizard-summary-tag-outline">${esc(d)}</span>`).join("");
    const groupName = state.input.groupName || "";
    const desc = (state.input.data && state.input.data.description) || "";
    const title = (state.input.title || "").trim();
    const location = (state.input.location || "").trim();
    const url = (state.input.url || "").trim();

    dom.stage.innerHTML = `
      <section class="wizard-step wizard-step-confirm">
        <header class="wizard-step-header">
          <h2 class="display-hero">${esc(stepCopy.title || "")}</h2>
        </header>
        <div class="wizard-summary">
          <div>
            <label class="field-label wizard-summary-tag">${esc((copy.labels && copy.labels.summaryType) || "Type")}</label>
            <p class="wizard-summary-value">${esc(typeLabel)}</p>
          </div>
          ${groupName ? `
            <div>
              <label class="field-label wizard-summary-tag">${esc((copy.labels && copy.labels.summaryGroup) || "Group")}</label>
              <p class="wizard-summary-value">${esc(groupName)}</p>
            </div>
          ` : ""}
          <div>
            <label class="field-label wizard-summary-tag">${esc((copy.labels && copy.labels.summaryDates) || "Dates")}</label>
            ${dateTags ? `<div class="wizard-summary-tags">${dateTags}</div>` : `<p class="wizard-summary-value">—</p>`}
          </div>
          ${title ? `
            <div>
              <label class="field-label wizard-summary-tag">${esc((copy.labels && copy.labels.summaryTitle) || "Title")}</label>
              <p class="wizard-summary-value">${esc(title)}</p>
            </div>
          ` : ""}
          ${location ? `
            <div>
              <label class="field-label wizard-summary-tag">${esc((copy.labels && copy.labels.summaryLocation) || "Location")}</label>
              <p class="wizard-summary-value">${esc(location)}</p>
            </div>
          ` : ""}
          ${url ? `
            <div>
              <label class="field-label wizard-summary-tag">${esc((copy.labels && copy.labels.summaryURL) || "URL")}</label>
              <p class="wizard-summary-value"><a href="${esc(url)}" target="_blank" rel="noopener noreferrer">${esc(url)}</a></p>
            </div>
          ` : ""}
          ${desc ? `
            <div>
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
