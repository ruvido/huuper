window.huuperRequestActions = (() => {
  function wait(ms) {
    return new Promise((resolve) => window.setTimeout(resolve, ms));
  }

  async function submitAndRedirect(config) {
    const button = config && config.button ? config.button : null;
    const minDurationMs = Number.isFinite(config && config.minDurationMs) ? config.minDurationMs : 1000;
    const startedAt = Date.now();
    const previousDisabled = button ? button.disabled : false;

    if (button) {
      button.disabled = true;
      button.classList.add("request-action-loading");
    }

    try {
      await window.huuperAuth.apiFetch(config.actionURL, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(config.body || {}),
      });

      const elapsed = Date.now() - startedAt;
      if (elapsed < minDurationMs) {
        await wait(minDurationMs - elapsed);
      }

      window.location.replace(String(config.redirectURL || "").trim() || "/");
    } catch (error) {
      if (button) {
        button.disabled = previousDisabled;
        button.classList.remove("request-action-loading");
      }
      throw error;
    }
  }

  return { submitAndRedirect };
})();
