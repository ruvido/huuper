(() => {
  const publicFields = window.huuperPublicFields;
  const loginForm = document.querySelector("#login-form");
  const loginEmailField = document.querySelector("#login-email");
  const loginEmailErrorNode = document.querySelector("#login-email-error");
  const loginPasswordField = document.querySelector("#login-password");
  const loginPasswordErrorNode = document.querySelector("#login-password-error");
  const loginErrorNode = document.querySelector("#login-error");
  const loginRequestLink = document.querySelector("#public-login-request-link");

  if (
    !loginForm ||
    !loginEmailField ||
    !loginEmailErrorNode ||
    !loginPasswordField ||
    !loginPasswordErrorNode ||
    !loginErrorNode ||
    !loginRequestLink ||
    !publicFields ||
    !window.huuperAuth
  ) {
    return;
  }

  const auth = window.huuperAuth.read();
  if (auth && auth.token) {
    window.location.replace(window.huuperAuth.redirectAfterLogin(auth.model && auth.model.admin === true ? "admin" : "me"));
    return;
  }

  function hideStatus(node) {
    node.hidden = true;
    node.textContent = "";
    node.classList.remove("is-error");
  }

  function setStatus(node, message, isError = false) {
    node.textContent = message;
    node.hidden = false;
    node.classList.toggle("is-error", isError);
  }

  loginEmailField.addEventListener("input", () => hideStatus(loginEmailErrorNode));
  loginPasswordField.addEventListener("input", () => hideStatus(loginPasswordErrorNode));

  loginForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    hideStatus(loginErrorNode);
    hideStatus(loginEmailErrorNode);
    hideStatus(loginPasswordErrorNode);

    const emailAdvance = publicFields.validateFieldOnAdvance({ key: "email", type: "email" }, loginEmailField.value);
    if (emailAdvance.error) {
      setStatus(loginEmailErrorNode, emailAdvance.error, true);
      return;
    }

    const email = emailAdvance.normalized || String(loginEmailField.value || "").trim().toLowerCase();
    const password = String(loginPasswordField.value || "");

    if (!email || !password) {
      if (!email) {
        setStatus(loginEmailErrorNode, "Enter a valid email address.", true);
      }
      if (!password) {
        setStatus(loginPasswordErrorNode, "Password is required.", true);
      }
      return;
    }

    try {
      const result = await window.huuperAuth.login(email, password);
      window.location.href = window.huuperAuth.redirectAfterLogin(result.scope);
    } catch (_) {
      setStatus(loginErrorNode, "Invalid credentials.", true);
    }
  });
})();
