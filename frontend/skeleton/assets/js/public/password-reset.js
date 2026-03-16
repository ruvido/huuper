(() => {
  const form = document.querySelector("#password-reset-form");
  const emailField = document.querySelector("#password-reset-email");
  const statusNode = document.querySelector("#password-reset-status");

  if (!form || !emailField || !statusNode) {
    return;
  }

  form.addEventListener("submit", async (event) => {
    event.preventDefault();
    statusNode.hidden = true;
    statusNode.textContent = "";

    const email = emailField.value.trim().toLowerCase();
    if (!email) {
      statusNode.textContent = "Email is required.";
      statusNode.hidden = false;
      return;
    }

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
