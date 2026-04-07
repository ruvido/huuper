(() => {
  const emailNode = document.querySelector("#profile-email");
  const telegramNode = document.querySelector("#profile-telegram");
  if (!emailNode || !telegramNode || !window.huuperAuth) {
    return;
  }

  const auth = window.huuperAuth.read();
  const model = auth && auth.model ? auth.model : {};
  const telegram = model.telegram || {};

  const email = typeof model.email === "string" ? model.email.trim() : "";
  const telegramName =
    typeof telegram.username === "string" && telegram.username.trim()
      ? telegram.username.trim()
      : typeof telegram.first_name === "string" && telegram.first_name.trim()
        ? telegram.first_name.trim()
        : "";

  emailNode.textContent = email || "Not available";
  telegramNode.textContent = telegramName || "";
})();
