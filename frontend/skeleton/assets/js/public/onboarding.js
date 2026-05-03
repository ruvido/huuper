(() => {
  const publicFields = window.appPublicFields;
  const publicCommon = window.appPublicCommon;

  const form = document.querySelector("#public-request-form");
  const stepNode = document.querySelector("#public-request-fields");
  const progressNode = document.querySelector("#public-request-progress");
  const fieldErrorNode = document.querySelector("#public-request-field-error");
  const statusNode = document.querySelector("#public-request-status");
  const loadingNode = document.querySelector("#public-request-loading");
  const submitButton = document.querySelector("#public-request-submit");
  const backButton = document.querySelector("#public-request-back");
  const topbarNode = document.querySelector("#public-request-topbar");
  const layoutNode = document.querySelector(".onboarding-layout");
  const actionsNode = document.querySelector("#public-request-actions");
  const startButton = document.querySelector("#public-request-start");
  const token = new URLSearchParams(window.location.search).get("token") || "";

  if (
    !form ||
    !stepNode ||
    !progressNode ||
    !fieldErrorNode ||
    !statusNode ||
    !loadingNode ||
    !submitButton ||
    !backButton ||
    !topbarNode ||
    !layoutNode ||
    !actionsNode ||
    !startButton ||
    !publicFields ||
    !publicCommon ||
    !window.appAuth
  ) {
    return;
  }

  const state = {
    email: "",
    fullName: "",
    userId: "",
    onboarding: null,
    profileFieldsByKey: new Map(),
    fields: [],
    values: {},
    fileFields: new Set(),
    stepIndex: -2,
    passwordMin: 0,
    ready: false,
    submitting: false,
  };
  let currentFieldErrorNode = null;
  let confirmationButtonNode = null;

  let confirmationErrorNode = null;

  function setStatus(message, hidden = false) {
    statusNode.textContent = message;
    statusNode.hidden = hidden || !message;
    statusNode.classList.toggle("is-error", !!message && !hidden);
  }

  function clearFieldStatus() {
    if (!currentFieldErrorNode) {
      return;
    }
    currentFieldErrorNode.hidden = true;
    currentFieldErrorNode.textContent = "";
  }

  function setLoading(message, visible) {
    loadingNode.textContent = message;
    loadingNode.hidden = !visible;
  }

  function passwordTooShortMessage() {
    const min = Number(state.passwordMin) || 0;
    if (min <= 0) {
      return "";
    }
    return min === 1 ? "Password must be at least 1 character." : `Password must be at least ${min} characters.`;
  }

  function text(value) {
    return publicCommon.text(value);
  }

  function currentStepNumber() {
    if (isStartScreen()) {
      return 0;
    }
    if (isPasswordScreen()) {
      return 1;
    }
    if (isConfirmationScreen()) {
      return state.fields.length + 2;
    }
    if (state.stepIndex >= 0) {
      return state.stepIndex + 2;
    }
    return hasStartPage() ? 0 : 1;
  }

  function syncStepToUrl() {
    const url = new URL(window.location.href);
    url.searchParams.set("step", String(currentStepNumber()));
    window.history.replaceState({}, "", url.pathname + url.search + url.hash);
  }

  function initialStepIndex() {
    const rawStep = new URLSearchParams(window.location.search).get("step");
    const parsedStep = Number(rawStep);
    if (!Number.isInteger(parsedStep) || parsedStep < 0) {
      return hasStartPage() ? -1 : -2;
    }

    if (parsedStep === 0) {
      return hasStartPage() ? -1 : -2;
    }
    if (parsedStep === 1) {
      return -2;
    }
    if (hasConfirmationPage() && parsedStep === state.fields.length + 2) {
      return state.fields.length;
    }

    const fieldIndex = parsedStep - 2;
    if (fieldIndex >= 0 && fieldIndex < state.fields.length) {
      return fieldIndex;
    }

    return hasStartPage() ? -1 : -2;
  }

  function hasStartPage() {
    return publicCommon.hasPageContent(state.onboarding && state.onboarding.start_page);
  }

  function hasConfirmationPage() {
    return publicCommon.hasPageContent(state.onboarding && state.onboarding.confirmation);
  }

  function isPasswordScreen() {
    return state.stepIndex === -2;
  }

  function isStartScreen() {
    return hasStartPage() && state.stepIndex === -1;
  }

  function isConfirmationScreen() {
    return hasConfirmationPage() && state.stepIndex === state.fields.length;
  }

  function currentField() {
    return state.fields[state.stepIndex] || null;
  }

  function totalSteps() {
    return Math.max(1, state.fields.length + 1);
  }

  function updateActions() {
    const passwordScreen = isPasswordScreen();
    const startScreen = isStartScreen();
    const fieldScreen = !passwordScreen && !startScreen && !isConfirmationScreen() && state.stepIndex < state.fields.length;
    const confirmationScreen = isConfirmationScreen();
    const field = currentField();
    const flowScreen = !!(fieldScreen && field && field.type === "select");
    const centerScreen = !startScreen && !flowScreen && (passwordScreen || fieldScreen || confirmationScreen);

    layoutNode.classList.toggle("onboarding-has-topbar", !startScreen && !confirmationScreen);
    layoutNode.classList.toggle("onboarding-screen-start", startScreen);
    layoutNode.classList.toggle("onboarding-screen-center", centerScreen);
    layoutNode.classList.toggle("onboarding-screen-flow", flowScreen);

    progressNode.hidden = startScreen || confirmationScreen;
    if (fieldScreen) {
      progressNode.hidden = false;
      progressNode.textContent = `Step ${state.stepIndex + 2} / ${totalSteps()}`;
    } else if (passwordScreen) {
      progressNode.hidden = false;
      progressNode.textContent = `Step 1 / ${totalSteps()}`;
    } else {
      progressNode.textContent = "";
    }

    backButton.hidden = (passwordScreen && !hasStartPage()) || (!fieldScreen && !confirmationScreen && !startScreen && !passwordScreen);
    backButton.disabled = state.submitting;
    topbarNode.hidden = startScreen || confirmationScreen;
    actionsNode.hidden = !startScreen;

    if (passwordScreen) {
      submitButton.textContent = state.submitting ? "Updating..." : "Next";
      submitButton.disabled = state.submitting;
      return;
    }

    if (startScreen) {
      const page = state.onboarding && state.onboarding.start_page ? state.onboarding.start_page : {};
      startButton.textContent = text(page.button) || "Start";
      startButton.disabled = false;
      return;
    }

    if (confirmationScreen) {
      const page = state.onboarding && state.onboarding.confirmation ? state.onboarding.confirmation : {};
      const confirmationLabel = text(page.button) || "Complete";
      submitButton.textContent = state.submitting ? "Submitting..." : confirmationLabel;
      submitButton.disabled = state.submitting;
      if (confirmationButtonNode) {
        confirmationButtonNode.textContent = state.submitting ? "Submitting..." : confirmationLabel;
        confirmationButtonNode.disabled = state.submitting;
      }
      return;
    }

    submitButton.textContent = state.submitting ? "Updating..." : "Next";
    submitButton.disabled = state.submitting || !field;
  }

  function renderPasswordScreen() {
    clearFieldStatus();
    currentFieldErrorNode = null;
    const minMessage = passwordTooShortMessage();
    stepNode.innerHTML = "";
    stepNode.className = "public-request-step public-request-step-start";

    const wrapper = document.createElement("div");
    wrapper.className = "public-request-field public-request-field-password";

    const heading = document.createElement("div");
    heading.className = "public-request-password-heading";

    const title = document.createElement("h2");
    title.className = "public-request-title";
    title.textContent = "Set your password";
    heading.appendChild(title);

    if (minMessage) {
      const subtitle = document.createElement("p");
      subtitle.className = "public-request-subtitle";
      subtitle.textContent = minMessage;
      heading.appendChild(subtitle);
    }

    wrapper.appendChild(heading);

    const stack = document.createElement("div");
    stack.className = "public-request-password-stack";

    const passwordErrorNode = document.createElement("p");
    passwordErrorNode.id = "onboarding-password-error";
    passwordErrorNode.className = "public-request-password-error meta-text is-error";
    passwordErrorNode.hidden = true;

    const passwordInput = document.createElement("input");
    passwordInput.id = "onboarding-password";
    passwordInput.type = "password";
    passwordInput.name = "password";
    passwordInput.placeholder = "••••••••";
    passwordInput.autocomplete = "new-password";
    passwordInput.required = true;
    passwordInput.classList.add("public-request-password-look");

    const passwordShell = publicFields.createFieldComponent({
      label: "New password",
      controlId: passwordInput.id,
      control: passwordInput,
      includeStatus: false,
      hideLabel: true,
    });
    stack.appendChild(passwordShell.root);

    const confirmInput = document.createElement("input");
    confirmInput.id = "onboarding-password-confirm";
    confirmInput.type = "password";
    confirmInput.name = "password_confirm";
    confirmInput.placeholder = "••••••••";
    confirmInput.autocomplete = "new-password";
    confirmInput.required = true;
    confirmInput.classList.add("public-request-password-look");

    const confirmShell = publicFields.createFieldComponent({
      label: "Confirm password",
      controlId: confirmInput.id,
      control: confirmInput,
      includeStatus: false,
      hideLabel: true,
    });
    stack.appendChild(confirmShell.root);

    const clearPasswordErrors = () => {
      if (passwordShell.statusNode) {
        passwordShell.statusNode.hidden = true;
        passwordShell.statusNode.textContent = "";
      }
      if (confirmShell.statusNode) {
        confirmShell.statusNode.hidden = true;
        confirmShell.statusNode.textContent = "";
      }
      setStatus("", true);
    };

    passwordInput.addEventListener("input", clearPasswordErrors);
    confirmInput.addEventListener("input", clearPasswordErrors);

    wrapper.appendChild(stack);
    wrapper.appendChild(passwordErrorNode);
    stepNode.appendChild(wrapper);
    currentFieldErrorNode = null;
    updateActions();
  }

  function renderStartScreen() {
    clearFieldStatus();
    currentFieldErrorNode = null;
    const page = state.onboarding && state.onboarding.start_page ? state.onboarding.start_page : {};
    if (window.appOnboardingComponent) {
      window.appOnboardingComponent.renderStartScreen({
        stepNode,
        startButton,
        page,
      });
    }
    updateActions();
  }

  function renderConfirmationScreen() {
    clearFieldStatus();
    currentFieldErrorNode = null;
    stepNode.className = "public-request-step public-request-step-start";
    const page = state.onboarding && state.onboarding.confirmation ? state.onboarding.confirmation : {};
    const title = text(page.title);
    const copy = text(page.text);
    const buttonLabel = text(page.button) || "Complete";

    stepNode.innerHTML = "";
    confirmationButtonNode = null;
    confirmationErrorNode = null;

    const wrapper = document.createElement("div");
    wrapper.className = "public-request-start-stack";

    const titleNode = document.createElement("h2");
    titleNode.className = "public-request-title public-request-title-start";
    titleNode.textContent = title;
    wrapper.appendChild(titleNode);

    const copyNode = document.createElement("p");
    copyNode.className = "public-request-copy";
    copyNode.innerHTML = publicCommon.escapeHTML(copy).replace(/\n/g, "<br>");
    wrapper.appendChild(copyNode);

    const actionRow = document.createElement("div");
    actionRow.className = "action-row";

    const confirmationButton = document.createElement("button");
    confirmationButton.type = "submit";
    confirmationButton.className = "primary";
    confirmationButton.textContent = buttonLabel;
    actionRow.appendChild(confirmationButton);
    confirmationButtonNode = confirmationButton;
    wrapper.appendChild(actionRow);

    const errorButton = document.createElement("button");
    errorButton.type = "button";
    errorButton.className = "onboarding-confirmation-error meta-text is-error";
    errorButton.hidden = true;
    errorButton.addEventListener("click", () => {
      state.stepIndex = hasStartPage() ? -1 : -2;
      setStatus("", true);
      render();
    });
    confirmationErrorNode = errorButton;
    wrapper.appendChild(errorButton);
    stepNode.appendChild(wrapper);
    updateActions();
  }

  function renderFieldScreen(field) {
    if (field && field.key === "avatar" && field.type === "file") {
      renderAvatarScreen(field);
      return;
    }

    const result = publicCommon.renderFieldScreen({
      stepNode,
      publicFields,
      field,
      value: state.values[field.key],
      values: state.values,
      mode: "onboarding",
      onValueChange: (key, nextValue) => {
        state.values[key] = nextValue;
      },
      onClearFieldStatus: clearFieldStatus,
      onUpdateActions: updateActions,
      setFieldErrorNode: (node) => {
        currentFieldErrorNode = node;
      },
      statusNode: fieldErrorNode,
      fullNamePlaceholder: "Nome e cognome",
    });
    currentFieldErrorNode = result.statusNode;
    updateActions();
  }

  function renderAvatarScreen(field) {
    clearFieldStatus();
    currentFieldErrorNode = fieldErrorNode;
    stepNode.innerHTML = "";
    stepNode.classList.remove("public-request-step-start");

    const wrapper = document.createElement("div");
    wrapper.className = "public-request-field public-request-field-password avatar-field";

    const preview = document.createElement("div");
    preview.className = "avatar-preview";
    const previewImg = document.createElement("img");
    previewImg.alt = "Anteprima avatar";
    preview.appendChild(previewImg);
    preview.hidden = true;
    wrapper.appendChild(preview);

    let previewUrl = null;
    const updatePreview = (file) => {
      if (previewUrl) {
        URL.revokeObjectURL(previewUrl);
        previewUrl = null;
      }
      if (file instanceof File) {
        previewUrl = URL.createObjectURL(file);
        previewImg.src = previewUrl;
        preview.hidden = false;
      } else {
        previewImg.removeAttribute("src");
        preview.hidden = true;
      }
    };

    if (state.values[field.key] instanceof File) {
      updatePreview(state.values[field.key]);
    }

    const heading = document.createElement("div");
    heading.className = "public-request-password-heading";

    const titleText = text(field.title);
    if (titleText !== "") {
      const title = document.createElement("h2");
      title.className = "public-request-title";
      title.textContent = titleText;
      heading.appendChild(title);
    }

    const subtitle = document.createElement("p");
    subtitle.className = "public-request-subtitle";
    subtitle.textContent = text(field.label) || "Carica una foto quadrata. Potrai zoomare e ritagliare prima di salvare.";
    heading.appendChild(subtitle);
    wrapper.appendChild(heading);

    const controlId = `${field.key}-onboarding`;
    const input = document.createElement("input");
    input.id = controlId;
    input.type = "file";
    input.name = field.key;
    input.accept = "image/*";

    const shell = publicFields.createFieldComponent({
      label: field.label || field.title || field.key,
      controlId,
      control: input,
      hideLabel: true,
      includeStatus: false,
    });

    input.addEventListener("change", async () => {
      const selected = input.files && input.files[0] ? input.files[0] : null;
      if (!selected) {
        return;
      }
      if (!window.appAvatarCropper) {
        fieldErrorNode.textContent = "Editor immagine non disponibile.";
        fieldErrorNode.hidden = false;
        return;
      }

      clearFieldStatus();

      try {
        const cropped = await window.appAvatarCropper.open(selected);
        if (!(cropped instanceof File)) {
          return;
        }
        state.values[field.key] = cropped;
        updatePreview(cropped);
        clearFieldStatus();
        updateActions();
      } catch (error) {
        const message = error && error.message ? error.message : "Impossibile ritagliare l'immagine.";
        fieldErrorNode.textContent = message;
        fieldErrorNode.hidden = false;
      }
    });

    wrapper.appendChild(shell.root);
    stepNode.appendChild(wrapper);
    updateActions();
  }

  function render() {
    form.hidden = false;
    syncStepToUrl();
    if (isPasswordScreen()) {
      renderPasswordScreen();
      return;
    }
    if (isStartScreen()) {
      renderStartScreen();
      return;
    }
    if (isConfirmationScreen()) {
      renderConfirmationScreen();
      return;
    }
    const field = currentField();
    if (field) {
      renderFieldScreen(field);
      return;
    }
    stepNode.innerHTML = "";
    updateActions();
  }

  function buildFields(profileSchema) {
    return publicCommon.buildFields({
      steps: state.onboarding && Array.isArray(state.onboarding.steps) ? state.onboarding.steps : [],
      profileFieldsByKey: state.profileFieldsByKey,
      includeFiles: true,
      mobileKey: "mobile",
    }).filter((field) => {
      const schemaField = state.profileFieldsByKey.get(field.key);
      return !!schemaField;
    });
  }

  function collectProfileData() {
    return publicCommon.collectFieldValues(state.fields, state.values, {
      includeFiles: false,
    });
  }

  async function loadFlow() {
    if (!text(token)) {
      throw new Error("missing_token");
    }

    const [onboardingResponse, profileResponse] = await Promise.all([
      fetch(`/api/public/onboarding/${encodeURIComponent(token)}`),
      fetch("/api/public/settings/profile_schema"),
    ]);

    if (!onboardingResponse.ok) {
      throw new Error("invalid_token");
    }
    if (!profileResponse.ok) {
      throw new Error("profile_schema_failed");
    }

    const onboardingPayload = await onboardingResponse.json();
    const profilePayload = await profileResponse.json();
    const profileSchema = profilePayload && profilePayload.data ? profilePayload.data : {};

    state.email = text(onboardingPayload.email).toLowerCase();
    state.fullName = text(onboardingPayload.full_name);
    state.userId = text(onboardingPayload.user_id);
    state.passwordMin = Number(onboardingPayload.password_min) || 0;
    state.onboarding = onboardingPayload.onboarding || null;
    state.values = {};
    state.profileFieldsByKey = new Map();
    for (const field of Array.isArray(profileSchema.fields) ? profileSchema.fields : []) {
      const key = text(field && field.key);
      if (key) {
        state.profileFieldsByKey.set(key, field);
      }
    }
    state.fields = buildFields(profileSchema);
    state.ready = true;
    state.stepIndex = initialStepIndex();

    if (!state.onboarding || state.fields.length === 0) {
      throw new Error("missing_onboarding");
    }
  }

  async function submitPassword() {
    const passwordField = document.querySelector("#onboarding-password");
    const passwordConfirmField = document.querySelector("#onboarding-password-confirm");
    const passwordStatusNode = document.querySelector("#onboarding-password-error");
    const password = text(passwordField && passwordField.value);
    const passwordConfirm = text(passwordConfirmField && passwordConfirmField.value);

    if (passwordStatusNode) {
      passwordStatusNode.hidden = true;
      passwordStatusNode.textContent = "";
    }

    if (!password || !passwordConfirm) {
      if (passwordStatusNode) {
        passwordStatusNode.textContent = !password
          ? "Password is required."
          : "Password confirmation is required.";
        passwordStatusNode.hidden = false;
      }
      return false;
    }
    if (password !== passwordConfirm) {
      if (passwordStatusNode) {
        passwordStatusNode.textContent = "Passwords do not match.";
        passwordStatusNode.hidden = false;
      }
      return false;
    }
    if (state.passwordMin > 0 && Array.from(password).length < state.passwordMin) {
      if (passwordStatusNode) {
        passwordStatusNode.textContent = passwordTooShortMessage();
        passwordStatusNode.hidden = false;
      } else {
        setStatus(passwordTooShortMessage(), false);
      }
      return false;
    }

    state.submitting = true;
    updateActions();

    try {
      const response = await fetch(`/api/public/onboarding/${encodeURIComponent(token)}`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          password,
          password_confirm: passwordConfirm,
        }),
      });
      if (!response.ok) {
        let message = "Unable to set your password.";
        try {
          const payload = await response.json();
          message = text(payload && payload.message) || message;
        } catch (_) {}
        throw new Error(message);
      }
      const payload = await response.json();
      const email = text(payload.email || state.email).toLowerCase();
      if (!email) {
        throw new Error("missing_email");
      }
      await window.appAuth.login(email, password);
      state.stepIndex = 0;
      state.submitting = false;
      setStatus("", true);
      render();
      return true;
    } catch (error) {
      state.submitting = false;
      updateActions();
      setStatus(error && error.message ? error.message : "Unable to set your password.", false);
      return false;
    }
  }

  async function finalizeOnboarding() {
    const auth = window.appAuth.read();
    if (!auth || !auth.token) {
      throw new Error("missing_auth");
    }

    const payload = collectProfileData();
    const formData = new FormData();
    formData.append("data", JSON.stringify(payload));
    for (const [key, value] of Object.entries(state.values)) {
      if (value instanceof File) {
        formData.append(key, value, value.name || `${key}.png`);
      }
    }

    const response = await fetch(`/api/public/onboarding/${encodeURIComponent(token)}/finalize`, {
      method: "POST",
      body: formData,
    });
    if (!response.ok) {
      throw new Error("failed_to_finalize_onboarding");
    }

    const result = await response.json();
    return result;
  }

  async function handleSubmit(event) {
    event.preventDefault();
    setStatus("", true);

    if (isPasswordScreen()) {
      await submitPassword();
      return;
    }

    if (isStartScreen()) {
      state.stepIndex = -2;
      render();
      return;
    }

    if (isConfirmationScreen()) {
      state.submitting = true;
      updateActions();
      try {
        await finalizeOnboarding();
        window.location.replace("/me/");
      } catch (_) {
        state.submitting = false;
        updateActions();
        if (confirmationErrorNode) {
          confirmationErrorNode.innerHTML = [
            "Oh no! There is an error.<br>",
            "Start again the onboarding to fix the error.",
            ' <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-arrow-counterclockwise" viewBox="0 0 16 16" aria-hidden="true" focusable="false">',
            '  <path fill-rule="evenodd" d="M8 3a5 5 0 1 1-4.546 2.914.5.5 0 0 0-.908-.417A6 6 0 1 0 8 2z"/>',
            '  <path d="M8 4.466V.534a.25.25 0 0 0-.41-.192L5.23 2.308a.25.25 0 0 0 0 .384l2.36 1.966A.25.25 0 0 0 8 4.466"/>',
            "</svg>",
          ].join("");
          confirmationErrorNode.hidden = false;
          confirmationErrorNode.classList.add("is-error");
        }
      }
      return;
    }

    const field = currentField();
    if (!field) {
      return;
    }

    const fieldAdvance = publicFields.validateFieldOnAdvance(field, state.values[field.key]);
    if (fieldAdvance.error) {
      if (currentFieldErrorNode) {
        currentFieldErrorNode.textContent = fieldAdvance.error;
        currentFieldErrorNode.hidden = false;
      } else {
        setStatus(fieldAdvance.error, false);
      }
      return;
    }
    if (!publicCommon.isFieldComplete(field, state.values[field.key])) {
      const message = field.key === "avatar" && field.type === "file"
        ? "Load an image."
        : "This field is required.";
      if (currentFieldErrorNode) {
        currentFieldErrorNode.textContent = message;
        currentFieldErrorNode.hidden = false;
      } else {
        setStatus(message, false);
      }
      return;
    }
    if (fieldAdvance.normalized) {
      state.values[field.key] = fieldAdvance.normalized;
    }

    if (state.stepIndex >= state.fields.length - 1) {
      if (hasConfirmationPage()) {
        state.stepIndex = state.fields.length;
        render();
        return;
      }

      state.submitting = true;
      updateActions();
      try {
        await finalizeOnboarding();
        window.location.replace("/me/");
      } catch (_) {
        state.submitting = false;
        updateActions();
        if (confirmationErrorNode) {
          confirmationErrorNode.innerHTML = [
            "Oh no! There is an error.<br>",
            "Start again the onboarding to fix the error.",
            ' <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-arrow-counterclockwise" viewBox="0 0 16 16" aria-hidden="true" focusable="false">',
            '  <path fill-rule="evenodd" d="M8 3a5 5 0 1 1-4.546 2.914.5.5 0 0 0-.908-.417A6 6 0 1 0 8 2z"/>',
            '  <path d="M8 4.466V.534a.25.25 0 0 0-.41-.192L5.23 2.308a.25.25 0 0 0 0 .384l2.36 1.966A.25.25 0 0 0 8 4.466"/>',
            "</svg>",
          ].join("");
          confirmationErrorNode.hidden = false;
          confirmationErrorNode.classList.add("is-error");
        }
      }
      return;
    }

    state.stepIndex += 1;
    render();
  }

  backButton.addEventListener("click", () => {
    setStatus("", true);
    if (isConfirmationScreen()) {
      state.stepIndex = state.fields.length - 1;
      render();
      return;
    }
    if (isPasswordScreen()) {
      if (hasStartPage()) {
        state.stepIndex = -1;
        render();
      }
      return;
    }
    if (isStartScreen()) {
      return;
    }
    if (state.stepIndex > 0) {
      state.stepIndex -= 1;
      render();
      return;
    }
    if (state.stepIndex === 0) {
      state.stepIndex = -2;
      render();
    }
  });

  startButton.addEventListener("click", () => {
    setStatus("", true);
    if (isStartScreen()) {
      state.stepIndex = -2;
      render();
    }
  });

  form.addEventListener("submit", handleSubmit);

  loadFlow()
    .then(() => {
      setLoading("", false);
      if (isPasswordScreen()) {
        render();
        return;
      }
      form.hidden = false;
      render();
    })
    .catch(() => {
      setLoading("This onboarding link is invalid.", true);
      form.hidden = true;
    });
})();
