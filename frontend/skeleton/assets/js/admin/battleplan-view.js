(() => {
  const auth = window.huuperAuth;
  if (!auth) return;

  const PREFIX = "battleplan-view";
  const $ = (suffix) => document.getElementById(`${PREFIX}-${suffix}`);

  const dom = {
    page: document.querySelector(".wizard-page"),
    form: $("form"),
    stage: $("stage"),
    loading: $("loading"),
    sticky: $("sticky"),
    back: $("back"),
    nextLabel: $("next-label"),
    next: $("next"),
  };

  const bpCopy = (window.huuperCopy && window.huuperCopy.battleplan) || {};
  const copy = bpCopy.wizard || {};
  const labels = copy.labels || {};

  function esc(str) {
    return String(str || "")
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;")
      .replace(/'/g, "&#39;");
  }

  function cadenceLabel(cadence) {
    if (!cadence || !cadence.type || cadence.type === "paused") return "";
    const times = Number(cadence.times);
    if (cadence.type === "times_per_week" && Number.isFinite(times) && times > 0) return `${times}X`;
    if (cadence.type === "specific_days") {
      const order = ["mon", "tue", "wed", "thu", "fri", "sat", "sun"];
      const map = { mon: "M", tue: "T", wed: "W", thu: "T", fri: "F", sat: "S", sun: "S" };
      const days = Array.isArray(cadence.days) ? cadence.days : [];
      const parts = order.filter((d) => days.includes(d)).map((d) => map[d]);
      return parts.length ? `(${parts.join(",")})` : "";
    }
    if (cadence.type === "daily") return "W";
    return "";
  }

  function durationDays(bp) {
    if (!bp.start_date || !bp.end_date) return null;
    const s = new Date(bp.start_date);
    const e = new Date(bp.end_date);
    if (isNaN(s) || isNaN(e)) return null;
    return Math.round((e - s) / (1000 * 60 * 60 * 24));
  }

  function visibilityLabel(cfg, value) {
    const match = (cfg.visibility || []).find((v) => v.value === value);
    return match ? match.label : value || "";
  }

  function endDateLabel(bp) {
    if (!bp.end_date) return "";
    const d = new Date(bp.end_date);
    if (isNaN(d)) return "";
    const months = ["Jan","Feb","Mar","Apr","May","Jun","Jul","Aug","Sep","Oct","Nov","Dec"];
    return `${d.getDate()} ${months[d.getMonth()]} ${d.getFullYear()}`;
  }

  function renderSummary(bp, cfg, id) {
    const data = bp.data || {};
    const priority = data.priority || {};
    const priorityText = (priority.title || "").trim();
    const priorityWhy = (priority.why || "").trim();
    const days = durationDays(bp);
    const endLabel = endDateLabel(bp);
    const vis = visibilityLabel(cfg, bp.visibility);

    const pillarsSummary = (cfg.pillars || []).map((def, pillarIndex) => {
      const value = (data.pillars && data.pillars[def.key]) || { objective: "", routines: [] };
      const objective = (value.objective || "").trim();
      const routines = value.routines || [];
      const routineItems = routines
        .map((r) => {
          if (!r || !r.title) return "";
          if ((r.cadence || {}).type === "paused") return "";
          const title = String(r.title).trim();
          if (!title) return "";
          const label = cadenceLabel(r.cadence || {});
          return `<li><span>${esc(title)}</span>${label ? ` <span class="wizard-routine-count-label">${esc(label)}</span>` : ""}</li>`;
        })
        .filter(Boolean)
        .join("");
      return `
        <li class="wizard-pillars-item wizard-summary-edit-target" data-step="${pillarIndex + 1}" data-id="${esc(id)}">
          <div class="wizard-pillars-col wizard-pillars-col-label"><strong>${esc(def.label)}</strong></div>
          <div class="wizard-pillars-col wizard-pillars-col-content">
            ${objective ? `<p class="wizard-pillars-objective-inline">${esc(objective)}</p>` : ""}
            <ul class="wizard-summary-muted-card wizard-pillars-routines">${routineItems || "<li><span>-</span></li>"}</ul>
          </div>
        </li>
      `;
    }).join("");

    return `
      <section class="wizard-step wizard-step-confirm">
        <div class="wizard-summary">
          <div>
            <div class="wizard-summary-tags">
              ${days ? `<label class="field-label wizard-summary-tag wizard-summary-tag-outline">${days}${esc(copy.daysSuffix || "D")}</label>` : ""}
              ${vis ? `<label class="field-label wizard-summary-tag wizard-summary-tag-outline">${esc(vis).toUpperCase()}</label>` : ""}
              ${endLabel ? `<em class="wizard-summary-end-date">&mdash; ${esc(endLabel)}</em>` : ""}
            </div>
          </div>
          <h2 class="display-hero">${esc(priorityText || bpCopy.titleBattleplan || "Battleplan")}</h2>
          ${priorityWhy ? `<p class="wizard-view-why">${esc(priorityWhy)}</p>` : ""}
          <div>
            <ul class="wizard-pillars-summary">${pillarsSummary}</ul>
          </div>
        </div>
      </section>
    `;
  }

  async function init() {
    const params = new URLSearchParams(window.location.search);
    const id = params.get("id");
    if (!id) { window.location.href = "/admin/battleplan/"; return; }

    try {
      const [bp, settings] = await Promise.all([
        auth.apiFetch(`/api/me/battleplans/${encodeURIComponent(id)}`),
        auth.apiFetch("/api/me/settings/battleplan"),
      ]);
      const cfg = settings.data || settings;

      if (dom.loading) dom.loading.remove();

      // Apply same confirm classes the wizard uses
      if (dom.page) dom.page.classList.add("wizard-page-confirm");
      if (dom.form) {
        dom.form.classList.add("wizard-form-confirm");
        dom.form.hidden = false;
      }
      if (dom.stage) {
        dom.stage.classList.add("wizard-stage-confirm");
        dom.stage.innerHTML = renderSummary(bp, cfg, id);
      }


      // Reuse same back/next buttons from wizard-shell
      if (dom.back) {
        dom.back.textContent = (bpCopy.archiveLabel || "Archive").toUpperCase();
        dom.back.addEventListener("click", async () => {
          if (dom.back.disabled) return;
          dom.back.disabled = true;
          try {
            await auth.apiFetch(`/api/me/battleplans/${encodeURIComponent(id)}/status`, {
              method: "POST",
              headers: { "Content-Type": "application/json" },
              body: JSON.stringify({ status: "archived" }),
            });
            window.location.href = "/admin/battleplan/";
          } catch {
            dom.back.disabled = false;
          }
        });
      }
      if (dom.next) {
        dom.next.setAttribute("type", "button");
        dom.next.removeAttribute("form");
      }
      if (dom.nextLabel) dom.nextLabel.textContent = "EDIT";
      if (dom.next) {
        dom.next.addEventListener("click", () => {
          window.location.href = `/admin/battleplan/edit/?edit=${encodeURIComponent(id)}&step=2`;
        });
      }
      if (dom.sticky) dom.sticky.hidden = false;

    } catch {
      if (dom.loading) dom.loading.textContent = "Error loading battleplan.";
    }
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
