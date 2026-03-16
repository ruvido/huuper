window.huuperAuth = (() => {
  const storageKey = "huuper.auth";
  const loginEndpoint = "/api/collections/users/auth-with-password";

  async function login(email, password) {
    const response = await fetch(loginEndpoint, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        identity: email.trim().toLowerCase(),
        password,
      }),
    });

    if (!response.ok) {
      const error = new Error("auth_failed");
      error.status = response.status;
      throw error;
    }

    const auth = await response.json();
    const scope = auth.record && auth.record.admin === true ? "admin" : "me";
    persist({
      scope,
      token: auth.token,
      model: auth.record || null,
    });
    return { scope, auth };
  }

  function persist(payload) {
    localStorage.setItem(storageKey, JSON.stringify(payload));
  }

  function clear() {
    localStorage.removeItem(storageKey);
  }

  function read() {
    const raw = localStorage.getItem(storageKey);
    if (!raw) return null;
    try {
      return JSON.parse(raw);
    } catch (_) {
      clear();
      return null;
    }
  }

  function isAdmin(auth) {
    return !!(auth && auth.model && auth.model.admin === true);
  }

  function requireScope(expectedScope) {
    const auth = read();
    if (!auth || !auth.token) {
      window.location.href = "/";
      return null;
    }

    const actualScope = isAdmin(auth) ? "admin" : "me";
    if (actualScope !== expectedScope) {
      window.location.href = actualScope === "admin" ? "/admin/" : "/me/";
      return null;
    }

    return auth;
  }

  async function apiFetch(url, options = {}) {
    const auth = read();
    const headers = new Headers(options.headers || {});
    if (auth && auth.token) {
      headers.set("Authorization", `Bearer ${auth.token}`);
    }

    const response = await fetch(url, {
      ...options,
      headers,
    });

    if (!response.ok) {
      const error = new Error("request_failed");
      error.status = response.status;
      try {
        const payload = await response.json();
        error.message = payload.message || error.message;
        error.payload = payload;
      } catch (_) {
        // Ignore non-JSON error bodies.
      }
      throw error;
    }

    return response.json();
  }

  return {
    login,
    clear,
    read,
    requireScope,
    apiFetch,
  };
})();
