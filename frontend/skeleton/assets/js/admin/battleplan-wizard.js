(() => {
  const auth = window.huuperAuth;
  if (!auth) return;
  const copy = window.huuperCopy && window.huuperCopy.battleplan && window.huuperCopy.battleplan.wizard;
  if (!copy) return;

  const PREFIX = "battleplan-wizard";
  const DRAFT_STORAGE_KEY = "battleplan-wizard-draft-v1";
  const $ = (suffix) => document.getElementById(`${PREFIX}-${suffix}`);

  const dom = {
    page: document.querySelector(".wizard-page"),
    progress: document.querySelector(".step-progress"),
    progressLabel: $("progress-label"),
    progressState: $("progress-state"),
    progressFill: $("progress-fill"),
    aside: $("aside"),
    form: $("form"),
    stage: $("stage"),
    status: $("status"),
    loading: $("loading"),
    sticky: $("sticky"),
    back: $("back"),
    nextLabel: $("next-label"),
  };
  const cancelEditButton = document.querySelector("[data-battleplan-cancel-edit]");

  const editId = new URLSearchParams(window.location.search).get("edit") || null;
  const viewId = new URLSearchParams(window.location.search).get("view") || null;

  const state = {
    cfg: null,
    editId,
    viewId,
    hasExistingBattleplan: false,
    step: 0,
    confirmAttempted: false,
    submitting: false,
    ui: {
      collapsedRoutines: {},
      focusRoutine: null,
    },
    input: {
      start_date: new Date().toISOString().slice(0, 10),
      duration_days: 30,
      visibility: "",
      data: {
        priority: { title: "", why: "" },
        pillars: {},
      },
    },
  };

  function esc(value) {
    return String(value || "")
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;")
      .replaceAll("'", "&#39;");
  }

  function saveDraft() {
    try {
      const payload = {
        input: state.input,
        step: state.step,
      };
      window.localStorage.setItem(DRAFT_STORAGE_KEY, JSON.stringify(payload));
    } catch (_) {
      // no-op: localStorage unavailable
    }
  }

  function clearDraft() {
    try {
      window.localStorage.removeItem(DRAFT_STORAGE_KEY);
    } catch (_) {
      // no-op: localStorage unavailable
    }
  }

  function loadDraft() {
    try {
      const raw = window.localStorage.getItem(DRAFT_STORAGE_KEY);
      if (!raw) return null;
      const parsed = JSON.parse(raw);
      if (!parsed || typeof parsed !== "object") return null;
      return parsed;
    } catch (_) {
      return null;
    }
  }

  function setStatus(msg) {
    if (!dom.status) return;
    dom.status.textContent = msg || "";
    dom.status.hidden = !msg;
    dom.status.classList.toggle("is-error", !!msg);
  }

  function setLoading(visible) {
    dom.loading.hidden = !visible;
  }

  function sleep(ms) {
    return new Promise((resolve) => window.setTimeout(resolve, ms));
  }

  async function pulseNextLoading(ms = 220) {
    if (!dom.nextLabel) return;
    dom.nextLabel.classList.add("request-action-loading");
    await sleep(ms);
    dom.nextLabel.classList.remove("request-action-loading");
  }

  function showWizard() {
    setLoading(false);
    dom.form.hidden = false;
    dom.sticky.hidden = false;
  }

  function totalSteps() {
    return 3 + state.cfg.pillars.length; // intro + priority + N pillars + confirm
  }

  function introEnabledBySettings() {
    const intro = state.cfg && state.cfg.wizard && state.cfg.wizard.intro;
    if (!intro || typeof intro !== "object") return true;
    if (typeof intro.show === "boolean") return intro.show;
    return true;
  }

  function shouldShowIntro() {
    return introEnabledBySettings() && !state.hasExistingBattleplan;
  }

  function defaultDurationValue() {
    const byDefault = state.cfg.durations.find((item) => item && item.default);
    if (byDefault && Number.isFinite(byDefault.value)) return Number(byDefault.value);
    const first = state.cfg.durations.find((item) => item && Number.isFinite(item.value));
    return first ? Number(first.value) : 0;
  }

  function defaultVisibilityValue() {
    const byDefault = state.cfg.visibility.find((item) => item && item.default);
    if (byDefault && byDefault.value) return byDefault.value;
    const first = state.cfg.visibility.find((item) => item && item.value);
    return first ? first.value : "";
  }

  function defaultCadenceType() {
    const byDefault = state.cfg.cadences.find((item) => item && item.default && item.type);
    if (byDefault) return byDefault.type;
    const first = state.cfg.cadences.find((item) => item && item.type);
    return first ? first.type : "daily";
  }

  function minStep() {
    if (state.viewId) return totalSteps() - 1;
    if (state.editId) return 2; // pillars only; priority not editable via edit page
    return shouldShowIntro() ? 0 : 1;
  }

  function maxStep() {
    if (state.viewId) return totalSteps() - 1;
    return totalSteps() - 1;
  }

  function clampStep(step) {
    if (!Number.isFinite(step)) return minStep();
    return Math.max(minStep(), Math.min(maxStep(), step));
  }

  function isPillarStep() {
    return state.step >= 2 && state.step <= state.cfg.pillars.length + 1;
  }

  function internalStepFromURLStep(urlStep) {
    if (!Number.isFinite(urlStep)) return null;
    if (state.viewId) return null;
    if (urlStep < 1 || urlStep > state.cfg.pillars.length) return null;
    return urlStep + 1;
  }

  function urlStepFromInternalStep() {
    if (!isPillarStep()) return null;
    return state.step - 1;
  }

  function readStepFromURL() {
    const url = new URL(window.location.href);
    const raw = url.searchParams.get("step");
    if (!raw) return null;
    const parsed = Number.parseInt(raw, 10);
    if (!Number.isFinite(parsed)) return null;
    return internalStepFromURLStep(parsed);
  }

  function writeStepToURL() {
    const url = new URL(window.location.href);
    const urlStep = urlStepFromInternalStep();
    if (urlStep === null) {
      url.searchParams.delete("step");
    } else {
      url.searchParams.set("step", String(urlStep));
    }
    window.history.replaceState({}, "", url.toString());
  }

  function pad2(n) { return String(n).padStart(2, "0"); }

  function isConfirm() {
    return state.step === totalSteps() - 1;
  }

  function isSummaryView() {
    return isConfirm() && (state.viewId || state.editId);
  }

  function stateLabel() {
    if (isConfirm()) return state.cfg.wizard.confirmation.title.toUpperCase();
    if (state.step === 0) return state.cfg.wizard.intro.title.toUpperCase();
    if (state.step === 1) return state.cfg.priority.label.toUpperCase();
    const def = state.cfg.pillars[state.step - 2];
    return def ? def.label.toUpperCase() : "";
  }

  function syncProgress() {
    const n = state.cfg.pillars.length;
    const cur = isPillarStep() ? state.step - 1 : 0;
    dom.progressLabel.textContent = `${copy.progressPrefix}: ${pad2(cur)}/${pad2(n)}`;
    dom.progressState.textContent = stateLabel();
    dom.progressFill.style.width = `${(cur / n) * 100}%`;
    if (isConfirm()) {
      dom.nextLabel.textContent = state.cfg.wizard.confirmation.button.toUpperCase();
      return;
    }
    dom.nextLabel.textContent = state.step === 0 ? state.cfg.wizard.intro.button.toUpperCase() : copy.nextLabel.toUpperCase();
  }

  function ensurePillar(key) {
    if (!state.input.data.pillars[key]) {
      state.input.data.pillars[key] = { objective: "", routines: [] };
    }
    return state.input.data.pillars[key];
  }

  function newRoutine() {
    return { title: "", trigger: "", cadence: { type: defaultCadenceType() } };
  }

  function syncChrome() {
    const onIntro = state.step === 0 && shouldShowIntro();
    const onConfirm = isConfirm();
    const onSummaryView = isSummaryView();
    if (dom.page) dom.page.classList.toggle("wizard-page-intro", onIntro);
    if (dom.page) dom.page.classList.toggle("wizard-page-confirm", onConfirm && !onSummaryView);
    if (dom.page) dom.page.classList.toggle("wizard-page-has-progress", !onIntro && !onConfirm);
    if (dom.stage) dom.stage.classList.toggle("wizard-stage-confirm", onConfirm && !onSummaryView);
    if (dom.stage) dom.stage.classList.toggle("wizard-stage-view", onSummaryView);
    if (dom.form) dom.form.classList.toggle("wizard-form-confirm", onConfirm && !onSummaryView);
    if (dom.form) {
      dom.form.classList.toggle("wizard-form-view", onSummaryView);
      if (!onSummaryView) dom.form.classList.remove("wizard-form-view-overflow");
    }
    if (dom.progress) dom.progress.hidden = onIntro || isConfirm();
    if (dom.aside) dom.aside.hidden = onIntro || isConfirm();
    if (dom.back) dom.back.hidden = onIntro;
    if (dom.sticky) dom.sticky.classList.toggle("sticky-actions-intro", onIntro);
  }

  // ---------- Step rendering ----------

  function renderIntro() {
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

  function renderPriority() {
    const priority = state.cfg.priority;
    const labels = {
      title: priority.label,
      priorityField: copy.labels.priorityField,
      whyField: copy.labels.whyField,
      durationField: copy.labels.durationField,
      priorityPlaceholder: copy.placeholders.priorityTitle,
      whyPlaceholder: copy.placeholders.priorityWhy,
    };
    const descriptionHTML = esc(priority.description || "").replaceAll("\n", "<br>");
    const introBlock = `
      <div class="intro-quote">
        <p>${descriptionHTML}</p>
      </div>
    `;
    const durationOpts = state.cfg.durations.map((d) => `
      <label class="segmented-option">
        <input type="radio" name="duration" value="${d.value}" ${state.input.duration_days === d.value ? "checked" : ""} />
        <span>${d.value} ${copy.daysSuffix}</span>
      </label>
    `).join("");

    dom.stage.innerHTML = `
      <section class="wizard-step wizard-step-priority">
        <header class="wizard-step-header">
          <h2 class="display-hero">${esc(labels.title)}</h2>
          ${introBlock}
        </header>

        <div>
          <label class="field-label">${esc(labels.priorityField)}</label>
          <input class="field-input" type="text" name="priority_title" value="${esc(state.input.data.priority.title)}" placeholder="${esc(labels.priorityPlaceholder)}" />
        </div>

        <div>
          <label class="field-label">${esc(labels.whyField)}</label>
          <textarea class="field-input" name="priority_why" rows="4" placeholder="${esc(labels.whyPlaceholder)}">${esc(state.input.data.priority.why)}</textarea>
        </div>

        <div>
          <label class="field-label">${esc(labels.durationField)}</label>
          <div class="segmented">${durationOpts}</div>
        </div>
      </section>
    `;
  }

  function renderVisibilityAside() {
    if (!dom.aside) return;
    dom.aside.classList.remove("step-progress-aside-badges");
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
    dom.aside.innerHTML = `<div class="segmented">${opts}</div>`;
    dom.aside.hidden = false;
  }

  function highlightCurrentRoutines() {
    const box = dom.stage && dom.stage.querySelector(".wizard-routines");
    if (!box) return;
    box.classList.add("wizard-routines-highlight");
    window.setTimeout(() => box.classList.remove("wizard-routines-highlight"), 1400);
  }

  function routineChevronSVG(collapsed) {
    if (collapsed) {
      return `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="wizard-routine-chevron" viewBox="0 0 16 16" aria-hidden="true"><path fill-rule="evenodd" d="M1.646 4.646a.5.5 0 0 1 .708 0L8 10.293l5.646-5.647a.5.5 0 0 1 .708.708l-6 6a.5.5 0 0 1-.708 0l-6-6a.5.5 0 0 1 0-.708"/></svg>`;
    }
    return `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="wizard-routine-chevron" viewBox="0 0 16 16" aria-hidden="true"><path fill-rule="evenodd" d="M7.646 4.646a.5.5 0 0 1 .708 0l6 6a.5.5 0 0 1-.708.708L8 5.707l-5.646 5.647a.5.5 0 0 1-.708-.708z"/></svg>`;
  }

  function animateCollapseOtherRoutines(pillarKey, keepIdx) {
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

  function renderRoutine(pillarKey, idx, routine) {
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
    const cadenceReset = `
      <button type="button" class="segmented-option segmented-reset${routine.cadence.type === "paused" ? " is-selected" : ""}" data-action="pause-cadence" data-pillar="${esc(pillarKey)}" data-idx="${idx}" aria-label="Routine in pausa">
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
            <span>delete</span>
          </button>
        </div>
      </article>
    `;
  }

  function renderPillar(pillarIndex) {
    const def = state.cfg.pillars[pillarIndex];
    const value = ensurePillar(def.key);
    const routinesHTML = value.routines.map((r, idx) => renderRoutine(def.key, idx, r)).join("");
    const heroTitle = def.label || "Interiorita";
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

  function visibilityLabel(value) {
    const def = state.cfg.visibility.find((v) => v.value === value);
    return def ? def.label : value;
  }

  function routineCadenceLabel(cadence) {
    if (!cadence || !cadence.type) return "";
    if (cadence.type === "paused") return "";
    const cadenceTimes = Number(cadence.times);
    if (cadence.type === "times_per_week" && Number.isFinite(cadenceTimes) && cadenceTimes > 0) {
      return `${cadenceTimes}X`;
    }
    if (cadence.type === "specific_days") {
      const dayLabels = {
        mon: "M",
        tue: "T",
        wed: "W",
        thu: "T",
        fri: "F",
        sat: "S",
        sun: "S",
      };
      const days = Array.isArray(cadence.days) ? cadence.days : [];
      const ordered = ["mon", "tue", "wed", "thu", "fri", "sat", "sun"]
        .filter((day) => days.includes(day))
        .map((day) => dayLabels[day]);
      return ordered.length ? `(${ordered.join(",")})` : "";
    }
    if (cadence.type === "daily") return "W";
    return "";
  }

  function isRoutineComplete(routine) {
    if (!routine) return false;
    if (!(routine.title || "").trim()) return false;
    if (!(routine.trigger || "").trim()) return false;
    const cadence = routine.cadence || {};
    if (!cadence.type || cadence.type === "paused") return false;
    if (cadence.type === "specific_days") return Array.isArray(cadence.days) && cadence.days.length > 0;
    if (cadence.type === "times_per_week") {
      const times = Number(cadence.times);
      return Number.isFinite(times) && times >= 1 && times <= 7;
    }
    return true;
  }

  function hasPillarContent(value) {
    if (!value) return false;
    if ((value.objective || "").trim()) return true;
    return (value.routines || []).some((r) => {
      if (!r) return false;
      return !!((r.title || "").trim() || (r.trigger || "").trim() || (r.cadence && r.cadence.type && r.cadence.type !== "paused"));
    });
  }

  function isPillarComplete(value) {
    if (!value || !(value.objective || "").trim()) return false;
    return (value.routines || []).some((r) => isRoutineComplete(r));
  }

  function renderConfirm() {
    const c = state.cfg.wizard.confirmation;
    const priorityText = (state.input.data.priority.title || "").trim();
    const priorityWhy = (state.input.data.priority.why || "").trim();
    const priorityMissing = !priorityText;
    const summaryEditId = state.viewId || state.editId;
    const hasCompletePillar = state.cfg.pillars.some((def) => isPillarComplete(state.input.data.pillars[def.key]));
    const pillarsSummary = state.cfg.pillars.map((def, pillarIndex) => {
      const value = state.input.data.pillars[def.key] || { objective: "", routines: [] };
      const objective = (value.objective || "").trim();
      const routines = (value.routines || []);
      const missing = hasPillarContent(value) && !isPillarComplete(value);
      const routineItems = routines
        .map((r) => {
          if (!r || !r.title) return "";
          if ((r.cadence || {}).type === "paused") return "";
          const title = String(r.title).trim();
          if (!title) return "";
          const cadence = r.cadence || {};
          const cadenceLabel = routineCadenceLabel(cadence);
          return `<li><span>${esc(title)}</span>${cadenceLabel ? ` <span class="wizard-routine-count-label">${esc(cadenceLabel)}</span>` : ""}</li>`;
        })
        .filter((line) => !!line)
        .join("");
      return `
        <li class="wizard-pillars-item wizard-summary-edit-target" data-action="${summaryEditId ? "edit-pillar" : "go-step"}" data-step="${summaryEditId ? pillarIndex + 1 : pillarIndex + 2}">
          <div class="wizard-pillars-col wizard-pillars-col-label"><strong>${esc(def.label)}</strong></div>
          <div class="wizard-pillars-col wizard-pillars-col-content">
            ${objective ? `<p class="wizard-pillars-objective-inline">${esc(objective)}</p>` : ""}
            <ul class="wizard-summary-muted-card wizard-pillars-routines">${routineItems || "<li><span>-</span></li>"}</ul>
          </div>
          ${missing
            ? `<p class="wizard-missing-line${state.confirmAttempted ? " is-error" : ""}">${state.confirmAttempted ? "Error: Missing Objective" : "Missing Objective"}</p>`
            : ""}
        </li>
      `;
    }).join("");

    if (dom.aside) {
      dom.aside.classList.add("step-progress-aside-badges");
      dom.aside.innerHTML = `
        <span class="wizard-confirm-badge">${state.input.duration_days}${esc(copy.daysSuffix)}</span>
        <span class="wizard-confirm-badge">${esc(visibilityLabel(state.input.visibility)).toUpperCase()}</span>
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
          ? `<p class="wizard-missing-line${state.confirmAttempted ? " is-error" : ""}">${state.confirmAttempted ? "Error: Missing Priority" : "Missing Priority"}</p>`
          : `<p class="wizard-summary-value">${esc(priorityText)}</p>${priorityWhy ? `<p class="wizard-summary-muted-card wizard-summary-why">${esc(priorityWhy)}</p>` : ""}`);
    const priorityTargetAttrs = summaryView
      ? ""
      : ` class="wizard-summary-edit-target" data-action="go-step" data-step="1"`;

    dom.stage.innerHTML = `
      <section class="wizard-step wizard-step-confirm${summaryView ? " wizard-step-view" : ""}">
        <header class="wizard-step-header">
          ${heroHTML}
        </header>

        <div class="wizard-summary">
          <div${priorityTargetAttrs}>
            <div class="wizard-summary-tags">
              <label class="field-label wizard-summary-tag">${esc(copy.labels.summaryPriority)}</label>
              <label class="field-label wizard-summary-tag wizard-summary-tag-outline">${state.input.duration_days} days</label>
              <label class="field-label wizard-summary-tag wizard-summary-tag-outline">${esc(visibilityLabel(state.input.visibility)).toUpperCase()}</label>
            </div>
            ${priorityBodyHTML}
          </div>
          <div class="wizard-summary-pillars-block">
            <label class="field-label wizard-summary-tag">${esc(copy.labels.summaryPillars)}</label>
            ${state.confirmAttempted && !hasCompletePillar ? `<p class="wizard-missing-line is-error">Error: ${esc(copy.errors.completePillarRequired)}</p>` : ""}
            <ul class="wizard-pillars-summary">${pillarsSummary}</ul>
          </div>
        </div>
      </section>
    `;
  }

  function hasConfirmMissing() {
    const priorityMissing = !(state.input.data.priority.title || "").trim();
    if (priorityMissing) return true;
    let hasCompletePillar = false;
    for (const def of state.cfg.pillars) {
      const value = state.input.data.pillars[def.key] || { objective: "", routines: [] };
      if (isPillarComplete(value)) {
        hasCompletePillar = true;
        continue;
      }
      if (hasPillarContent(value)) return true;
    }
    return !hasCompletePillar;
  }

  function render() {
    syncChrome();
    syncProgress();
    writeStepToURL();
    setStatus("");
    if (state.step === 0) {
      renderIntro();
    } else if (state.step === 1) {
      renderVisibilityAside();
      renderPriority();
    } else if (state.step <= state.cfg.pillars.length + 1) {
      renderVisibilityAside();
      renderPillar(state.step - 2);
    } else {
      renderConfirm();
    }
    if (state.ui.focusRoutine && dom.stage) {
      const { pillarKey, idx } = state.ui.focusRoutine;
      const input = dom.stage.querySelector(`[name="routine_title_${pillarKey}_${idx}"]`);
      if (input) {
        input.scrollIntoView({ behavior: "smooth", block: "center" });
        input.focus({ preventScroll: true });
      }
      state.ui.focusRoutine = null;
    }
    syncViewOverflow();
  }

  function syncViewOverflow() {
    if (!isSummaryView() || !dom.form) return;
    window.requestAnimationFrame(() => {
      const needsBottomSpace = document.documentElement.scrollHeight > window.innerHeight;
      dom.form.classList.toggle("wizard-form-view-overflow", needsBottomSpace);
    });
  }

  // ---------- Form sync ----------

  function syncFromDOM() {
    const visAside = dom.aside && dom.aside.querySelector('[name="visibility_aside"]:checked');
    if (visAside) state.input.visibility = visAside.value;
    if (state.step === 1) {
      const titleEl = dom.stage.querySelector('[name="priority_title"]');
      const whyEl = dom.stage.querySelector('[name="priority_why"]');
      const dur = dom.stage.querySelector('[name="duration"]:checked');
      if (titleEl) state.input.data.priority.title = titleEl.value.trim();
      if (whyEl) state.input.data.priority.why = whyEl.value.trim();
      if (dur) state.input.duration_days = Number(dur.value);
    } else if (state.step >= 2 && state.step <= state.cfg.pillars.length + 1) {
      const def = state.cfg.pillars[state.step - 2];
      const value = ensurePillar(def.key);
      const objEl = dom.stage.querySelector(`[name="pillar_objective_${def.key}"]`);
      if (objEl) value.objective = objEl.value.trim();
      value.routines.forEach((r, idx) => {
        const titleEl = dom.stage.querySelector(`[name="routine_title_${def.key}_${idx}"]`);
        const trigEl = dom.stage.querySelector(`[name="routine_trigger_${def.key}_${idx}"]`);
        const cadType = dom.stage.querySelector(`[name="cadence_type_${def.key}_${idx}"]:checked`);
        if (titleEl) r.title = titleEl.value.trim();
        if (trigEl) r.trigger = trigEl.value.trim();
        if (cadType) {
          if (cadType.value === "specific_days") {
            const days = Array.from(dom.stage.querySelectorAll(`[name="cadence_day_${def.key}_${idx}"]:checked`)).map((el) => el.value);
            r.cadence = { type: "specific_days", days };
          } else if (cadType.value === "times_per_week") {
            const timesEl = dom.stage.querySelector(`[name="cadence_times_${def.key}_${idx}"]`);
            r.cadence = { type: "times_per_week", times: Number(timesEl ? timesEl.value : 1) || 1 };
          } else {
            r.cadence = { type: "daily" };
          }
        }
      });
    }
    if (!state.editId) saveDraft();
  }

  function validateStep() {
    if (state.step === 0) {
      return "";
    }
    if (state.step === 1) {
      if (!state.input.data.priority.title) return copy.errors.priorityRequired;
      if (!state.input.data.priority.why) return copy.errors.priorityWhyRequired;
      if (!state.cfg.durations.some((item) => item.value === state.input.duration_days)) return copy.errors.invalidDuration;
      if (!state.cfg.visibility.find((v) => v.value === state.input.visibility)) return copy.errors.invalidVisibility;
    } else if (state.step >= 2 && state.step <= state.cfg.pillars.length + 1) {
      const def = state.cfg.pillars[state.step - 2];
      const value = ensurePillar(def.key);
      for (const r of value.routines) {
        if (!r.title) return copy.errors.routineTitleRequired;
        if (!r.trigger) return copy.errors.routineTriggerRequired;
        if (r.cadence.type === "specific_days" && (!r.cadence.days || !r.cadence.days.length)) return copy.errors.selectAtLeastOneDay;
        if (r.cadence.type === "times_per_week" && (!r.cadence.times || r.cadence.times < 1 || r.cadence.times > 7)) return copy.errors.timesPerWeekRange;
      }
    }
    return "";
  }

  function buildPayload() {
    const pillars = {};
    for (const [key, value] of Object.entries(state.input.data.pillars)) {
      if (!value) continue;
      if (!value.objective && !(value.routines || []).length) continue;
      pillars[key] = value;
    }
    return {
      start_date: state.input.start_date,
      duration_days: state.input.duration_days,
      visibility: state.input.visibility,
      data: {
        priority: state.input.data.priority,
        pillars,
      },
    };
  }

  function pruneIncompleteRoutinesForPillar(pillarKey) {
    const value = ensurePillar(pillarKey);
    const before = (value.routines || []).length;
    value.routines = (value.routines || []).filter((r) => {
      if (!r) return false;
      const title = String(r.title || "").trim();
      const trigger = String(r.trigger || "").trim();
      return title !== "" && trigger !== "";
    });
    return before - value.routines.length;
  }

  function removeRoutineAt(pillarKey, idx) {
    const value = ensurePillar(pillarKey);
    value.routines.splice(idx, 1);
    const next = {};
    for (const [k, v] of Object.entries(state.ui.collapsedRoutines)) {
      if (!k.startsWith(`${pillarKey}:`)) {
        next[k] = v;
        continue;
      }
      const oldIdx = Number(k.split(":")[1]);
      if (oldIdx < idx) next[k] = v;
      if (oldIdx > idx) next[`${pillarKey}:${oldIdx - 1}`] = v;
    }
    state.ui.collapsedRoutines = next;
    render();
  }

  function confirmRoutineRemoval(pillarKey, idx) {
    const sheet = window.huuperActionSheet;
    if (!sheet || typeof sheet.open !== "function") {
      removeRoutineAt(pillarKey, idx);
      return;
    }
    const deleteCopy = copy.actions.deleteRoutine;

    sheet.open({
      title: deleteCopy.title,
      meta: deleteCopy.meta,
      actions: [],
      footerAction: {
        label: deleteCopy.confirmLabel,
        tone: "danger",
        onSelect: async () => {
          removeRoutineAt(pillarKey, idx);
          if (typeof sheet.close === "function") sheet.close();
        },
      },
    });
  }

  function cancelEdit() {
    window.location.href = "/admin/battleplan/";
  }

  function confirmCancelEdit() {
    const sheet = window.huuperActionSheet;
    const cancelCopy = copy.actions.cancelEdit;
    if (!sheet || typeof sheet.open !== "function") {
      cancelEdit();
      return;
    }

    sheet.open({
      title: cancelCopy.title,
      actions: [],
      footerAction: {
        label: cancelCopy.confirmLabel,
        tone: "danger",
        onSelect: () => {
          cancelEdit();
        },
      },
    });
  }

  async function submit() {
    if (state.submitting) return;
    state.submitting = true;
    setStatus("");
    try {
      const url = state.editId
        ? `/api/me/battleplans/${encodeURIComponent(state.editId)}`
        : "/api/me/battleplans";
      const result = await auth.apiFetch(url, {
        method: state.editId ? "PATCH" : "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(buildPayload()),
      });
      const id = state.editId || (result && result.id);
      clearDraft();
      window.location.href = id ? `/admin/battleplan/view/?view=${encodeURIComponent(id)}` : "/admin/battleplan/";
    } catch (err) {
      if (isConfirm()) {
        state.confirmAttempted = true;
        render();
      } else {
        setStatus(copy.errors.saveFailed);
      }
      state.submitting = false;
    }
  }

  // ---------- Wiring ----------

  dom.back.addEventListener("click", () => {
    syncFromDOM();
    if (state.editId && isPillarStep()) {
      if (state.step > 2) {
        state.step -= 1;
        render();
        return;
      }
      window.location.href = `/admin/battleplan/view/?view=${encodeURIComponent(state.editId)}`;
      return;
    }
    if (state.step <= 0 || (!shouldShowIntro() && state.step <= 1)) {
      window.location.href = "/admin/battleplan/";
      return;
    }
    state.step -= 1;
    render();
  });

  if (cancelEditButton) {
    cancelEditButton.addEventListener("click", () => {
      confirmCancelEdit();
    });
  }

  dom.form.addEventListener("submit", async (e) => {
    e.preventDefault();
    await pulseNextLoading();
    syncFromDOM();
    if (state.step >= 2 && state.step <= state.cfg.pillars.length + 1) {
      const def = state.cfg.pillars[state.step - 2];
      const removed = pruneIncompleteRoutinesForPillar(def.key);
      if (removed > 0) {
        setStatus(`Removed ${removed} incomplete routine${removed > 1 ? "s" : ""} (missing title or trigger).`);
        render();
        highlightCurrentRoutines();
      }
    }
    const err = validateStep();
    if (err) { setStatus(err); return; }
    if (isConfirm()) {
      if (hasConfirmMissing()) {
        state.confirmAttempted = true;
        render();
        return;
      }
      submit();
      return;
    }
    state.confirmAttempted = false;
    state.step += 1;
    render();
  });

  dom.stage.addEventListener("click", async (e) => {
    const target = e.target.closest("[data-action]");
    if (!target) return;
    const action = target.dataset.action;
    if (action === "add-routine") {
      await pulseNextLoading();
      syncFromDOM();
      const key = target.dataset.pillar;
      const removed = pruneIncompleteRoutinesForPillar(key);
      const value = ensurePillar(key);
      const newIdx = value.routines.length;
      value.routines.push(newRoutine());
      state.ui.collapsedRoutines[`${key}:${newIdx}`] = false;
      state.ui.focusRoutine = { pillarKey: key, idx: newIdx };
      render();
      if (removed > 0) {
        setStatus(`Removed ${removed} incomplete routine${removed > 1 ? "s" : ""} (missing title or trigger).`);
        highlightCurrentRoutines();
      }
      animateCollapseOtherRoutines(key, newIdx);
    } else if (action === "toggle-routine") {
      syncFromDOM();
      const key = target.dataset.pillar;
      const idx = Number(target.dataset.idx);
      const collapseKey = `${key}:${idx}`;
      const value = ensurePillar(key);
      const currentCollapsed = Object.prototype.hasOwnProperty.call(state.ui.collapsedRoutines, collapseKey)
        ? !!state.ui.collapsedRoutines[collapseKey]
        : true;
      if (currentCollapsed) {
        for (let i = 0; i < value.routines.length; i += 1) {
          state.ui.collapsedRoutines[`${key}:${i}`] = true;
        }
        state.ui.collapsedRoutines[collapseKey] = false;
      } else {
        state.ui.collapsedRoutines[collapseKey] = true;
      }
      render();
    } else if (action === "remove-routine") {
      syncFromDOM();
      const key = target.dataset.pillar;
      const idx = Number(target.dataset.idx);
      confirmRoutineRemoval(key, idx);
    } else if (action === "pause-cadence") {
      syncFromDOM();
      const key = target.dataset.pillar;
      const idx = Number(target.dataset.idx);
      const value = ensurePillar(key);
      const routine = value.routines[idx];
      if (!routine) return;
      routine.cadence = { type: "paused" };
      render();
    } else if (action === "go-step") {
      syncFromDOM();
      const step = Number(target.dataset.step);
      if (!Number.isFinite(step)) return;
      state.step = clampStep(step);
      state.confirmAttempted = false;
      render();
    } else if (action === "edit-pillar") {
      const step = Number(target.dataset.step);
      const id = state.viewId || state.editId;
      if (!id || !Number.isFinite(step)) return;
      window.location.href = `/admin/battleplan/edit/?edit=${encodeURIComponent(id)}&step=${encodeURIComponent(String(step))}`;
    }
  });

  dom.stage.addEventListener("change", (e) => {
    if (!e.target.matches('[name^="cadence_type_"]')) return;
    syncFromDOM();
    render();
  });

  dom.form.addEventListener("input", () => {
    syncFromDOM();
  });

  dom.form.addEventListener("change", (e) => {
    if (e.target && e.target.name === "visibility_aside") {
      syncFromDOM();
      return;
    }
    syncFromDOM();
  });

  // ---------- Init ----------

  function validateCopy(c) {
    if (!c || !c.labels || !c.placeholders || !c.actions || !c.errors || !c.dayShort) {
      throw new Error("battleplan copy incomplete");
    }
    if (!c.actions.removeRoutineAria) throw new Error("battleplan copy missing actions.removeRoutineAria");
    if (!c.actions.expandRoutineAria) throw new Error("battleplan copy missing actions.expandRoutineAria");
    if (!c.actions.collapseRoutineAria) throw new Error("battleplan copy missing actions.collapseRoutineAria");
    if (!c.actions.deleteRoutine || typeof c.actions.deleteRoutine !== "object") throw new Error("battleplan copy missing actions.deleteRoutine");
    if (!c.actions.deleteRoutine.title) throw new Error("battleplan copy missing actions.deleteRoutine.title");
    if (!c.actions.deleteRoutine.meta) throw new Error("battleplan copy missing actions.deleteRoutine.meta");
    if (!c.actions.deleteRoutine.cancelLabel) throw new Error("battleplan copy missing actions.deleteRoutine.cancelLabel");
    if (!c.actions.deleteRoutine.confirmLabel) throw new Error("battleplan copy missing actions.deleteRoutine.confirmLabel");
    if (!c.actions.cancelEdit || typeof c.actions.cancelEdit !== "object") throw new Error("battleplan copy missing actions.cancelEdit");
    if (!c.actions.cancelEdit.title) throw new Error("battleplan copy missing actions.cancelEdit.title");
    if (!c.actions.cancelEdit.keepLabel) throw new Error("battleplan copy missing actions.cancelEdit.keepLabel");
    if (!c.actions.cancelEdit.confirmLabel) throw new Error("battleplan copy missing actions.cancelEdit.confirmLabel");
    if (!c.errors.priorityRequired) throw new Error("battleplan copy missing errors.priorityRequired");
    if (!c.errors.priorityWhyRequired) throw new Error("battleplan copy missing errors.priorityWhyRequired");
    if (!c.errors.completePillarRequired) throw new Error("battleplan copy missing errors.completePillarRequired");
  }

  function validateSettings(cfg) {
    if (!cfg || !cfg.priority || !cfg.wizard || !cfg.wizard.intro || !cfg.wizard.confirmation) {
      throw new Error("battleplan settings incomplete");
    }
    if (introEnabledBySettings()) {
      if (!cfg.wizard.intro.title) throw new Error("battleplan settings missing wizard.intro.title");
    if (!cfg.wizard.intro.button) throw new Error("battleplan settings missing wizard.intro.button");
    }
    if (!cfg.wizard.confirmation.title) throw new Error("battleplan settings missing wizard.confirmation.title");
    if (!cfg.wizard.confirmation.button) throw new Error("battleplan settings missing wizard.confirmation.button");
    if (!cfg.priority.label) throw new Error("battleplan settings missing priority.label");
    if (!cfg.priority.description) throw new Error("battleplan settings missing priority.description");
    if (!Array.isArray(cfg.durations) || cfg.durations.length === 0) throw new Error("battleplan settings missing durations");
    if (!cfg.durations.some((item) => item.default)) throw new Error("battleplan settings missing default duration");
    for (const item of cfg.durations) {
      if (!Number.isFinite(item.value)) throw new Error("battleplan settings invalid durations.value");
    }
    if (!Array.isArray(cfg.visibility) || cfg.visibility.length === 0) throw new Error("battleplan settings missing visibility");
    if (!cfg.visibility.some((item) => item.default)) throw new Error("battleplan settings missing default visibility");
  }

  async function init() {
    try {
      validateCopy(copy);
      const activeId = state.editId || state.viewId;
      const fetchList = activeId
        ? auth.apiFetch(`/api/me/battleplans/${encodeURIComponent(activeId)}`)
        : auth.apiFetch("/api/me/battleplans?per_page=1");
      const [settings, existingOrBp] = await Promise.all([
        auth.apiFetch("/api/me/settings/battleplan"),
        fetchList,
      ]);
      state.cfg = settings.data;
      if (state.editId || state.viewId) {
        state.hasExistingBattleplan = true;
        const bp = existingOrBp;
        if (bp.visibility) state.input.visibility = bp.visibility;
        if (bp.start_date) state.input.start_date = bp.start_date.slice(0, 10);
        const bpData = bp.data || {};
        if (bpData.priority) {
          state.input.data.priority.title = String(bpData.priority.title || "");
          state.input.data.priority.why = String(bpData.priority.why || "");
        }
        if (bpData.pillars && typeof bpData.pillars === "object") {
          for (const def of (state.cfg.pillars || [])) {
            const src = bpData.pillars[def.key];
            if (!src) continue;
            const target = ensurePillar(def.key);
            target.objective = String(src.objective || "");
            if (Array.isArray(src.routines)) {
              target.routines = src.routines.map((r) => ({
                title: String((r && r.title) || ""),
                trigger: String((r && r.trigger) || ""),
                cadence: (r && r.cadence && typeof r.cadence === "object") ? r.cadence : { type: defaultCadenceType() },
              }));
            }
          }
        }
      } else {
        state.hasExistingBattleplan = !!(existingOrBp && Array.isArray(existingOrBp.items) && existingOrBp.items.length > 0);
      }
      validateSettings(state.cfg);
      state.input.visibility = state.input.visibility || defaultVisibilityValue();
      for (const p of state.cfg.pillars) {
        ensurePillar(p.key);
      }
      if (dom.back) dom.back.textContent = copy.backLabel.toUpperCase();
      if (!state.editId && !state.viewId) state.input.duration_days = defaultDurationValue();
      const draft = (state.editId || state.viewId) ? null : loadDraft();
      if (draft && draft.input && typeof draft.input === "object") {
        if (typeof draft.input.start_date === "string") state.input.start_date = draft.input.start_date;
        if (Number.isFinite(draft.input.duration_days) && state.cfg.durations.some((item) => item.value === draft.input.duration_days)) {
          state.input.duration_days = draft.input.duration_days;
        }
        if (typeof draft.input.visibility === "string" && state.cfg.visibility.some((v) => v.value === draft.input.visibility)) {
          state.input.visibility = draft.input.visibility;
        }
        if (draft.input.data && typeof draft.input.data === "object") {
          if (draft.input.data.priority && typeof draft.input.data.priority === "object") {
            state.input.data.priority.title = String(draft.input.data.priority.title || "");
            state.input.data.priority.why = String(draft.input.data.priority.why || "");
          }
          if (draft.input.data.pillars && typeof draft.input.data.pillars === "object") {
            for (const def of state.cfg.pillars) {
              const src = draft.input.data.pillars[def.key];
              if (!src || typeof src !== "object") continue;
              const target = ensurePillar(def.key);
              target.objective = String(src.objective || "");
              if (Array.isArray(src.routines)) {
                target.routines = src.routines.map((r) => ({
                  title: String((r && r.title) || ""),
                  trigger: String((r && r.trigger) || ""),
                  cadence: (r && r.cadence && typeof r.cadence === "object") ? r.cadence : { type: defaultCadenceType() },
                }));
              }
            }
          }
        }
      }
      const stepFromURL = readStepFromURL();
      if (stepFromURL === null) {
        state.step = minStep();
      } else {
        state.step = clampStep(stepFromURL);
      }
      if (dom.progress && dom.status && dom.status.parentElement !== dom.progress) {
        dom.progress.appendChild(dom.status);
        dom.status.classList.add("wizard-progress-status");
      }
      if (!state.editId && !state.viewId) saveDraft();
      showWizard();
      renderVisibilityAside();
      render();
      if (state.viewId) {
        const bpCopy = (window.huuperCopy && window.huuperCopy.battleplan) || {};
        if (dom.back) {
          dom.back.textContent = (bpCopy.archiveLabel || "Archive").toUpperCase();
          dom.back.addEventListener("click", async () => {
            if (dom.back.disabled) return;
            dom.back.disabled = true;
            try {
              await auth.apiFetch(`/api/me/battleplans/${encodeURIComponent(state.viewId)}/status`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ status: "archived" }),
              });
              window.location.href = "/admin/battleplan/";
            } catch { dom.back.disabled = false; }
          });
        }
        const nextBtn = document.getElementById(`${PREFIX}-next`);
        const nextLabel = document.getElementById(`${PREFIX}-next-label`);
        if (nextLabel) nextLabel.textContent = "EDIT";
        if (nextBtn) {
          nextBtn.setAttribute("type", "button");
          nextBtn.removeAttribute("form");
          nextBtn.onclick = () => {
            window.location.href = `/admin/battleplan/edit/?edit=${encodeURIComponent(state.viewId)}&step=1`;
          };
        }
      }
    } catch (err) {
      dom.loading.hidden = true;
      dom.stage.innerHTML = `
        <section class="empty-state empty-state-icon-only" aria-label="Battleplan">
          <svg class="empty-state-icon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor" aria-hidden="true"><path d="M8 15A7 7 0 1 1 8 1a7 7 0 0 1 0 14m0 1A8 8 0 1 0 8 0a8 8 0 0 0 0 16"/><path d="M8 13A5 5 0 1 1 8 3a5 5 0 0 1 0 10m0 1A6 6 0 1 0 8 2a6 6 0 0 0 0 12"/><path d="M8 11a3 3 0 1 1 0-6 3 3 0 0 1 0 6m0 1a4 4 0 1 0 0-8 4 4 0 0 0 0 8"/><path d="M9.5 8a1.5 1.5 0 1 1-3 0 1.5 1.5 0 0 1 3 0"/></svg>
          <p class="empty-state-label">${esc(copy.emptyStateLabel)}</p>
        </section>
      `;
      dom.form.hidden = false;
    }
  }

  init();
})();
