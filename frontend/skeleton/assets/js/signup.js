(() => {
  const form = document.querySelector("#signup-form");
  const emailField = document.querySelector("#signup-email");
  const passwordField = document.querySelector("#signup-password");
  const passwordConfirmField = document.querySelector("#signup-password-confirm");
  const statusNode = document.querySelector("#signup-status");

  if (!form || !emailField || !passwordField || !passwordConfirmField || !statusNode) {
    return;
  }

  form.addEventListener("submit", async (event) => {
    event.preventDefault();
    statusNode.hidden = true;
    statusNode.textContent = "";

    const email = emailField.value.trim().toLowerCase();
    const password = passwordField.value;
    const passwordConfirm = passwordConfirmField.value;

    if (!email || !password || !passwordConfirm) {
      statusNode.textContent = "All fields are required.";
      statusNode.hidden = false;
      return;
    }

    try {
      const response = await fetch("/api/collections/users/records", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          email,
          password,
          passwordConfirm,
        }),
      });

      if (!response.ok) {
        throw new Error("signup_failed");
      }

      window.location.href = "/";
    } catch (_) {
      statusNode.textContent = "Signup failed.";
      statusNode.hidden = false;
    }
  });
})();
