(() => {
  const publicFields = window.appPublicFields;
  const publicCommon = window.appPublicCommon;

  const signupSettingsURL = "/api/public/settings/signup";
  const profileSchemaURL = "/api/public/settings/profile_schema";
  const emailOTPURL = "/api/public/requests/email-otp";
  const emailOTPVerifyURL = "/api/public/requests/email-otp/verify";
  const submitURL = "/api/public/requests";

  const requestForm = document.querySelector("#public-request-form");
  const requestProgressNode = document.querySelector("#public-request-progress");
  const requestFieldErrorNode = document.querySelector("#public-request-field-error");
  const requestStepNode = document.querySelector("#public-request-fields");
  const requestStatusNode = document.querySelector("#public-request-status");
  const requestLoadingNode = document.querySelector("#public-request-loading");
  const requestTopbarNode = document.querySelector("#public-request-topbar");
  const layoutNode = document.querySelector(".onboarding-layout");
  const requestSubmitButton = document.querySelector("#public-request-submit");
  const requestStartButton = document.querySelector("#public-request-start");
  const requestBackButton = document.querySelector("#public-request-back");
  const requestActionsNode = document.querySelector("#public-request-actions");

  if (
    !requestForm ||
    !requestProgressNode ||
    !requestFieldErrorNode ||
    !requestStepNode ||
    !requestStatusNode ||
    !requestLoadingNode ||
    !requestTopbarNode ||
    !layoutNode ||
    !requestSubmitButton ||
    !requestStartButton ||
    !requestBackButton ||
    !requestActionsNode ||
    !publicFields ||
    !publicCommon ||
    !window.appAuth
  ) {
    return;
  }

  if (requestActionsNode.parentNode && requestStatusNode.parentNode === requestActionsNode.parentNode) {
    requestActionsNode.insertAdjacentElement("afterend", requestStatusNode);
  }

  const auth = window.appAuth.read();
  if (auth && auth.token) {
    window.location.replace(window.appAuth.redirectAfterLogin(auth.model && auth.model.admin === true ? "admin" : "me"));
    return;
  }

  const requestState = {
    startPage: null,
    confirmation: null,
    success: null,
    signupSteps: [],
    profileFieldsByKey: new Map(),
    fields: [],
    values: {},
    stepIndex: 0,
    submitLabel: "Send request",
    showConfirmation: false,
    showSuccess: false,
    submitting: false,
    otpToken: "",
    otpCode: "",
    otpEmail: "",
    emailVerificationPending: false,
    emailVerified: false,
    otpResending: false,
    otpResendCooldownEnd: 0,
    otpFailed: false,
    ready: false,
  };
  let requestFieldStatusNode = null;
  let resendCooldownTimer = null;

  function setStatus(node, message, isError = false) {
    node.textContent = message;
    if (node.hasAttribute("hidden")) node.removeAttribute("hidden");
    node.classList.toggle("is-error", isError);
  }

  function hideStatus(node) {
    node.textContent = "";
    if (node.hasAttribute("hidden")) node.removeAttribute("hidden");
    node.classList.remove("is-error");
  }

  function escapeHTML(value) {
    return String(value || "")
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;")
      .replaceAll("'", "&#039;");
  }

  function clearFieldStatus() {
    if (!requestFieldStatusNode) {
      return;
    }
    hideStatus(requestFieldStatusNode);
  }

  function setLoading(message, visible) {
    requestLoadingNode.textContent = message;
    requestLoadingNode.hidden = !visible;
    requestForm.hidden = visible;
  }

  function hasStartPage() {
    return publicCommon.hasPageContent(requestState.startPage);
  }

  function isStartScreen() {
    return hasStartPage() && requestState.stepIndex === -1;
  }

  function isConfirmationScreen() {
    return requestState.showConfirmation;
  }

  function isSuccessScreen() {
    return requestState.showSuccess;
  }

  function currentField() {
    return requestState.fields[requestState.stepIndex] || null;
  }

  function currentStepNumber() {
    if (isConfirmationScreen() || isSuccessScreen()) {
      return null;
    }
    if (requestState.stepIndex >= 0 && requestState.stepIndex < requestState.fields.length) {
      return requestState.stepIndex + 1;
    }
    return null;
  }

  function syncStepToUrl() {
    const url = new URL(window.location.href);
    if (isConfirmationScreen() || isSuccessScreen()) {
      window.history.replaceState({}, "", "/");
      return;
    }
    const step = currentStepNumber();
    if (step === null) {
      url.searchParams.delete("step");
    } else {
      url.searchParams.set("step", String(step));
    }
    window.history.replaceState({}, "", url.pathname + url.search + url.hash);
  }

  function initialStepIndex() {
    return hasStartPage() ? -1 : 0;
  }

  function updateActions() {
    const startScreen = isStartScreen();
    const confirmationScreen = isConfirmationScreen();
    const successScreen = isSuccessScreen();
    const field = currentField();
    const fieldScreen = !startScreen && !confirmationScreen && !successScreen && requestState.stepIndex < requestState.fields.length;
    const flowScreen = !!(fieldScreen && field && field.type === "select");
    const centerScreen = !startScreen && !flowScreen && (fieldScreen || confirmationScreen);

    layoutNode.classList.toggle("onboarding-has-topbar", fieldScreen);
    layoutNode.classList.toggle("onboarding-screen-start", startScreen);
    layoutNode.classList.toggle("onboarding-screen-center", centerScreen);
    layoutNode.classList.toggle("onboarding-screen-flow", flowScreen);

    requestBackButton.hidden = fieldScreen
      ? (!hasStartPage() && requestState.stepIndex <= 0)
      : !confirmationScreen;
    requestBackButton.disabled = requestState.submitting;
    requestActionsNode.hidden = !(startScreen || confirmationScreen);
    requestTopbarNode.hidden = startScreen || confirmationScreen || successScreen;

    if (confirmationScreen && requestStartButton.parentNode !== requestActionsNode) {
      requestActionsNode.appendChild(requestStartButton);
    }

    if (startScreen) {
      requestProgressNode.hidden = true;
      requestProgressNode.textContent = "";
      requestStartButton.textContent = String(requestState.startPage.button || "Start").trim() || "Start";
      requestStartButton.disabled = false;
      requestSubmitButton.classList.remove("request-action-loading");
      return;
    }

    if (confirmationScreen) {
      requestProgressNode.hidden = true;
      requestProgressNode.textContent = "";
      requestStartButton.textContent = requestState.submitting
        ? String(requestCopy().submitting || "").trim()
        : String(requestState.confirmation && requestState.confirmation.button || requestState.submitLabel || "").trim();
      requestStartButton.disabled = requestState.submitting;
      requestSubmitButton.classList.remove("request-action-loading");
      return;
    }

    if (successScreen) {
      requestProgressNode.hidden = true;
      requestProgressNode.textContent = "";
      requestSubmitButton.classList.remove("request-action-loading");
      return;
    }

    requestProgressNode.hidden = false;
    requestProgressNode.textContent = `Step ${requestState.stepIndex + 1} / ${requestState.fields.length}`;
    const emailLabel = emailStepSubmitLabel(field);
    requestSubmitButton.textContent = emailLabel || (
      requestState.stepIndex >= requestState.fields.length - 1 && !requestState.confirmation
        ? requestState.submitLabel
        : "Next"
    );
    requestSubmitButton.disabled = requestState.submitting || !field;
    requestSubmitButton.classList.toggle("request-action-loading", requestState.submitting && isEmailField(field));
  }

  function isEmailField(field) {
    return !!field && field.key === "email";
  }

  function currentEmail() {
    return String(requestState.values.email || "").trim().toLowerCase();
  }

  function emailStepSubmitLabel(field) {
    if (!isEmailField(field)) {
      return "";
    }
    if (requestState.emailVerified && requestState.otpEmail === currentEmail()) {
      return "Next";
    }
    return "Verify";
  }

  function renderStartScreen() {
    requestFieldStatusNode = null;
    if (window.appOnboardingComponent) {
      window.appOnboardingComponent.renderStartScreen({
        stepNode: requestStepNode,
        startButton: requestStartButton,
        page: requestState.startPage || {},
      });
    }
    updateActions();
  }

  function renderConfirmationScreen() {
    requestFieldStatusNode = null;
    if (window.appOnboardingComponent) {
      window.appOnboardingComponent.renderCenteredScreen({
        stepNode: requestStepNode,
        page: requestState.confirmation || {},
        titleClass: "public-request-title public-request-title-start",
      });
    }
    updateActions();
  }

  function renderSuccessScreen() {
    requestFieldStatusNode = null;
    const page = requestState.success || {};
    requestStepNode.className = "public-request-step public-request-step-start";
    requestStepNode.innerHTML = "";

    const wrapper = document.createElement("div");
    wrapper.className = "public-request-start-stack public-request-success-stack";

    const checkWrap = document.createElement("div");
    checkWrap.className = "public-request-success-check";
    checkWrap.innerHTML = `<svg viewBox="0 0 16 16" fill="currentColor" xmlns="http://www.w3.org/2000/svg" aria-hidden="true"><path d="M16 8A8 8 0 1 1 0 8a8 8 0 0 1 16 0zm-3.97-3.03a.75.75 0 0 0-1.08.022L7.477 9.417 5.384 7.323a.75.75 0 0 0-1.06 1.06L6.97 11.03a.75.75 0 0 0 1.079-.02l3.992-4.99a.75.75 0 0 0-.01-1.05z"/></svg>`;
    wrapper.appendChild(checkWrap);

    const titleNode = document.createElement("h2");
    titleNode.className = "public-request-title public-request-title-start";
    titleNode.textContent = String(page.title || "").trim();
    wrapper.appendChild(titleNode);

    const copyNode = document.createElement("p");
    copyNode.className = "public-request-copy";
    const copyText = String(page.text || "").trim();
    copyNode.innerHTML = window.appPublicCommon.escapeHTML(copyText).replace(/\n/g, "<br>");
    wrapper.appendChild(copyNode);

    const buttonLabel = String(page.button || "").trim();
    const buttonUrl = String(page.button_url || "").trim();
    if (buttonLabel) {
      const actionRow = document.createElement("div");
      actionRow.className = "action-row";
      if (buttonUrl) {
        const link = document.createElement("a");
        link.className = "primary";
        link.href = buttonUrl;
        link.textContent = buttonLabel;
        actionRow.appendChild(link);
      } else {
        const btn = document.createElement("button");
        btn.type = "button";
        btn.className = "primary";
        btn.textContent = buttonLabel;
        actionRow.appendChild(btn);
      }
      wrapper.appendChild(actionRow);
    }

    requestStepNode.appendChild(wrapper);
    updateActions();
  }

  function renderFieldScreen(field) {
    const result = publicCommon.renderFieldScreen({
      stepNode: requestStepNode,
      publicFields,
      field,
      value: requestState.values[field.key],
      values: requestState.values,
      mode: "onboarding",
      onValueChange: (key, nextValue) => {
        requestState.values[key] = nextValue;
        if (key === "email" && currentEmail() !== requestState.otpEmail) {
          requestState.otpToken = "";
          requestState.otpCode = "";
          requestState.otpEmail = "";
          requestState.emailVerificationPending = false;
          requestState.emailVerified = false;
        }
      },
      onClearFieldStatus: clearFieldStatus,
      onUpdateActions: updateActions,
      setFieldErrorNode: (node) => {
        requestFieldStatusNode = node;
      },
      statusNode: requestFieldErrorNode,
      useInlineError: true,
      fullNamePlaceholder: "Nome e cognome",
    });
    requestFieldStatusNode = result.statusNode;
    if (requestFieldStatusNode && requestFieldStatusNode.hasAttribute("hidden")) {
      requestFieldStatusNode.removeAttribute("hidden");
    }
    if (isEmailField(field)) {
      renderEmailVerificationBlock();
    }
    updateActions();
  }

  function renderEmailVerificationBlock() {
    const c = requestCopy();
    const fieldNode = requestStepNode.querySelector(".public-request-field");
    if (!fieldNode || !requestState.emailVerificationPending) {
      return;
    }

    const label = String(c.otpLabel || "").trim();
    const emailInput = fieldNode.querySelector('input[name="email"]');
    if (emailInput) {
      emailInput.readOnly = true;
      emailInput.disabled = true;
      emailInput.classList.add("public-request-otp-email-muted");
    }

    const changeEmailButton = document.createElement("button");
    changeEmailButton.type = "button";
    changeEmailButton.className = "public-request-otp-change";
    changeEmailButton.textContent = String(c.changeEmail || "Change email");
    fieldNode.appendChild(changeEmailButton);

    const block = document.createElement("div");
    block.className = "public-request-otp";
    block.innerHTML = `
      <span class="form-field-label public-request-otp-label">${escapeHTML(label)}</span>
      <div class="public-request-otp-control">
        <input class="public-request-otp-input" type="text" inputmode="numeric" autocomplete="one-time-code" maxlength="4" value="${escapeHTML(requestState.otpCode)}" aria-label="${escapeHTML(label)}" />
        <div class="public-request-otp-slots" aria-hidden="true">
          ${renderOTPSlots(requestState.otpCode)}
        </div>
        <div class="public-request-otp-state" aria-hidden="true">
          <span class="public-request-otp-spinner" role="status" aria-label="${escapeHTML(c.otpChecking || "Checking code...")}"></span>
          <svg class="public-request-otp-check" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor" aria-hidden="true"><path d="M16 8A8 8 0 1 1 0 8a8 8 0 0 1 16 0zm-3.97-3.03a.75.75 0 0 0-1.08.022L7.477 9.417 5.384 7.323a.75.75 0 0 0-1.06 1.06L6.97 11.03a.75.75 0 0 0 1.079-.02l3.992-4.99a.75.75 0 0 0-.01-1.05z"/></svg>
        </div>
      </div>
      <p class="public-request-subtitle">${escapeHTML(c.otpText || "")}</p>
      <button class="public-request-otp-resend" type="button"${requestState.otpFailed ? "" : " hidden"}>${escapeHTML(c.resendCode || "Resend code")}</button>
    `;
    fieldNode.appendChild(block);
    const inlineError = fieldNode.querySelector(".public-request-password-error");
    if (inlineError) {
      fieldNode.appendChild(inlineError);
      requestFieldStatusNode = inlineError;
    }

    const resendButton = block.querySelector(".public-request-otp-resend");
    if (resendButton) {
      resendButton.addEventListener("click", () => {
        resendEmailOTP();
      });
    }
    updateResendButton();
    startResendCooldownTimer();

    changeEmailButton.addEventListener("click", () => {
      requestState.otpToken = "";
      requestState.otpCode = "";
      requestState.otpEmail = "";
      requestState.emailVerificationPending = false;
      requestState.emailVerified = false;
      requestState.otpResendCooldownEnd = 0;
      requestState.otpFailed = false;
      stopResendCooldownTimer();
      renderStep();
      setTimeout(() => {
        const nextEmailInput = requestStepNode.querySelector('input[name="email"]');
        if (nextEmailInput) {
          nextEmailInput.focus();
        }
      }, 0);
    });

    const input = block.querySelector("input");
    if (input) {
      const handleOTPChange = () => {
        requestState.otpCode = String(input.value || "").replace(/\D/g, "").slice(0, 4);
        if (input.value !== requestState.otpCode) {
          input.value = requestState.otpCode;
        }
        updateOTPSlots(block);
        clearFieldStatus();
        if (requestState.otpCode.length === 4 && !requestState.submitting) {
          setTimeout(() => {
            if (requestState.otpCode.length === 4 && !requestState.submitting) {
              requestForm.requestSubmit
                ? requestForm.requestSubmit()
                : requestForm.dispatchEvent(new Event("submit", { cancelable: true }));
            }
          }, 800);
        }
      };
      input.addEventListener("input", handleOTPChange);
      input.addEventListener("paste", (event) => {
        const pasted = (event.clipboardData || window.clipboardData)?.getData("text") || "";
        const digits = pasted.replace(/\D/g, "").slice(0, 4);
        if (!digits) return;
        event.preventDefault();
        input.value = digits;
        handleOTPChange();
      });
    }
  }

  function setEmailInputMuted(muted) {
    const emailInput = requestStepNode.querySelector('input[name="email"]');
    if (!emailInput) return;
    emailInput.readOnly = muted;
    emailInput.disabled = muted;
    emailInput.classList.toggle("public-request-otp-email-muted", muted);
  }

  function updateResendButton() {
    const button = requestStepNode.querySelector(".public-request-otp-resend");
    if (!button) return;
    const c = requestCopy();
    const baseLabel = String(c.resendCode || "Resend code");
    if (requestState.otpResending) {
      button.disabled = true;
      button.classList.add("is-loading");
      button.textContent = String(c.otpSending || "Sending...");
      return;
    }
    button.classList.remove("is-loading");
    const remaining = Math.ceil((requestState.otpResendCooldownEnd - Date.now()) / 1000);
    if (remaining > 0) {
      button.disabled = true;
      button.textContent = `${baseLabel} (${remaining}s)`;
    } else {
      button.disabled = false;
      button.textContent = baseLabel;
      stopResendCooldownTimer();
    }
  }

  function startResendCooldownTimer() {
    stopResendCooldownTimer();
    if (requestState.otpResendCooldownEnd <= Date.now()) return;
    resendCooldownTimer = setInterval(updateResendButton, 1000);
  }

  function stopResendCooldownTimer() {
    if (resendCooldownTimer) {
      clearInterval(resendCooldownTimer);
      resendCooldownTimer = null;
    }
  }

  async function resendEmailOTP() {
    if (requestState.otpResending) return;
    if (Date.now() < requestState.otpResendCooldownEnd) return;
    requestState.otpResending = true;
    clearFieldStatus();
    updateResendButton();
    try {
      const payload = collectRequestPayload();
      const response = await fetch(emailOTPURL, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email: payload.email || "" }),
      });
      if (!response.ok) throw new Error("otp_failed");
      const responsePayload = await response.json();
      requestState.otpToken = String(responsePayload.otp_token || "").trim();
      requestState.otpCode = "";
      requestState.otpEmail = currentEmail();
      requestState.emailVerificationPending = true;
      requestState.emailVerified = false;
      requestState.otpResendCooldownEnd = Date.now() + 60000;
      const input = requestStepNode.querySelector(".public-request-otp-input");
      if (input) input.value = "";
      const block = requestStepNode.querySelector(".public-request-otp");
      if (block) updateOTPSlots(block);
    } catch (_) {
      setStatus(requestFieldStatusNode || requestStatusNode, requestCopy().otpSendError || "", true);
    }
    requestState.otpResending = false;
    updateResendButton();
    startResendCooldownTimer();
  }

  function renderOTPSlots(code) {
    const value = String(code || "");
    const cursor = Math.min(value.length, 3);
    return Array.from({ length: 4 }, (_, index) => {
      const classes = ["public-request-otp-slot"];
      if (value[index]) classes.push("is-filled");
      if (index === cursor && !value[index]) classes.push("is-current");
      return `<span class="${classes.join(" ")}">${escapeHTML(value[index] || "")}</span>`;
    }).join("");
  }

  function updateOTPSlots(root) {
    const slots = Array.from(root.querySelectorAll(".public-request-otp-slot"));
    const value = String(requestState.otpCode || "");
    const cursor = Math.min(value.length, 3);
    for (let index = 0; index < slots.length; index += 1) {
      slots[index].textContent = value[index] || "";
      slots[index].classList.toggle("is-filled", !!value[index]);
      slots[index].classList.toggle("is-current", index === cursor && !value[index]);
    }
  }

  function renderStep() {
    syncStepToUrl();

    if (isStartScreen()) {
      renderStartScreen();
      return;
    }
    if (isConfirmationScreen()) {
      renderConfirmationScreen();
      return;
    }
    if (isSuccessScreen()) {
      renderSuccessScreen();
      return;
    }
    const field = currentField();
    if (!field) {
      requestStepNode.innerHTML = "";
      updateActions();
      return;
    }
    renderFieldScreen(field);
  }

  function buildRequestFields() {
    return publicCommon.buildFields({
      steps: requestState.signupSteps,
      profileFieldsByKey: requestState.profileFieldsByKey,
      includeFiles: false,
      mobileKey: "mobile",
    });
  }

  function collectRequestPayload() {
    return publicCommon.collectFieldValues(requestState.fields, requestState.values, {
      includeFiles: false,
    });
  }

  function requestCopy() {
    return (window.appCopy && window.appCopy.ui && window.appCopy.ui.requests) || {};
  }

  function notificationEmailAccepted(payload) {
    const notifications = payload && typeof payload.notifications === "object" ? payload.notifications : {};
    const submitted = notifications && typeof notifications.request_submitted_email === "object" ? notifications.request_submitted_email : null;
    return !submitted || submitted.accepted !== false;
  }

  async function sendEmailOTP() {
    requestState.submitting = true;
    updateActions();
    try {
      const payload = collectRequestPayload();
      const response = await fetch(emailOTPURL, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email: payload.email || "" }),
      });
      const responsePayload = await response.json().catch(() => ({}));
      if (!response.ok) {
        const errMsg = (responsePayload && (responsePayload.message || responsePayload.error)) || "";
        const err = new Error("otp_failed");
        err.backendMessage = errMsg;
        throw err;
      }
      requestState.otpToken = String(responsePayload.otp_token || "").trim();
      requestState.otpCode = "";
      requestState.otpEmail = currentEmail();
      requestState.emailVerificationPending = true;
      requestState.emailVerified = false;
      requestState.showConfirmation = false;
      requestState.submitting = false;
      requestState.otpResendCooldownEnd = Date.now() + 60000;
      requestState.otpFailed = false;
      renderStep();
    } catch (err) {
      requestState.submitting = false;
      updateActions();
      const copy = requestCopy();
      const errorsMap = (copy && copy.otpErrors) || {};
      const rawCode = (err && err.backendMessage) ? String(err.backendMessage) : "";
      const code = rawCode.trim().replace(/\.+$/, "").toLowerCase();
      const friendly = (code && errorsMap[code]) || copy.otpSendError || "";
      setStatus(requestFieldStatusNode || requestStatusNode, friendly, true);
      const emailInput = requestStepNode.querySelector('input[name="email"]');
      if (emailInput) {
        setTimeout(() => emailInput.focus(), 0);
      }
    }
  }

  async function verifyEmailOTP() {
    requestState.submitting = true;
    updateActions();
    const block = requestStepNode.querySelector(".public-request-otp");
    if (block) {
      block.classList.add("is-verifying");
      block.classList.remove("is-verified");
    }
    try {
      const response = await fetch(emailOTPVerifyURL, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          email: currentEmail(),
          otp_token: requestState.otpToken,
          otp_code: requestState.otpCode,
        }),
      });
      if (!response.ok) {
        throw new Error("invalid_email_otp");
      }
      requestState.emailVerificationPending = false;
      requestState.emailVerified = true;
      if (block) {
        block.classList.remove("is-verifying");
        block.classList.add("is-verified");
      }
      await new Promise((resolve) => setTimeout(resolve, 1000));
      requestState.submitting = false;
      return true;
    } catch (_) {
      requestState.submitting = false;
      requestState.otpCode = "";
      requestState.otpFailed = true;
      updateActions();
      setStatus(requestFieldStatusNode || requestStatusNode, requestCopy().otpInvalidError || "", true);
      const input = requestStepNode.querySelector(".public-request-otp-input");
      if (input) {
        input.value = "";
      }
      if (block) {
        block.classList.remove("is-verifying");
        block.classList.remove("is-verified");
        updateOTPSlots(block);
      }
      const resendButton = requestStepNode.querySelector(".public-request-otp-resend");
      if (resendButton) {
        resendButton.hidden = false;
        updateResendButton();
        startResendCooldownTimer();
      }
      return false;
    }
  }

  async function submitRequest() {
    requestState.submitting = true;
    updateActions();

    try {
      const submitPayload = collectRequestPayload();
      submitPayload.otp_token = requestState.otpToken;
      submitPayload.otp_code = requestState.otpCode;
      const response = await fetch(submitURL, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(submitPayload),
      });
      if (!response.ok) {
        let code = "";
        try {
          const errorPayload = await response.json();
          code = String(errorPayload && errorPayload.message || "").trim();
        } catch (_) {}
        throw new Error(code || "request_failed");
      }
      const responsePayload = await response.json();

      requestState.submitting = false;
      requestState.values = {};
      requestState.showConfirmation = false;
      requestState.otpToken = "";
      requestState.otpCode = "";
      requestState.otpEmail = "";
      requestState.emailVerificationPending = false;
      requestState.emailVerified = false;

      if (!notificationEmailAccepted(responsePayload)) {
        requestState.stepIndex = hasStartPage() ? -1 : 0;
        renderStep();
        setStatus(requestStatusNode, requestCopy().submittedEmailNotAccepted || "", true);
        return;
      }

      if (requestState.success) {
        requestState.showSuccess = true;
        renderStep();
        return;
      }

      requestState.stepIndex = hasStartPage() ? -1 : 0;
      renderStep();
    } catch (err) {
      requestState.submitting = false;
      updateActions();
      const c = requestCopy();
      const code = String(err && err.message || "");
      const otpError = code.includes("email_otp");
      const message = otpError ? c.otpInvalidError : c.submitError;
      setStatus(requestStatusNode, message || "", true);
    }
  }

  async function loadRequestFlow() {
    setLoading("Loading link...", true);

    const signupResponse = await fetch(signupSettingsURL);
    if (!signupResponse.ok) {
      throw new Error("settings_failed");
    }

    const signupPayload = await signupResponse.json();
    const signupSettings = signupPayload && signupPayload.data ? signupPayload.data : {};
    requestState.startPage = signupSettings.start_page || null;
    requestState.confirmation = signupSettings.confirmation || null;
    requestState.success = signupSettings.success || null;
    requestState.signupSteps = Array.isArray(signupSettings.steps) ? signupSettings.steps : [];
    requestState.profileFieldsByKey = new Map();
    requestState.fields = [];
    requestState.values = {};
    requestState.stepIndex = signupSettings.start_page ? -1 : 0;
    requestState.showConfirmation = false;
    requestState.showSuccess = false;
    requestState.otpToken = "";
    requestState.otpCode = "";
    requestState.otpEmail = "";
    requestState.emailVerificationPending = false;
    requestState.emailVerified = false;
    requestState.submitting = false;
    requestState.submitLabel = String(signupSettings.submit_label || "Send request").trim();
    requestState.ready = false;

    const schemaResponse = await fetch(profileSchemaURL);
    if (!schemaResponse.ok) {
      throw new Error("profile_schema_failed");
    }

    const schemaPayload = await schemaResponse.json();
    const profileSchema = schemaPayload && schemaPayload.data ? schemaPayload.data : {};
    const rawFields = Array.isArray(profileSchema.fields) ? profileSchema.fields : [];
    requestState.profileFieldsByKey = new Map(rawFields.map((field) => [String(field.key || "").trim(), field]));
    requestState.fields = buildRequestFields();
    requestState.ready = true;

    if (requestState.fields.length === 0) {
      throw new Error("missing_fields");
    }

    requestState.stepIndex = initialStepIndex();
    renderStep();
    setLoading("", false);
  }

  requestBackButton.addEventListener("click", () => {
    hideStatus(requestStatusNode);
    if (isConfirmationScreen()) {
      requestState.showConfirmation = false;
      renderStep();
      return;
    }
    if (requestState.stepIndex > 0) {
      requestState.stepIndex -= 1;
    } else if (requestState.stepIndex === 0 && hasStartPage()) {
      requestState.stepIndex = -1;
    }
    renderStep();
  });

  requestStartButton.addEventListener("click", async () => {
    hideStatus(requestStatusNode);

    if (isStartScreen()) {
      if (!requestState.ready) {
        setStatus(requestStatusNode, "Loading...", false);
        return;
      }
      requestState.stepIndex = 0;
      renderStep();
      return;
    }

    if (!isConfirmationScreen()) {
      if (isSuccessScreen()) {
        window.location.href = String(requestState.success && requestState.success.url || "/").trim() || "/";
      }
      return;
    }

    try {
      await submitRequest();
    } catch (_) {}
  });

  requestForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    hideStatus(requestStatusNode);

    if (isConfirmationScreen()) {
      await submitRequest();
      return;
    }

    const field = currentField();
    if (!field) {
      return;
    }

    const fieldAdvance = publicFields.validateFieldOnAdvance(field, requestState.values[field.key]);
    if (fieldAdvance.error) {
      setStatus(requestFieldStatusNode || requestStatusNode, fieldAdvance.error, true);
      return;
    }
    if (!publicCommon.isFieldComplete(field, requestState.values[field.key])) {
      setStatus(requestFieldStatusNode || requestStatusNode, "This field is required.", true);
      return;
    }
    if (fieldAdvance.normalized) {
      requestState.values[field.key] = fieldAdvance.normalized;
    }

    if (isEmailField(field)) {
      const email = currentEmail();
      if (requestState.emailVerified && requestState.otpEmail === email) {
        // already verified for the current email; continue below
      } else if (!requestState.emailVerificationPending || requestState.otpEmail !== email) {
        requestState.otpToken = "";
        requestState.otpCode = "";
        requestState.otpEmail = "";
        requestState.emailVerificationPending = false;
        requestState.emailVerified = false;
        await sendEmailOTP();
        return;
      } else if (!requestState.otpToken || requestState.otpCode.length !== 4) {
        setStatus(requestFieldStatusNode || requestStatusNode, requestCopy().otpInvalidError || "", true);
        return;
      } else {
        if (!await verifyEmailOTP()) {
          return;
        }
      }
    }

    if (requestState.stepIndex >= requestState.fields.length - 1) {
      if (requestState.confirmation) {
        requestState.showConfirmation = true;
        renderStep();
        return;
      }
      requestState.showConfirmation = true;
      requestState.confirmation = {
        title: "",
        text: "",
        button: requestState.submitLabel,
      };
      renderStep();
      return;
    }

    requestState.stepIndex += 1;
    renderStep();
  });

  loadRequestFlow().catch(() => {
    setLoading("", false);
    setStatus(requestStatusNode, "Signup unavailable.", true);
  });
})();
