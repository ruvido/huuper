// Landing page for a registration that went through without a payment to
// make: the guest pre-registration waiting for a call back, and the member on
// a retreat with no deposit. Leaving them on the form with a line of text
// underneath reads like nothing happened — a submitted form should land
// somewhere. Wording comes from window.appCopy.retreats.public.registered.
(() => {
  const titleNode = document.querySelector("#retreat-registered-title");
  const messageNode = document.querySelector("#retreat-registered-message");
  const backNode = document.querySelector("#retreat-registered-back");
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
  // Anything unexpected is treated as the guest case: it is the one that
  // promises nothing the reader has not already been told.
  const status = params.get("status") === "member" ? "member" : "guest";

  titleNode.textContent = text(`registered.${status}.title`);
  messageNode.textContent = text(`registered.${status}.body`);

  if (backNode) {
    if (slug) {
      backNode.href = "/retreat/" + encodeURIComponent(slug);
    }
    backNode.textContent = text(slug ? "registered.backToRetreat" : "registered.backHome");
  }
})();
