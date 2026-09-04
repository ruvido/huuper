window.appEventDetail = (() => {
  function tabLabel(label, count) {
    return `<span class="section-tab-text">${window.appListPage.escapeHTML(label)}</span><span class="section-tab-count">${window.appListPage.escapeHTML(String(count))}</span>`;
  }

  function eventsCopy() {
    return (window.appCopy && window.appCopy.events) || {};
  }

  function detailCopy() {
    return eventsCopy().detail || {};
  }

  function registrationEnabled(event) {
    if (!event || typeof event.registration === "undefined") return false;
    return event.registration !== false && event.registration !== null;
  }

  function isPastEvent(event) {
    const raw = String((event && event.start_date) || "").trim();
    if (!raw) return false;
    const parsed = new Date(raw);
    if (Number.isNaN(parsed.getTime())) return false;
    return parsed.getTime() < Date.now();
  }

  function currentUserID() {
    try {
      const raw = localStorage.getItem("app.auth");
      const auth = raw ? JSON.parse(raw) : null;
      return (auth && auth.model && auth.model.id) || "";
    } catch (_) {
      return "";
    }
  }

  function occurrenceDate(occurrence) {
    const raw = String((occurrence && occurrence.date) || "").trim();
    if (!raw) return "";
    return raw.slice(0, 10);
  }

  function renderOccurrences(occurrences, canEdit) {
    const c = detailCopy();
    const esc = window.appListPage.escapeHTML;
    const cad = window.appEventsCadence || {};
    const items = Array.isArray(occurrences) ? occurrences : [];
    if (items.length === 0) return "";
    const rows = items.map((occurrence) => {
      const date = occurrenceDate(occurrence);
      const cancelled = occurrence && occurrence.cancelled === true;
      const past = occurrence && occurrence.past === true;
      const dateLabel = cad.shortDate ? cad.shortDate(date) : date;
      const status = cancelled ? (c.cancelledLabel || "Cancelled") : (past ? (c.pastLabel || "Past") : "");
      const action = canEdit && !cancelled && !past
        ? `<button type="button" class="wizard-btn wizard-btn-outline" data-cancel-occurrence="${esc(date)}">${esc(c.cancelOccurrenceLabel || "Cancel this one")}</button>`
        : "";
      return `
        <li class="event-occurrence-row${cancelled ? " event-occurrence-row-cancelled" : ""}${past ? " event-occurrence-row-past" : ""}">
          <span class="event-occurrence-date">${esc(dateLabel || date)}</span>
          ${status ? `<span class="event-occurrence-status">${esc(status)}</span>` : ""}
          ${action ? `<span class="event-occurrence-action">${action}</span>` : ""}
        </li>
      `;
    }).join("");
    return `
      <h3 class="event-occurrences-title">${esc(c.occurrencesLabel || "Occurrences")}</h3>
      <ul class="event-occurrences-list">${rows}</ul>
    `;
  }

  async function postCancelOccurrence(scope, eventID, date) {
    return window.appAuth.apiFetch(`/api/${scope}/events/${encodeURIComponent(eventID)}/cancel-occurrence`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ date }),
    });
  }

  // Inserts (or returns the existing) host element used for the per-user
  // RSVP section (register/unregister buttons). Lives between the summary
  // card and the tabs section so it reads as the primary detail action.
  function ensureRsvpHost(summaryNode) {
    let host = document.getElementById("event-rsvp");
    if (host) return host;
    host = document.createElement("section");
    host.id = "event-rsvp";
    host.className = "event-rsvp";
    host.hidden = true;
    if (summaryNode && summaryNode.parentNode) {
      summaryNode.parentNode.insertBefore(host, summaryNode.nextSibling);
    }
    return host;
  }

  // Hosts the post-event attendance roster (admin/assistant) or read-only
  // status (attendee). Inserted just below the tabs so it doesn't collide
  // with the active-registration list.
  function ensureVisibilityHost(summaryNode) {
    let host = document.getElementById("event-visibility");
    if (host) return host;
    host = document.createElement("section");
    host.id = "event-visibility";
    host.className = "event-visibility";
    if (summaryNode && summaryNode.parentNode) {
      summaryNode.parentNode.insertBefore(host, summaryNode.nextSibling);
    }
    return host;
  }

  async function postSetActive(eventID, active) {
    return window.appAuth.apiFetch(`/api/admin/events/${encodeURIComponent(eventID)}/active`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ active }),
    });
  }

  function renderVisibilityToggle(summaryNode, event, eventID, reload) {
    const host = ensureVisibilityHost(summaryNode);
    const isActive = event.active !== false;
    const esc = window.appListPage.escapeHTML;
    const c = detailCopy();
    const statusLabel = isActive ? (c.publishedLabel || "Published") : (c.draftLabel || "Draft (admin only)");
    const actionLabel = isActive ? (c.unpublishAction || "Move to draft") : (c.publishAction || "Publish");
    host.innerHTML = `
      <article class="detail-card event-visibility-card">
        <strong>${esc(statusLabel)}</strong>
        <button type="button" class="wizard-btn wizard-btn-outline" data-toggle-active>${esc(actionLabel)}</button>
      </article>
    `;
    const button = host.querySelector("[data-toggle-active]");
    if (button) {
      button.addEventListener("click", async () => {
        button.disabled = true;
        try {
          await postSetActive(eventID, !isActive);
          await reload();
        } catch (_) {
          button.disabled = false;
        }
      });
    }
  }

  function ensureAttendanceHost(referenceNode) {
    let host = document.getElementById("event-attendance");
    if (host) return host;
    host = document.createElement("section");
    host.id = "event-attendance";
    host.className = "event-attendance";
    host.hidden = true;
    if (referenceNode && referenceNode.parentNode) {
      referenceNode.parentNode.insertBefore(host, referenceNode);
    }
    return host;
  }

  async function postRegister(scope, eventID) {
    return window.appAuth.apiFetch(`/api/${scope}/events/${encodeURIComponent(eventID)}/register`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({}),
    });
  }

  async function postUnregister(scope, eventID) {
    return window.appAuth.apiFetch(`/api/${scope}/events/${encodeURIComponent(eventID)}/unregister`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({}),
    });
  }

  async function postAttendance(scope, registrationID, attended) {
    return window.appAuth.apiFetch(`/api/${scope}/registrations/${encodeURIComponent(registrationID)}/attendance`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ attended }),
    });
  }

  function renderRsvpButtons(state) {
    const c = detailCopy();
    const esc = window.appListPage.escapeHTML;
    if (state.past) return "";
    if (!state.registrationEnabled) return "";
    if (state.registered) {
      return `
        <div class="event-rsvp-state">
          <p class="event-rsvp-status">${esc(c.registrationConfirmed || "")}</p>
          <button type="button" class="wizard-btn wizard-btn-outline" data-action="unregister">${esc(c.unregisterLabel || "")}</button>
        </div>
      `;
    }
    return `
      <div class="event-rsvp-state">
        <button type="button" class="wizard-btn wizard-btn-primary" data-action="register">${esc(c.registerLabel || "")}</button>
      </div>
    `;
  }

  function renderAttendanceRow(item, scope, esc) {
    const c = detailCopy();
    const attended = (item && item.data && typeof item.data.attended !== "undefined") ? item.data.attended : null;
    const attendedRaw = (item && typeof item.attended !== "undefined") ? item.attended : attended;
    const isPresent = attendedRaw === true;
    const isAbsent = attendedRaw === false;
    return `
      <li class="event-attendance-row" data-id="${esc(item.id || "")}">
        <span class="event-attendance-name">${esc(item.full_name || item.email || item.id || "")}</span>
        <span class="event-attendance-actions">
          <button type="button" class="wizard-btn wizard-btn-outline${isPresent ? " is-active" : ""}" data-attendance="present" data-scope="${esc(scope)}">${esc(c.markPresent || "Present")}</button>
          <button type="button" class="wizard-btn wizard-btn-outline${isAbsent ? " is-active" : ""}" data-attendance="absent" data-scope="${esc(scope)}">${esc(c.markAbsent || "Absent")}</button>
          <button type="button" class="wizard-btn wizard-btn-outline" data-attendance="clear" data-scope="${esc(scope)}">${esc(c.clearAttendance || "Clear")}</button>
        </span>
      </li>
    `;
  }

  function renderAttendanceSection(state, registrations) {
    const c = detailCopy();
    const esc = window.appListPage.escapeHTML;
    if (!state.past) return "";
    if (!state.registrationEnabled) return "";
    if (state.canManage) {
      const rows = registrations.map((reg) => renderAttendanceRow(reg, state.scope, esc)).join("");
      return `
        <h3 class="event-attendance-title">${esc(c.pastAttendanceTitle || c.attendanceTitle || "")}</h3>
        <ul class="event-attendance-list">${rows}</ul>
      `;
    }
    // Attendee read-only view — find their own registration if present.
    const myID = currentUserID();
    const mine = registrations.find((reg) => reg && reg.user_id === myID);
    if (!mine) return "";
    const attended = mine && (mine.attended === true || (mine.data && mine.data.attended === true));
    const missed = mine && (mine.attended === false || (mine.data && mine.data.attended === false));
    if (!attended && !missed) return "";
    const label = attended ? (c.attendedLabel || "") : (c.missedLabel || "");
    return `<p class="event-attendance-readonly">${esc(label)}</p>`;
  }

  function bindRsvpHandlers(host, state, eventID, refresh) {
    host.querySelectorAll("button[data-action]").forEach((btn) => {
      btn.addEventListener("click", async () => {
        const action = btn.getAttribute("data-action");
        const previousDisabled = btn.disabled;
        btn.disabled = true;
        try {
          if (action === "register") {
            await postRegister(state.scope, eventID);
          } else if (action === "unregister") {
            await postUnregister(state.scope, eventID);
          }
          if (typeof refresh === "function") refresh();
        } catch (_err) {
          btn.disabled = previousDisabled;
          const c = detailCopy();
          const msg = action === "register" ? (c.registerError || "") : (c.unregisterError || "");
          if (msg) {
            const note = document.createElement("p");
            note.className = "event-rsvp-error";
            note.textContent = msg;
            host.appendChild(note);
          }
        }
      });
    });
  }

  function bindAttendanceHandlers(host, refresh) {
    host.querySelectorAll("button[data-attendance]").forEach((btn) => {
      btn.addEventListener("click", async () => {
        const row = btn.closest("[data-id]");
        const id = row ? row.getAttribute("data-id") : "";
        if (!id) return;
        const mode = btn.getAttribute("data-attendance");
        const scope = btn.getAttribute("data-scope") || "me";
        let payload = null;
        if (mode === "present") payload = true;
        else if (mode === "absent") payload = false;
        else payload = null;
        const previousDisabled = btn.disabled;
        btn.disabled = true;
        try {
          await postAttendance(scope, id, payload);
          if (typeof refresh === "function") refresh();
        } catch (_err) {
          btn.disabled = previousDisabled;
          const c = detailCopy();
          if (c.attendanceError) {
            const note = document.createElement("p");
            note.className = "event-attendance-error";
            note.textContent = c.attendanceError;
            host.appendChild(note);
          }
        }
      });
    });
  }

  function bindOccurrenceHandlers(host, state, eventID, refresh) {
    host.querySelectorAll("button[data-cancel-occurrence]").forEach((btn) => {
      btn.addEventListener("click", () => {
        const date = btn.getAttribute("data-cancel-occurrence") || "";
        if (!date) return;
        const c = detailCopy();
        const dialog = c.cancelOccurrenceConfirm || {};
        window.appActionSheet.open({
          title: dialog.title || "Cancel this occurrence?",
          actions: [],
          footerAction: {
            label: dialog.confirmLabel || c.cancelOccurrenceLabel || "Cancel this one",
            tone: "danger",
            onSelect: async () => {
              await postCancelOccurrence(state.scope, eventID, date);
              if (typeof refresh === "function") refresh();
            },
          },
        });
      });
    });
  }

  function init(config) {
    const summaryNode = document.querySelector("#event-summary");
    const topbarTitleNode = document.querySelector(".top-bar-title");
    const tabsNode = document.querySelector("#event-tabs");
    const tabUsersNode = document.querySelector("#event-tab-users");
    const tabGuestsNode = document.querySelector("#event-tab-guests");
    const tabPendingNode = document.querySelector("#event-tab-pending");
    const registrationsNode = document.querySelector("#event-registrations");
    const occurrencesNode = document.querySelector("#event-occurrences");
    const cancelledBlockNode = document.querySelector("#event-cancelled-block");
    const cancelledToggleNode = document.querySelector("#event-cancelled-toggle");
    const cancelledLabelNode = document.querySelector("#event-cancelled-label");
    const cancelledCountNode = document.querySelector("#event-cancelled-count");
    const cancelledListNode = document.querySelector("#event-cancelled-list");
    if (!summaryNode || !occurrencesNode || !tabsNode || !tabUsersNode || !tabGuestsNode || !tabPendingNode || !registrationsNode || !cancelledBlockNode || !cancelledToggleNode || !cancelledLabelNode || !cancelledCountNode || !cancelledListNode || !window.appAuth || !window.appListPage || !window.appEventSummary || !window.appListItem || !window.appActionSheet) {
      return;
    }

    const id = window.appListPage.queryParam("id");
    if (!id) {
      summaryNode.hidden = false;
      summaryNode.innerHTML = `<article class="detail-card"><strong>Missing event.</strong></article>`;
      return;
    }

    function load() {
      return window.appAuth.apiFetch(config.detailURL(id)).then(render).catch(() => {
        summaryNode.hidden = false;
        summaryNode.innerHTML = `<article class="detail-card"><strong>Event unavailable.</strong></article>`;
      });
    }

    function render(payload) {
      const event = payload.event || {};
      const location = window.appListPage.text(event.location);
      const duration = window.appListPage.text(((event.data || {}).duration));
      const description = window.appListPage.text(((event.data || {}).description));
      const url = window.appListPage.text(event.url);
      if (topbarTitleNode && event.title) {
        topbarTitleNode.textContent = event.title;
      }
      summaryNode.hidden = false;
      summaryNode.innerHTML = window.appEventSummary.render(location, duration, description, url);

      const past = isPastEvent(event);
      const scope = config.scope || "me";
      if (scope === "admin") {
        renderVisibilityToggle(summaryNode, event, id, load);
      }
      const canManage = config.canManageRegistrations === true;
      const canEdit = payload.can_edit === true || config.canEdit === true;
      const registered = payload.registered === true;
      const state = { past, scope, canManage, registered, registrationEnabled: registrationEnabled(event) };

      const occurrencesHTML = renderOccurrences(payload.occurrences, canEdit);
      if (occurrencesHTML) {
        occurrencesNode.innerHTML = occurrencesHTML;
        occurrencesNode.hidden = false;
        bindOccurrenceHandlers(occurrencesNode, state, id, load);
      } else {
        occurrencesNode.innerHTML = "";
        occurrencesNode.hidden = true;
      }

      const rsvpHost = ensureRsvpHost(summaryNode);
      const rsvpHTML = renderRsvpButtons(state);
      if (rsvpHTML) {
        rsvpHost.innerHTML = rsvpHTML;
        rsvpHost.hidden = false;
        bindRsvpHandlers(rsvpHost, state, id, load);
      } else {
        rsvpHost.innerHTML = "";
        rsvpHost.hidden = true;
      }

      const allRegistrations = Array.isArray(payload.registrations) ? payload.registrations : [];
      const attendanceHost = ensureAttendanceHost(tabsNode);
      const attendanceHTML = renderAttendanceSection(state, allRegistrations);
      if (attendanceHTML) {
        attendanceHost.innerHTML = attendanceHTML;
        attendanceHost.hidden = false;
        if (canManage) bindAttendanceHandlers(attendanceHost, load);
      } else {
        attendanceHost.innerHTML = "";
        attendanceHost.hidden = true;
      }

      if (config.includeRegistrations && registrationsNode) {
        const items = Array.isArray(payload.registrations) ? payload.registrations : [];
        const pending = Array.isArray(payload.pending_registrations) ? payload.pending_registrations : [];
        const cancelled = Array.isArray(payload.cancelled_registrations) ? payload.cancelled_registrations : [];
        if (items.length > 0 || pending.length > 0 || cancelled.length > 0) {
          const users = items.filter((item) => item && item.is_user);
          const guests = items.filter((item) => !item || !item.is_user);
          const pendingItems = pending;
          const cancelledItems = cancelled;
          const allowMenu = config.canManageRegistrations === true;
          const showPending = config.showPending === true;
          const showCancelled = config.showCancelled === true;
          let cancelledRendered = false;
          tabsNode.hidden = false;
          tabsNode.style.display = "";
          registrationsNode.hidden = false;
          tabUsersNode.innerHTML = tabLabel("Users", users.length);
          tabGuestsNode.innerHTML = tabLabel("Guests", guests.length);
          tabPendingNode.innerHTML = tabLabel("Pending", pendingItems.length);
          if (showCancelled) {
            cancelledBlockNode.hidden = cancelledItems.length === 0;
            cancelledLabelNode.textContent = "Cancelled";
            cancelledCountNode.textContent = String(cancelledItems.length);
            cancelledToggleNode.setAttribute("aria-expanded", "false");
            cancelledToggleNode.classList.remove("event-secondary-toggle-open");
            cancelledListNode.hidden = true;
            cancelledListNode.style.display = "none";
            cancelledToggleNode.onclick = null;
          } else {
            cancelledBlockNode.hidden = true;
            cancelledBlockNode.style.display = "none";
            cancelledListNode.hidden = true;
            cancelledListNode.style.display = "none";
            cancelledToggleNode.onclick = null;
          }

          const renderItems = (collection) => {
            window.appListPage.renderList(registrationsNode, collection, (item) => {
              const href = item.is_user && item.user_id ? `/${config.scope || "me"}/user/?id=${encodeURIComponent(item.user_id)}` : "";
              return window.appListItem.renderAttendee(item, href, { menu: allowMenu });
            });
            bindRowLinks(registrationsNode);
            if (allowMenu) {
              bindMenus(registrationsNode);
            }
          };

          const renderCancelled = () => {
            window.appListPage.renderList(cancelledListNode, cancelledItems, (item) => {
              const href = item.is_user && item.user_id ? `/${config.scope || "me"}/user/?id=${encodeURIComponent(item.user_id)}` : "";
              return window.appListItem.renderAttendee(item, href, { menu: allowMenu });
            });
            bindRowLinks(cancelledListNode);
            if (allowMenu) {
              bindMenus(cancelledListNode);
            }
          };

          const showTab = (name) => {
            tabUsersNode.classList.toggle("section-tab-current", name === "users");
            tabGuestsNode.classList.toggle("section-tab-current", name === "guests");
            tabPendingNode.classList.toggle("section-tab-current", name === "pending");
            if (name === "users") {
              renderItems(users);
              return;
            }
            if (name === "guests") {
              renderItems(guests);
              return;
            }
            renderItems(pendingItems);
          };

          const hasUsers = users.length > 0;
          const hasGuests = guests.length > 0;
          const hasPending = showPending && pendingItems.length > 0;

          tabUsersNode.hidden = !hasUsers;
          tabGuestsNode.hidden = !hasGuests;
          tabPendingNode.hidden = !hasPending;
          tabUsersNode.style.display = hasUsers ? "" : "none";
          tabGuestsNode.style.display = hasGuests ? "" : "none";
          tabPendingNode.style.display = hasPending ? "" : "none";

          if (hasUsers) {
            showTab("users");
          } else if (hasGuests) {
            showTab("guests");
          } else if (hasPending) {
            showTab("pending");
          } else {
            window.appListPage.renderList(registrationsNode, [], () => "");
            tabsNode.hidden = true;
            tabsNode.style.display = "none";
          }

          tabUsersNode.addEventListener("click", () => showTab("users"));
          tabGuestsNode.addEventListener("click", () => showTab("guests"));
          if (hasPending) {
            tabPendingNode.addEventListener("click", () => showTab("pending"));
          }
          if (showCancelled && cancelledItems.length > 0) {
            cancelledToggleNode.onclick = () => {
              const expanded = cancelledToggleNode.getAttribute("aria-expanded") === "true";
              const next = !expanded;
              if (next && !cancelledRendered) {
                renderCancelled();
                cancelledRendered = true;
              }
              cancelledToggleNode.setAttribute("aria-expanded", next ? "true" : "false");
              cancelledToggleNode.classList.toggle("event-secondary-toggle-open", next);
              cancelledListNode.hidden = !next;
              cancelledListNode.style.display = next ? "" : "none";
              if (next) {
                cancelledBlockNode.scrollIntoView({ block: "start", behavior: "smooth" });
              }
            };
          }
        } else {
          tabsNode.hidden = true;
          cancelledBlockNode.hidden = true;
        }
      }
    }

    load();
  }

  function bindMenus(root) {
    root.querySelectorAll(".row-menu-trigger").forEach((button) => {
      if (button.dataset.bound === "true") {
        return;
      }
      button.dataset.bound = "true";
      button.addEventListener("click", (event) => {
        event.preventDefault();
        event.stopPropagation();
        let attendee = {};
        try {
          attendee = JSON.parse(button.dataset.attendee || "{}");
        } catch (_error) {
          attendee = {};
        }
        window.appActionSheet.open({
          actions: [
            {
              label: "Cancel",
              tone: "danger",
              onSelect: () => {
                document.dispatchEvent(new CustomEvent("app:attendee-cancel", { detail: attendee }));
              },
            },
          ],
        });
      });
    });
  }

  function bindRowLinks(root) {
    root.querySelectorAll(".list-item-linkable").forEach((row) => {
      if (row.dataset.linkBound === "true") {
        return;
      }
      row.dataset.linkBound = "true";
      row.addEventListener("click", (event) => {
        if (event.target.closest(".row-menu-trigger")) {
          return;
        }
        const href = row.getAttribute("data-href");
        if (href) {
          window.location.href = href;
        }
      });
    });
  }

  return { init };
})();
