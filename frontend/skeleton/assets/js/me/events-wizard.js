// Events wizard orchestrator. Owns `state` and DOM event wiring; delegates
// pure helpers to ns.config, render fns to ns.render, and side-effect
// actions to ns.actions. Each split file publishes onto
// window.appEventsWizard. The same file is loaded by both /me and /admin
// wizard pages — `state.scope` is derived from the URL prefix.
(() => {
  const auth = window.appAuth;
  if (!auth) return;
  const copyAll = window.appCopy && window.appCopy.events;
  const copy = copyAll && copyAll.wizard;
  if (!copy) return;

  window.appEventsWizard = window.appEventsWizard || {};
  const ns = window.appEventsWizard;
  if (!ns.config || !ns.render || !ns.actions) return;

  const cfgmod = ns.config;
  const rndmod = ns.render;
  const actmod = ns.actions;

  const PREFIX = "events-wizard";
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

  // Mirror the measured .step-progress height onto .wizard-page as
  // --step-progress-h so CSS can reserve matching padding-top. The bar is
  // position:fixed at the viewport top, so without this the page content
  // would slide under it.
  function syncProgressHeight() {
    if (!dom.page || !dom.progress) return;
    const h = dom.progress.hidden ? 0 : Math.ceil(dom.progress.getBoundingClientRect().height);
    dom.page.style.setProperty("--step-progress-h", `${h}px`);
  }
  if (dom.progress && typeof ResizeObserver !== "undefined") {
    const ro = new ResizeObserver(syncProgressHeight);
    ro.observe(dom.progress);
  }

  // /me/events/... vs /admin/events/...
  const pathParts = window.location.pathname.split("/").filter(Boolean);
  const scope = pathParts[0] === "admin" ? "admin" : "me";
  const basePath = `/${scope}`;

  function readEditIdFromURL() {
    const params = new URLSearchParams(window.location.search);
    const fromQuery = params.get("id");
    if (fromQuery) return fromQuery;
    const idx = pathParts.findIndex((part, i) => part === "edit" && pathParts[i - 1] === "events");
    if (idx >= 0 && pathParts[idx + 1]) return decodeURIComponent(pathParts[idx + 1]);
    return null;
  }

  const editId = readEditIdFromURL();

  const state = {
    scope,
    cfg: null,
    isAdmin: scope === "admin",
    editId,
    submitting: false,
    step: 0,
    fieldErrors: {},
    availableGroups: [],
    loadedEvent: null,
    input: {
      type: "",
      group: "",
      groupName: "",
      title: "",
      location: "",
      url: "",
      schedule: "once",
      startDate: cfgmod.todayISO(),
      startTime: "",
      endDate: "",
      weekday: cfgmod.weekdayFromDate(cfgmod.todayISO()),
      monthPosition: "1st",
      monthDay: cfgmod.weekdayFromDate(cfgmod.todayISO()),
      count: 3,
      data: { description: "" },
    },
  };

  function setStatus(msg) {
    if (!dom.status) return;
    dom.status.textContent = msg || "";
    dom.status.hidden = !msg;
    dom.status.classList.toggle("is-error", !!msg);
  }

  function setLoading(visible) {
    if (dom.loading) dom.loading.hidden = !visible;
  }

  function showWizard() {
    setLoading(false);
    if (dom.form) dom.form.hidden = false;
    if (dom.sticky) dom.sticky.hidden = false;
  }

  // Step plan when CREATING:
  //   0: type
  //   1: group (skipped when type does not require group)
  //   2: dates
  //   3: details (skipped when the selected type has no required detail fields)
  //   4: confirm
  // Step plan when EDITING (single occurrence):
  //   0: details (we treat editing as details-only; reschedule is a separate action)
  //   1: confirm
  function stepKeys() {
    if (state.editId) return ["details", "confirm"];
    const td = cfgmod.typeDef(state.cfg, state.input.type);
    const requiresGroup = cfgmod.required(td, "group");
    const requiresDetails = cfgmod.hasAnyRequired(td, cfgmod.DETAIL_FIELDS);
    const keys = ["type"];
    // Group step is only injected when the chosen type actually requires it.
    // (Init defaults state.input.type to the first allowed type, so we don't
    // need to defensively include the group step for an "unknown" type.)
    if (requiresGroup) keys.push("group");
    keys.push("dates");
    if (requiresDetails) keys.push("details");
    keys.push("confirm");
    return keys;
  }

  function currentStepKey() {
    const keys = stepKeys();
    return keys[Math.max(0, Math.min(state.step, keys.length - 1))];
  }

  function totalSteps() {
    return stepKeys().length;
  }

  function isConfirm() {
    return currentStepKey() === "confirm";
  }

  function syncProgress() {
    const total = totalSteps();
    const cur = state.step;
    const stepLabel = (copy.steps && copy.steps[currentStepKey()] && copy.steps[currentStepKey()].title) || "";
    if (dom.progressLabel) {
      dom.progressLabel.textContent = `${(copy.progressPrefix || "STEP")}: ${cfgmod.pad2(cur + 1)}/${cfgmod.pad2(total)}`;
    }
    if (dom.progressState) dom.progressState.textContent = String(stepLabel).toUpperCase();
    if (dom.progressFill) dom.progressFill.style.width = `${((cur + 1) / total) * 100}%`;
    if (isConfirm()) {
      if (dom.nextLabel) dom.nextLabel.textContent = (state.editId ? (copy.saveLabel || "Save") : (copy.confirmLabel || "Create")).toUpperCase();
      rndmod.setNextIcon("edit", PREFIX);
    } else {
      if (dom.nextLabel) dom.nextLabel.textContent = (copy.nextLabel || "Next").toUpperCase();
      rndmod.setNextIcon("default", PREFIX);
    }
  }

  function syncChrome() {
    const onConfirm = isConfirm();
    if (dom.page) {
      dom.page.classList.toggle("wizard-page-confirm", onConfirm);
      dom.page.classList.toggle("wizard-page-has-progress", !onConfirm);
    }
    if (dom.stage) dom.stage.classList.toggle("wizard-stage-confirm", onConfirm);
    if (dom.form) dom.form.classList.toggle("wizard-form-confirm", onConfirm);
    if (dom.progress) dom.progress.hidden = onConfirm;
    if (dom.aside) dom.aside.hidden = true;
    syncProgressHeight();
    if (dom.back) dom.back.hidden = false;
  }

  function makeRenderCtx() {
    return { state, copy, dom, PREFIX };
  }

  function render() {
    syncChrome();
    syncProgress();
    setStatus("");
    const errSlot = document.getElementById(`${PREFIX}-top-error`);
    if (errSlot) errSlot.textContent = "";
    const ctx = makeRenderCtx();
    const key = currentStepKey();
    if (key === "type") rndmod.renderTypeStep(ctx);
    else if (key === "group") rndmod.renderGroupStep(ctx);
    else if (key === "dates") rndmod.renderDatesStep(ctx);
    else if (key === "details") rndmod.renderDetailsStep(ctx);
    else rndmod.renderConfirm(ctx);
    if (state.editId) renderEditFooterActions();
  }

  // Inject Cancel + Reschedule buttons in edit mode below the form.
  // These buttons live inside the stage so they re-render with each step.
  function renderEditFooterActions() {
    if (!dom.stage) return;
    const wrap = document.createElement("div");
    wrap.className = "wizard-routines";
    const cancelLabel = (copy.actions && copy.actions.cancelEvent) || "Cancel event";
    const rescheduleLabel = (copy.actions && copy.actions.reschedule) || "Reschedule";
    wrap.innerHTML = `
      <button type="button" class="wizard-btn wizard-btn-outline" data-action="reschedule">${rndmod.esc(rescheduleLabel)}</button>
      <button type="button" class="wizard-btn wizard-btn-danger" data-action="cancel-event">${rndmod.esc(cancelLabel)}</button>
    `;
    dom.stage.appendChild(wrap);
  }

  // ---------- Sync from DOM ----------

  function syncFromDOM() {
    if (!dom.stage) return;
    const key = currentStepKey();
    if (key === "type") {
      const sel = dom.stage.querySelector('input[name="event_type"]:checked');
      if (sel) {
        const newType = sel.value;
        if (newType !== state.input.type) {
          state.input.type = newType;
          const td = cfgmod.typeDef(state.cfg, newType);
          cfgmod.resetFieldsForType(state.input, td);
        }
      }
    } else if (key === "group") {
      const sel = dom.stage.querySelector('input[name="event_group"]:checked');
      if (sel) {
        state.input.group = sel.value;
        const found = (state.availableGroups || []).find((g) => g.id === sel.value);
        state.input.groupName = found ? (found.name || "") : "";
      }
    } else if (key === "dates") {
      const scheduleEl = dom.stage.querySelector('input[name="event_schedule"]:checked');
      if (scheduleEl) state.input.schedule = scheduleEl.value;
      const startEl = dom.stage.querySelector('input[name="event_start_date"]');
      if (startEl) {
        state.input.startDate = startEl.value;
        if (!state.input.weekday) state.input.weekday = cfgmod.weekdayFromDate(startEl.value);
        if (!state.input.monthDay) state.input.monthDay = cfgmod.weekdayFromDate(startEl.value);
      }
      const startTimeEl = dom.stage.querySelector('input[name="event_start_time"]');
      if (startTimeEl) state.input.startTime = startTimeEl.value;
      else state.input.startTime = "";
      const endEl = dom.stage.querySelector('input[name="event_end_date"]');
      if (endEl) state.input.endDate = endEl.value;
      else state.input.endDate = "";
      const weekdayEl = dom.stage.querySelector('input[name="event_weekday"]:checked');
      if (weekdayEl) state.input.weekday = weekdayEl.value;
      const monthPositionEl = dom.stage.querySelector('input[name="event_month_position"]:checked');
      if (monthPositionEl) state.input.monthPosition = monthPositionEl.value;
      const monthDayEl = dom.stage.querySelector('input[name="event_month_day"]:checked');
      if (monthDayEl) state.input.monthDay = monthDayEl.value;
      const countEl = dom.stage.querySelector('input[name="event_count"]:checked');
      if (countEl) state.input.count = Number(countEl.value) || 3;
    } else if (key === "details") {
      const titleEl = dom.stage.querySelector('input[name="event_title"]');
      const locationEl = dom.stage.querySelector('input[name="event_location"]');
      const urlEl = dom.stage.querySelector('input[name="event_url"]');
      const descEl = dom.stage.querySelector('textarea[name="event_description"]');
      if (titleEl) state.input.title = titleEl.value.trim();
      if (locationEl) state.input.location = locationEl.value.trim();
      if (urlEl) state.input.url = urlEl.value.trim();
      if (descEl) {
        state.input.data = state.input.data || {};
        state.input.data.description = descEl.value;
      } else if (state.input.data) {
        state.input.data.description = "";
      }
    }
  }

  function validateStep() {
    state.fieldErrors = {};
    const key = currentStepKey();
    if (key === "type") {
      if (!state.input.type) return copy.errors.typeRequired || "Type required";
      const allowed = cfgmod.allowedTypes(state.cfg, state.isAdmin);
      if (!allowed.find((t) => t.value === state.input.type)) return copy.errors.typeInvalid || "Invalid type";
    } else if (key === "group") {
      const td = cfgmod.typeDef(state.cfg, state.input.type);
      const requires = cfgmod.required(td, "group");
      if (requires && !state.input.group) return copy.errors.groupRequired || "Group required";
    } else if (key === "dates") {
      if (!String(state.input.startDate || "").trim()) {
        state.fieldErrors.start_date = true;
        return copy.errors.startDateRequired || "Start date required";
      }
      if (!cfgmod.isValidISODate(state.input.startDate)) {
        state.fieldErrors.start_date = true;
        return copy.errors.startDateInvalid || "Start date must be valid.";
      }
      const td = cfgmod.typeDef(state.cfg, state.input.type);
      if (cfgmod.required(td, "time") && !String(state.input.startTime || "").trim()) {
        state.fieldErrors.time = true;
        return copy.errors.timeRequired || "Time required";
      }
      if (cfgmod.required(td, "time") && !cfgmod.isValidTimeText(state.input.startTime)) {
        state.fieldErrors.time = true;
        return copy.errors.timeInvalid || "Time must be valid.";
      }
      if (cfgmod.required(td, "end_date") && !String(state.input.endDate || "").trim()) {
        state.fieldErrors.end_date = true;
        return copy.errors.endDateRequired || "End date required";
      }
      if (cfgmod.required(td, "end_date") && !cfgmod.isValidISODate(state.input.endDate)) {
        state.fieldErrors.end_date = true;
        return copy.errors.endDateInvalid || "End date must be valid.";
      }
      if (state.input.endDate && !cfgmod.isValidISODate(state.input.endDate)) return copy.errors.endDateInvalid || "Invalid end date";
      if (state.input.endDate && state.input.endDate < state.input.startDate) return copy.errors.endDateBeforeStart || "End date must be after start date";
      if (!["once", "weekly", "monthly"].includes(state.input.schedule)) return copy.errors.scheduleInvalid || "Invalid schedule";
      if (!cfgmod.COUNT_OPTIONS.includes(Number(state.input.count))) return copy.errors.countInvalid || "Invalid count";
      if (state.input.schedule === "weekly" && !cfgmod.WEEKDAYS.includes(state.input.weekday)) return copy.errors.weekdayInvalid || "Invalid weekday";
      if (state.input.schedule === "monthly") {
        if (!cfgmod.MONTH_POSITIONS.includes(state.input.monthPosition)) return copy.errors.monthPositionInvalid || "Invalid month position";
        if (!cfgmod.WEEKDAYS.includes(state.input.monthDay)) return copy.errors.weekdayInvalid || "Invalid weekday";
      }
    } else if (key === "details") {
      const td = cfgmod.typeDef(state.cfg, state.input.type);
      if (cfgmod.required(td, "title") && !state.input.title) {
        state.fieldErrors.title = true;
        return copy.errors.titleRequired || "Title required";
      }
      if (cfgmod.required(td, "location") && !state.input.location) {
        state.fieldErrors.location = true;
        return copy.errors.locationRequired || "Location required";
      }
      if (cfgmod.required(td, "url") && !state.input.url) {
        state.fieldErrors.url = true;
        return copy.errors.urlRequired || "URL required";
      }
      if (cfgmod.required(td, "description") && !(state.input.data && String(state.input.data.description || "").trim())) {
        state.fieldErrors.description = true;
        return copy.errors.descriptionRequired || "Description required";
      }
      if (state.input.url && !cfgmod.normalizeURL(state.input.url)) {
        state.fieldErrors.url = true;
        return copy.errors.urlInvalid || "Invalid URL";
      }
    }
    return "";
  }

  function buildCreatePayload() {
    return cfgmod.eventPayload(state.input, cfgmod.typeDef(state.cfg, state.input.type));
  }

  function buildUpdatePayload() {
    const payload = cfgmod.eventPayload(state.input, cfgmod.typeDef(state.cfg, state.input.type));
    return {
      title: payload.title || "",
      location: payload.location || "",
      url: payload.url || "",
      data: payload.data || {},
    };
  }

  async function submit() {
    if (state.submitting) return;
    state.submitting = true;
    setStatus("");
    try {
      if (state.editId) {
        await auth.apiFetch(`/api/${state.scope}/events/${encodeURIComponent(state.editId)}`, {
          method: "PATCH",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(buildUpdatePayload()),
        });
        // After save, navigate back to the detail view of this event.
        window.location.href = `${basePath}/event/?id=${encodeURIComponent(state.editId)}`;
        return;
      }
      const result = await auth.apiFetch(`/api/${state.scope}/events`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(buildCreatePayload()),
      });
      // On success, navigate to the events list (or to detail of first item).
      const items = (result && Array.isArray(result.items)) ? result.items : [];
      if (items.length === 1 && items[0].id) {
        window.location.href = `${basePath}/event/?id=${encodeURIComponent(items[0].id)}`;
      } else {
        window.location.href = `${basePath}/events/`;
      }
    } catch (err) {
      const msg = (err && err.message) || (copy.errors && copy.errors.saveFailed) || "Save failed";
      setStatus(msg);
      state.submitting = false;
    }
  }

  // ---------- Wiring ----------

  if (dom.back) {
    dom.back.addEventListener("click", () => {
      syncFromDOM();
      if (state.step <= 0) {
        window.location.href = `${basePath}/events/`;
        return;
      }
      state.step -= 1;
      render();
    });
  }

  if (cancelEditButton) {
    cancelEditButton.addEventListener("click", () => {
      window.location.href = `${basePath}/events/`;
    });
  }

  if (dom.form) {
    dom.form.addEventListener("blur", (e) => {
      const target = e.target;
      if (!target || target.tagName !== "INPUT") return;
      if (!["event_start_date", "event_start_time", "event_end_date"].includes(target.name)) return;
      syncFromDOM();
      if (currentStepKey() !== "dates") return;
      const err = validateStep();
      if (!err) return;
      setStatus(err);
      render();
      setStatus(err);
    }, true);

    dom.form.addEventListener("submit", (e) => {
      e.preventDefault();
      syncFromDOM();
      const err = validateStep();
      if (err) { setStatus(err); render(); setStatus(err); return; }
      if (isConfirm()) { submit(); return; }
      state.step += 1;
      render();
    });

    dom.form.addEventListener("change", async (e) => {
      syncFromDOM();
      if (e.target && e.target.name === "event_type") {
        // Type change resets group and may change the step plan. We also
        // refetch groups in case the user switches to a type that needs them.
        await loadGroupsFor(state.input.type);
        render();
      }
    });
  }

  if (dom.stage) {
    dom.stage.addEventListener("click", (e) => {
      const target = e.target.closest("[data-action]");
      if (!target) return;
      const action = target.dataset.action;
      if (action === "cancel-event") {
        if (!state.editId) return;
        actmod.openCancelDialog({
          auth,
          scope: state.scope,
          eventID: state.editId,
          onDone: () => { window.location.href = `${basePath}/events/`; },
        });
      } else if (action === "reschedule") {
        if (!state.editId || !state.loadedEvent) return;
        actmod.openRescheduleDialog({
          auth,
          scope: state.scope,
          eventID: state.editId,
          currentISO: state.loadedEvent.start_date || "",
          onDone: () => { window.location.href = `${basePath}/event/?id=${encodeURIComponent(state.editId)}`; },
        });
      }
    });
  }

  // ---------- Init ----------

  async function loadGroupsFor(typeValue) {
    // Only need groups when the type requires them.
    const td = cfgmod.typeDef(state.cfg, typeValue);
    if (!cfgmod.required(td, "group")) {
      state.availableGroups = [];
      return;
    }
    try {
      const url = state.scope === "admin" ? "/api/admin/groups" : "/api/me/groups";
      const payload = await auth.apiFetch(url);
      const items = (payload && Array.isArray(payload.items)) ? payload.items : [];
      // For non-admin members, only groups where they're assistant can host meetups.
      const filtered = state.isAdmin
        ? items
        : items.filter((g) => !!g && (g.is_assistant === true));
      state.availableGroups = filtered.map((g) => ({ id: g.id, name: g.name }));
    } catch (_) {
      state.availableGroups = [];
    }
  }

  async function loadEventForEdit() {
    if (!state.editId) return;
    const url = `/api/${state.scope}/events/${encodeURIComponent(state.editId)}`;
    let payload;
    try {
      payload = await auth.apiFetch(url);
    } catch (_) {
      // Surface a meaningful, user-facing error instead of letting the raw
      // 404/500 message bubble up into the generic "Wizard unavailable" box.
      const msg = (copy.errors && copy.errors.eventNotFound) || "Event not found.";
      throw new Error(msg);
    }
    const ev = (payload && payload.event) || {};
    state.loadedEvent = ev;
    state.input.type = ev.type || "";
    state.input.title = ev.title || "";
    state.input.location = ev.location || "";
    state.input.url = ev.url || "";
    if (ev.data && typeof ev.data === "object") {
      state.input.data = state.input.data || {};
      state.input.data.description = String(ev.data.description || "");
    }
    if (ev.start_date) {
      const raw = String(ev.start_date);
      state.input.startDate = raw.slice(0, 10);
      const m = raw.match(/[T ](\d{2}):(\d{2})/);
      if (m) state.input.startTime = `${m[1]}:${m[2]}`;
    }
    if (ev.end_date) state.input.endDate = String(ev.end_date).slice(0, 10);
    if (ev.cadence) applyCadenceToInput(String(ev.cadence));
    if (ev.count) state.input.count = Number(ev.count) || 3;
    if (ev.group) {
      state.input.group = ev.group;
    }
  }

  function detectIsAdmin() {
    try {
      const raw = localStorage.getItem("app.auth");
      const parsed = raw ? JSON.parse(raw) : null;
      return !!(parsed && parsed.model && parsed.model.admin === true);
    } catch (_) { return false; }
  }

  async function init() {
    try {
      cfgmod.validateCopy(copyAll);
      state.isAdmin = state.scope === "admin" || detectIsAdmin();
      const settings = await auth.apiFetch(`/api/${state.scope}/settings/eventflow`);
      state.cfg = settings && settings.data ? settings.data : settings;
      cfgmod.validateSettings(state.cfg);

      if (state.editId) {
        await loadEventForEdit();
      } else {
        // For brand-new events, default the type to the first the user can
        // create. Defaulting (instead of leaving it empty) keeps the step
        // plan correct from first render — otherwise stepKeys() would inject
        // a "group" step that may not be needed once a real type is chosen.
        const allowed = cfgmod.allowedTypes(state.cfg, state.isAdmin);
        if (allowed.length > 0) state.input.type = allowed[0].value;
      }

      // Pre-load groups (for create flow when meetup is the chosen type, or
      // always for admin since they may switch types). For edit mode we
      // skip the group step entirely so we don't need this.
      if (!state.editId) {
        await loadGroupsFor(state.input.type || "meetup");
      }

      if (dom.back) dom.back.textContent = (copy.backLabel || "Back").toUpperCase();
      showWizard();
      render();
    } catch (err) {
      setLoading(false);
      if (dom.stage) {
        dom.stage.innerHTML = `
          <section class="empty-state empty-state-icon-only" aria-label="Events">
            <p class="empty-state-label">${rndmod.esc((err && err.message) || (copy.emptyStateLabel || "Wizard unavailable"))}</p>
          </section>
        `;
      }
      if (dom.form) dom.form.hidden = false;
    }
  }

  function applyCadenceToInput(cadence) {
    if (!cadence || cadence === "once") {
      state.input.schedule = "once";
      return;
    }
    if (cadence.startsWith("weekly:")) {
      state.input.schedule = "weekly";
      const day = cadence.slice("weekly:".length);
      if (cfgmod.WEEKDAYS.includes(day)) state.input.weekday = day;
      return;
    }
    if (cadence.startsWith("monthly:")) {
      state.input.schedule = "monthly";
      const parts = cadence.slice("monthly:".length).split("-");
      if (cfgmod.MONTH_POSITIONS.includes(parts[0])) state.input.monthPosition = parts[0];
      if (cfgmod.WEEKDAYS.includes(parts[1])) state.input.monthDay = parts[1];
    }
  }

  init();
})();
