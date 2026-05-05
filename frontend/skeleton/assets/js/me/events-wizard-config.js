// Pure helpers for the events wizard. No DOM mutation, no state mutation.
// All functions take their dependencies as parameters; nothing reads from
// outer closures or implicit globals. The orchestrator imports these via the
// shared `window.appEventsWizard` namespace.
(() => {
  window.appEventsWizard = window.appEventsWizard || {};
  const ns = window.appEventsWizard;

  function pad2(n) { return String(n).padStart(2, "0"); }

  const WEEKDAYS = ["sun", "mon", "tue", "wed", "thu", "fri", "sat"];
  const MONTH_POSITIONS = ["1st", "2nd", "3rd", "4th", "5th", "last"];
  const COUNT_OPTIONS = [1, 3, 6, 12];

  function typeDef(cfg, value) {
    if (!cfg || !Array.isArray(cfg.types)) return null;
    return cfg.types.find((t) => t && t.value === value) || null;
  }

  function allowedTypes(cfg, isAdmin) {
    if (!cfg || !Array.isArray(cfg.types)) return [];
    return cfg.types.filter((t) => {
      if (!t || !t.value) return false;
      const creators = Array.isArray(t.creators) ? t.creators : [];
      if (creators.includes("admin_or_assistant")) return true;
      return creators.includes(isAdmin ? "admin" : "assistant");
    });
  }

  function required(type, field) {
    return !!(type && type.required && type.required[field] === true);
  }

  function isValidISODate(value) {
    if (!value) return false;
    const m = String(value).match(/^\d{4}-\d{2}-\d{2}$/);
    if (!m) return false;
    const d = new Date(`${value}T12:00:00`);
    return !Number.isNaN(d.getTime());
  }

  function todayISO() {
    const d = new Date();
    const yyyy = d.getFullYear();
    const mm = pad2(d.getMonth() + 1);
    const dd = pad2(d.getDate());
    return `${yyyy}-${mm}-${dd}`;
  }

  function combineDateTime(dateISO, timeHHMM) {
    if (!isValidISODate(dateISO)) return "";
    const time = /^\d{2}:\d{2}$/.test(String(timeHHMM || "")) ? timeHHMM : "00:00";
    return `${dateISO}T${time}:00`;
  }

  function weekdayFromDate(startISO) {
    if (!isValidISODate(startISO)) return "mon";
    const d = new Date(`${startISO}T12:00:00`);
    return WEEKDAYS[d.getDay()] || "mon";
  }

  function monthPositionFromDate(startISO) {
    if (!isValidISODate(startISO)) return "1st";
    const d = new Date(`${startISO}T12:00:00`);
    const day = d.getDate();
    const nth = Math.floor((day - 1) / 7) + 1;
    const nextWeek = new Date(d.getFullYear(), d.getMonth(), day + 7, 12, 0, 0);
    if (nextWeek.getMonth() !== d.getMonth()) return "last";
    return `${nth}${nth === 1 ? "st" : nth === 2 ? "nd" : nth === 3 ? "rd" : "th"}`;
  }

  function cadenceFromInput(input) {
    const schedule = input && input.schedule ? input.schedule : "once";
    if (schedule === "weekly") return `weekly:${weekdayFromDate(input.startDate)}`;
    if (schedule === "monthly") return `monthly:${monthPositionFromDate(input.startDate)}-${weekdayFromDate(input.startDate)}`;
    return "once";
  }

  function generateScheduleDates(startISO, cadence, count) {
    const out = [];
    if (!isValidISODate(startISO)) return out;
    const safeCount = Math.max(1, Math.min(Number(count) || 1, 12));
    const [yy, mm, dd] = startISO.split("-").map((s) => Number(s));
    const base = new Date(yy, mm - 1, dd, 12, 0, 0);
    if (!cadence || cadence === "once") return [startISO];
    for (let i = 0; i < safeCount; i += 1) {
      const next = new Date(base.getTime());
      if (cadence.startsWith("weekly:")) {
        next.setDate(base.getDate() + i * 7);
      } else if (cadence.startsWith("monthly:")) {
        const parts = cadence.slice("monthly:".length).split("-");
        const pos = parts[0] || "1";
        const day = WEEKDAYS.indexOf(parts[1] || "mon");
        const anchor = new Date(base.getFullYear(), base.getMonth() + i, 1, 12, 0, 0);
        const weekday = day >= 0 ? day : 1;
        if (pos === "last") {
          const last = new Date(anchor.getFullYear(), anchor.getMonth() + 1, 0, 12, 0, 0);
          last.setDate(last.getDate() - ((last.getDay() - weekday + 7) % 7));
          next.setTime(last.getTime());
        } else {
        const nthMap = { "1st": 1, "2nd": 2, "3rd": 3, "4th": 4, "5th": 5 };
        const nth = nthMap[pos] || 1;
          const offset = (weekday - anchor.getDay() + 7) % 7;
          next.setTime(anchor.getTime());
          next.setDate(1 + offset + (nth - 1) * 7);
          if (next.getMonth() !== anchor.getMonth()) next.setDate(next.getDate() - 7);
        }
      } else {
        return [startISO];
      }
      const yyyy = next.getFullYear();
      const m2 = pad2(next.getMonth() + 1);
      const d2 = pad2(next.getDate());
      out.push(`${yyyy}-${m2}-${d2}`);
    }
    return out;
  }

  function validateCopy(c) {
    if (!c) throw new Error("events copy missing");
    if (!c.wizard) throw new Error("events copy missing wizard");
    const w = c.wizard;
    if (!w.steps) throw new Error("events copy missing wizard.steps");
    if (!w.labels) throw new Error("events copy missing wizard.labels");
    if (!w.placeholders) throw new Error("events copy missing wizard.placeholders");
    if (!w.actions) throw new Error("events copy missing wizard.actions");
    if (!w.errors) throw new Error("events copy missing wizard.errors");
  }

  function validateSettings(cfg) {
    if (!cfg || !Array.isArray(cfg.types) || cfg.types.length === 0) {
      throw new Error("eventflow settings missing types");
    }
  }

  ns.config = {
    WEEKDAYS,
    MONTH_POSITIONS,
    COUNT_OPTIONS,
    pad2,
    typeDef,
    allowedTypes,
    required,
    isValidISODate,
    todayISO,
    combineDateTime,
    weekdayFromDate,
    monthPositionFromDate,
    cadenceFromInput,
    generateScheduleDates,
    validateCopy,
    validateSettings,
  };
})();
