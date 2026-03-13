(() => {
  const form = document.querySelector("#login-form");
  const emailField = document.querySelector("#login-email");
  const passwordField = document.querySelector("#login-password");
  const errorNode = document.querySelector("#login-error");

  if (!form || !emailField || !passwordField || !errorNode || !window.huuperAuth) {
    return;
  }

  form.addEventListener("submit", async (event) => {
    event.preventDefault();
    errorNode.hidden = true;
    errorNode.textContent = "";

    const email = emailField.value.trim();
    const password = passwordField.value;

    if (!email || !password) {
      errorNode.textContent = "Email e password sono obbligatorie.";
      errorNode.hidden = false;
      return;
    }

    try {
      const result = await window.huuperAuth.login(email, password);
      window.location.href = result.scope === "admin" ? "/admin/" : "/me/";
    } catch (_) {
      errorNode.textContent = "Credenziali non valide.";
      errorNode.hidden = false;
    }
  });
})();
