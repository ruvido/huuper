// Shared cadence helpers for the events list/detail render. Mirrors backend
// internal/events/cadence.go: parseable strings ("once"/"weekly:mon"/...),
// occurrence expansion, next-upcoming with cancelled_dates skip.
// Published as window.appEventsCadence so render modules can stay decoupled.
(() => {
  window.appEventsCadence = window.appEventsCadence || {};
  const ns = window.appEventsCadence;

  const MONTHS = ["Jan","Feb","Mar","Apr","May","Jun","Jul","Aug","Sep","Oct","Nov","Dec"];

  function shortDate(value) {
    if (!value) return "";
    const d = new Date(value);
    if (Number.isNaN(d.getTime())) return "";
    return `${d.getDate()} ${MONTHS[d.getMonth()]} ${d.getFullYear()}`;
  }

  function cadenceLabel(cadence) {
    if (!cadence || cadence === "once") return "";
    if (cadence.startsWith("weekly:")) return "WEEKLY";
    if (cadence.startsWith("monthly:")) return "MONTHLY";
    return cadence.toUpperCase();
  }

  function parseWeekday(token) {
    const map = { sun: 0, mon: 1, tue: 2, wed: 3, thu: 4, fri: 5, sat: 6 };
    return map[(token || "").toLowerCase()] ?? -1;
  }

  function nthWeekdayOfMonth(year, monthIndex, weekday, nthToken) {
    if (nthToken === "last") {
      const lastDay = new Date(year, monthIndex + 1, 0);
      const offset = (lastDay.getDay() - weekday + 7) % 7;
      return lastDay.getDate() - offset;
    }
    const nth = parseInt(nthToken, 10);
    if (!nth) return 1;
    const first = new Date(year, monthIndex, 1);
    const offset = (weekday - first.getDay() + 7) % 7;
    let day = 1 + offset + 7 * (nth - 1);
    const maxDays = new Date(year, monthIndex + 1, 0).getDate();
    while (day > maxDays) day -= 7;
    return day;
  }

  function computeOccurrences(start, cadence, count) {
    if (!start || !(start instanceof Date) || Number.isNaN(start.getTime())) return [];
    const c = cadence || "once";
    if (c === "once") return [new Date(start)];
    const n = Math.max(1, Number(count) || 1);
    if (c.startsWith("weekly:")) {
      const out = [];
      for (let i = 0; i < n; i++) {
        const d = new Date(start);
        d.setDate(d.getDate() + 7 * i);
        out.push(d);
      }
      return out;
    }
    if (c.startsWith("monthly:")) {
      const spec = c.slice("monthly:".length).split("-");
      if (spec.length !== 2) return [new Date(start)];
      const wd = parseWeekday(spec[1]);
      if (wd < 0) return [new Date(start)];
      const baseHours = start.getHours();
      const baseMins = start.getMinutes();
      const out = [];
      for (let i = 0; i < n; i++) {
        const anchor = new Date(start.getFullYear(), start.getMonth() + i, 1);
        const day = nthWeekdayOfMonth(anchor.getFullYear(), anchor.getMonth(), wd, spec[0]);
        out.push(new Date(anchor.getFullYear(), anchor.getMonth(), day, baseHours, baseMins));
      }
      return out;
    }
    return [new Date(start)];
  }

  // nextOccurrence(item) → ISO date string of the soonest upcoming non-cancelled
  // occurrence, or the latest non-cancelled past occurrence as fallback. Returns
  // empty string when no schedule can be derived.
  function nextOccurrence(item) {
    if (!item || !item.start_date) return "";
    const start = new Date(item.start_date);
    if (Number.isNaN(start.getTime())) return "";
    const cancelled = new Set();
    const data = item.data || {};
    const list = Array.isArray(data.cancelled_dates) ? data.cancelled_dates : [];
    list.forEach((c) => {
      const d = new Date(c);
      if (!Number.isNaN(d.getTime())) cancelled.add(d.toISOString().slice(0, 10));
    });
    const occurrences = computeOccurrences(start, item.cadence || "once", item.count);
    const now = Date.now();
    let lastFuture = "";
    for (const occ of occurrences) {
      const key = occ.toISOString().slice(0, 10);
      if (cancelled.has(key)) continue;
      if (occ.getTime() >= now) return occ.toISOString();
      lastFuture = occ.toISOString();
    }
    return lastFuture;
  }

  // Build the list-meta string for an event row. Three cases:
  //   - cadence != once: "NEXT 4 MAY · WEEKLY × 3 · ROMA"
  //   - cadence == once + end_date diverso da start: "15 MAY → 17 MAY · ROMA"
  //   - cadence == once: "15 MAY · ROMA"
  function metaText(item) {
    const start = shortDate(item.start_date);
    const end = shortDate(item.end_date);
    const location = (item.location || "").trim();
    const cadence = item.cadence || "once";
    if (cadence !== "once") {
      const next = shortDate(nextOccurrence(item));
      const label = cadenceLabel(cadence);
      const count = item.count ? `× ${item.count}` : "";
      return [
        next ? `NEXT ${next}` : "",
        [label, count].filter(Boolean).join(" "),
        location.toUpperCase(),
      ].filter(Boolean).join(" · ");
    }
    const dateText = (start && end && start !== end) ? `${start} → ${end}` : start;
    return [dateText, location.toUpperCase()].filter(Boolean).join(" · ");
  }

  ns.shortDate = shortDate;
  ns.cadenceLabel = cadenceLabel;
  ns.computeOccurrences = computeOccurrences;
  ns.nextOccurrence = nextOccurrence;
  ns.metaText = metaText;
})();
