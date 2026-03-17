window.huuperEventDetail = (() => {
  function tabLabel(label, count) {
    return `<span class="section-tab-text">${window.huuperListPage.escapeHTML(label)}</span><span class="section-tab-count">${window.huuperListPage.escapeHTML(String(count))}</span>`;
  }

  function init(config) {
    const summaryNode = document.querySelector("#event-summary");
    const topbarTitleNode = document.querySelector(".top-bar-title");
    const tabsNode = document.querySelector("#event-tabs");
    const tabUsersNode = document.querySelector("#event-tab-users");
    const tabGuestsNode = document.querySelector("#event-tab-guests");
    const tabPendingNode = document.querySelector("#event-tab-pending");
    const registrationsNode = document.querySelector("#event-registrations");
    const cancelledBlockNode = document.querySelector("#event-cancelled-block");
    const cancelledToggleNode = document.querySelector("#event-cancelled-toggle");
    const cancelledLabelNode = document.querySelector("#event-cancelled-label");
    const cancelledCountNode = document.querySelector("#event-cancelled-count");
    const cancelledListNode = document.querySelector("#event-cancelled-list");
    if (!summaryNode || !tabsNode || !tabUsersNode || !tabGuestsNode || !tabPendingNode || !registrationsNode || !cancelledBlockNode || !cancelledToggleNode || !cancelledLabelNode || !cancelledCountNode || !cancelledListNode || !window.huuperAuth || !window.huuperListPage || !window.huuperEventCard || !window.huuperUserCard || !window.huuperActionSheet) {
      return;
    }

    const id = window.huuperListPage.queryParam("id");
    if (!id) {
      summaryNode.hidden = false;
      summaryNode.innerHTML = `<article class="event-summary-card"><strong>Missing event.</strong></article>`;
      return;
    }

    window.huuperAuth.apiFetch(config.detailURL(id)).then((payload) => {
      const event = payload.event || {};
      const location = window.huuperListPage.text(((event.data || {}).location));
      const duration = window.huuperListPage.text(((event.data || {}).duration));
      if (topbarTitleNode && event.title) {
        topbarTitleNode.textContent = event.title;
      }
      summaryNode.hidden = false;
      summaryNode.innerHTML = window.huuperEventCard.summaryLines(location, duration);

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
            window.huuperListPage.renderList(registrationsNode, collection, (item) => {
              const href = item.is_user && item.user_id ? `/${config.scope || "me"}/user/?id=${encodeURIComponent(item.user_id)}` : "";
              return window.huuperUserCard.renderAttendee(item, href, { menu: allowMenu });
            });
            bindRowLinks(registrationsNode);
            if (allowMenu) {
              bindMenus(registrationsNode);
            }
          };

          const renderCancelled = () => {
            window.huuperListPage.renderList(cancelledListNode, cancelledItems, (item) => {
              const href = item.is_user && item.user_id ? `/${config.scope || "me"}/user/?id=${encodeURIComponent(item.user_id)}` : "";
              return window.huuperUserCard.renderAttendee(item, href, { menu: allowMenu });
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
            window.huuperListPage.renderList(registrationsNode, [], () => "");
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
    }).catch(() => {
      summaryNode.hidden = false;
      summaryNode.innerHTML = `<article class="event-summary-card"><strong>Event unavailable.</strong></article>`;
    });
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
        window.huuperActionSheet.open({
          actions: [
            {
              label: "Cancel",
              tone: "danger",
              onSelect: () => {
                document.dispatchEvent(new CustomEvent("huuper:attendee-cancel", { detail: attendee }));
              },
            },
          ],
        });
      });
    });
  }

  function bindRowLinks(root) {
    root.querySelectorAll(".user-row-linkable").forEach((row) => {
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
