(() => {
  const publicFields = window.appPublicFields;
  const form = document.querySelector("#password-reset-form");
  const emailField = document.querySelector("#password-reset-email");
  const emailErrorNode = document.querySelector("#password-reset-email-error");
  const statusNode = document.querySelector("#password-reset-status");

  if (!form || !emailField || !emailErrorNode || !statusNode || !publicFields) {
    return;
  }

  emailField.addEventListener("input", () => {
    emailErrorNode.hidden = true;
    emailErrorNode.textContent = "";
  });

  form.addEventListener("submit", async (event) => {
    event.preventDefault();
    statusNode.hidden = true;
    statusNode.textContent = "";
    emailErrorNode.hidden = true;
    emailErrorNode.textContent = "";

    const fieldAdvance = publicFields.validateFieldOnAdvance({ key: "email", type: "email" }, emailField.value);
    if (fieldAdvance.error) {
      emailErrorNode.textContent = fieldAdvance.error;
      emailErrorNode.hidden = false;
      return;
    }
    const email = fieldAdvance.normalized || emailField.value.trim().toLowerCase();

    try {
      const response = await fetch("/api/collections/users/request-password-reset", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ email }),
      });

      if (!response.ok) {
        throw new Error("password_reset_failed");
      }

      statusNode.textContent = "Request sent.";
      statusNode.hidden = false;
    } catch (_) {
      statusNode.textContent = "Request failed.";
      statusNode.hidden = false;
    }
  });
})();
