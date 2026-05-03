window.appPublicCommon = (() => {
  let tabsInitialized = false;

  function text(value) {
    return String(value || "").trim();
  }

  function escapeHTML(value) {
    return String(value || "")
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;")
      .replaceAll("'", "&#39;");
  }

  function optionParts(option) {
    const raw = text(option);
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
    const type = text(field && field.type).toLowerCase();
    if (type === "email") {
      return "email";
    }
    if (type === "phone") {
      return "tel";
    }
    return "text";
  }

  function isCustomSelectOption(value) {
    return String(value || "").includes(":input");
  }

  function parseBoolean(value) {
    if (typeof value === "boolean") {
      return value;
    }
    if (typeof value === "number") {
      return value !== 0;
    }
    const normalized = text(value).toLowerCase();
    if (!normalized) {
      return null;
    }
    if (["true", "1", "yes", "on"].includes(normalized)) {
      return true;
    }
    if (["false", "0", "no", "off"].includes(normalized)) {
      return false;
    }
    return null;
  }

  function hasPageContent(page) {
    return !!(page && (text(page.title) || text(page.text) || text(page.button)));
  }

  function renderPageScreen(stepNode, page, options = {}) {
    if (!stepNode) {
      return;
    }
    const className = text(options.className) || "public-request-step-start";
    const titleClass = text(options.titleClass) || "public-request-title public-request-title-start";
    const copyClass = text(options.copyClass) || "public-request-copy";
    const title = escapeHTML(text(page && page.title));
    const copy = escapeHTML(text(page && page.text)).replace(/\n/g, "<br>");

    stepNode.innerHTML = `
      <div class="${className}">
        <h2 class="${titleClass}">${title}</h2>
        <p class="${copyClass}">${copy}</p>
      </div>
    `;
  }

  function buildFields(options) {
    const {
      steps,
      profileFieldsByKey,
      includeFiles = true,
      mobileKey = "mobile",
    } = options || {};

    const resolved = [];
    const stepList = Array.isArray(steps) ? steps : [];
    const fieldMap = profileFieldsByKey instanceof Map ? profileFieldsByKey : new Map();

    for (const step of stepList) {
      const key = text(step && step.field);
      if (!key) {
        continue;
      }
      const schemaField = fieldMap.get(key);
      if (!schemaField) {
        continue;
      }
      const rawType = text(schemaField.type).toLowerCase();
      const type = key === mobileKey && rawType === "text" ? "phone" : rawType;
      if (!type || (!includeFiles && type === "file")) {
        continue;
      }
      const stepUnique = parseBoolean(step && step.unique);
      const schemaUnique = parseBoolean(schemaField.unique);
      const isSingleSelect = stepUnique !== null ? stepUnique : schemaUnique !== null ? schemaUnique : false;
      const resolvedLabel = text(step.label || schemaField.label || step.title || schemaField.title || key);
      const resolvedTitle = text(step.title || schemaField.title || resolvedLabel || key);

      resolved.push({
        key,
        type,
        title: resolvedTitle,
        label: resolvedLabel,
        default: schemaField.default,
        required: schemaField.required !== false,
        min: Number.isFinite(schemaField.min) ? schemaField.min : 0,
        max: Number.isFinite(schemaField.max) ? schemaField.max : 0,
        options: Array.isArray(schemaField.options) ? schemaField.options : [],
        multiple: rawType === "select" ? !isSingleSelect : false,
      });
    }

    return resolved;
  }

  function isFieldComplete(field, value) {
    if (!field) {
      return false;
    }

    const raw = text(value);
    if (field.type === "file") {
      return value instanceof File || raw !== "";
    }
    if (field.type === "select") {
      if (field.multiple) {
        const items = Array.isArray(value) ? value : [];
        if (field.min > 0 && items.length < field.min) {
          return false;
        }
        return items.length > 0;
      }
      return raw !== "";
    }
    if (field.type === "textarea" || field.type === "text" || field.type === "email" || field.type === "phone") {
      return raw !== "";
    }
    return false;
  }

  function collectFieldValues(fields, values, options = {}) {
    const {
      includeFiles = false,
      selectCustomSuffix = "__custom",
    } = options;

    const payload = {};
    for (const field of Array.isArray(fields) ? fields : []) {
      const value = values ? values[field.key] : undefined;
      if (field.type === "file") {
        if (includeFiles && value instanceof File) {
          payload[field.key] = value;
        }
        continue;
      }
      if (field.type === "select") {
        const selectedValues = Array.isArray(value) ? value : value ? [value] : [];
        const normalizedValues = selectedValues.map((item) => (
          isCustomSelectOption(item)
            ? text(values[`${field.key}${selectCustomSuffix}`] || "")
            : text(item)
        )).filter((item) => item !== "");

        payload[field.key] = field.multiple ? normalizedValues : (normalizedValues[0] || "");
        continue;
      }
      payload[field.key] = text(value);
    }
    return payload;
  }

  function renderFieldScreen(options) {
    const {
      stepNode,
      publicFields,
      field,
      value,
      values,
      mode = "public",
      onValueChange,
      onClearFieldStatus,
      onUpdateActions,
      setFieldErrorNode,
      statusNode = null,
      useInlineError = false,
      fullNamePlaceholder = "Nome e cognome",
    } = options || {};

    if (!stepNode || !publicFields || !field) {
      return;
    }

    stepNode.innerHTML = "";
    stepNode.className = "public-request-step";
    stepNode.classList.remove("public-request-step-start");
    if (typeof setFieldErrorNode === "function") {
      setFieldErrorNode(null);
    }

    const wrapper = document.createElement("div");
    wrapper.className = "public-request-field public-request-field-password";
    let inlineErrorNode = null;
    const usePasswordLikeSingleField = useInlineError && field.type !== "select" && field.type !== "file";

    const heading = document.createElement("div");
    heading.className = "public-request-password-heading";

    const titleText = text(field.title);
    if (titleText !== "") {
      const title = document.createElement("h2");
      title.className = "public-request-title";
      title.textContent = titleText;
      heading.appendChild(title);
    }

    const subtitleText = text(field.label);
    if (subtitleText && subtitleText !== titleText) {
      const subtitle = document.createElement("p");
      subtitle.className = "public-request-subtitle";
      subtitle.textContent = subtitleText;
      heading.appendChild(subtitle);
    }

    wrapper.appendChild(heading);
    if (useInlineError) {
      inlineErrorNode = document.createElement("p");
      inlineErrorNode.className = usePasswordLikeSingleField
        ? "public-request-password-error meta-text is-error"
        : "public-request-field-error meta-text is-error";
      inlineErrorNode.hidden = true;
    }
    if (typeof setFieldErrorNode === "function") {
      const useTopbarError = !!(statusNode && !(useInlineError && field.type !== "select"));
      setFieldErrorNode(useTopbarError ? statusNode : inlineErrorNode);
    }

    const idSuffix = mode === "onboarding" ? "-onboarding" : "-public";
    const customSuffix = `${field.key}__custom${idSuffix}`;
    let input = null;
    let shell = null;

    function notifyChange(nextValue, rerender = false) {
      if (typeof onValueChange === "function") {
        onValueChange(field.key, nextValue);
      }
      if (typeof onClearFieldStatus === "function") {
        onClearFieldStatus();
      }
      if (typeof onUpdateActions === "function") {
        onUpdateActions();
      }
      if (rerender) {
        renderFieldScreen({
          stepNode,
          publicFields,
          field,
          value: nextValue,
          values,
          mode,
          onValueChange,
          onClearFieldStatus,
          onUpdateActions,
          setFieldErrorNode,
          statusNode,
          useInlineError,
          fullNamePlaceholder,
        });
      }
    }

    if (field.type === "textarea") {
      const controlId = `${field.key}${idSuffix}`;
      input = document.createElement("textarea");
      input.id = controlId;
      input.name = field.key;
      input.value = text(value);
      if (usePasswordLikeSingleField) {
        input.classList.add("public-request-password-look");
      }
      shell = publicFields.createFieldComponent({
        label: field.label || field.title || field.key,
        controlId,
        control: input,
        errorId: `${field.key}-error`,
        hideLabel: true,
        includeStatus: false,
      });
      if (field.key === "full_name") {
        input.placeholder = fullNamePlaceholder;
      }
      input.addEventListener("input", () => notifyChange(input.value));
      if (usePasswordLikeSingleField) {
        const stack = document.createElement("div");
        stack.className = "public-request-password-stack";
        stack.appendChild(shell.root);
        wrapper.appendChild(stack);
      } else {
        wrapper.appendChild(shell.root);
      }
    } else if (field.type === "select") {
      const selected = field.multiple
        ? (Array.isArray(value) ? value.map((item) => text(item)) : [text(value)])
        : [text(Array.isArray(value) ? value[0] : value)];
      const optionsNode = document.createElement("div");
      optionsNode.className = "public-option-list";

      for (const option of Array.isArray(field.options) ? field.options : []) {
        const parts = optionParts(option);
        if (!parts.value) {
          continue;
        }
        const button = document.createElement("button");
        button.type = "button";
        button.className = "public-option";
        if (selected.includes(parts.value)) {
          button.classList.add("public-option-current");
        }
        button.dataset.optionValue = parts.value;
        button.textContent = parts.label;
        button.addEventListener("click", () => {
          const optionValue = button.dataset.optionValue || "";
          let nextValue;
          if (field.multiple) {
            const nextValues = Array.isArray(values[field.key]) ? [...values[field.key]] : [];
            if (nextValues.includes(optionValue)) {
              nextValue = nextValues.filter((item) => item !== optionValue);
            } else {
              nextValue = [...nextValues, optionValue];
            }
          } else {
            nextValue = optionValue;
          }
          values[field.key] = nextValue;
          if (!isCustomSelectOption(field.multiple ? nextValue[0] || "" : nextValue)) {
            values[`${field.key}__custom`] = "";
          }
          if (typeof onClearFieldStatus === "function") {
            onClearFieldStatus();
          }
          if (typeof onUpdateActions === "function") {
            onUpdateActions();
          }
          renderFieldScreen({
            stepNode,
            publicFields,
            field,
            value: nextValue,
            values,
            mode,
            onValueChange,
            onClearFieldStatus,
            onUpdateActions,
            setFieldErrorNode,
            fullNamePlaceholder,
          });
        });
        optionsNode.appendChild(button);
      }

      shell = publicFields.createFieldComponent({
        label: field.label || field.title || field.key,
        control: optionsNode,
        errorId: `${field.key}-error`,
        hideLabel: true,
        includeStatus: false,
      });
      wrapper.appendChild(shell.root);

      if (selected.some(isCustomSelectOption)) {
        const customInput = document.createElement("input");
        customInput.id = customSuffix;
        customInput.type = "text";
        customInput.name = `${field.key}__custom`;
        customInput.value = text(values[`${field.key}__custom`] || "");
        customInput.placeholder = "Specifica";
        const customShell = publicFields.createFieldComponent({
          label: "Specifica",
          controlId: customInput.id,
          control: customInput,
          includeStatus: false,
          hideLabel: true,
        });
        customShell.root.classList.add("public-request-custom-input");
        customInput.classList.add("public-request-password-look");
        customInput.addEventListener("input", () => {
          values[`${field.key}__custom`] = customInput.value;
          if (typeof onClearFieldStatus === "function") {
            onClearFieldStatus();
          }
          if (typeof onUpdateActions === "function") {
            onUpdateActions();
          }
        });
        optionsNode.appendChild(customShell.root);
        requestAnimationFrame(() => {
          customShell.root.scrollIntoView({ block: "center", behavior: "smooth" });
          customInput.focus();
          const end = customInput.value.length;
          try {
            customInput.setSelectionRange(end, end);
          } catch (_) {}
        });
      }
    } else if (field.type === "file") {
      const controlId = `${field.key}${idSuffix}`;
      input = document.createElement("input");
      input.id = controlId;
      input.type = "file";
      input.name = field.key;
      shell = publicFields.createFieldComponent({
        label: field.label || field.title || field.key,
        controlId,
        control: input,
        errorId: `${field.key}-error`,
        hideLabel: true,
        includeStatus: false,
      });
      input.addEventListener("change", () => {
        notifyChange(input.files && input.files[0] ? input.files[0] : null);
      });
      wrapper.appendChild(shell.root);
    } else {
      const controlId = `${field.key}${idSuffix}`;
      input = document.createElement("input");
      input.id = controlId;
      input.type = inputTypeFor(field);
      input.name = field.key;
      input.value = field.type === "phone"
        ? publicFields.normalizePhoneDisplay(value || "+39")
        : text(value);
      if (input.type === "email") {
        input.autocomplete = "email";
      }
      if (input.type === "tel") {
        input.autocomplete = "tel";
      }
      if (usePasswordLikeSingleField) {
        input.classList.add("public-request-password-look");
      }
      shell = publicFields.createFieldComponent({
        label: field.label || field.title || field.key,
        controlId,
        control: input,
        errorId: `${field.key}-error`,
        hideLabel: true,
        includeStatus: false,
      });
      if (field.key === "full_name") {
        input.placeholder = fullNamePlaceholder;
      }
      input.addEventListener("input", () => {
        const nextValue = field.type === "phone" ? publicFields.normalizePhoneDisplay(input.value) : input.value;
        if (typeof onValueChange === "function") {
          onValueChange(field.key, nextValue);
        }
        if (field.type === "phone") {
          input.value = nextValue;
        }
        if (typeof onClearFieldStatus === "function") {
          onClearFieldStatus();
        }
        if (typeof onUpdateActions === "function") {
          onUpdateActions();
        }
      });
      if (field.type === "phone") {
        input.addEventListener("focus", () => {
          const end = input.value.length;
          try {
            input.setSelectionRange(end, end);
          } catch (_) {}
        });
      }
      if (usePasswordLikeSingleField) {
        const stack = document.createElement("div");
        stack.className = "public-request-password-stack";
        stack.appendChild(shell.root);
        wrapper.appendChild(stack);
      } else {
        wrapper.appendChild(shell.root);
      }
    }

    if (!input) {
      if (inlineErrorNode && field.type !== "select") {
        wrapper.appendChild(inlineErrorNode);
      }
      stepNode.appendChild(wrapper);
      return { statusNode: (statusNode && field.type === "select") ? statusNode : inlineErrorNode };
    }

    if (inlineErrorNode && field.type !== "select") {
      wrapper.appendChild(inlineErrorNode);
    }
    stepNode.appendChild(wrapper);
    input.focus();
    return { statusNode: (statusNode && field.type === "select") ? statusNode : inlineErrorNode };
  }

  function initTabs() {
    if (tabsInitialized) {
      return;
    }

    const tabNodes = Array.from(document.querySelectorAll("[data-public-mode]"));
    const paneNodes = Array.from(document.querySelectorAll("[data-public-pane]"));
    if (paneNodes.length === 0) {
      return;
    }

    const defaultMode = document.body.dataset.publicdefaultmode === "login" ? "login" : "request";
    if (tabNodes.length === 0) {
      for (const paneNode of paneNodes) {
        paneNode.hidden = paneNode.dataset.publicPane !== defaultMode;
      }
      return;
    }

    tabsInitialized = true;
    
    const modePath = {
      request: "/signup/",
      login: "/",
    };

    function normalizePath(path) {
      const raw = text(path);
      if (!raw) {
        return "/";
      }
      const normalized = raw.endsWith("/") && raw.length > 1 ? raw.slice(0, -1) : raw;
      return normalized || "/";
    }

    for (const tabNode of tabNodes) {
      tabNode.addEventListener("click", () => {
        const targetMode = tabNode.dataset.publicMode;
        const targetPath = modePath[targetMode];
        if (targetPath && normalizePath(window.location.pathname) !== normalizePath(targetPath)) {
          window.location.assign(targetPath);
          return;
        }

        for (const otherTab of tabNodes) {
          const active = otherTab === tabNode;
          otherTab.classList.toggle("section-tab-current", active);
          otherTab.setAttribute("aria-selected", active ? "true" : "false");
        }
        for (const paneNode of paneNodes) {
          paneNode.hidden = paneNode.dataset.publicPane !== tabNode.dataset.publicMode;
        }
      });
    }

    const initialTab = tabNodes.find((node) => node.dataset.publicMode === defaultMode) || tabNodes[0];
    if (initialTab) {
      initialTab.dispatchEvent(new Event("click"));
    }
  }

  initTabs();

  return {
    text,
    escapeHTML,
    optionParts,
    inputTypeFor,
    hasPageContent,
    renderPageScreen,
    buildFields,
    isFieldComplete,
    collectFieldValues,
    renderFieldScreen,
    initTabs,
  };
})();
