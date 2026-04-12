(() => {
  const signupSettingsURL = "/api/public/settings/signup";
  const profileSchemaURL = "/api/public/settings/profile_schema";
  const submitURL = "/api/public/requests";

  const requestForm = document.querySelector("#public-request-form");
  const requestProgressNode = document.querySelector("#public-request-progress");
  const requestStepNode = document.querySelector("#public-request-fields");
  const requestStatusNode = document.querySelector("#public-request-status");
  const requestTopbarNode = document.querySelector("#public-request-topbar");
  const requestSubmitButton = document.querySelector("#public-request-submit");
  const requestStartButton = document.querySelector("#public-request-start");
  const requestBackButton = document.querySelector("#public-request-back");
  const requestActionsNode = document.querySelector("#public-request-actions");
  const publicTabsNode = document.querySelector("#public-entry-tabs");
  const loginForm = document.querySelector("#login-form");
  const loginEmailField = document.querySelector("#login-email");
  const loginPasswordField = document.querySelector("#login-password");
  const loginErrorNode = document.querySelector("#login-error");
  const tabNodes = Array.from(document.querySelectorAll("[data-public-mode]"));
  const paneNodes = Array.from(document.querySelectorAll("[data-public-pane]"));
  const defaultMode = document.body.dataset.publicdefaultmode === "login" ? "login" : "request";

  if (
    !requestForm ||
    !requestProgressNode ||
    !requestStepNode ||
    !requestStatusNode ||
    !requestTopbarNode ||
    !requestSubmitButton ||
    !requestStartButton ||
    !requestBackButton ||
    !requestActionsNode ||
    !publicTabsNode ||
    !loginForm ||
    !loginEmailField ||
    !loginPasswordField ||
    !loginErrorNode ||
    !window.huuperAuth
  ) {
    return;
  }

  const auth = window.huuperAuth.read();
  if (auth && auth.token) {
    window.location.replace(window.huuperAuth.redirectAfterLogin(auth.model && auth.model.admin === true ? "admin" : "me"));
    return;
  }

  const requestState = {
    startPage: null,
    confirmation: null,
    success: null,
    fields: [],
    signupSteps: [],
    profileFieldsByKey: new Map(),
    values: {},
    stepIndex: 0,
    submitLabel: "Send request",
    showConfirmation: false,
    showSuccess: false,
    submitting: false,
    ready: false,
  };

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
      requestStatusNode.textContent = "Request failed.";
      requestStatusNode.hidden = false;
    }
  }

  function showMode(mode) {
    for (const tabNode of tabNodes) {
      const active = tabNode.dataset.publicMode === mode;
      tabNode.classList.toggle("section-tab-current", active);
      tabNode.setAttribute("aria-selected", active ? "true" : "false");
    }
    for (const paneNode of paneNodes) {
      paneNode.hidden = paneNode.dataset.publicPane !== mode;
    }
  }

  function optionParts(option) {
    const raw = String(option || "").trim();
    if (!raw) {
      return { value: "", label: "" };
    }
    const suffixIndex = raw.indexOf(":");
    if (suffixIndex < 0) {
      return { value: raw, label: raw };
    }
    return {
      value: raw,
      label: raw.slice(0, suffixIndex).trim(),
    };
  }

  function inputTypeFor(field) {
    const type = String(field.type || "").trim().toLowerCase();
    if (type === "email") {
      return "email";
    }
    if (type === "phone") {
      return "tel";
    }
    return "text";
  }

  function currentField() {
    return requestState.fields[requestState.stepIndex] || null;
  }

  function hasStartPage() {
    return !!(requestState.startPage && (requestState.startPage.title || requestState.startPage.text || requestState.startPage.button));
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

  function isFieldComplete(field) {
    if (!field) {
      return false;
    }
    const value = requestState.values[field.key];
    if (field.type === "select") {
      const items = Array.isArray(value) ? value : [];
      const min = Number.isFinite(field.min) ? field.min : 1;
      if (items.length < min) {
        return false;
      }
      if (!items.some((item) => String(item).includes(":input"))) {
        return true;
      }
      return String(requestState.values[`${field.key}__custom`] || "").trim() !== "";
    }
    if (field.type === "textarea" || field.type === "text" || field.type === "email" || field.type === "phone") {
      return String(value || "").trim() !== "";
    }
    return false;
  }

  function updateActions() {
    requestStatusNode.hidden = true;
    const startScreen = isStartScreen();
    const confirmationScreen = isConfirmationScreen();
    const successScreen = isSuccessScreen();
    requestBackButton.hidden = !hasStartPage() && requestState.stepIndex <= 0;
    requestForm.classList.toggle("public-request-form-start", startScreen || confirmationScreen || successScreen);
    requestForm.classList.toggle("public-request-form-flow", !startScreen && !confirmationScreen && !successScreen);
    requestActionsNode.hidden = !startScreen && !confirmationScreen && !successScreen;
    requestTopbarNode.hidden = startScreen || confirmationScreen || successScreen;
    publicTabsNode.classList.toggle("public-entry-tabs-hidden", !startScreen);
    publicTabsNode.hidden = !startScreen;
    requestStartButton.classList.toggle("is-loading", confirmationScreen && requestState.submitting);

    if (startScreen) {
      requestProgressNode.hidden = true;
      requestProgressNode.textContent = "";
      requestStartButton.textContent = String(requestState.startPage.button || "Start").trim() || "Start";
      requestStartButton.disabled = false;
      return;
    }

    if (confirmationScreen) {
      requestProgressNode.hidden = true;
      requestStartButton.textContent = requestState.submitting
        ? "Invio..."
        : String(requestState.confirmation.button || requestState.submitLabel || "Send request").trim();
      requestStartButton.disabled = requestState.submitting;
      requestProgressNode.textContent = "";
      return;
    }

    if (successScreen) {
      requestProgressNode.hidden = true;
      requestProgressNode.textContent = "";
      requestStartButton.textContent = String(requestState.success.button || "Continua").trim() || "Continua";
      requestStartButton.disabled = false;
      return;
    }

    requestProgressNode.hidden = false;
    requestProgressNode.textContent = `Step ${requestState.stepIndex + 1} / ${requestState.fields.length}`;
    requestSubmitButton.textContent = requestState.stepIndex >= requestState.fields.length - 1 && !requestState.confirmation
      ? requestState.submitLabel
      : "Next";
    requestSubmitButton.disabled = !isFieldComplete(currentField());
  }

  function renderStartScreen() {
    requestStepNode.innerHTML = "";
    requestStepNode.classList.add("public-request-step-start");

    const title = document.createElement("h2");
    title.className = "public-request-title public-request-title-start";
    title.textContent = String(requestState.startPage.title || "").trim();

    const text = document.createElement("p");
    text.className = "public-request-copy";
    text.innerHTML = String(requestState.startPage.text || "").trim().replace(/\n/g, "<br>");

    requestStepNode.append(title, text);
    updateActions();
  }

  function createSelectableOption(field, option, selectedValues) {
    const parts = optionParts(option);
    if (!parts.value) {
      return null;
    }

    const button = document.createElement("button");
    button.type = "button";
    button.className = "public-option";
    if (selectedValues.includes(parts.value)) {
      button.classList.add("public-option-current");
    }
    button.dataset.optionValue = parts.value;
    button.textContent = parts.label;
    return button;
  }

  function renderFieldScreen(field) {
    requestStepNode.innerHTML = "";
    requestStepNode.classList.remove("public-request-step-start");

    const wrapper = document.createElement("div");
    wrapper.className = "public-request-field";

    const title = document.createElement("h2");
    title.className = "public-request-title";
    title.textContent = String(field.title || field.label || "").trim();
    wrapper.appendChild(title);

    if (field.label && String(field.label).trim() && String(field.label).trim() !== String(field.title || "").trim()) {
      const subtitle = document.createElement("p");
      subtitle.className = "public-request-copy public-request-copy-subtle";
      subtitle.textContent = String(field.label).trim();
      wrapper.appendChild(subtitle);
    }

    if (field.type === "textarea") {
      const textarea = document.createElement("textarea");
      textarea.name = field.key;
      textarea.value = String(requestState.values[field.key] || "");
      wrapper.appendChild(textarea);
      textarea.addEventListener("input", () => {
        requestState.values[field.key] = textarea.value;
        updateActions();
      });
      requestStepNode.appendChild(wrapper);
      textarea.focus();
      updateActions();
      return;
    }

    if (field.type === "select") {
      const selectedValues = Array.isArray(requestState.values[field.key]) ? requestState.values[field.key] : [];
      const optionsNode = document.createElement("div");
      optionsNode.className = "public-option-list";

      for (const option of field.options) {
        const optionNode = createSelectableOption(field, option, selectedValues);
        if (!optionNode) {
          continue;
        }
        optionNode.addEventListener("click", () => {
          const optionValue = optionNode.dataset.optionValue || "";
          let nextValues = Array.isArray(requestState.values[field.key]) ? [...requestState.values[field.key]] : [];
          const isMultiple = !Number.isFinite(field.max) || field.max > 1;

          if (isMultiple) {
            if (nextValues.includes(optionValue)) {
              nextValues = nextValues.filter((item) => item !== optionValue);
            } else {
              nextValues.push(optionValue);
            }
          } else {
            nextValues = [optionValue];
          }

          requestState.values[field.key] = nextValues;
          if (!nextValues.some((item) => String(item).includes(":input"))) {
            requestState.values[`${field.key}__custom`] = "";
          }
          renderFieldScreen(field);
        });
        optionsNode.appendChild(optionNode);
      }

      wrapper.appendChild(optionsNode);

      if (selectedValues.some((item) => String(item).includes(":input"))) {
        const customInput = document.createElement("input");
        customInput.type = "text";
        customInput.name = `${field.key}__custom`;
        customInput.className = "public-request-custom-input";
        customInput.value = String(requestState.values[`${field.key}__custom`] || "");
        customInput.placeholder = "Specifica";
        customInput.addEventListener("input", () => {
          requestState.values[`${field.key}__custom`] = customInput.value;
          updateActions();
        });
        wrapper.appendChild(customInput);
      }

      requestStepNode.appendChild(wrapper);
      updateActions();
      return;
    }

    const input = document.createElement("input");
    input.type = inputTypeFor(field);
    input.name = field.key;
    input.value = String(requestState.values[field.key] || field.default || "");
    if (input.type === "email") {
      input.placeholder = "name@domain.com";
      input.autocomplete = "email";
    }
    if (input.type === "tel") {
      input.autocomplete = "tel";
    }
    wrapper.appendChild(input);
    input.addEventListener("input", () => {
      requestState.values[field.key] = input.value;
      updateActions();
    });

    requestStepNode.appendChild(wrapper);
    input.focus();
    updateActions();
  }

  function renderConfirmationScreen() {
    requestStepNode.innerHTML = "";
    requestStepNode.classList.add("public-request-step-start");

    const title = document.createElement("h2");
    title.className = "public-request-title public-request-title-start";
    title.textContent = String(requestState.confirmation.title || "").trim();

    const text = document.createElement("p");
    text.className = "public-request-copy";
    text.innerHTML = String(requestState.confirmation.text || "").trim().replace(/\n/g, "<br>");

    requestStepNode.append(title, text);
    updateActions();
  }

  function renderSuccessScreen() {
    requestStepNode.innerHTML = "";
    requestStepNode.classList.add("public-request-step-start");

    const title = document.createElement("h2");
    title.className = "public-request-title public-request-title-start";
    title.textContent = String(requestState.success.title || "").trim();

    const text = document.createElement("p");
    text.className = "public-request-copy";
    text.innerHTML = String(requestState.success.text || "").trim().replace(/\n/g, "<br>");

    requestStepNode.append(title, text);
    updateActions();
  }

  function renderStep() {
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
    const resolved = [];

    for (const step of requestState.signupSteps) {
      const key = String(step && step.field || "").trim();
      if (!key) {
        continue;
      }
      const schemaField = requestState.profileFieldsByKey.get(key);
      if (!schemaField) {
        continue;
      }
      const type = String(schemaField.type || "").trim().toLowerCase();
      if (type === "file") {
        continue;
      }
      resolved.push({
        key,
        type,
        title: String(step.title || schemaField.title || schemaField.label || "").trim(),
        label: String(step.label || schemaField.label || "").trim(),
        default: schemaField.default,
        min: Number(schemaField.min),
        max: Number(schemaField.max),
        options: Array.isArray(schemaField.options) ? schemaField.options : [],
      });
    }

    return resolved;
  }

  function collectRequestPayload() {
    const payload = {};
    for (const field of requestState.fields) {
      if (field.type === "select") {
        const values = Array.isArray(requestState.values[field.key]) ? requestState.values[field.key] : [];
        if (values.length === 0) {
          payload[field.key] = "";
        } else if (values.length === 1) {
          const value = values[0];
          if (String(value).includes(":input")) {
            payload[field.key] = String(requestState.values[`${field.key}__custom`] || "").trim();
          } else {
            payload[field.key] = value;
          }
        } else {
          payload[field.key] = values.map((value) => (
            String(value).includes(":input")
              ? String(requestState.values[`${field.key}__custom`] || "").trim()
              : value
          ));
        }
        continue;
      }
      payload[field.key] = String(requestState.values[field.key] || "").trim();
    }
    return payload;
  }

  async function loadRequestFlow() {
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

    if (isStartScreen()) {
      renderStep();
    }

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

    if (!isStartScreen()) {
      renderStep();
    }
  }

  requestBackButton.addEventListener("click", () => {
    requestStatusNode.hidden = true;
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

  for (const tabNode of tabNodes) {
    tabNode.addEventListener("click", () => {
      showMode(tabNode.dataset.publicMode === "login" ? "login" : "request");
    });
  }

  requestStartButton.addEventListener("click", async () => {
    requestStatusNode.hidden = true;
    requestStatusNode.textContent = "";

    if (isStartScreen()) {
      if (!requestState.ready) {
        requestStatusNode.textContent = "Loading...";
        requestStatusNode.hidden = false;
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
    requestStatusNode.hidden = true;
    requestStatusNode.textContent = "";

    if (isConfirmationScreen()) {
      await submitRequest();
      return;
    }

    if (!isFieldComplete(currentField())) {
      updateActions();
      return;
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

  loginForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    loginErrorNode.hidden = true;
    loginErrorNode.textContent = "";

    const email = loginEmailField.value.trim();
    const password = loginPasswordField.value;

    if (!email || !password) {
      loginErrorNode.textContent = "Email and password are required.";
      loginErrorNode.hidden = false;
      return;
    }

    try {
      const result = await window.huuperAuth.login(email, password);
      window.location.href = window.huuperAuth.redirectAfterLogin(result.scope);
    } catch (_) {
      loginErrorNode.textContent = "Invalid credentials.";
      loginErrorNode.hidden = false;
    }
  });

  showMode(defaultMode);
  loadRequestFlow().catch(() => {
    requestStatusNode.textContent = "Signup unavailable.";
    requestStatusNode.hidden = false;
  });
})();
