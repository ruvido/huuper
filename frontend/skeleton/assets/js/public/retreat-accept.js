// Landing page for the approve-from-email link. The organiser taps the button
// in the admin notification, the backend approves and redirects here with the
// outcome: this page only has to say, in words, what just happened. Wording
// comes from window.appCopy.retreats.public.accept — nothing is written here.
(() => {
  const titleNode = document.querySelector("#retreat-accept-title");
  const messageNode = document.querySelector("#retreat-accept-message");
  const backNode = document.querySelector("#retreat-accept-back");
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
  const slug = (params.get("retreat") || "").trim();
  // An unknown status is treated like an unusable link rather than shown raw.
  const known = ["approved", "active", "already", "invalid", "failed"];
  const status = params.get("status");
  const key = known.includes(status) ? status : "invalid";

  titleNode.textContent = text(`accept.${key}.title`);
  messageNode.textContent = text(`accept.${key}.body`);

  if (backNode) {
    if (slug) {
      backNode.href = "/retreat/" + encodeURIComponent(slug);
    }
    backNode.textContent = text(slug ? "accept.backToRetreat" : "accept.backHome");
  }
})();
