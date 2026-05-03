(() => {
  const publicFields = window.appPublicFields;
  const publicCommon = window.appPublicCommon;

  const signupSettingsURL = "/api/public/settings/signup";
  const profileSchemaURL = "/api/public/settings/profile_schema";
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
    ready: false,
  };
  let requestFieldStatusNode = null;

  function setStatus(node, message, isError = false) {
    node.textContent = message;
    node.hidden = false;
    node.classList.toggle("is-error", isError);
  }

  function hideStatus(node) {
    node.hidden = true;
    node.textContent = "";
    node.classList.remove("is-error");
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
    const rawStep = new URLSearchParams(window.location.search).get("step");
    const parsedStep = Number(rawStep);
    if (!Number.isInteger(parsedStep) || parsedStep < 1) {
      return hasStartPage() ? -1 : 0;
    }
    const fieldIndex = parsedStep - 1;
    if (fieldIndex >= 0 && fieldIndex < requestState.fields.length) {
      return fieldIndex;
    }
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

    requestBackButton.hidden = (!hasStartPage() && requestState.stepIndex <= 0) || (!fieldScreen && !confirmationScreen);
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
      return;
    }

    if (confirmationScreen) {
      requestProgressNode.hidden = true;
      requestProgressNode.textContent = "";
      requestStartButton.textContent = requestState.submitting
        ? "Invio..."
        : String(requestState.confirmation && requestState.confirmation.button || requestState.submitLabel || "Send request").trim();
      requestStartButton.disabled = requestState.submitting;
      return;
    }

    if (successScreen) {
      requestProgressNode.hidden = true;
      requestProgressNode.textContent = "";
      return;
    }

    requestProgressNode.hidden = false;
    requestProgressNode.textContent = `Step ${requestState.stepIndex + 1} / ${requestState.fields.length}`;
    requestSubmitButton.textContent = requestState.stepIndex >= requestState.fields.length - 1 && !requestState.confirmation
      ? requestState.submitLabel
      : "Next";
    requestSubmitButton.disabled = requestState.submitting || !field;
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
    if (window.appOnboardingComponent) {
      window.appOnboardingComponent.renderCenteredScreen({
        stepNode: requestStepNode,
        page: requestState.success || {},
        titleClass: "public-request-title public-request-title-start",
      });
    }
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
    updateActions();
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

  async function submitRequest() {
    requestState.submitting = true;
    updateActions();

    try {
      const payload = collectRequestPayload();
      const response = await fetch(submitURL, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
      });
      if (!response.ok) {
        throw new Error("request_failed");
      }

      requestState.submitting = false;
      requestState.values = {};
      requestState.showConfirmation = false;

      if (requestState.success) {
        requestState.showSuccess = true;
        renderStep();
        return;
      }

      requestState.stepIndex = hasStartPage() ? -1 : 0;
      renderStep();
    } catch (_) {
      requestState.submitting = false;
      updateActions();
      setStatus(requestStatusNode, "There was a problem. Please reload the page.", true);
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
