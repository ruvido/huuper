(() => {
  const statusNode = document.getElementById("admin-status");
  if (!statusNode || !window.appAuth || !window.appListPage) {
    return;
  }

  const esc = window.appListPage.escapeHTML;
  const setStatus = (msg) => window.appListPage.setStatus(statusNode, msg);

  const nodes = {
    hero: document.getElementById("admin-hero"),
    heroTitle: document.getElementById("admin-hero-title"),
    blockMetrics: document.getElementById("block-metrics"),
    metricsUpdated: document.getElementById("metrics-updated"),
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
    (window.appCopy && window.appCopy.ui && window.appCopy.ui.admin && window.appCopy.ui.admin.dashboard) || {};
  const metricLabels = copy.metrics || {};

  const METRICS = [
    { key: "utenti", label: metricLabels.users, color: "accent", axis: "left" },
    { key: "angeli", label: metricLabels.guardians, color: "muted", axis: "right" },
    { key: "richieste", label: metricLabels.requests, color: "text", axis: "right" },
  ];

  let payload = null;
  let currentPassion = "sports";
  let lastLoadedAt = null;

  function formatAgo(ms) {
    const sec = Math.max(0, Math.floor(ms / 1000));
    if (sec < 60) return "aggiornato ora";
    const min = Math.floor(sec / 60);
    if (min < 60) return `aggiornato ${min} min fa`;
    const hr = Math.floor(min / 60);
    if (hr < 24) return `aggiornato ${hr} ${hr === 1 ? "ora" : "ore"} fa`;
    const day = Math.floor(hr / 24);
    return `aggiornato ${day} ${day === 1 ? "giorno" : "giorni"} fa`;
  }

  function updateAgo() {
    if (!nodes.metricsUpdated || !lastLoadedAt) return;
    nodes.metricsUpdated.textContent = formatAgo(Date.now() - lastLoadedAt);
  }

  function show(node, hasData) {
    if (node) node.hidden = !hasData;
  }

  function renderHero(users, groups) {
    const total = users.total || 0;
    const groupsTotal = (groups && groups.total) || 0;
    const regions = (users.byRegion || []).filter((r) => r.name !== "nd").length;
    const template = copy.heroTemplate || "Il Branco sono {users},<br>attivi in {groups} e {regions}";
    const html = template
      .replace("{users}", `<a href="/admin/users/"><b>${esc(total)} uomini</b></a>`)
      .replace("{groups}", `<a href="/admin/groups/"><b>${esc(groupsTotal)} gruppi</b></a>`)
      .replace("{regions}", `${esc(regions)} regioni`);
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
    const yTicks = 5;
    const yIntervals = yTicks - 1;
    nodes.mcSvg.setAttribute("viewBox", `0 0 ${W} ${H}`);
    const pad = { top: 14, bottom: 24, left: 32, right: 32 };
    const xs = [];
    for (let i = 0; i < 6; i++) {
      xs.push(pad.left + ((W - pad.left - pad.right) * i) / 5);
    }
    const plotH = H - pad.top - pad.bottom;

    const buildAxis = (vals) => {
      const finite = vals.filter((v) => Number.isFinite(v));
      if (!finite.length) {
        return { min: 0, max: yIntervals, step: 1 };
      }
      let min = Math.min(...finite);
      let max = Math.max(...finite);
      if (min === max) {
        const pad = Math.max(1, Math.abs(min) * 0.1);
        min -= pad;
        max += pad;
      }
      const span = Math.max(1, max - min);
      const step = Math.max(1, Math.ceil(span / yIntervals));
      const free = yIntervals * step - span;
      let axisMin = Math.max(0, Math.floor(min - free / 2));
      let axisMax = axisMin + yIntervals * step;
      if (axisMax < max) {
        axisMax = Math.ceil(max);
        axisMin = Math.max(0, axisMax - yIntervals * step);
        axisMax = axisMin + yIntervals * step;
      }
      return { min: axisMin, max: axisMax, step };
    };

    const formatTick = (value) => String(Math.max(0, value));

    const leftVals = METRICS.filter((s) => s.axis === "left").flatMap((s) => weekly[s.key] || []);
    const rightVals = METRICS.filter((s) => s.axis === "right").flatMap((s) => weekly[s.key] || []);
    const left = buildAxis(leftVals);
    const right = buildAxis(rightVals);
    const toY = (value, axis) => {
      const range = axis.max - axis.min || 1;
      return H - pad.bottom - ((value - axis.min) / range) * plotH;
    };

    const ticks = [];
    const tickLabels = [];
    for (let i = 0; i < yTicks; i++) {
      const y = H - pad.bottom - (i / yIntervals) * plotH;
      const leftValue = left.min + i * left.step;
      const rightValue = right.min + i * right.step;
      ticks.push(
        `<line x1="${pad.left}" x2="${W - pad.right}" y1="${y.toFixed(1)}" y2="${y.toFixed(
          1
        )}" stroke="var(--line)" stroke-dasharray="2 4"></line>`
      );
      tickLabels.push(
        `<text x="${pad.left - 6}" y="${(y + 3).toFixed(
          1
        )}" text-anchor="end" fill="var(--text-faint)" font-size="9" font-family="var(--font-family-ui)">${formatTick(
          leftValue
        )}</text>`,
        `<text x="${W - pad.right + 6}" y="${(y + 3).toFixed(
          1
        )}" text-anchor="start" fill="var(--text-faint)" font-size="9" font-family="var(--font-family-ui)">${formatTick(
          rightValue
        )}</text>`
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
      const axis = s.axis === "left" ? left : right;
      const pts = vals.map((v, i) => {
        const y = toY(v, axis);
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
      payload = await window.appAuth.apiFetch("/api/admin/summary");
      const users = payload.users || {};
      const events = payload.events || { total: 0 };

      lastLoadedAt = Date.now();
      updateAgo();
      renderHero(users, payload.groups);
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
  setInterval(updateAgo, 60000);
})();
