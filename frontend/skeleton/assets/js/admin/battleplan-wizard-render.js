// Render functions for the battleplan wizard. These functions take a `ctx`
// object containing { state, copy, dom, PREFIX } and read from it explicitly.
// No closure over outer state — call sites must build a fresh ctx and pass it.
(() => {
  window.appBattleplanWizard = window.appBattleplanWizard || {};
  const ns = window.appBattleplanWizard;
  if (!ns.config) return; // config module must load first

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

  function routineChevronSVG(collapsed) {
    if (collapsed) {
      return `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="wizard-routine-chevron" viewBox="0 0 16 16" aria-hidden="true"><path fill-rule="evenodd" d="M1.646 4.646a.5.5 0 0 1 .708 0L8 10.293l5.646-5.647a.5.5 0 0 1 .708.708l-6 6a.5.5 0 0 1-.708 0l-6-6a.5.5 0 0 1 0-.708"/></svg>`;
    }
    return `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="wizard-routine-chevron" viewBox="0 0 16 16" aria-hidden="true"><path fill-rule="evenodd" d="M7.646 4.646a.5.5 0 0 1 .708 0l6 6a.5.5 0 0 1-.708.708L8 5.707l-5.646 5.647a.5.5 0 0 1-.708-.708z"/></svg>`;
  }

  function highlightCurrentRoutines(dom) {
    const box = dom.stage && dom.stage.querySelector(".wizard-routines");
    if (!box) return;
    box.classList.add("wizard-routines-highlight");
    window.setTimeout(() => box.classList.remove("wizard-routines-highlight"), 1400);
  }

  function animateCollapseOtherRoutines(ctx, pillarKey, keepIdx) {
    const { dom, state, copy } = ctx;
    if (!dom.stage) return;
    const cards = dom.stage.querySelectorAll(`.wizard-routine[data-pillar="${pillarKey}"]`);
    cards.forEach((card) => {
      const idx = Number(card.getAttribute("data-routine-idx"));
      if (!Number.isFinite(idx) || idx === keepIdx) return;
      card.classList.add("is-collapsed");
      state.ui.collapsedRoutines[`${pillarKey}:${idx}`] = true;
      const trash = card.querySelector(".wizard-routine-delete");
      if (trash) trash.remove();
      const toggle = card.querySelector(".wizard-routine-toggle");
      if (toggle) {
        toggle.setAttribute("aria-label", copy.actions.expandRoutineAria);
        toggle.innerHTML = routineChevronSVG(true);
      }
    });
  }

  function renderIntro(ctx) {
    const { state, dom } = ctx;
    const intro = state.cfg.wizard.intro;
    const introText = esc(intro.text).replaceAll("\n", "<br>");
    dom.stage.innerHTML = `
      <section class="wizard-step wizard-step-intro">
        <div class="public-request-start-stack wizard-intro-stack">
          <h2 class="public-request-title public-request-title-start public-request-title-start-lg wizard-intro-title">${esc(intro.title)}</h2>
          <p class="public-request-copy wizard-intro-copy">${introText}</p>
        </div>
      </section>
    `;
  }

  function renderPriority(ctx) {
    const { state, dom, copy } = ctx;
    const priority = cfgmod.priorityCopy(state);
    const labels = {
      title: priority.title || "",
      priorityField: copy.labels.priorityField,
      whyField: copy.labels.whyField,
      priorityPlaceholder: copy.placeholders.priorityTitle,
      whyPlaceholder: copy.placeholders.priorityWhy,
    };
    const descriptionHTML = esc(priority.text || "").replaceAll("\n", "<br>");
    const introBlock = `
      <div class="intro-quote">
        <p>${descriptionHTML}</p>
      </div>
    `;
    dom.stage.innerHTML = `
      <section class="wizard-step wizard-step-priority">
        <header class="wizard-step-header">
          <h2 class="display-hero">${esc(labels.title)}</h2>
          ${introBlock}
        </header>

        <div>
          <label class="field-label">${esc(labels.priorityField)}</label>
          <input class="field-input${state.fieldErrors.priorityTitle ? " field-input-error" : ""}" type="text" name="priority_title" value="${esc(state.input.data.priority.title)}" placeholder="${esc(labels.priorityPlaceholder)}" />
        </div>

        <div>
          <label class="field-label">${esc(labels.whyField)}</label>
          <textarea class="field-input${state.fieldErrors.priorityWhy ? " field-input-error" : ""}" name="priority_why" rows="4" placeholder="${esc(labels.whyPlaceholder)}">${esc(state.input.data.priority.why)}</textarea>
        </div>
      </section>
    `;
  }

  function renderVisibilityAside(ctx) {
    const { state, dom } = ctx;
    if (!dom.aside) return;
    dom.aside.classList.remove("step-progress-aside-badges");
    dom.aside.classList.add("step-progress-aside-controls");
    const items = [...state.cfg.visibility].sort((a, b) => {
      if (a.value === "group" && b.value !== "group") return 1;
      if (b.value === "group" && a.value !== "group") return -1;
      return 0;
    });
    const opts = items.map((item) => `
      <label class="segmented-option">
        <input type="radio" name="visibility_aside" value="${esc(item.value)}" ${state.input.visibility === item.value ? "checked" : ""} />
        <span>${esc(item.label)}</span>
      </label>
    `).join("");
    dom.aside.innerHTML = `
      <div class="segmented step-progress-visibility-options">${opts}</div>
    `;
    dom.aside.hidden = false;
  }

  function renderEndDate(ctx) {
    const { state, dom, copy } = ctx;
    const customSelected = !!state.input.use_end_date;
    const customDate = state.input.end_date || "";
    const sharedCopy = (window.appCopy && window.appCopy.battleplan && window.appCopy.battleplan.wizardShared) || {};
    const titleText = sharedCopy.endDateTitle || copy.labels.durationField;
    const descText = sharedCopy.endDateDescription || "";
    const placeholder = sharedCopy.endDatePlaceholder || "";
    const introBlock = descText ? `<div class="intro-quote"><p>${esc(descText)}</p></div>` : "";
    const durationOpts = state.cfg.durations.map((item) => `
      <label class="wizard-end-date-preset segmented-option">
        <input type="radio" name="duration_choice" value="${item.value}" ${!customSelected && state.input.duration_days === item.value ? "checked" : ""} />
        <span>${item.value}${esc(copy.daysSuffix)}</span>
      </label>
    `).join("");
    dom.stage.innerHTML = `
      <section class="wizard-step wizard-step-end-date">
        <header class="wizard-step-header">
          <h2 class="display-hero">${esc(titleText)}</h2>
          ${introBlock}
        </header>
        <div class="wizard-end-date-presets segmented">${durationOpts}</div>
        <div class="wizard-end-date-custom segmented-option${customSelected ? "" : " is-empty"}">
          <input type="radio" name="duration_choice" value="custom" ${customSelected ? "checked" : ""} />
          <input type="text" name="duration_end_date" inputmode="numeric" autocomplete="off" placeholder="${esc(placeholder)}" maxlength="10" value="${esc(customDate)}" />
        </div>
      </section>
    `;
  }

  function renderRoutine(ctx, pillarKey, idx, routine) {
    const { state, copy } = ctx;
    const collapseKey = `${pillarKey}:${idx}`;
    const collapsed = Object.prototype.hasOwnProperty.call(state.ui.collapsedRoutines, collapseKey)
      ? !!state.ui.collapsedRoutines[collapseKey]
      : true;
    const routineTitle = String((routine && routine.title) || "").trim();
    const cadenceOpts = state.cfg.cadences.filter((c) => c && c.type !== "paused").map((c) => `
      <label class="segmented-option">
        <input type="radio" name="cadence_type_${pillarKey}_${idx}" value="${esc(c.type)}" ${routine.cadence.type === c.type ? "checked" : ""} />
        <span>${esc(c.label)}</span>
      </label>
    `).join("");
    const sharedCopy = (window.appCopy && window.appCopy.battleplan && window.appCopy.battleplan.wizardShared) || {};
    const cadenceReset = `
      <button type="button" class="segmented-option segmented-reset${routine.cadence.type === "paused" ? " is-selected" : ""}" data-action="pause-cadence" data-pillar="${esc(pillarKey)}" data-idx="${idx}" aria-label="${esc(sharedCopy.pauseRoutineAria || "")}">
        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" viewBox="0 0 16 16" aria-hidden="true">
          <path d="M6 3.5a.5.5 0 0 1 .5.5v8a.5.5 0 0 1-1 0V4a.5.5 0 0 1 .5-.5m4 0a.5.5 0 0 1 .5.5v8a.5.5 0 0 1-1 0V4a.5.5 0 0 1 .5-.5"/>
        </svg>
      </button>
    `;

    let extra = "";
    if (routine.cadence.type === "specific_days") {
      const days = routine.cadence.days || [];
      const dayLabels = [
        ["mon", copy.dayShort.mon],
        ["tue", copy.dayShort.tue],
        ["wed", copy.dayShort.wed],
        ["thu", copy.dayShort.thu],
        ["fri", copy.dayShort.fri],
        ["sat", copy.dayShort.sat],
        ["sun", copy.dayShort.sun],
      ];
      extra = `
        <div class="day-toggles">
          ${dayLabels.map(([key, label]) => `
            <label class="segmented-option day-toggle">
              <input type="checkbox" name="cadence_day_${pillarKey}_${idx}" value="${key}" ${days.includes(key) ? "checked" : ""} />
              <span>${label}</span>
            </label>
          `).join("")}
        </div>
      `;
    } else if (routine.cadence.type === "times_per_week") {
      extra = `
        <div class="day-toggles cadence-times-card">
          <label class="segmented-option day-toggle cadence-times-box">
            <input class="cadence-times-input" type="number" min="1" max="7" name="cadence_times_${pillarKey}_${idx}" value="${routine.cadence.times || 1}" />
          </label>
        </div>
      `;
    }

    return `
      <article class="wizard-routine${collapsed ? " is-collapsed" : ""}${routine.cadence.type === "paused" ? " is-paused" : ""}" data-routine-idx="${idx}" data-pillar="${esc(pillarKey)}">
        <header class="wizard-routine-hdr" data-action="toggle-routine" data-pillar="${esc(pillarKey)}" data-idx="${idx}">
          <div class="wizard-routine-left">
            <button type="button" class="wizard-routine-collapse" data-action="toggle-routine" data-pillar="${esc(pillarKey)}" data-idx="${idx}" aria-expanded="${collapsed ? "false" : "true"}">
              <span>${routine.cadence.type === "paused" ? "⏸ " : ""}${esc(copy.routineWord)} ${idx + 1}</span>
              ${collapsed && routineTitle ? `<span class="wizard-routine-collapsed-title">${esc(routineTitle)}</span>` : ""}
            </button>
          </div>
          <button type="button" class="wizard-routine-toggle" data-action="toggle-routine" data-pillar="${esc(pillarKey)}" data-idx="${idx}" aria-label="${collapsed ? esc(copy.actions.expandRoutineAria) : esc(copy.actions.collapseRoutineAria)}">
            ${routineChevronSVG(collapsed)}
          </button>
        </header>
        <div class="wizard-routine-body">
          <div>
            <label class="field-label">${esc(copy.labels.titleField)}</label>
            <input class="field-input" type="text" name="routine_title_${pillarKey}_${idx}" value="${esc(routine.title)}" placeholder="${esc(copy.placeholders.routineTitle)}" />
          </div>
          <div>
            <label class="field-label">${esc(copy.labels.triggerField)}</label>
            <input class="field-input" type="text" name="routine_trigger_${pillarKey}_${idx}" value="${esc(routine.trigger)}" placeholder="${esc(copy.placeholders.routineTrigger)}" />
          </div>
          <div>
            <label class="field-label">${esc(copy.labels.cadenceField)}</label>
            <div class="segmented segmented-cadence">${cadenceReset}${cadenceOpts}</div>
          </div>
          ${extra}
          <button type="button" class="wizard-btn wizard-btn-danger" data-action="remove-routine" data-pillar="${esc(pillarKey)}" data-idx="${idx}">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" viewBox="0 0 16 16" aria-hidden="true">
              <path d="M16 8A8 8 0 1 1 0 8a8 8 0 0 1 16 0M5.354 4.646a.5.5 0 1 0-.708.708L7.293 8l-2.647 2.646a.5.5 0 0 0 .708.708L8 8.707l2.646 2.647a.5.5 0 0 0 .708-.708L8.707 8l2.647-2.646a.5.5 0 0 0-.708-.708L8 7.293z"/>
            </svg>
            <span>${esc(sharedCopy.deleteRoutineButtonLabel || "")}</span>
          </button>
        </div>
      </article>
    `;
  }

  function renderPillar(ctx, pillarIndex) {
    const { state, dom, copy, ensurePillar } = ctx;
    const def = state.cfg.pillars[pillarIndex];
    const value = ensurePillar(def.key);
    const routinesHTML = value.routines.map((r, idx) => renderRoutine(ctx, def.key, idx, r)).join("");
    const sharedCopy = (window.appCopy && window.appCopy.battleplan && window.appCopy.battleplan.wizardShared) || {};
    const heroTitle = def.label || sharedCopy.pillarFallbackLabel || "";
    const introBlock = def.description ? `<cite class="intro-quote">${esc(def.description)}</cite>` : "";

    dom.stage.innerHTML = `
      <section class="wizard-step wizard-step-priority">
        <header class="wizard-step-header">
          <h2 class="display-hero">${esc(heroTitle)}</h2>
          ${introBlock}
        </header>

        <div>
          <label class="field-label">${esc(copy.labels.objectiveField)}</label>
          <textarea class="field-input" name="pillar_objective_${def.key}" rows="3" placeholder="${esc(copy.placeholders.objective)}">${esc(value.objective)}</textarea>
        </div>

        <div class="wizard-routines">${routinesHTML}</div>

        <button type="button" class="inline-add" data-action="add-routine" data-pillar="${esc(def.key)}">${esc(copy.actions.addRoutine)}</button>
      </section>
    `;
  }

  function dateOnly(value) {
    const raw = String(value || "").trim();
    if (!raw) return "";
    return raw.slice(0, 10);
  }

  function renderNotesHTML(notes, label) {
    if (!Array.isArray(notes) || notes.length === 0) return "";
    const items = notes.map((note) => {
      const noteText = String((note && note.text) || "").trim();
      if (!noteText) return "";
      const who = String((note && note.by) || "").trim();
      const when = dateOnly(note && note.at);
      const metaParts = [];
      if (when) {
        metaParts.push(`\u2014 ${when}`);
      }
      if (who) {
        metaParts.push(who);
      }
      const meta = metaParts.join(" \u2022 ");
      return `
        <div class="battleplan-note-item">
          <span class="battleplan-note-marker" aria-hidden="true"></span>
          <p class="battleplan-note-text">
            ${esc(noteText)}
            ${meta ? `<span class="battleplan-note-meta">${esc(meta)}</span>` : ""}
          </p>
        </div>
      `;
    }).filter(Boolean).join("");
    return items ? `
      <div>
        <label class="field-label wizard-summary-tag">${esc(label || "")}</label>
        <div class="battleplan-notes-history">${items}</div>
      </div>
    ` : "";
  }

  function renderConfirm(ctx) {
    const { state, dom, copy, PREFIX, isSummaryView } = ctx;
    const c = state.cfg.wizard.confirmation;
    const sharedCopy = (window.appCopy && window.appCopy.battleplan && window.appCopy.battleplan.wizardShared) || {};
    const priorityText = (state.input.data.priority.title || "").trim();
    const priorityWhy = (state.input.data.priority.why || "").trim();
    const priorityMissing = !priorityText;
    const summaryEditId = state.viewId || state.editId;
    const hasCompletePillar = state.cfg.pillars.some((def) => cfgmod.isPillarComplete(state.input.data.pillars[def.key]));
    const pillarsSummary = state.cfg.pillars.map((def, pillarIndex) => {
      const value = state.input.data.pillars[def.key] || { objective: "", routines: [] };
      const objective = (value.objective || "").trim();
      const routines = (value.routines || []);
      const routineItems = routines
        .map((r) => {
          if (!r || !r.title) return "";
          if ((r.cadence || {}).type === "paused") return "";
          const title = String(r.title).trim();
          if (!title) return "";
          const cadence = r.cadence || {};
          const cadenceLabel = cfgmod.routineCadenceLabel(cadence);
          return `<li><span>${esc(title)}</span>${cadenceLabel ? ` <span class="wizard-routine-count-label">${esc(cadenceLabel)}</span>` : ""}</li>`;
        })
        .filter((line) => !!line)
        .join("");
      return `
        <li class="wizard-pillars-item wizard-summary-edit-target" data-action="${summaryEditId ? "edit-pillar" : "go-step"}" data-step="${summaryEditId ? pillarIndex + 1 : pillarIndex + 2}">
          <div class="wizard-pillars-col wizard-pillars-col-label"><strong>${esc(def.label)}</strong></div>
          <div class="wizard-pillars-col wizard-pillars-col-content">
            ${objective ? `<p class="wizard-pillars-objective-inline">${esc(objective)}</p>` : ""}
            ${routineItems ? `<ul class="wizard-summary-muted-card wizard-pillars-routines">${routineItems}</ul>` : ""}
          </div>
        </li>
      `;
    }).join("");

    let topErrorMessage = "";
    if (state.confirmAttempted) {
      if (priorityMissing) {
        topErrorMessage = copy.errors.priorityRequired;
      } else {
        for (const def of state.cfg.pillars) {
          const v = state.input.data.pillars[def.key] || { objective: "", routines: [] };
          const obj = (v.objective || "").trim();
          const validRoutine = (v.routines || []).some((r) => cfgmod.isRoutineComplete(r));
          if (!obj && !validRoutine) continue;
          if (!obj) { topErrorMessage = copy.errors.pillarObjectiveRequired; break; }
          if (!validRoutine) { topErrorMessage = copy.errors.pillarRoutineRequired; break; }
        }
        if (!topErrorMessage && !hasCompletePillar) {
          topErrorMessage = copy.errors.completePillarRequired;
        }
      }
    }

    if (dom.aside) {
      dom.aside.classList.add("step-progress-aside-badges");
      dom.aside.classList.remove("step-progress-aside-controls");
      dom.aside.innerHTML = `
        <span class="wizard-confirm-badge">${state.input.duration_days}${esc(copy.daysSuffix)}</span>
        <span class="wizard-confirm-badge">${esc(cfgmod.visibilityLabel(state.input.visibility, state.cfg)).toUpperCase()}</span>
      `;
      dom.aside.hidden = false;
    }

    const summaryView = isSummaryView();
    const heroHTML = summaryView
      ? `<h2 class="display-hero">${esc(priorityText)}</h2>`
      : (c.text ? `<h2 class="display-hero">${esc(c.text)}</h2>` : "");
    const priorityBodyHTML = summaryView
      ? (priorityWhy ? `<p class="wizard-view-why">${esc(priorityWhy)}</p>` : "")
      : (priorityMissing
          ? `<p class="wizard-missing-line">${esc(sharedCopy.missingPriorityLabel || "")}</p>`
          : `<p class="wizard-summary-value">${esc(priorityText)}</p>${priorityWhy ? `<p class="wizard-summary-muted-card wizard-summary-why">${esc(priorityWhy)}</p>` : ""}`);
    const priorityTargetAttrs = summaryView
      ? ""
      : ` class="wizard-summary-edit-target" data-action="go-step" data-step="1"`;
    const viewCopy = (window.appCopy && window.appCopy.battleplan && window.appCopy.battleplan.view) || {};
    const notesHTML = summaryView ? renderNotesHTML(state.input.data.notes, viewCopy.notesLabel) : "";

    const topErrorSlot = document.getElementById(`${PREFIX}-top-error`);
    if (topErrorSlot) {
      const errorPrefix = sharedCopy.errorPrefix || "";
      topErrorSlot.textContent = topErrorMessage ? `${errorPrefix}: ${topErrorMessage}` : "";
    }

    dom.stage.innerHTML = `
      <section class="wizard-step wizard-step-confirm${summaryView ? " wizard-step-view" : ""}">
        <header class="wizard-step-header">
          ${heroHTML}
        </header>

        <div class="wizard-summary">
          <div${priorityTargetAttrs}>
            <div class="wizard-summary-tags">
              <label class="field-label wizard-summary-tag">${esc(copy.labels.summaryPriority)}</label>
              <label class="field-label wizard-summary-tag wizard-summary-tag-outline">${esc((sharedCopy.summaryDurationDays || "").replace("{count}", String(state.input.duration_days)))}</label>
              <label class="field-label wizard-summary-tag wizard-summary-tag-outline">${esc(cfgmod.visibilityLabel(state.input.visibility, state.cfg)).toUpperCase()}</label>
              ${cfgmod.willSaveAsDraft(state) ? `<label class="field-label wizard-summary-tag wizard-summary-tag-draft">${esc(sharedCopy.summaryDraftBadge || "")}</label>` : ""}
            </div>
            ${priorityBodyHTML}
          </div>
          <div class="wizard-summary-pillars-block">
            <label class="field-label wizard-summary-tag">${esc(copy.labels.summaryPillars)}</label>
            <ul class="wizard-pillars-summary">${pillarsSummary}</ul>
          </div>
          ${notesHTML}
        </div>
      </section>
    `;
  }

  ns.render = {
    esc,
    setNextIcon,
    routineChevronSVG,
    highlightCurrentRoutines,
    animateCollapseOtherRoutines,
    renderIntro,
    renderPriority,
    renderEndDate,
    renderVisibilityAside,
    renderRoutine,
    renderPillar,
    renderConfirm,
  };
})();
