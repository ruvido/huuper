(() => {
  const PHONE_MIN_TOTAL_DIGITS = 9;
  const FIELD_REQUIRED_MESSAGE = "Field cannot be empty.";
  const FULL_NAME_REQUIRED_MESSAGE = "Both First and Last name are required";
  const EMAIL_INVALID_MESSAGE = "Enter a valid email address.";
  const PHONE_INVALID_MESSAGE = "Enter a valid phone number.";

  function text(value) {
    return String(value || "").trim();
  }

  function capitalizeWord(value) {
    const chars = Array.from(text(value));
    if (chars.length === 0) {
      return "";
    }

    const [first, ...rest] = chars;
    return [first.toLocaleUpperCase(), ...rest.map((char) => char.toLocaleLowerCase())].join("");
  }

  function normalizeFullName(value) {
    const parts = text(value).split(/\s+/).filter(Boolean);
    if (parts.length < 2) {
      return "";
    }
    return parts.map(capitalizeWord).join(" ");
  }

  function isValidEmail(value) {
    const input = document.createElement("input");
    input.type = "email";
    input.value = text(value);
    if (!input.checkValidity()) {
      return false;
    }
    const parts = text(value).split("@");
    if (parts.length !== 2) {
      return false;
    }
    return parts[1].includes(".");
  }

  function normalizeEmail(value) {
    const normalized = text(value).toLowerCase();
    if (!normalized || !isValidEmail(normalized)) {
      return "";
    }
    return normalized;
  }

  function normalizePhoneDisplay(value) {
    const raw = text(value).replace(/\s+/g, "");
    if (!raw.startsWith("+")) {
      return text(value);
    }
    if (raw.length <= 3) {
      return raw === "+" ? raw : `${raw} `;
    }
    return `${raw.slice(0, 3)} ${raw.slice(3)}`;
  }

  function normalizePhone(value) {
    const raw = text(value);
    if (!raw) {
      return "";
    }

    const compact = raw.replace(/\s+/g, "");
    if (!compact.startsWith("+")) {
      return "";
    }

    const digits = compact.slice(1);
    if (!/^\d+$/.test(digits)) {
      return "";
    }

    if (digits.length < PHONE_MIN_TOTAL_DIGITS) {
      return "";
    }

    return `+${digits.slice(0, 2)} ${digits.slice(2)}`;
  }

  function validateFieldOnAdvance(field, value) {
    const key = text(field && field.key);
    const type = text(field && field.type).toLowerCase();
    const required = field && field.required !== false;

    if (type === "file") {
      if (value instanceof File || value instanceof Blob) {
        return { error: "", normalized: value };
      }
      return { error: required ? FIELD_REQUIRED_MESSAGE : "", normalized: null };
    }

    const raw = text(value);

    if (key === "full_name") {
      if (!raw) {
        return { error: FULL_NAME_REQUIRED_MESSAGE, normalized: "" };
      }
      const normalized = normalizeFullName(raw);
      if (!normalized) {
        return { error: FULL_NAME_REQUIRED_MESSAGE, normalized: "" };
      }
      return { error: "", normalized };
    }

    if (key === "email" || type === "email") {
      if (!raw) {
        return { error: required ? FIELD_REQUIRED_MESSAGE : "", normalized: "" };
      }
      const normalized = normalizeEmail(raw);
      if (!normalized) {
        return { error: EMAIL_INVALID_MESSAGE, normalized: "" };
      }
      return { error: "", normalized };
    }

    if (type === "select") {
      if (!raw) {
        return { error: required ? "Select at least one choice" : "", normalized: value || "" };
      }
      return { error: "", normalized: value };
    }

    if (key === "mobile" || type === "phone") {
      if (!raw) {
        return { error: required ? FIELD_REQUIRED_MESSAGE : "", normalized: "" };
      }
      const normalized = normalizePhone(raw);
      if (!normalized) {
        return { error: PHONE_INVALID_MESSAGE, normalized: "" };
      }
      return { error: "", normalized };
    }

    if (!raw && required) {
      return { error: FIELD_REQUIRED_MESSAGE, normalized: "" };
    }

    return { error: "", normalized: raw };
  }

  function createFieldComponent(options) {
    const root = document.createElement("div");
    root.className = "form-field";
    if (options && options.className) {
      root.classList.add(options.className);
    }

    const labelNode = document.createElement(options && options.controlId ? "label" : "div");
    labelNode.className = "form-field-label";
    if (options && options.controlId) {
      labelNode.htmlFor = options.controlId;
    }

    if (!options || !options.hideLabel) {
      const labelText = document.createElement("span");
      labelText.textContent = text(options && options.label);
      labelNode.appendChild(labelText);
    } else if (options && options.control && text(options.label)) {
      options.control.setAttribute("aria-label", text(options.label));
    }

    if (options && options.control) {
      labelNode.appendChild(options.control);
    }

    root.appendChild(labelNode);

    let statusNode = null;
    if (!options || options.includeStatus !== false) {
      const statusSlot = document.createElement("div");
      statusSlot.className = "form-field-status-slot";

      statusNode = document.createElement("p");
      statusNode.className = "form-field-status meta-text is-error";
      statusNode.hidden = true;
      if (options && options.errorId) {
        statusNode.id = options.errorId;
      }

      statusSlot.appendChild(statusNode);
      root.appendChild(statusSlot);
    }

    return { root, statusNode, labelNode };
  }

  window.huuperPublicFields = {
    createFieldComponent,
    validateFieldOnAdvance,
    normalizePhoneDisplay,
  };
})();
