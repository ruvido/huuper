window.huuperAuth = (() => {
  const storageKey = "huuper.auth";
  const loginEndpoint = "/api/collections/users/auth-with-password";
  const refreshEndpoint = "/api/collections/users/auth-refresh";

  function currentPath() {
    return window.location.pathname + window.location.search + window.location.hash;
  }

  function normalizeRedirectTarget(value) {
    const raw = String(value || "").trim();
    if (!raw || raw.startsWith("//")) {
      return "";
    }

    try {
      const url = new URL(raw, window.location.origin);
      if (url.origin !== window.location.origin) {
        return "";
      }
      return url.pathname + url.search + url.hash;
    } catch (_) {
      return "";
    }
  }

  function scopeHome(scope) {
    return scope === "admin" ? "/admin/" : "/me/";
  }

  function nextFromQuery() {
    const params = new URLSearchParams(window.location.search);
    return normalizeRedirectTarget(params.get("next"));
  }

  function scopeRedirectPath(scope, next) {
    const normalizedNext = normalizeRedirectTarget(next);
    if (!normalizedNext) {
      return "";
    }

    if (scope === "admin" && normalizedNext.startsWith("/me/")) {
      return "/admin/" + normalizedNext.slice("/me/".length);
    }

    if (scope === "me" && normalizedNext.startsWith("/admin/")) {
      return "/me/" + normalizedNext.slice("/admin/".length);
    }

    if (normalizedNext.startsWith(scope === "admin" ? "/admin/" : "/me/")) {
      return normalizedNext;
    }

    return "";
  }

  function redirectToLogin(next = "") {
    const target = new URL("/", window.location.origin);
    const normalizedNext = normalizeRedirectTarget(next) || normalizeRedirectTarget(currentPath());
    if (normalizedNext) {
      target.searchParams.set("next", normalizedNext);
    }
    window.location.href = target.pathname + target.search + target.hash;
  }

  function redirectAfterLogin(scope) {
    const next = nextFromQuery();
    const scopedNext = scopeRedirectPath(scope, next);
    if (scopedNext) {
      return scopedNext;
    }
    return scopeHome(scope);
  }

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
      redirectToLogin(currentPath());
      return null;
    }

    const actualScope = isAdmin(auth) ? "admin" : "me";
    if (actualScope !== expectedScope) {
      const target = scopeRedirectPath(actualScope, currentPath());
      window.location.href = target || (actualScope === "admin" ? "/admin/" : "/me/");
      return null;
    }

    refresh().then((nextAuth) => {
      if (!nextAuth || !nextAuth.token) {
        redirectToLogin(currentPath());
        return;
      }
      const nextScope = isAdmin(nextAuth) ? "admin" : "me";
      if (nextScope !== expectedScope) {
        const target = scopeRedirectPath(nextScope, currentPath());
        window.location.href = target || (nextScope === "admin" ? "/admin/" : "/me/");
      }
    }).catch(() => {
      clear();
      redirectToLogin(currentPath());
    });

    return auth;
  }

  async function refresh() {
    const auth = read();
    if (!auth || !auth.token) {
      return null;
    }

    const response = await fetch(refreshEndpoint, {
      method: "POST",
      headers: {
        "Authorization": `Bearer ${auth.token}`,
      },
    });

    if (!response.ok) {
      const error = new Error("auth_refresh_failed");
      error.status = response.status;
      throw error;
    }

    const payload = await response.json();
    const nextAuth = {
      scope: payload.record && payload.record.admin === true ? "admin" : "me",
      token: payload.token,
      model: payload.record || null,
    };
    persist(nextAuth);
    return nextAuth;
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
    refresh,
    requireScope,
    apiFetch,
    redirectAfterLogin,
    redirectToLogin,
    normalizeRedirectTarget,
    scopeRedirectPath,
  };
})();
