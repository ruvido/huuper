// Review page for the links in the organiser's emails. They land here with a
// token; the page fetches the request, shows it in full and asks. Opened on
// "confirm" it offers to approve; opened with ?action=archive it goes
// straight to putting the request aside with a note. Only the buttons change
// anything — opening a link never does, so a mail client prefetching it
// approves nobody. Wording comes from window.appCopy.retreats.public.accept.
(() => {
  const titleNode = document.querySelector("#retreat-accept-title");
  const messageNode = document.querySelector("#retreat-accept-message");
  const reviewNode = document.querySelector("#retreat-accept-review");
  const noteFieldNode = document.querySelector("#retreat-accept-note-field");
  const noteLabelNode = document.querySelector("#retreat-accept-note-label");
  const noteInput = document.querySelector("#retreat-accept-note");
  const choiceNode = document.querySelector("#retreat-accept-choice");
  const yesButton = document.querySelector("#retreat-accept-yes");
  const archiveYesButton = document.querySelector("#retreat-accept-archive-yes");
  const archiveButton = document.querySelector("#retreat-accept-archive");
  const noButton = document.querySelector("#retreat-accept-no");
  const actionsNode = document.querySelector("#retreat-accept-actions");
  const backNode = document.querySelector("#retreat-accept-back");
  if (!titleNode || !messageNode || !reviewNode) {
    return;
  }

  function text(path) {
    const parts = String(path || "").split(".");
    let node = (window.appCopy && window.appCopy.retreats && window.appCopy.retreats.public) || {};
    for (const part of parts) {
      if (!node || typeof node !== "object") return "";
      node = node[part];
    }
    return typeof node === "string" ? node : "";
  }

  const params = new URLSearchParams(window.location.search);
  const token = (params.get("token") || "").trim();
  const opensOnArchive = params.get("action") === "archive";
  const known = ["approved", "active", "already", "archived", "invalid", "failed", "declined"];

  let current = null;

  function addField(label, value) {
    if (!value) return;
    const block = document.createElement("div");
    block.className = "data-block";
    const labelNode = document.createElement("div");
    labelNode.className = "data-label";
    labelNode.textContent = label;
    const valueNode = document.createElement("div");
    valueNode.className = "data-value";
    valueNode.textContent = value;
    block.appendChild(labelNode);
    block.appendChild(valueNode);
    reviewNode.appendChild(block);
  }

  // The end state of every path through this page: one sentence on what
  // happened and a way out. The review and its buttons are gone by then —
  // except the organiser's own note on an archived request, which is the
  // one thing worth keeping on screen.
  function showOutcome(status, slug, note) {
    const key = known.includes(status) ? status : "invalid";
    titleNode.textContent = text(`accept.${key}.title`);
    messageNode.textContent = text(`accept.${key}.body`);
    reviewNode.innerHTML = "";
    reviewNode.hidden = true;
    noteFieldNode.hidden = true;
    choiceNode.hidden = true;
    if (note) {
      addField(text("accept.review.note"), note);
      reviewNode.hidden = false;
    }
    if (backNode) {
      if (slug) {
        backNode.href = "/retreat/" + encodeURIComponent(slug);
      }
      backNode.textContent = text(slug ? "accept.backToRetreat" : "accept.backHome");
    }
    actionsNode.hidden = false;
  }

  function renderFields(view) {
    const retreat = view.retreat || {};
    const registration = view.registration || {};
    reviewNode.innerHTML = "";
    addField(text("signup.guestFields.fullName"), registration.full_name);
    addField(text("accept.review.email"), registration.email);
    addField(text("signup.guestFields.mobile"), registration.mobile);
    addField(text("signup.guestFields.birthYear"), registration.birth_year);
    addField(text("signup.guestFields.maritalStatus"), registration.marital_status);
    addField(text("signup.guestFields.provenance"), registration.provenance);
    addField(text("accept.review.receivedOn"), registration.received_on);
    addField(text("accept.review.previous"), view.previous);
    if (retreat.dates) {
      addField(text("stats.when"), retreat.dates);
    }
    reviewNode.hidden = false;
  }

  function showConfirm(view) {
    const retreat = view.retreat || {};
    titleNode.textContent = text("accept.review.title");
    messageNode.textContent = text("accept.review.body").replace("{retreat}", retreat.title || "");
    renderFields(view);
    noteFieldNode.hidden = true;
    yesButton.hidden = false;
    archiveYesButton.hidden = true;
    archiveButton.hidden = false;
    choiceNode.hidden = false;
  }

  function showArchive(view) {
    const retreat = view.retreat || {};
    titleNode.textContent = text("accept.archive.title");
    messageNode.textContent = text("accept.archive.body").replace("{retreat}", retreat.title || "");
    renderFields(view);
    noteLabelNode.textContent = text("accept.review.note");
    noteInput.placeholder = text("accept.archive.notePlaceholder");
    noteFieldNode.hidden = false;
    yesButton.hidden = true;
    archiveYesButton.hidden = false;
    archiveButton.hidden = true;
    choiceNode.hidden = false;
    noteInput.focus();
  }

  async function decide(path, body, fallback) {
    const slug = ((current || {}).retreat || {}).slug;
    [yesButton, archiveYesButton, archiveButton, noButton].forEach((button) => { button.disabled = true; });
    try {
      const response = await fetch(path, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });
      const payload = await response.json().catch(() => ({}));
      showOutcome(payload.status || (response.ok ? fallback : "failed"), slug, body.note);
    } catch (_) {
      showOutcome("failed", slug);
    }
  }

  yesButton.textContent = text("accept.review.yes");
  archiveYesButton.textContent = text("accept.archive.yes");
  archiveButton.textContent = text("accept.review.archive");
  noButton.textContent = text("accept.review.no");

  yesButton.addEventListener("click", () => decide("/api/public/retreats/accept", { token }, "approved"));
  archiveButton.addEventListener("click", () => { if (current) showArchive(current); });
  archiveYesButton.addEventListener("click", () => decide("/api/public/retreats/archive", { token, note: noteInput.value.trim() }, "archived"));
  noButton.addEventListener("click", () => showOutcome("declined", ((current || {}).retreat || {}).slug));

  if (!token) {
    showOutcome("invalid");
    return;
  }

  titleNode.textContent = text("accept.loading");
  fetch("/api/public/retreats/accept/request?token=" + encodeURIComponent(token))
    .then((response) => response.json().catch(() => ({})).then((body) => ({ ok: response.ok, body })))
    .then(({ ok, body }) => {
      const slug = (body.retreat || {}).slug;
      if (!ok) {
        showOutcome(body.status || "invalid");
        return;
      }
      if (body.status === "rejected" || body.status === "cancelled") {
        showOutcome("archived", slug, body.note);
        return;
      }
      if (body.status !== "pending") {
        showOutcome("already", slug);
        return;
      }
      current = body;
      if (opensOnArchive) {
        showArchive(body);
      } else {
        showConfirm(body);
      }
    })
    .catch(() => showOutcome("failed"));
})();
