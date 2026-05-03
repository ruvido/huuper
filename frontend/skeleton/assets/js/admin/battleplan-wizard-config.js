// Pure helpers for the battleplan wizard. No DOM mutation, no state mutation.
// All functions take their dependencies as parameters; nothing reads from
// outer closures or implicit globals. The orchestrator imports these via the
// shared `window.appBattleplanWizard` namespace.
(() => {
  window.appBattleplanWizard = window.appBattleplanWizard || {};
  const ns = window.appBattleplanWizard;

  function pad2(n) { return String(n).padStart(2, "0"); }

  function priorityCopy(state) {
    const p = (state && state.cfg && state.cfg.priority) || {};
    return (state.editId || state.viewId) ? (p.edit || {}) : (p.new || {});
  }

  // True when an existing draft record is loaded for editing (vs an active plan).
  function isEditingDraft(state) {
    return !!(state && state.editId && state.loadedStatus === "draft");
  }

  // True when the next save will create/keep a draft record:
  //  - editing an existing draft, or
  //  - creating a brand-new plan while an active one already exists.
  function willSaveAsDraft(state) {
    if (state.editId) return isEditingDraft(state);
    return !!state.hasExistingActive;
  }

  function defaultDurationValue(cfg) {
    const byDefault = cfg.durations.find((item) => item && item.default);
    if (byDefault && Number.isFinite(byDefault.value)) return Number(byDefault.value);
    const first = cfg.durations.find((item) => item && Number.isFinite(item.value));
    return first ? Number(first.value) : 0;
  }

  function defaultVisibilityValue(cfg) {
    const byDefault = cfg.visibility.find((item) => item && item.default);
    if (byDefault && byDefault.value) return byDefault.value;
    const first = cfg.visibility.find((item) => item && item.value);
    return first ? first.value : "";
  }

  function defaultCadenceType(cfg) {
    const byDefault = cfg.cadences.find((item) => item && item.default && item.type);
    if (byDefault) return byDefault.type;
    const first = cfg.cadences.find((item) => item && item.type);
    return first ? first.type : "daily";
  }

  function introEnabledBySettings(cfg) {
    const intro = cfg && cfg.wizard && cfg.wizard.intro;
    if (!intro || typeof intro !== "object") return true;
    if (typeof intro.show === "boolean") return intro.show;
    return true;
  }

  function shouldShowIntro(state) {
    return introEnabledBySettings(state.cfg) && !state.hasExistingBattleplan;
  }

  function visibilityLabel(value, cfg) {
    const def = cfg.visibility.find((v) => v.value === value);
    return def ? def.label : value;
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

  function isPillarComplete(value) {
    if (!value || !(value.objective || "").trim()) return false;
    return (value.routines || []).some((r) => isRoutineComplete(r));
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
    if (!c.confirmLabel) throw new Error("battleplan copy missing confirmLabel");
    if (!c.saveDraftLabel) throw new Error("battleplan copy missing saveDraftLabel");
  }

  function validateSettings(cfg) {
    if (!cfg || !cfg.priority || !cfg.wizard || !cfg.wizard.intro || !cfg.wizard.confirmation) {
      throw new Error("battleplan settings incomplete");
    }
    if (introEnabledBySettings(cfg)) {
      if (!cfg.wizard.intro.title) throw new Error("battleplan settings missing wizard.intro.title");
      if (!cfg.wizard.intro.button) throw new Error("battleplan settings missing wizard.intro.button");
    }
    if (!cfg.wizard.confirmation.title) throw new Error("battleplan settings missing wizard.confirmation.title");
    if (!cfg.priority.new || typeof cfg.priority.new !== "object") throw new Error("battleplan settings missing priority.new");
    if (!cfg.priority.new.title) throw new Error("battleplan settings missing priority.new.title");
    if (!cfg.priority.new.text) throw new Error("battleplan settings missing priority.new.text");
    if (!cfg.priority.edit || typeof cfg.priority.edit !== "object") throw new Error("battleplan settings missing priority.edit");
    if (!cfg.priority.edit.title) throw new Error("battleplan settings missing priority.edit.title");
    if (!cfg.priority.edit.text) throw new Error("battleplan settings missing priority.edit.text");
    if (!Array.isArray(cfg.durations) || cfg.durations.length === 0) throw new Error("battleplan settings missing durations");
    if (!cfg.durations.some((item) => item.default)) throw new Error("battleplan settings missing default duration");
    for (const item of cfg.durations) {
      if (!Number.isFinite(item.value)) throw new Error("battleplan settings invalid durations.value");
    }
    if (!Array.isArray(cfg.visibility) || cfg.visibility.length === 0) throw new Error("battleplan settings missing visibility");
    if (!cfg.visibility.some((item) => item.default)) throw new Error("battleplan settings missing default visibility");
  }

  ns.config = {
    pad2,
    priorityCopy,
    isEditingDraft,
    willSaveAsDraft,
    defaultDurationValue,
    defaultVisibilityValue,
    defaultCadenceType,
    introEnabledBySettings,
    shouldShowIntro,
    visibilityLabel,
    isRoutineComplete,
    isPillarComplete,
    routineCadenceLabel,
    validateCopy,
    validateSettings,
  };
})();
