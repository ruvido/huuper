// Battleplan wizard orchestrator. Owns `state` and the DOM event wiring;
// delegates pure helpers to ns.config, render fns to ns.render, and view-page
// action wiring to ns.actions. Each split file publishes onto
// window.appBattleplanWizard; we read from it here.
(() => {
  const auth = window.appAuth;
  if (!auth) return;
  const copy = window.appCopy && window.appCopy.battleplan && window.appCopy.battleplan.wizard;
  if (!copy) return;

  window.appBattleplanWizard = window.appBattleplanWizard || {};
  const ns = window.appBattleplanWizard;
  if (!ns.config || !ns.render || !ns.actions) return; // split modules required

  const cfgmod = ns.config;
  const rndmod = ns.render;
  const actmod = ns.actions;
  const esc = rndmod.esc;

  const PREFIX = "battleplan-wizard";
  // Legacy localStorage draft key — only used by clearDraft() to wipe stale
  // entries from earlier auto-resume behaviour. No new drafts are written here.
  const LEGACY_DRAFT_STORAGE_KEY = "battleplan-wizard-draft-v1";
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
    next: $("next"),
    nextLabel: $("next-label"),
  };
  const cancelEditButton = document.querySelector("[data-wizard-cancel-edit]");

  // The .step-progress is position:fixed at the viewport top. Write its
  // measured height onto .wizard-page as --step-progress-h so the page can
  // reserve matching padding-top via CSS. Height varies with the visibility
  // aside, so a ResizeObserver keeps the value accurate.
  function syncProgressHeight() {
    if (!dom.page || !dom.progress) return;
    const h = dom.progress.hidden ? 0 : Math.ceil(dom.progress.getBoundingClientRect().height);
    dom.page.style.setProperty("--step-progress-h", `${h}px`);
  }
  if (dom.progress && typeof ResizeObserver !== "undefined") {
    const ro = new ResizeObserver(syncProgressHeight);
    ro.observe(dom.progress);
  }

  function idFromPath(kind) {
    const parts = window.location.pathname.split("/").filter(Boolean);
    const idx = parts.findIndex((part, i) => part === kind && parts[i - 1] === "battleplan");
    return idx >= 0 && parts[idx + 1] ? decodeURIComponent(parts[idx + 1]) : null;
  }

  const basePath = () => window.appBattleplan.basePath();

  const params = new URLSearchParams(window.location.search);
  const editId = idFromPath("edit") || params.get("edit") || null;
  const viewId = idFromPath("view") || params.get("view") || null;

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
      focusEndDateCustom: false,
    },
    fieldErrors: {},
    input: {
      start_date: new Date().toISOString().slice(0, 10),
      duration_days: 30,
      end_date: "",
      use_end_date: false,
      visibility: "",
      data: {
        priority: { title: "", why: "" },
        pillars: {},
      },
    },
  };

  function clearDraft() {
    try {
      window.localStorage.removeItem(LEGACY_DRAFT_STORAGE_KEY);
    } catch (_) {
      // no-op: localStorage unavailable
    }
  }

  function setStatus(msg) {
    if (!dom.status) return;
    dom.status.textContent = msg || "";
    dom.status.hidden = !msg;
    dom.status.classList.toggle("is-error", !!msg);
  }

  // Format the "removed N incomplete routine(s)" status from copy templates.
  // Reads from appCopy each call so a copy update takes effect immediately.
  function formatRemovedStatus(count) {
    const shared = (window.appCopy && window.appCopy.battleplan && window.appCopy.battleplan.wizardShared) || {};
    const tmpl = shared.removedRoutinesStatus || {};
    if (count === 1) return tmpl.singular || "";
    return (tmpl.plural || "").replace("{count}", String(count));
  }

  function setLoading(visible) {
    dom.loading.hidden = !visible;
  }

  function showWizard() {
    setLoading(false);
    dom.form.hidden = false;
    dom.sticky.hidden = false;
  }

  function totalSteps() {
    return 4 + state.cfg.pillars.length; // intro + priority + N pillars + endDate + confirm
  }

  function minStep() {
    if (state.viewId) return totalSteps() - 1;
    // Drafts behave like a fresh new plan: start from intro/priority, allow editing all steps.
    if (state.editId && !cfgmod.isEditingDraft(state)) return 2;
    return cfgmod.shouldShowIntro(state) ? 0 : 1;
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

  function isEndDateStep() {
    return state.step === state.cfg.pillars.length + 2;
  }

  function internalStepFromURLStep(urlStep) {
    if (!Number.isFinite(urlStep)) return null;
    if (state.viewId) return null;
    if (urlStep < 1 || urlStep > state.cfg.pillars.length + 1) return null;
    return urlStep + 1;
  }

  function urlStepFromInternalStep() {
    if (isPillarStep() || isEndDateStep()) return state.step - 1;
    return null;
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

  function isConfirm() {
    return state.step === totalSteps() - 1;
  }

  function isSummaryView() {
    return isConfirm() && (state.viewId || state.editId);
  }

  function stateLabel() {
    if (isConfirm()) return state.cfg.wizard.confirmation.title.toUpperCase();
    if (state.step === 0) return state.cfg.wizard.intro.title.toUpperCase();
    if (state.step === 1) return String(cfgmod.priorityCopy(state).title || "").toUpperCase();
    if (isEndDateStep()) {
      const sharedCopy = (window.appCopy && window.appCopy.battleplan && window.appCopy.battleplan.wizardShared) || {};
      return String(sharedCopy.endDateTitle || copy.labels.durationField).toUpperCase();
    }
    const def = state.cfg.pillars[state.step - 2];
    return def ? def.label.toUpperCase() : "";
  }

  function syncProgress() {
    const n = state.cfg.pillars.length + 1; // pillars + end date
    let cur = 0;
    if (isPillarStep()) cur = state.step - 1;
    else if (isEndDateStep()) cur = n;
    dom.progressLabel.textContent = `${copy.progressPrefix}: ${cfgmod.pad2(cur)}/${cfgmod.pad2(n)}`;
    dom.progressState.textContent = stateLabel();
    dom.progressFill.style.width = `${(cur / n) * 100}%`;
    if (isConfirm()) {
      dom.nextLabel.textContent = (cfgmod.willSaveAsDraft(state) ? copy.saveDraftLabel : copy.confirmLabel).toUpperCase();
      rndmod.setNextIcon("edit", PREFIX);
      return;
    }
    dom.nextLabel.textContent = state.step === 0 ? state.cfg.wizard.intro.button.toUpperCase() : copy.nextLabel.toUpperCase();
    rndmod.setNextIcon("default", PREFIX);
  }

  function ensurePillar(key) {
    if (!state.input.data.pillars[key]) {
      state.input.data.pillars[key] = { objective: "", routines: [] };
    }
    return state.input.data.pillars[key];
  }

  function newRoutine() {
    return { title: "", trigger: "", cadence: { type: cfgmod.defaultCadenceType(state.cfg) } };
  }

  function syncChrome() {
    const onIntro = state.step === 0 && cfgmod.shouldShowIntro(state);
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
    if (dom.progress) {
      dom.progress.hidden = state.step === 0 || isConfirm();
      dom.progress.classList.toggle("step-progress-aside-only", state.step === 1);
    }
    if (dom.aside) dom.aside.hidden = state.step === 0 || isConfirm();
    syncProgressHeight();
    if (dom.back) dom.back.hidden = onIntro;
    if (dom.sticky) dom.sticky.classList.toggle("sticky-actions-intro", onIntro);
  }

  // Build the rendering context object passed to render module functions.
  // Built fresh per render so split renderers never close over stale data.
  function makeRenderCtx() {
    return {
      state,
      copy,
      dom,
      PREFIX,
      ensurePillar,
      isSummaryView,
    };
  }

  function render() {
    syncChrome();
    syncProgress();
    writeStepToURL();
    setStatus("");
    const errSlot = document.getElementById(`${PREFIX}-top-error`);
    if (errSlot) errSlot.textContent = "";
    const ctx = makeRenderCtx();
    if (state.step === 0) {
      rndmod.renderIntro(ctx);
    } else if (state.step === 1) {
      rndmod.renderVisibilityAside(ctx);
      rndmod.renderPriority(ctx);
    } else if (isPillarStep()) {
      rndmod.renderVisibilityAside(ctx);
      rndmod.renderPillar(ctx, state.step - 2);
    } else if (isEndDateStep()) {
      rndmod.renderVisibilityAside(ctx);
      rndmod.renderEndDate(ctx);
    } else {
      rndmod.renderConfirm(ctx);
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
    if (state.ui.focusEndDateCustom && dom.stage) {
      const input = dom.stage.querySelector('[name="duration_end_date"]');
      if (input) {
        input.focus({ preventScroll: true });
        const len = input.value.length;
        try { input.setSelectionRange(len, len); } catch (_) {}
      }
      state.ui.focusEndDateCustom = false;
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
      if (titleEl) state.input.data.priority.title = titleEl.value.trim();
      if (whyEl) state.input.data.priority.why = whyEl.value.trim();
    } else if (isEndDateStep()) {
      const choice = dom.stage.querySelector('[name="duration_choice"]:checked');
      const dateInput = dom.stage.querySelector('[name="duration_end_date"]');
      if (choice) {
        if (choice.value === "custom") {
          state.input.use_end_date = true;
          const raw = dateInput ? dateInput.value.trim() : "";
          state.input.end_date = raw;
          if (/^\d{4}-\d{2}-\d{2}$/.test(raw)) {
            const start = new Date(state.input.start_date + "T00:00:00Z");
            const end = new Date(raw + "T00:00:00Z");
            if (!isNaN(end.getTime())) {
              const diff = Math.round((end - start) / 86400000);
              if (diff > 0) state.input.duration_days = diff;
            }
          }
        } else {
          state.input.use_end_date = false;
          state.input.duration_days = Number(choice.value);
          state.input.end_date = "";
        }
      }
    } else if (isPillarStep()) {
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
  }

  function validateStep() {
    state.fieldErrors = {};
    if (state.step === 0) {
      return "";
    }
    if (state.step === 1) {
      if (!state.input.data.priority.title) {
        state.fieldErrors.priorityTitle = true;
        return copy.errors.priorityRequired;
      }
      if (!state.input.data.priority.why) {
        state.fieldErrors.priorityWhy = true;
        return copy.errors.priorityWhyRequired;
      }
      if (!state.cfg.visibility.find((v) => v.value === state.input.visibility)) return copy.errors.invalidVisibility;
    } else if (isEndDateStep()) {
      if (state.input.use_end_date) {
        if (!/^\d{4}-\d{2}-\d{2}$/.test(state.input.end_date)) return copy.errors.invalidDuration;
        const endTs = Date.parse(state.input.end_date + "T00:00:00Z");
        const startTs = Date.parse(state.input.start_date + "T00:00:00Z");
        if (isNaN(endTs) || !(endTs > startTs)) return copy.errors.invalidDuration;
        const days = Math.round((endTs - startTs) / 86400000);
        if (days > rndmod.MAX_END_DATE_DAYS) return copy.errors.invalidDuration;
      } else if (!state.cfg.durations.some((item) => item.value === state.input.duration_days)) {
        return copy.errors.invalidDuration;
      }
    } else if (isPillarStep()) {
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
      end_date: state.input.use_end_date ? state.input.end_date : "",
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
    const sheet = window.appActionSheet;
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
    window.location.href = `${basePath()}/`;
  }

  function confirmCancelEdit() {
    const sheet = window.appActionSheet;
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

  function hasConfirmMissing() {
    const priorityMissing = !(state.input.data.priority.title || "").trim();
    if (priorityMissing) return true;
    let hasCompletePillar = false;
    for (const def of state.cfg.pillars) {
      const value = state.input.data.pillars[def.key] || { objective: "", routines: [] };
      if (cfgmod.isPillarComplete(value)) {
        hasCompletePillar = true;
        break;
      }
    }
    return !hasCompletePillar;
  }

  async function submit() {
    if (state.submitting) return;
    state.submitting = true;
    setStatus("");
    try {
      const url = state.editId
        ? `/api/me/battleplans/${encodeURIComponent(state.editId)}`
        : "/api/me/battleplans";
      const payload = buildPayload();
      // New plan + active already exists → save as draft (active is unique per user).
      if (!state.editId && state.hasExistingActive) {
        payload.status = "draft";
      }
      await auth.apiFetch(url, {
        method: state.editId ? "PATCH" : "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      clearDraft();
      window.location.href = `${basePath()}/`;
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

  // Build the deps object for the actions module. Frozen per call so view-page
  // buttons always read the latest state through closure on `state` itself.
  function makeActionDeps() {
    return { auth, state, dom, basePath, PREFIX, buildPayload };
  }

  // ---------- Wiring ----------

  dom.back.addEventListener("click", () => {
    syncFromDOM();
    // Non-draft edits jump back to /view/ once they hit the pillar boundary.
    // Drafts and new plans walk back step-by-step to the list.
    if (state.editId && !cfgmod.isEditingDraft(state) && isPillarStep()) {
      if (state.step > 2) {
        state.step -= 1;
        render();
        return;
      }
      window.location.href = `${basePath()}/view/${encodeURIComponent(state.editId)}/`;
      return;
    }
    if (state.step <= 0 || (!cfgmod.shouldShowIntro(state) && state.step <= 1)) {
      window.location.href = `${basePath()}/`;
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
    syncFromDOM();
    if (isPillarStep()) {
      const def = state.cfg.pillars[state.step - 2];
      const removed = pruneIncompleteRoutinesForPillar(def.key);
      if (removed > 0) {
        setStatus(formatRemovedStatus(removed));
        render();
        rndmod.highlightCurrentRoutines(dom);
      }
    }
    const err = validateStep();
    if (err) {
      render();
      setStatus(err);
      const errorEl = dom.stage.querySelector(".field-input-error");
      if (errorEl && typeof errorEl.scrollIntoView === "function") {
        errorEl.scrollIntoView({ behavior: "smooth", block: "center" });
      }
      return;
    }
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
        setStatus(formatRemovedStatus(removed));
        rndmod.highlightCurrentRoutines(dom);
      }
      rndmod.animateCollapseOtherRoutines(makeRenderCtx(), key, newIdx);
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
    } else if (action === "edit-pillar" || action === "edit-end-date") {
      const step = Number(target.dataset.step);
      const id = state.viewId || state.editId;
      if (!id || !Number.isFinite(step)) return;
      window.location.href = `${basePath()}/edit/${encodeURIComponent(id)}/?step=${encodeURIComponent(String(step))}`;
    }
  });

  dom.stage.addEventListener("change", (e) => {
    if (!e.target.matches('[name^="cadence_type_"]')) return;
    syncFromDOM();
    render();
  });

  dom.form.addEventListener("input", (e) => {
    if (e.target && e.target.name === "priority_why") {
      state.fieldErrors.priorityWhy = false;
      e.target.classList.remove("field-input-error");
    }
    if (e.target && e.target.name === "priority_title") {
      state.fieldErrors.priorityTitle = false;
      e.target.classList.remove("field-input-error");
    }
    syncFromDOM();
  });

  dom.form.addEventListener("focusin", (e) => {
    const el = e.target;
    if (!el) return;
    const tag = el.tagName;
    if (tag !== "INPUT" && tag !== "TEXTAREA") return;
    if (el.type === "radio" || el.type === "checkbox") return;
    // Wait for mobile keyboard to expand (and viewport to shrink) before scrolling.
    window.setTimeout(() => {
      if (document.activeElement !== el) return;
      if (typeof el.scrollIntoView === "function") {
        el.scrollIntoView({ behavior: "smooth", block: "center" });
      }
    }, 300);
  });

  if (dom.aside) {
    dom.aside.addEventListener("change", (e) => {
      if (!e.target || e.target.name !== "visibility_aside") return;
      syncFromDOM();
      render();
    });
  }

  dom.stage.addEventListener("click", (e) => {
    const mask = e.target.closest('.wizard-end-date-mask');
    if (mask) {
      const inp = mask.querySelector('input[type="text"]');
      if (inp && document.activeElement !== inp) inp.focus();
      return;
    }
    const prompt = e.target.closest('[data-action="open-end-date-custom"]');
    if (!prompt) return;
    state.input.use_end_date = true;
    state.ui.focusEndDateCustom = true;
    render();
  });

  dom.stage.addEventListener("input", (e) => {
    if (!e.target || e.target.name !== "duration_end_date") return;
    const input = e.target;
    const isDelete = typeof e.inputType === "string" && e.inputType.startsWith("delete");
    const digits = input.value.replace(/\D/g, "").slice(0, 8);
    let formatted = digits.slice(0, 4);
    const yearDone = isDelete ? digits.length > 4 : digits.length >= 4;
    const monthDone = isDelete ? digits.length > 6 : digits.length >= 6;
    if (yearDone) formatted += "-" + digits.slice(4, 6);
    if (monthDone) formatted += "-" + digits.slice(6, 8);
    if (input.value !== formatted) {
      input.value = formatted;
      const len = formatted.length;
      try { input.setSelectionRange(len, len); } catch (_) {}
    }
    syncEndDateMask(input);
    state.input.use_end_date = true;
    syncFromDOM();
  });

  function syncEndDateMask(input) {
    const template = "YYYY-MM-DD";
    const len = input.value.length;
    input.style.width = len + "ch";
    const ghost = input.parentElement && input.parentElement.querySelector(".wizard-end-date-mask-ghost");
    if (ghost) ghost.textContent = template.slice(Math.min(len, template.length));
    const summary = dom.stage.querySelector(".wizard-end-date-summary");
    if (summary) {
      const sharedCopy = (window.appCopy && window.appCopy.battleplan && window.appCopy.battleplan.wizardShared) || {};
      const result = rndmod.computeEndDateSummary(state, input.value, sharedCopy);
      const textEl = summary.querySelector(".wizard-end-date-summary-text");
      if (textEl) textEl.textContent = result.text;
      summary.classList.toggle("is-error", !!(result.text && result.error));
      summary.classList.toggle("is-ok", !!(result.text && !result.error));
    }
  }

  dom.stage.addEventListener("change", (e) => {
    if (!e.target || e.target.name !== "duration_choice") return;
    syncFromDOM();
    render();
  });

  dom.form.addEventListener("change", (e) => {
    if (e.target && e.target.name === "visibility_aside") {
      syncFromDOM();
      return;
    }
    if (e.target && e.target.name === "duration") {
      syncFromDOM();
      render();
      return;
    }
    syncFromDOM();
  });

  // ---------- Init ----------

  async function init() {
    try {
      cfgmod.validateCopy(copy);
      const [settings, list] = await Promise.all([
        auth.apiFetch("/api/me/settings/battleplan"),
        auth.apiFetch("/api/me/battleplans?per_page=200"),
      ]);
      state.cfg = settings.data;
      const items = (list && Array.isArray(list.items)) ? list.items : [];
      state.hasExistingBattleplan = items.length > 0;
      state.hasExistingActive = items.some((it) => it && it.status === "active");
      if (state.editId || state.viewId) {
        const id = state.editId || state.viewId;
        const bp = items.find((it) => it && it.id === id);
        if (!bp) throw new Error("battleplan not found");
        state.loadedStatus = bp.status || "";
        if (bp.visibility) state.input.visibility = bp.visibility;
        if (bp.start_date) state.input.start_date = bp.start_date.slice(0, 10);
        const bpData = bp.data || {};
        if (bpData.priority) {
          state.input.data.priority.title = String(bpData.priority.title || "");
          state.input.data.priority.why = String(bpData.priority.why || "");
        }
        if (Array.isArray(bpData.notes)) {
          state.input.data.notes = bpData.notes.map((note) => ({
            text: String((note && note.text) || ""),
            at: String((note && note.at) || ""),
            by: String((note && note.by) || ""),
          }));
        }
        if (bpData.pillars && typeof bpData.pillars === "object") {
          for (const def of (state.cfg.pillars || [])) {
            const src = bpData.pillars[def.key];
            if (!src) continue;
            const target = ensurePillar(def.key);
            target.objective = String(src.objective || "");
            if (Array.isArray(src.routines)) {
              target.routines = src.routines.map((r) => ({
                id: String((r && r.id) || ""),
                title: String((r && r.title) || ""),
                trigger: String((r && r.trigger) || ""),
                cadence: (r && r.cadence && typeof r.cadence === "object") ? r.cadence : { type: cfgmod.defaultCadenceType(state.cfg) },
                created: String((r && r.created) || ""),
              }));
            }
          }
        }
      }
      cfgmod.validateSettings(state.cfg);
      state.input.visibility = state.input.visibility || cfgmod.defaultVisibilityValue(state.cfg);
      for (const p of state.cfg.pillars) {
        ensurePillar(p.key);
      }
      if (dom.back) dom.back.textContent = copy.backLabel.toUpperCase();
      if (!state.editId && !state.viewId) {
        state.input.duration_days = cfgmod.defaultDurationValue(state.cfg);
        // /new/ always starts fresh — server-side drafts replace the old
        // localStorage auto-resume behaviour, which would silently re-populate
        // fields the user did not expect.
        clearDraft();
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
      showWizard();
      rndmod.renderVisibilityAside(makeRenderCtx());
      render();
      if (state.viewId) actmod.configureViewActions(makeActionDeps());
    } catch (err) {
      dom.loading.hidden = true;
      const ariaLabel = (window.appCopy && window.appCopy.battleplan && window.appCopy.battleplan.titleBattleplan) || "";
      dom.stage.innerHTML = `
        <section class="empty-state empty-state-icon-only" aria-label="${esc(ariaLabel)}">
          <svg class="empty-state-icon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor" aria-hidden="true"><path d="M8 15A7 7 0 1 1 8 1a7 7 0 0 1 0 14m0 1A8 8 0 1 0 8 0a8 8 0 0 0 0 16"/><path d="M8 13A5 5 0 1 1 8 3a5 5 0 0 1 0 10m0 1A6 6 0 1 0 8 2a6 6 0 0 0 0 12"/><path d="M8 11a3 3 0 1 1 0-6 3 3 0 0 1 0 6m0 1a4 4 0 1 0 0-8 4 4 0 0 0 0 8"/><path d="M9.5 8a1.5 1.5 0 1 1-3 0 1.5 1.5 0 0 1 3 0"/></svg>
          <p class="empty-state-label">${esc(copy.emptyStateLabel)}</p>
        </section>
      `;
      dom.form.hidden = false;
    }
  }

  init();
})();
