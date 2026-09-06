// Public retreat page. All user-facing strings come from
// window.appCopy.retreats.public (frontend/skeleton/copy/retreats.js) — never
// hardcode copy in here. Retreat content itself (intro, faq, lists, ...) comes
// from the record's `data` JSON and is rendered verbatim.
(() => {
  const publicFields = window.appPublicFields;
  const auth = window.appAuth;

  const loadingNode = document.querySelector("#retreat-loading");
  const errorNode = document.querySelector("#retreat-error");
  const bodyNode = document.querySelector("#retreat-body");

  const navLinksNode = document.querySelector("#retreat-nav-links");

  const kickerNode = document.querySelector("#retreat-kicker");
  const titleNode = document.querySelector("#retreat-title");
  const leadNode = document.querySelector("#retreat-lead");

  const statStripNode = document.querySelector("#retreat-stat-strip");

  const aboutHeadingNode = document.querySelector("#retreat-about-heading");
  const introNode = document.querySelector("#retreat-intro");
  const highlightsNode = document.querySelector("#retreat-highlights");

  const gallerySectionNode = document.querySelector("#retreat-gallery-section");
  const galleryNode = document.querySelector("#retreat-gallery");

  const placeSectionNode = document.querySelector("#luogo");
  const placeHeadingNode = document.querySelector("#retreat-place-heading");
  const locationBodyNode = document.querySelector("#retreat-location-body");
  const arrivalNode = document.querySelector("#retreat-arrival");

  const infoSectionNode = document.querySelector("#info");
  const packingListNode = document.querySelector("#retreat-packing-list");
  const includedListNode = document.querySelector("#retreat-included-list");
  const notIncludedListNode = document.querySelector("#retreat-not-included-list");

  const deadlineNode = document.querySelector("#retreat-register-deadline");
  const signupHeadingNode = document.querySelector("#retreat-signup-heading");
  const signupBodyNode = document.querySelector("#retreat-signup-body");
  const emailField = document.querySelector("#retreat-email");
  const emailErrorNode = document.querySelector("#retreat-email-error");
  const form = document.querySelector("#retreat-register-form");
  const statusNode = document.querySelector("#retreat-register-status");
  const submitButton = document.querySelector("#retreat-register-submit");
  const contactNode = document.querySelector("#retreat-contact");

  const startNode = document.querySelector("#retreat-start");
  const startButton = document.querySelector("#retreat-start-button");
  const choiceNode = document.querySelector("#retreat-choice");
  const choiceQuestionNode = document.querySelector("#retreat-choice-question");
  const choiceOrNode = document.querySelector("#retreat-choice-or");
  const choiceMemberButton = document.querySelector("#retreat-choice-member");
  const choiceGuestButton = document.querySelector("#retreat-choice-guest");
  const ctasNode = document.querySelector(".retreat-signup-ctas");

  const fieldsLoadingNode = document.querySelector("#retreat-fields-loading");
  const guestFieldsNode = document.querySelector("#retreat-guest-fields");

  const signupCardNode = document.querySelector("#retreat-signup-card");
  const priceListNode = document.querySelector("#retreat-price-list");

  const faqSectionNode = document.querySelector("#faq");
  const faqNode = document.querySelector("#retreat-faq");

  if (!loadingNode || !errorNode || !bodyNode || !form || !submitButton || !auth) {
    return;
  }

  function copyRoot() {
    return (window.appCopy && window.appCopy.retreats && window.appCopy.retreats.public) || {};
  }

  // Reads a dotted path out of the copy object, e.g. text("stats.when").
  function text(path) {
    const parts = String(path || "").split(".");
    let node = copyRoot();
    for (const part of parts) {
      if (!node || typeof node !== "object") return "";
      node = node[part];
    }
    return typeof node === "string" ? node : "";
  }

  function fill(template, values) {
    return String(template || "").replace(/\{(\w+)\}/g, (match, key) => (
      Object.prototype.hasOwnProperty.call(values, key) ? String(values[key]) : match
    ));
  }

  // Populates every static label declared in the markup as data-copy="path".
  function applyStaticCopy() {
    document.querySelectorAll("[data-copy]").forEach((node) => {
      node.textContent = text(node.getAttribute("data-copy"));
    });
  }

  function escapeHTML(value) {
    return String(value == null ? "" : value).replace(/[&<>"']/g, (ch) => (
      { "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[ch]
    ));
  }

  function slugFromURL() {
    const segments = window.location.pathname.split("/").filter(Boolean);
    return segments.length >= 2 ? segments[1] : "";
  }

  function dateLocale() {
    return text("dateLocale") || undefined;
  }

  function formatDate(raw) {
    if (!raw) return "";
    const date = new Date(raw);
    if (Number.isNaN(date.getTime())) return "";
    return date.toLocaleDateString(dateLocale(), { day: "numeric", month: "long", year: "numeric" });
  }

  // Compact range: "8 – 11 October 2026" rather than repeating the month
  // on both sides of the dash.
  function formatDateRange(start, end) {
    const startDate = start ? new Date(start) : null;
    const endDate = end ? new Date(end) : null;
    const startValid = startDate && !Number.isNaN(startDate.getTime());
    const endValid = endDate && !Number.isNaN(endDate.getTime());
    if (!startValid) return "";
    if (!endValid || startDate.toDateString() === endDate.toDateString()) {
      return formatDate(start);
    }
    const sameMonth = startDate.getMonth() === endDate.getMonth()
      && startDate.getFullYear() === endDate.getFullYear();
    if (sameMonth) {
      const monthYear = endDate.toLocaleDateString(dateLocale(), { month: "long", year: "numeric" });
      return `${startDate.getDate()} – ${endDate.getDate()} ${monthYear}`;
    }
    return `${formatDate(start)} – ${formatDate(end)}`;
  }

  // toFixed() always formats with a dot, whatever the page language. The rest
  // of the page already formats dates through dateLocale(); amounts follow.
  // Only the number is localised: the trailing " €" keeps the suffix position
  // the page already uses (it-IT currency style would move it to the front).
  function formatMoney(cents) {
    const n = Number(cents || 0);
    if (!n) return "";
    const amount = (n / 100).toLocaleString(dateLocale(), {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    });
    return amount + " €";
  }

  function durationLabel(start, end) {
    if (!start || !end) return "";
    const startDate = new Date(start);
    const endDate = new Date(end);
    if (Number.isNaN(startDate.getTime()) || Number.isNaN(endDate.getTime())) return "";
    const days = Math.max(1, Math.round((endDate - startDate) / 86400000) + 1);
    const nights = Math.max(0, days - 1);
    const dayWord = days === 1 ? text("duration.day") : text("duration.days");
    if (nights === 0) return `${days} ${dayWord}`;
    const nightWord = nights === 1 ? text("duration.night") : text("duration.nights");
    return `${days} ${dayWord} · ${nights} ${nightWord}`;
  }

  function galleryURL(retreat, filename) {
    return "/api/files/retreats/" + encodeURIComponent(retreat.id) + "/" + encodeURIComponent(filename);
  }

  function showError(message) {
    loadingNode.hidden = true;
    errorNode.textContent = message;
    errorNode.hidden = false;
  }

  function renderNav(sections) {
    const links = [
      ["evento", text("nav.event"), sections.evento],
      ["luogo", text("nav.place"), sections.luogo],
      ["info", text("nav.info"), sections.info],
      ["faq", text("nav.faq"), sections.faq],
    ].filter(([, , visible]) => visible);
    if (links.length === 0) {
      navLinksNode.innerHTML = "";
      return;
    }
    navLinksNode.innerHTML = links.map(([id, label]) => (
      `<a href="#${id}" data-scroll-target="#${id}">${escapeHTML(label)}</a>`
    )).join("");
  }

  function renderStatStrip(retreat, data) {
    const stats = [];
    const when = formatDateRange(retreat.start_date, retreat.end_date);
    if (when) stats.push([text("stats.when"), when, data.meeting_time_note || ""]);
    if (retreat.location) stats.push([text("stats.where"), retreat.location, ""]);
    const duration = durationLabel(retreat.start_date, retreat.end_date);
    if (duration) stats.push([text("stats.duration"), duration, ""]);
    if (data.price_cents) stats.push([text("stats.fee"), formatMoney(data.price_cents), ""]);

    if (stats.length === 0) {
      statStripNode.hidden = true;
      return;
    }
    statStripNode.hidden = false;
    statStripNode.innerHTML = stats.map(([label, value, note]) => (
      `<div class="retreat-stat">
        <div class="retreat-stat-label">${escapeHTML(label)}</div>
        <div class="retreat-stat-value">${escapeHTML(value)}</div>
        ${note ? `<div class="retreat-stat-note">${escapeHTML(note)}</div>` : ""}
      </div>`
    )).join("");
  }

  function renderHighlights(highlights) {
    if (!Array.isArray(highlights) || highlights.length === 0) {
      highlightsNode.innerHTML = "";
      return;
    }
    highlightsNode.innerHTML = highlights.map((item) => (
      `<div class="retreat-feat-cell">
        <div class="retreat-feat-title">${escapeHTML(item && item.title)}</div>
        <div class="retreat-feat-body">${escapeHTML(item && item.text)}</div>
      </div>`
    )).join("");
  }

  function renderGallery(retreat) {
    const files = Array.isArray(retreat.gallery) ? retreat.gallery.filter(Boolean) : [];
    if (files.length === 0) {
      gallerySectionNode.hidden = true;
      galleryNode.innerHTML = "";
      return false;
    }
    gallerySectionNode.hidden = false;
    if (files.length === 1) {
      galleryNode.innerHTML = `<img class="retreat-gallery-image" src="${escapeHTML(galleryURL(retreat, files[0]))}" alt="" loading="lazy" />`;
      return true;
    }
    const tall = files[0];
    const rest = files.slice(1, 3);
    galleryNode.innerHTML = (
      `<div class="retreat-gallery-tall">
        <img class="retreat-gallery-image" src="${escapeHTML(galleryURL(retreat, tall))}" alt="" loading="lazy" />
      </div>
      <div class="retreat-gallery-col">
        ${rest.map((filename) => (
          `<img class="retreat-gallery-image" src="${escapeHTML(galleryURL(retreat, filename))}" alt="" loading="lazy" />`
        )).join("")}
      </div>`
    );
    return true;
  }

  function renderArrival(arrival) {
    const entries = [
      [text("place.arrivalCar"), arrival && arrival.auto],
      [text("place.arrivalTrain"), arrival && arrival.train],
      [text("place.arrivalCarpooling"), arrival && arrival.carpooling],
    ].filter(([, value]) => value);
    if (entries.length === 0) {
      arrivalNode.innerHTML = "";
      return false;
    }
    arrivalNode.innerHTML = entries.map(([label, value]) => (
      `<div class="retreat-arrival-row">
        <span class="retreat-arrival-key">${escapeHTML(label)}</span>
        <span class="retreat-arrival-val">${escapeHTML(value)}</span>
      </div>`
    )).join("");
    return true;
  }

  function renderListBlock(node, title, items, mutedItems) {
    if (!Array.isArray(items) || items.length === 0) {
      node.hidden = true;
      node.innerHTML = "";
      return false;
    }
    const itemClass = mutedItems ? "retreat-info-item retreat-info-item-muted" : "retreat-info-item";
    node.hidden = false;
    node.innerHTML = `<div class="retreat-info-column-title">${escapeHTML(title)}</div><ul class="retreat-info-item-list">${
      items.map((item) => `<li class="${itemClass}">${escapeHTML(item)}</li>`).join("")
    }</ul>`;
    return true;
  }

  function renderFAQ(faq) {
    if (!Array.isArray(faq) || faq.length === 0) {
      faqSectionNode.hidden = true;
      return;
    }
    faqSectionNode.hidden = false;
    faqNode.innerHTML = faq.map((item) => (
      `<div class="retreat-faq-cell">
        <p class="retreat-faq-question">${escapeHTML(item && item.question)}</p>
        <p class="retreat-faq-answer">${escapeHTML(item && item.answer)}</p>
      </div>`
    )).join("");
  }

  // The seats counter is deliberately not rendered: how many spots are left is
  // organiser-side information, not something the public page advertises.
  function renderSignupCard(data) {
    let hasAny = false;

    // { label, amount, note } — the note stays a separate field so a long
    // suffix can never push the amount out of its column.
    const priceRows = [];
    if (data.price_cents) {
      priceRows.push({ label: text("signup.feeRow"), amount: formatMoney(data.price_cents) });
    }
    if (data.deposit_cents) {
      priceRows.push({
        label: text("signup.depositRow"),
        amount: formatMoney(data.deposit_cents),
        note: text("signup.depositSuffix"),
      });
      if (data.price_cents && data.price_cents > data.deposit_cents) {
        const balance = data.price_cents - data.deposit_cents;
        priceRows.push({
          label: text("signup.balanceRow"),
          amount: formatMoney(balance),
          note: text("signup.balanceSuffix"),
        });
      }
    }
    if (data.min_age_note) {
      priceRows.push({ label: text("signup.minAgeRow"), amount: data.min_age_note });
    }
    if (priceRows.length > 0) {
      hasAny = true;
      priceListNode.hidden = false;
      priceListNode.innerHTML = priceRows.map((row) => (
        `<div class="retreat-price-row">
          <div class="retreat-price-row-head">
            <span class="retreat-price-label">${escapeHTML(row.label)}</span>
            <span class="retreat-price-amount">${escapeHTML(row.amount)}</span>
          </div>
          ${row.note ? `<div class="retreat-price-note">${escapeHTML(row.note)}</div>` : ""}
        </div>`
      )).join("");
    } else {
      priceListNode.hidden = true;
      priceListNode.innerHTML = "";
    }

    hasPrices = hasAny;
    signupCardNode.hidden = !hasAny;
  }

  function renderSignupBody(data) {
    const parts = [];
    parts.push(text("signup.capacityNote"));
    if (data.deposit_cents) {
      parts.push(text("signup.depositNote"));
    }
    if (parts.length === 0) {
      parts.push(text("signup.genericNote"));
    }
    defaultSignupBody = parts.filter(Boolean).join(" ");
    signupBodyNode.textContent = defaultSignupBody;
  }

  // Guests (non-members) submit a pre-registration, so we collect what an
  // organiser needs to call them back. Members are already on file: they only
  // confirm. Field keys match the shared validator in core/public-fields.js
  // (full_name / mobile have dedicated rules there) — see also onboarding.
  const GUEST_FIELDS = [
    { key: "full_name", type: "text", label: "fullName", placeholder: "fullNamePlaceholder", autocomplete: "name" },
    { key: "birth_year", type: "number", label: "birthYear", placeholder: "birthYearPlaceholder" },
    { key: "provenance", type: "text", label: "provenance", placeholder: "provenancePlaceholder", autocomplete: "address-level2" },
    { key: "mobile", type: "phone", label: "mobile", placeholder: "mobilePlaceholder", hint: "mobileHint", autocomplete: "tel" },
  ];

  const guestControls = {};

  function guestText(key) {
    return text("signup.guestFields." + key);
  }

  function minBirthYear() { return 1900; }
  function maxBirthYear() { return new Date().getFullYear() - 16; }

  function validateBirthYear(raw) {
    const value = String(raw || "").trim();
    if (!/^\d{4}$/.test(value)) return "";
    const year = Number(value);
    if (year < minBirthYear() || year > maxBirthYear()) return "";
    return value;
  }

  let guestFieldsBuilt = false;

  // Built once, up front: rebuilding them on click made the fields flash in.
  function buildGuestFields() {
    if (!publicFields || !guestFieldsNode || guestFieldsBuilt) return;
    guestFieldsBuilt = true;

    GUEST_FIELDS.forEach((field) => {
      const controlId = "retreat-" + field.key.replace(/_/g, "-");
      const input = document.createElement("input");
      input.id = controlId;
      input.name = field.key;
      input.type = field.type === "phone" ? "tel" : (field.type === "number" ? "text" : "text");
      if (field.type === "number") {
        input.inputMode = "numeric";
      }
      input.placeholder = guestText(field.placeholder);
      input.required = true;
      if (field.autocomplete) input.autocomplete = field.autocomplete;

      const component = publicFields.createFieldComponent({
        label: guestText(field.label),
        control: input,
        controlId,
        errorId: controlId + "-error",
      });

      if (field.hint) {
        const hint = document.createElement("p");
        hint.className = "retreat-field-hint meta-text";
        hint.textContent = guestText(field.hint);
        component.root.appendChild(hint);
      }

      guestFieldsNode.appendChild(component.root);
      guestControls[field.key] = { input, statusNode: component.statusNode, field };

      input.addEventListener("input", () => {
        if (field.type === "phone" && publicFields.normalizePhoneDisplay) {
          const caretAtEnd = input.selectionStart === input.value.length;
          const display = publicFields.normalizePhoneDisplay(input.value);
          if (display !== input.value) {
            input.value = display;
            if (caretAtEnd) input.setSelectionRange(display.length, display.length);
          }
        }
        if (component.statusNode) {
          component.statusNode.hidden = true;
          component.statusNode.textContent = "";
        }
      });
    });
  }

  // Returns the collected values, or null if any field is invalid (errors are
  // rendered in place, next to the offending input).
  function collectGuestFields() {
    let firstInvalid = null;
    const values = {};

    GUEST_FIELDS.forEach((field) => {
      const control = guestControls[field.key];
      if (!control) return;

      const outcome = publicFields.validateFieldOnAdvance(
        { key: field.key, type: field.type, required: true },
        control.input.value,
      );

      let error = outcome.error;
      let normalized = outcome.normalized;

      if (!error && field.key === "birth_year") {
        normalized = validateBirthYear(control.input.value);
        if (!normalized) error = guestText("birthYearError");
      }

      if (error) {
        if (control.statusNode) {
          control.statusNode.textContent = error;
          control.statusNode.hidden = false;
        }
        if (!firstInvalid) firstInvalid = control.input;
        return;
      }
      values[field.key] = normalized;
    });

    if (firstInvalid) {
      firstInvalid.focus();
      return null;
    }
    return values;
  }

  // Three states for the panel, all sharing one form so the price summary
  // always sits above whichever action is offered:
  //   choice  — visitor hasn't said who they are yet
  //   member  — logged in: one button, nothing to fill in
  //   guest   — not a member: the pre-registration fields
  let hasPrices = false;
  let defaultSignupBody = "";

  function hideAllActions() {
    fieldsLoadingNode.hidden = true;
    startNode.hidden = true;
    choiceNode.hidden = true;
    guestFieldsNode.hidden = true;
    ctasNode.hidden = true;
    signupCardNode.hidden = true;
  }

  // Each step replaces the one before it in the same slot, with a short
  // spinner in between: the price summary is what you decide on, once decided
  // it makes way for the next question rather than pushing it off-screen.
  function transitionTo(render) {
    hideAllActions();
    fieldsLoadingNode.hidden = false;
    window.setTimeout(() => {
      fieldsLoadingNode.hidden = true;
      render();
    }, 300);
  }

  // Step 1: the price summary and a single call to action.
  function showStart() {
    hideAllActions();
    signupBodyNode.textContent = defaultSignupBody;
    signupCardNode.hidden = !hasPrices;
    startButton.textContent = text("signup.choice.startCta");
    startNode.hidden = false;
  }

  // Step 2: which door are you coming through?
  function showChoice() {
    hideAllActions();
    signupBodyNode.textContent = defaultSignupBody;
    choiceQuestionNode.textContent = text("signup.choice.question");
    choiceMemberButton.textContent = text("signup.choice.memberCta");
    choiceOrNode.textContent = text("signup.choice.or");
    choiceGuestButton.textContent = text("signup.choice.guestCta");
    choiceNode.hidden = false;
  }

  function showMemberForm(data) {
    hideAllActions();
    signupCardNode.hidden = !hasPrices;
    ctasNode.hidden = false;
    submitButton.textContent = data.deposit_cents > 0
      ? text("signup.submitMemberDeposit")
      : text("signup.submitMember");
  }

  // Step 3: the pre-registration fields.
  function showGuestForm() {
    hideAllActions();
    signupBodyNode.textContent = text("signup.guestNote") || defaultSignupBody;
    buildGuestFields();
    guestFieldsNode.hidden = false;
    ctasNode.hidden = false;
    submitButton.textContent = text("signup.submitGuest");
  }

  applyStaticCopy();

  const slug = slugFromURL();
  if (!slug) {
    showError(text("notFound"));
    return;
  }

  // A token in localStorage isn't proof of a live session. Without this check
  // an expired one still renders the member path, and the submit then fails
  // as a guest with "fill in the required fields" — over a form that shows
  // no fields at all.
  let session = auth.read();
  let isMember = !!(session && session.token);

  const sessionChecked = isMember
    ? auth.refresh()
      .then((next) => {
        session = next;
        isMember = !!(next && next.token);
      })
      .catch(() => {
        auth.clear();
        session = null;
        isMember = false;
      })
    : Promise.resolve();

  sessionChecked.then(() => fetch("/api/public/retreats/" + encodeURIComponent(slug)))
    .then((response) => {
      if (!response.ok) throw new Error("not_found");
      return response.json();
    })
    .then((payload) => {
      const retreat = payload.retreat || {};
      const data = retreat.data || {};
      // The seat count never leaves the server: the API only reports whether
      // the retreat can still take registrations.
      const full = !!payload.full;

      // hero
      const kickerBits = [];
      const kicker = retreat.tagline || retreat.title || text("hero.kickerFallback");
      kickerBits.push(`<span class="retreat-kicker-pill">${escapeHTML(kicker)}</span>`);
      const dateRegion = [formatDateRange(retreat.start_date, retreat.end_date), data.region].filter(Boolean).join(" · ");
      if (dateRegion) {
        kickerBits.push(`<span class="retreat-kicker-meta">${escapeHTML(dateRegion.toUpperCase())}</span>`);
      }
      kickerNode.innerHTML = kickerBits.join("");

      titleNode.textContent = retreat.title || "";
      leadNode.textContent = data.lead || "";
      leadNode.hidden = !data.lead;

      renderStatStrip(retreat, data);

      // event section
      const hasHighlights = Array.isArray(data.highlights) && data.highlights.length > 0;
      const eventoVisible = !!(data.intro || hasHighlights);
      aboutHeadingNode.textContent = text("event.heading");
      aboutHeadingNode.hidden = !eventoVisible;
      introNode.textContent = data.intro || "";
      introNode.hidden = !data.intro;
      renderHighlights(data.highlights);
      document.querySelector("#evento").hidden = !eventoVisible;

      renderGallery(retreat);

      // place section
      const arrivalVisible = renderArrival(data.arrival);
      const placeVisible = !!(data.location_body || arrivalVisible);
      placeHeadingNode.textContent = data.place_name || retreat.location || "";
      locationBodyNode.textContent = data.location_body || "";
      locationBodyNode.hidden = !data.location_body;
      placeSectionNode.hidden = !placeVisible;

      // practical info
      const hasPacking = renderListBlock(packingListNode, text("info.packing"), data.packing_list, false);
      const hasIncluded = renderListBlock(includedListNode, text("info.included"), data.included, false);
      const hasNotIncluded = renderListBlock(notIncludedListNode, text("info.notIncluded"), data.not_included, true);
      infoSectionNode.hidden = !(hasPacking || hasIncluded || hasNotIncluded);

      // registration
      if (data.registration_deadline) {
        deadlineNode.textContent = `${text("signup.deadlinePrefix")} ${formatDate(data.registration_deadline).toUpperCase()}`;
        deadlineNode.hidden = false;
      } else {
        deadlineNode.hidden = true;
      }
      signupHeadingNode.textContent = retreat.title || "";
      renderSignupBody(data);
      renderSignupCard(data);

      if (data.contact_email) {
        contactNode.href = "mailto:" + data.contact_email;
        contactNode.hidden = false;
      } else {
        contactNode.hidden = true;
      }

      renderFAQ(data.faq);

      renderNav({
        evento: eventoVisible,
        luogo: placeVisible,
        info: !infoSectionNode.hidden,
        faq: !faqSectionNode.hidden,
      });

      // registration state
      if (!retreat.active) {
        form.hidden = true;
        statusNode.textContent = text("signup.notOpen");
        statusNode.hidden = false;
      } else if (full) {
        form.hidden = true;
        statusNode.textContent = text("signup.full");
        statusNode.hidden = false;
      } else if (isMember) {
        buildGuestFields();
        showMemberForm(data);
      } else {
        showStart();
      }

      loadingNode.hidden = true;
      bodyNode.hidden = false;

      // The panel is rendered after the fetch, so the browser has nothing to
      // scroll to when it first handles the #iscrizione anchor.
      if (window.location.hash === "#iscrizione") {
        const target = document.querySelector("#iscrizione");
        if (target) {
          window.requestAnimationFrame(() => {
            target.scrollIntoView({ behavior: "auto", block: "start" });
          });
        }
      }
    })
    .catch(() => {
      showError(text("notFound"));
    });

  document.addEventListener("click", (event) => {
    const target = event.target.closest && event.target.closest("[data-scroll-target]");
    if (!target) return;
    const dest = document.querySelector(target.getAttribute("data-scroll-target"));
    if (!dest) return;
    event.preventDefault();
    dest.scrollIntoView({ behavior: "smooth", block: "start" });
  });

  if (emailField) {
    emailField.addEventListener("input", () => {
      emailErrorNode.hidden = true;
      emailErrorNode.textContent = "";
    });
  }

  if (startButton) {
    startButton.addEventListener("click", () => transitionTo(showChoice));
  }

  if (choiceMemberButton) {
    choiceMemberButton.addEventListener("click", () => {
      // Comes back to this page once logged in (see publicReturnPrefixes in
      // core/auth.js), instead of landing on /me/.
      auth.redirectToLogin(window.location.pathname + window.location.search + "#iscrizione");
    });
  }

  if (choiceGuestButton) {
    choiceGuestButton.addEventListener("click", () => {
      transitionTo(() => {
        showGuestForm();
        if (emailField) emailField.focus();
      });
    });
  }

  form.addEventListener("submit", async (event) => {
    event.preventDefault();
    statusNode.hidden = true;
    statusNode.textContent = "";

    let email = "";
    let registrationData = { source: "retreat_page" };

    if (isMember) {
      email = session.model && session.model.email ? session.model.email : "";
    } else if (publicFields && emailField) {
      const fieldAdvance = publicFields.validateFieldOnAdvance({ key: "email", type: "email" }, emailField.value);
      if (fieldAdvance.error) {
        emailErrorNode.textContent = fieldAdvance.error;
        emailErrorNode.hidden = false;
        return;
      }
      email = fieldAdvance.normalized || emailField.value.trim().toLowerCase();

      const guestValues = collectGuestFields();
      if (!guestValues) return;
      registrationData = Object.assign(registrationData, guestValues);
    }

    // Registering hits Stripe and the mail server, so the wait is real and
    // visible. Without this the button just goes dead and people click again.
    submitButton.disabled = true;
    submitButton.classList.add("is-busy");
    try {
      const payload = await auth.apiFetch("/api/public/retreats/" + encodeURIComponent(slug) + "/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, data: registrationData }),
      });

      if (payload.checkout_url) {
        window.location.href = payload.checkout_url;
        return;
      }

      // Nothing left to pay: land on a page that says so, instead of leaving
      // the reader in front of the form they just filled in.
      const outcome = isMember ? "member" : "guest";
      window.location.href = "/retreat-registered/?status=" + outcome +
        "&retreat=" + encodeURIComponent(slug);
      return;
    } catch (error) {
      const code = (error && error.payload && error.payload.message) || "";
      const messages = {
        already_submitted: text("signup.errors.alreadySubmitted"),
        retreat_closed: text("signup.errors.closed"),
        retreat_full: text("signup.errors.full"),
        invalid_email: text("signup.errors.invalidEmail"),
        missing_guest_fields: text("signup.errors.missingGuestFields"),
      };
      statusNode.textContent = messages[code] || text("signup.errors.generic");
      statusNode.hidden = false;
      // Only here, not in a `finally`: both success paths navigate away, and
      // `finally` would run before the browser leaves, flashing the button
      // back to life on a form the reader is already done with.
      submitButton.disabled = false;
      submitButton.classList.remove("is-busy");
    }
  });
})();
