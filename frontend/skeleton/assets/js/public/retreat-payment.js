// Result page after a Stripe checkout. Wording comes from
// window.appCopy.retreats.public.payment, and a retreat can override it
// through payment_success_title / payment_success_body (and the cancelled
// pair) in its own `data` — nothing user-facing is written here.
(() => {
  const titleNode = document.querySelector("#retreat-payment-title");
  const messageNode = document.querySelector("#retreat-payment-message");
  const backNode = document.querySelector("#retreat-payment-back");
  if (!titleNode || !messageNode) {
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
  const paid = params.get("status") === "success";
  const slug = (params.get("retreat") || "").trim();
  const key = paid ? "success" : "cancelled";

  function render(overrides) {
    const data = overrides || {};
    titleNode.textContent = data[`payment_${key}_title`] || text(`payment.${key}.title`);
    messageNode.textContent = data[`payment_${key}_body`] || text(`payment.${key}.body`);
    // The retreat is the useful place to go back to, paid or not: home says
    // nothing about what just happened.
    if (backNode) {
      if (slug) {
        backNode.href = "/retreat/" + encodeURIComponent(slug);
      }
      backNode.textContent = text(slug ? "payment.backToRetreat" : "payment.backHome");
    }
  }

  if (!slug) {
    render(null);
    return;
  }

  fetch("/api/public/retreats/" + encodeURIComponent(slug))
    .then((response) => (response.ok ? response.json() : null))
    .then((payload) => render(payload && payload.retreat ? payload.retreat.data : null))
    .catch(() => render(null));
})();
