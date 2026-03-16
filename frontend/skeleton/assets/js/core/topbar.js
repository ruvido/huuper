(() => {
  const storageKey = "huuper.auth";
  const profileNode = document.querySelector("[data-topbar-profile]");
  if (!profileNode) {
    return;
  }

  let auth = null;
  try {
    const raw = localStorage.getItem(storageKey);
    auth = raw ? JSON.parse(raw) : null;
  } catch (_) {
    auth = null;
  }

  const model = auth && auth.model ? auth.model : null;
  if (!model) {
    return;
  }

  const fullName =
    (model.data && typeof model.data.full_name === "string" && model.data.full_name.trim()) ||
    (typeof model.email === "string" && model.email.trim()) ||
    "Profile";

  const avatar =
    typeof model.avatar === "string" && model.avatar.trim()
      ? `/api/files/users/${encodeURIComponent(model.id)}/${encodeURIComponent(model.avatar.trim())}`
      : "";

  if (avatar) {
    profileNode.innerHTML = `<img class="topbar-avatar-image" src="${avatar}" alt="${fullName}" />`;
    return;
  }

  const parts = String(fullName).trim().split(/\s+/).filter(Boolean).slice(0, 2);
  const initials = parts.length > 0 ? parts.map((part) => part[0] || "").join("").toUpperCase() : "?";
  profileNode.innerHTML = `<span class="topbar-avatar-fallback">${initials}</span>`;
})();
