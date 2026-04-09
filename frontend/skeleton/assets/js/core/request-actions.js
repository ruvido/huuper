window.huuperRequestActions = (() => {
  async function submitAndRedirect(config) {
    await window.huuperAuth.apiFetch(config.actionURL, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(config.body || {}),
    });

    window.location.replace(String(config.redirectURL || "").trim() || "/");
  }

  return { submitAndRedirect };
})();
