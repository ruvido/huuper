// Review page for the approve-from-email link. The organiser taps the button
// in the admin notification and lands here with a token; the page fetches the
// request, shows it in full and asks. Only "yes" changes anything — opening
// the link never does, so a mail client prefetching it approves nobody.
// Wording comes from window.appCopy.retreats.public.accept — nothing is
// written here.
(() => {
  const titleNode = document.querySelector("#retreat-accept-title");
  const messageNode = document.querySelector("#retreat-accept-message");
  const reviewNode = document.querySelector("#retreat-accept-review");
  const choiceNode = document.querySelector("#retreat-accept-choice");
  const yesButton = document.querySelector("#retreat-accept-yes");
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
  const known = ["approved", "active", "already", "archived", "invalid", "failed", "declined"];

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

  function showReview(view) {
    const retreat = view.retreat || {};
    const registration = view.registration || {};

    titleNode.textContent = text("accept.review.title");
    messageNode.textContent = text("accept.review.body").replace("{retreat}", retreat.title || "");

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

    yesButton.textContent = text("accept.review.yes");
    noButton.textContent = text("accept.review.no");
    choiceNode.hidden = false;

    yesButton.addEventListener("click", async () => {
      yesButton.disabled = true;
      noButton.disabled = true;
      try {
        const response = await fetch("/api/public/retreats/accept", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ token }),
        });
        const body = await response.json().catch(() => ({}));
        showOutcome(body.status || (response.ok ? "approved" : "failed"), retreat.slug);
      } catch (_) {
        showOutcome("failed", retreat.slug);
      }
    }, { once: true });

    noButton.addEventListener("click", () => showOutcome("declined", retreat.slug), { once: true });
  }

  if (!token) {
    showOutcome("invalid");
    return;
  }

  titleNode.textContent = text("accept.loading");
  fetch("/api/public/retreats/accept/request?token=" + encodeURIComponent(token))
    .then((response) => response.json().catch(() => ({})).then((body) => ({ ok: response.ok, body })))
    .then(({ ok, body }) => {
      if (!ok) {
        showOutcome(body.status || "invalid");
        return;
      }
      if (body.status === "rejected" || body.status === "cancelled") {
        showOutcome("archived", (body.retreat || {}).slug, body.note);
        return;
      }
      if (body.status !== "pending") {
        showOutcome("already", (body.retreat || {}).slug);
        return;
      }
      showReview(body);
    })
    .catch(() => showOutcome("failed"));
})();
