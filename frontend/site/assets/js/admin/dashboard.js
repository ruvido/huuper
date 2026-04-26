(() => {
  const statusNode = document.getElementById("admin-status");
  if (!statusNode || !window.huuperAuth || !window.huuperListPage) {
    return;
  }

  const esc = window.huuperListPage.escapeHTML;
  const setStatus = (msg) => window.huuperListPage.setStatus(statusNode, msg);

  const nodes = {
    hero: document.getElementById("admin-hero"),
    heroTitle: document.getElementById("admin-hero-title"),
    blockMetrics: document.getElementById("block-metrics"),
    mcLegend: document.getElementById("mc-legend"),
    mcSvg: document.getElementById("mc-svg"),
    mcAxis: document.getElementById("mc-axis"),
    blockRegions: document.getElementById("block-regions"),
    regionsChart: document.getElementById("regions-chart"),
    regionsMeta: document.getElementById("regions-meta"),
    regionsTotal: document.getElementById("regions-total"),
    blockAge: document.getElementById("block-age"),
    ageChart: document.getElementById("age-chart"),
    ageMeta: document.getElementById("age-meta"),
    ageTotal: document.getElementById("age-total"),
    blockMarital: document.getElementById("block-marital"),
    maritalChart: document.getElementById("marital-chart"),
    maritalMeta: document.getElementById("marital-meta"),
    maritalTotal: document.getElementById("marital-total"),
    blockWork: document.getElementById("block-work"),
    workChart: document.getElementById("work-chart"),
    workTotal: document.getElementById("work-total"),
    blockPassions: document.getElementById("block-passions"),
    passionChart: document.getElementById("passion-chart"),
    passionTabs: document.getElementById("passion-tabs"),
    blockNext: document.getElementById("block-next-event"),
    nextLink: document.getElementById("next-event-link"),
    nextTitle: document.getElementById("next-event-title"),
    nextMeta: document.getElementById("next-event-meta"),
  };

  const copy =
    (window.huuperCopy && window.huuperCopy.ui && window.huuperCopy.ui.admin && window.huuperCopy.ui.admin.dashboard) || {};
  const metricLabels = copy.metrics || {};

  const METRICS = [
    { key: "utenti", label: metricLabels.utenti || "Utenti", color: "accent", axis: "left" },
    { key: "richieste", label: metricLabels.richieste || "Richieste", color: "text", axis: "right" },
    { key: "angeli", label: metricLabels.angeli || "Angeli", color: "muted", axis: "right" },
  ];

  let payload = null;
  let currentPassion = "sports";

  function show(node, hasData) {
    if (node) node.hidden = !hasData;
  }

  function renderHero(users) {
    const total = users.total || 0;
    const regions = (users.byRegion || []).filter((r) => r.name !== "nd").length;
    const template = copy.heroTemplate || "Il Branco è di {users} presenti su {regions} regioni";
    const html = template
      .replace("{users}", `<b>${esc(total)} uomini</b>`)
      .replace("{regions}", `${esc(regions)}`);
    nodes.heroTitle.innerHTML = html;
    show(nodes.hero, true);
  }

  function renderMetrics(series) {
    if (!series || !Array.isArray(series.labels)) return;
    const weekly = series.weekly || {};
    const totals = series.totals || {};
    const delta = series.delta || {};

    nodes.mcLegend.innerHTML = METRICS.map((s) => {
      const total = Number(totals[s.key] || 0);
      const d = Number(delta[s.key] || 0);
      const deltaHTML = d > 0 ? `<em class="mc-delta">+${d}</em>` : "";
      return `
        <div class="mc-item" data-color="${s.color}">
          <span class="mc-head">
            <span class="mc-dot" aria-hidden="true"></span>
            <span class="mc-lbl">${esc(s.label)}</span>
          </span>
          <span class="mc-val">${total}${deltaHTML}</span>
        </div>`;
    }).join("");

    const W = Math.max(280, Math.round(nodes.mcSvg.clientWidth || 340));
    const H = 180;
    nodes.mcSvg.setAttribute("viewBox", `0 0 ${W} ${H}`);
    const pad = { top: 14, bottom: 24, left: 32, right: 32 };
    const xs = [];
    for (let i = 0; i < 6; i++) {
      xs.push(pad.left + ((W - pad.left - pad.right) * i) / 5);
    }
    const plotH = H - pad.top - pad.bottom;

    const niceAxis = (v) => {
      if (v <= 5) return { step: 1, max: 5 };
      if (v <= 10) return { step: 2, max: 10 };
      if (v <= 25) return { step: 5, max: 25 };
      if (v <= 50) return { step: 10, max: 50 };
      if (v <= 100) return { step: 20, max: 100 };
      const mag = Math.pow(10, Math.floor(Math.log10(v)));
      const step = Math.ceil(v / (5 * mag)) * mag;
      return { step, max: step * 5 };
    };

    const leftVals = METRICS.filter((s) => s.axis === "left").flatMap((s) => weekly[s.key] || []);
    const rightVals = METRICS.filter((s) => s.axis === "right").flatMap((s) => weekly[s.key] || []);
    const left = niceAxis(Math.max(1, ...leftVals));
    const right = niceAxis(Math.max(1, ...rightVals));
    const maxLeft = left.max;
    const maxRight = right.max;

    const ticks = [];
    const tickLabels = [];
    for (let i = 0; i <= 5; i++) {
      const y = H - pad.bottom - (i / 5) * plotH;
      ticks.push(
        `<line x1="${pad.left}" x2="${W - pad.right}" y1="${y.toFixed(1)}" y2="${y.toFixed(
          1
        )}" stroke="var(--line)" stroke-dasharray="2 4"></line>`
      );
      tickLabels.push(
        `<text x="${pad.left - 6}" y="${(y + 3).toFixed(
          1
        )}" text-anchor="end" fill="var(--text-faint)" font-size="9" font-family="var(--font-family-ui)">${
          i * left.step
        }</text>`,
        `<text x="${W - pad.right + 6}" y="${(y + 3).toFixed(
          1
        )}" text-anchor="start" fill="var(--text-faint)" font-size="9" font-family="var(--font-family-ui)">${
          i * right.step
        }</text>`
      );
    }

    const strokes = {
      accent: "var(--accent)",
      text: "var(--text)",
      muted: "var(--text-dim)",
    };

    const paths = METRICS.map((s) => {
      const vals = weekly[s.key] || [0, 0, 0, 0, 0, 0];
      const stroke = strokes[s.color] || "var(--text-muted)";
      const width = 4;
      const dotR = 5;
      const maxScale = s.axis === "left" ? maxLeft : maxRight;
      const pts = vals.map((v, i) => {
        const y = H - pad.bottom - (v / maxScale) * plotH;
        return { x: xs[i], y };
      });
      const line = pts.map((p) => `${p.x.toFixed(1)},${p.y.toFixed(1)}`).join(" ");
      const dots = pts
        .map(
          (p) =>
            `<circle cx="${p.x.toFixed(1)}" cy="${p.y.toFixed(1)}" r="${dotR}" style="fill:${stroke}"></circle>`
        )
        .join("");
      return `<polyline points="${line}" style="fill:none;stroke:${stroke};stroke-width:${width};stroke-linecap:round;stroke-linejoin:round"></polyline>${dots}`;
    });

    nodes.mcSvg.innerHTML = ticks.join("") + tickLabels.join("") + paths.join("");

    nodes.mcAxis.innerHTML = series.labels.map((l) => `<span>${esc(l)}</span>`).join("");

    show(nodes.blockMetrics, true);
  }

  function hbarRow(name, count, width, variant, href) {
    const cls = variant ? `hbar-row hbar-row--${variant}` : "hbar-row";
    const fillStyle = width > 0 ? `style="width:${width}%"` : `style="width:0"`;
    const n = Number(count || 0);
    const inner = `
      <span class="hbar-name">${esc(name)}</span>
      <span class="hbar-track"><span class="hbar-track-fill" ${fillStyle}></span></span>
      <span class="hbar-val">${n}</span>`;
    if (href) {
      return `<a class="${cls}" href="${esc(href)}">${inner}</a>`;
    }
    return `<div class="${cls}">${inner}</div>`;
  }

  function filterHref(param, value) {
    if (!value) return "";
    if (value.startsWith("Altre ") || value.startsWith("Altri ")) return "";
    return `/admin/users/?${param}=${encodeURIComponent(value)}`;
  }

  function renderHbarList(target, items, { topAccent = true, muteName = null, filterParam = null } = {}) {
    const max = items.reduce((a, b) => Math.max(a, b.count), 0);
    target.innerHTML = items
      .map((it, i) => {
        const width = max > 0 ? Math.round((it.count / max) * 100) : 0;
        let variant = null;
        if (topAccent && i === 0 && it.count > 0) variant = "top";
        if (muteName && it.name === muteName) variant = "rest";
        const href = filterParam ? filterHref(filterParam, it.name) : "";
        return hbarRow(it.name, it.count, width, variant, href);
      })
      .join("");
  }

  function renderRegions(users) {
    const items = users.byRegion || [];
    if (items.length === 0) return;
    const visible = items.slice(0, 7);
    const rest = items.slice(7);
    const restTotal = rest.reduce((a, b) => a + b.count, 0);
    const list = rest.length
      ? [...visible, { name: `Altre ${rest.length} regioni`, count: restTotal }]
      : visible;

    renderHbarList(nodes.regionsChart, list, {
      topAccent: true,
      muteName: rest.length ? `Altre ${rest.length} regioni` : null,
      filterParam: "region",
    });

    const regionCount = items.filter((r) => r.name !== "nd").length;
    nodes.regionsMeta.textContent = `${regionCount} regioni`;
    nodes.regionsTotal.textContent = `${users.total || 0} utenti`;
    show(nodes.blockRegions, true);
  }

  function renderAge(users) {
    const items = users.byAge || [];
    if (items.length === 0) return;
    renderHbarList(nodes.ageChart, items, {
      topAccent: false,
      muteName: "nd",
      filterParam: "age",
    });
    const muted = items.find((it) => it.name === "nd" && it.count === 0);
    if (muted) muted.count = 0;
    nodes.ageMeta.textContent = users.avgAge ? `media ${users.avgAge}a` : "";
    nodes.ageTotal.textContent = `${users.total || 0} utenti`;

    let peakIdx = -1;
    let peak = 0;
    items.forEach((it, i) => {
      if (it.name !== "nd" && it.count > peak) {
        peak = it.count;
        peakIdx = i;
      }
    });
    if (peakIdx >= 0) {
      const rows = nodes.ageChart.querySelectorAll(".hbar-row");
      if (rows[peakIdx]) rows[peakIdx].classList.add("hbar-row--top");
    }
    show(nodes.blockAge, true);
  }

  function renderMarital(users) {
    const items = users.marital || [];
    if (items.length === 0) return;
    renderHbarList(nodes.maritalChart, items, {
      topAccent: true,
      muteName: "nd",
      filterParam: "marital",
    });
    nodes.maritalTotal.textContent = `${users.total || 0} utenti`;
    nodes.maritalMeta.textContent = `${items.filter((i) => i.name !== "nd").length} categorie`;
    show(nodes.blockMarital, true);
  }

  function renderWork(users) {
    const items = users.work || [];
    if (items.length === 0) return;
    const visible = items.slice(0, 6);
    const rest = items.slice(6);
    const restTotal = rest.reduce((a, b) => a + b.count, 0);
    const list = rest.length
      ? [...visible, { name: `Altri ${rest.length}`, count: restTotal }]
      : visible;
    renderHbarList(nodes.workChart, list, {
      topAccent: true,
      muteName: rest.length ? `Altri ${rest.length}` : "nd",
      filterParam: "work",
    });
    nodes.workTotal.textContent = `${users.total || 0} utenti`;
    show(nodes.blockWork, true);
  }

  function renderPassions(kind) {
    currentPassion = kind;
    const items = (payload.users && payload.users[kind]) || [];
    if (items.length === 0) {
      nodes.passionChart.innerHTML = `<p class="meta-text">Nessun dato.</p>`;
    } else {
      const filterParam =
        kind === "sports" ? "sport" : kind === "interests" ? "interest" : "skill";
      renderHbarList(nodes.passionChart, items, { topAccent: true, filterParam });
    }
    nodes.passionTabs
      .querySelectorAll("button")
      .forEach((b) => b.classList.toggle("active", b.dataset.p === kind));
    show(nodes.blockPassions, true);
  }

  function renderNextEvent(next) {
    if (!next || !next.id) return;
    const date = next.event_date ? new Date(next.event_date) : null;
    const dateLabel =
      date && !Number.isNaN(date.getTime())
        ? date.toLocaleDateString("it-IT", { day: "2-digit", month: "short", year: "numeric" })
        : "";
    nodes.nextLink.href = `/admin/event/?id=${encodeURIComponent(next.id)}`;
    nodes.nextTitle.textContent = next.title || "Evento";
    const parts = [];
    if (dateLabel) parts.push(dateLabel);
    parts.push(`${next.registrations || 0} iscritti`);
    if (next.pending > 0) parts.push(`${next.pending} in attesa`);
    nodes.nextMeta.textContent = parts.join(" · ");
    show(nodes.blockNext, true);
  }

  async function load() {
    try {
      payload = await window.huuperAuth.apiFetch("/api/admin/summary");
      const users = payload.users || {};
      const events = payload.events || { total: 0 };

      renderHero(users);
      renderMetrics(payload.series);
      renderRegions(users);
      renderAge(users);
      renderMarital(users);
      renderWork(users);
      renderPassions(currentPassion);
      renderNextEvent(events.next);

      setStatus("");
    } catch (_) {
      setStatus("Summary unavailable.");
    }
  }

  if (nodes.passionTabs) {
    nodes.passionTabs.addEventListener("click", (e) => {
      const btn = e.target.closest("button[data-p]");
      if (btn && payload) renderPassions(btn.dataset.p);
    });
  }

  if (nodes.mcSvg && typeof ResizeObserver !== "undefined") {
    let lastWidth = 0;
    const ro = new ResizeObserver(() => {
      if (!payload || !payload.series) return;
      const w = Math.round(nodes.mcSvg.clientWidth || 0);
      if (Math.abs(w - lastWidth) < 4) return;
      lastWidth = w;
      renderMetrics(payload.series);
    });
    ro.observe(nodes.mcSvg);
  }

  load();
})();
