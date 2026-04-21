window.huuperRequestActions = (() => {
  function text(value) {
    if (window.huuperListPage && typeof window.huuperListPage.text === "function") {
      return window.huuperListPage.text(value);
    }
    return String(value || "").trim();
  }

  function errorMessage(error, fallback = "Action unavailable.") {
    const payload = error && typeof error === "object" ? error.payload : null;
    const payloadMessage = text(payload && payload.message);
    if (payloadMessage) {
      return payloadMessage;
    }

    const data = payload && typeof payload.data === "object" && payload.data ? payload.data : null;
    if (data) {
      const keys = Object.keys(data).filter((key) => text(key) !== "");
      if (keys.length > 0) {
        return keys[0];
      }
    }

    const direct = text(error && error.message);
    if (direct && direct !== "request_failed") {
      return direct;
    }

    return fallback;
  }

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

  return { submitAndRedirect, errorMessage };
})();
