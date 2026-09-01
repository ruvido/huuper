window.appRequestDetail = (() => {
  function text(value) {
    return window.appListPage.text(value);
  }

  function escapeHTML(value) {
    return window.appListPage.escapeHTML(value);
  }

  function actionText(value) {
    return window.appRequestItem.actionText(value);
  }

  function copy() {
    return (window.appCopy && window.appCopy.ui && window.appCopy.ui.requests) || {};
  }

  function row(label, value) {
    const rendered = text(value);
    if (!rendered) {
      return "";
    }
    return `<p class="request-row"><span>${escapeHTML(label)}:</span> <strong>${escapeHTML(rendered)}</strong></p>`;
  }

  function dateOnly(value) {
    const raw = text(value);
    if (!raw) {
      return "";
    }
    const parsed = new Date(raw);
    if (Number.isNaN(parsed.getTime())) {
      return raw;
    }
    return new Intl.DateTimeFormat("it-IT", {
      day: "2-digit",
      month: "2-digit",
      year: "numeric",
    }).format(parsed);
  }

  function titleCase(value) {
    const raw = text(value);
    if (!raw) {
      return "";
    }
    return raw
      .toLowerCase()
      .split(/\s+/)
      .filter(Boolean)
      .map((token) => token
        .split("-")
        .map((part) => part ? `${part.charAt(0).toUpperCase()}${part.slice(1)}` : part)
        .join("-"))
      .join(" ");
  }

  function personLabel(value) {
    const raw = text(value).trim();
    if (!raw) {
      return "";
    }
    if (!raw.includes("@")) {
      return raw;
    }
    const localPart = raw.split("@")[0].trim();
    if (!localPart) {
      return raw;
    }
    const spaced = localPart.replace(/[._-]+/g, " ");
    return titleCase(spaced) || raw;
  }

  function completedByLine(who) {
    const cleanWho = personLabel(who);
    if (!cleanWho) {
      return "";
    }
    return `Completed by ${cleanWho}`;
  }

  function processNote(rawHTML, rawText) {
    const html = text(rawHTML);
    if (html) {
      return html;
    }
    const plain = text(rawText);
    if (!plain) {
      return "";
    }
    return `<p>${escapeHTML(plain)}</p>`;
  }

  function requestFlowFromPayload(payload) {
    if (payload && typeof payload.request_flow === "object" && payload.request_flow) {
      const flowData = payload.request_flow.data;
      if (flowData && typeof flowData === "object") {
        return flowData;
      }
    }
    return {};
  }

  function renderTimelineStep(options) {
    const title = text(options.title);
    const date = dateOnly(options.date);
    const valueHTML = typeof options.valueHTML === "string" ? options.valueHTML : "";
    const value = text(options.valueText);
    const subtitle = text(options.subtitle);
    const subtitleClass = text(options.subtitleClass);
    const completedBy = completedByLine(options.completedBy);
    const note = processNote(options.noteHTML, options.noteText);
    const noteClass = text(options.noteClass);
    const noteBodyClass = text(options.noteBodyClass) || "request-process-note-body";
    const flowNoteHTML = text(options.flowNoteHTML);
    const flowNoteText = text(options.flowNoteText);
    const infoNoteHTML = text(options.infoNoteHTML);
    const infoNoteText = text(options.infoNoteText);
    const hasVisibleInfoNoteInTimeline = Boolean(flowNoteHTML || flowNoteText);
    const completed = options.completed === true;
    const current = options.current === true;
    const classes = ["request-process-step"];
    if (options.variantClass) {
      classes.push(text(options.variantClass));
    }
    if (completed) {
      classes.push("request-process-step-done");
    } else if (current) {
      classes.push("request-process-step-current");
    } else {
      classes.push("request-process-step-todo");
    }
    return `
      <article class="${classes.join(" ")}">
        <span class="request-process-rail request-process-rail-top" aria-hidden="true"></span>
        <span class="request-process-marker" aria-hidden="true"></span>
        <span class="request-process-rail request-process-rail-bottom" aria-hidden="true"></span>
        <div class="request-process-content">
          <div class="request-process-head">
            ${title ? `
              <div class="request-process-label-col">
                <div class="request-process-title-row">
                  <p class="request-detail-term request-process-label">${escapeHTML(title)}</p>
                  ${date ? `
                    <div class="request-process-head-meta">
                      <p class="request-process-date">${escapeHTML(date)}</p>
                    </div>
                  ` : ""}
                  ${current && !completed && !hasVisibleInfoNoteInTimeline && (infoNoteHTML || infoNoteText) ? `<button type="button" class="request-process-info" data-info-title="${escapeHTML(title)}"${infoNoteHTML ? ` data-info-html="${escapeHTML(encodeURIComponent(infoNoteHTML))}"` : ""}${infoNoteText ? ` data-info-text="${escapeHTML(encodeURIComponent(infoNoteText))}"` : ""} aria-label="${escapeHTML(`${title} info`)}">i</button>` : ""}
                </div>
              </div>
            ` : ""}
          </div>
          ${(flowNoteHTML || flowNoteText || note || valueHTML || value || subtitle) ? `
            <div class="request-process-body">
              ${flowNoteHTML ? `<div class="request-process-flow-note">${flowNoteHTML}</div>` : flowNoteText ? `<p class="request-process-flow-note">${escapeHTML(flowNoteText)}</p>` : ""}
              ${note ? `<div class="request-process-note${noteClass ? ` ${noteClass}` : ""}"><div class="${noteBodyClass}">${note}</div></div>` : ""}
              ${!note && valueHTML ? `<div class="request-process-value">${valueHTML}</div>` : !note && value ? `<p class="request-process-value">${escapeHTML(value)}</p>` : ""}
              ${!note && subtitle ? `<p class="request-process-subtitle${subtitleClass ? ` ${subtitleClass}` : ""}">${escapeHTML(subtitle)}</p>` : ""}
            </div>
          ` : ""}
          ${completedBy ? `<p class="request-process-signature">${escapeHTML(completedBy)}</p>` : ""}
        </div>
      </article>
    `;
  }

  function flowNoteForStep(step) {
    const html = text(step && step.notes_html);
    if (html) {
      return { html, text: "" };
    }
    return {
      html: "",
      text: text(step && step.notes),
    };
  }

  function renderMentoringNotesHTML(mentoringData) {
    const notes = Array.isArray(mentoringData && mentoringData.notes) ? mentoringData.notes : [];
    const items = notes.map((note) => {
      const noteText = text(note && note.text);
      if (!noteText) {
        return "";
      }
      const who = personLabel(note && note.by);
      const when = dateOnly(note && note.at);
      const metaParts = [];
      if (when) {
        metaParts.push(`\u2014 ${when}`);
      }
      if (who) {
        metaParts.push(who);
      }
      const meta = metaParts.join(" \u2022 ");
      return `
        <div>
          <p>
            ${escapeHTML(noteText)}
            ${meta ? ` <span class="request-notes-meta-inline">${escapeHTML(meta)}</span>` : ""}
          </p>
        </div>
      `;
    }).filter(Boolean).join("");
    if (!items) {
      return "";
    }
    return `<div class="request-note-list request-notes">${items}</div>`;
  }

  function renderProcessApproval(payload) {
    const data = payload && typeof payload.data === "object" ? payload.data : {};
    const flow = requestFlowFromPayload(payload);
    const flowSteps = Array.isArray(flow.steps) ? flow.steps : [];
    const flowTitle = text(flow.title);
    const mentoringData = data.mentoring && typeof data.mentoring === "object" ? data.mentoring : {};
    const mentoringDoneAt = text(mentoringData.done_at);
    const mentoringDoneBy = text(mentoringData.done_by);
    const mentoringNotesHTML = renderMentoringNotesHTML(mentoringData);
    const assignmentGroupData = data.assign_group && typeof data.assign_group === "object" ? data.assign_group : {};
    const guardianData = data.guardian && typeof data.guardian === "object" ? data.guardian : {};
    const groupName = titleCase(payload.group_name);
    const groupAssignedAt = text(assignmentGroupData.assigned_at);
    const groupAssignedBy = text(assignmentGroupData.assigned_by);
    const stepCompleted = (action) => {
      switch (action) {
        case "assign_group":
          return groupAssignedAt !== "" || text(data.group) !== "";
        case "assign_guardian":
          return text(guardianData.assigned_at) !== "";
        case "mentoring":
          return mentoringDoneAt !== "";
        case "group_approved":
          return text(data.group_approved_at) !== "";
        case "admin_approved":
          return text(data.admin_approved_at) !== "";
        default:
          return false;
      }
    };
    const currentIndex = flowSteps.findIndex((step) => !stepCompleted(text(step && step.action)));
    const normalizedCurrentIndex = currentIndex < 0 ? flowSteps.length : currentIndex;
    const stepStateForAction = (action, step, index) => {
      const isCurrent = index === normalizedCurrentIndex;
      const flowNote = flowNoteForStep(step);
      switch (action) {
        case "assign_group":
          {
            const completed = groupAssignedAt !== "" || text(data.group) !== "";
            const renderedGroupName = groupName || text(payload.group_name);
            return {
              completed,
              current: isCurrent,
              date: groupAssignedAt,
              subtitle: "",
              completedBy: groupAssignedBy,
              flowNoteHTML: completed ? "" : flowNote.html,
              flowNoteText: completed ? "" : flowNote.text,
              variantClass: "request-process-step-group",
              valueHTML: "",
              noteClass: "request-process-note-content",
              noteBodyClass: "request-process-note-content-body",
              noteHTML: completed && renderedGroupName ? `<p><strong>${escapeHTML(renderedGroupName)}</strong></p>` : "",
              noteText: "",
            };
          }
        case "assign_guardian":
          {
            const guardianName = personLabel(guardianData.name);
            const completed = text(guardianData.assigned_at) !== "";
            const renderedGuardianName = guardianName || personLabel(payload.guardian_name);
            return {
              completed,
              current: isCurrent,
              date: guardianData.assigned_at,
              subtitle: "",
              completedBy: guardianData.assigned_by || payload.guardian_name,
              flowNoteHTML: completed ? "" : flowNote.html,
              flowNoteText: completed ? "" : flowNote.text,
              valueHTML: "",
              noteClass: "request-process-note-content",
              noteBodyClass: "request-process-note-content-body",
              noteHTML: completed && renderedGuardianName ? `<p><strong>${escapeHTML(renderedGuardianName)}</strong></p>` : "",
              noteText: "",
            };
          }
        case "mentoring":
          return {
            completed: mentoringDoneAt !== "",
            current: isCurrent,
            date: mentoringDoneAt,
            completedBy: mentoringDoneBy,
            flowNoteHTML: mentoringNotesHTML ? "" : undefined,
            flowNoteText: mentoringNotesHTML ? "" : undefined,
            noteBodyClass: "request-process-note-content-body",
            noteHTML: mentoringNotesHTML,
            noteText: "",
          };
        case "group_approved":
          return {
            completed: text(data.group_approved_at) !== "",
            current: isCurrent,
            date: data.group_approved_at,
            subtitle: text(data.group_approved_at) !== "" ? "Approved" : "",
            subtitleClass: "request-process-subtitle-content",
            completedBy: data.group_approved_by,
            flowNoteHTML: text(data.group_approved_at) !== "" ? "" : undefined,
            flowNoteText: text(data.group_approved_at) !== "" ? "" : undefined,
            noteText: "",
          };
        case "admin_approved":
          return {
            completed: text(data.admin_approved_at) !== "",
            current: isCurrent,
            date: data.admin_approved_at,
            subtitle: text(data.admin_approved_at) !== "" ? "Approved" : "",
            subtitleClass: "request-process-subtitle-content",
            completedBy: data.admin_approved_by,
            flowNoteHTML: text(data.admin_approved_at) !== "" ? "" : undefined,
            flowNoteText: text(data.admin_approved_at) !== "" ? "" : undefined,
            noteText: "",
          };
        default:
          return {
            completed: false,
            current: false,
          };
      }
    };

    const steps = flowSteps.map((step, index) => {
      const action = text(step && step.action);
      if (!action) {
        return "";
      }
      const state = stepStateForAction(action, step, index);
      const displayTitle = text(step && step.label);
      const stepFlowNoteHTML = state.flowNoteHTML !== undefined ? state.flowNoteHTML : text(step && step.notes_html);
      const stepFlowNoteText = state.flowNoteText !== undefined ? state.flowNoteText : text(step && step.notes);
      return renderTimelineStep({
        title: displayTitle,
        flowNoteHTML: stepFlowNoteHTML,
        flowNoteText: stepFlowNoteText,
        infoNoteHTML: text(step && step.notes_html),
        infoNoteText: text(step && step.notes),
        completed: state.completed === true,
        current: state.current === true,
        date: state.date,
        subtitle: state.subtitle,
        completedBy: state.completedBy,
        noteHTML: state.noteHTML,
        noteText: state.noteText,
        valueHTML: state.valueHTML,
        valueText: state.valueText,
      });
    }).filter(Boolean).join("");

    return `
      <section class="request-process">
        ${flowTitle ? `<p class="request-section-title request-process-title">${escapeHTML(flowTitle)}</p>` : ""}
        <div class="request-process-list">${steps}</div>
      </section>
    `;
  }

  function init(config) {
    const statusNode = document.querySelector("#request-status");
    const summaryNode = document.querySelector("#request-summary");
    const workflowNode = document.querySelector("#request-workflow");
    const archiveButtonNode = document.querySelector("[data-request-archive]");
    const unarchiveButtonNode = document.querySelector("[data-request-unarchive]");
    if (!statusNode || !summaryNode || !workflowNode || !window.appAuth || !window.appListPage || !window.appRequestItem || !window.appActionSheet || !window.appRequestActions || !window.appRequestNoteSheet) {
      return;
    }

    document.body.classList.add("request-page");
    const id = window.appListPage.queryParam("id");
    if (!id) {
      window.appListPage.setStatus(statusNode, "Missing request.");
      return;
    }

    function bindArchiveButton(buttonNode) {
      if (!buttonNode) {
        return;
      }
      buttonNode.addEventListener("click", () => {
        const archiveDialog = copy().archiveDialog || {};
        window.appRequestNoteSheet.open({
          title: archiveDialog.title || "Are you sure you want to archive this request?",
          submitLabel: archiveDialog.submitLabel || "Archive",
          submitTone: "danger",
          emptyStatus: archiveDialog.emptyStatus || "Write reason.",
          statusNode,
          onSubmit: async (reason) => {
            await window.appRequestActions.submitAndRedirect({
              actionURL: config.actionURL(id),
              body: {
                action: "archive",
                reason,
              },
              redirectURL: config.requestsURL,
            });
          },
        });
      });
    }

    function bindUnarchiveButton(buttonNode) {
      if (!buttonNode) {
        return;
      }
      buttonNode.addEventListener("click", async () => {
        await window.appRequestActions.submitAndRedirect({
          actionURL: config.actionURL(id),
          body: {
            action: "unarchive",
          },
          redirectURL: config.requestsURL,
        });
      });
    }

    if (archiveButtonNode) {
      archiveButtonNode.hidden = true;
      bindArchiveButton(archiveButtonNode);
    }
    if (unarchiveButtonNode) {
      unarchiveButtonNode.hidden = true;
      bindUnarchiveButton(unarchiveButtonNode);
    }

    function applyPayload(payload) {
      summaryNode.hidden = false;
      summaryNode.innerHTML = window.appRequestItem.renderDetail(payload);
      const workflow = payload && typeof payload.workflow === "object" ? payload.workflow : {};
      if (archiveButtonNode) {
        archiveButtonNode.hidden = workflow.can_archive !== true;
      }
      if (unarchiveButtonNode) {
        unarchiveButtonNode.hidden = workflow.can_unarchive !== true;
      }
      renderWorkflow(payload);
      window.appListPage.setStatus(statusNode, "");
    }

    async function refreshDetail() {
      const payload = await window.appAuth.apiFetch(config.detailURL(id));
      applyPayload(payload);
    }

    function renderWorkflow(payload) {
      const workflow = payload && typeof payload.workflow === "object" ? payload.workflow : {};
      const canTakeAction = workflow.can_take_pending_action === true;
      const canUnarchive = workflow.can_unarchive === true;
      const processApproval = renderProcessApproval(payload);
      if (!canTakeAction && !processApproval && !canUnarchive) {
        workflowNode.hidden = true;
        workflowNode.innerHTML = "";
        return;
      }

      const requiredField = workflow.required_field || "";
      const action = text(workflow.pending_action);
      const payloadData = payload && typeof payload.data === "object" && payload.data ? payload.data : {};
      const mentoringState = payloadData.mentoring && typeof payloadData.mentoring === "object" ? payloadData.mentoring : {};
      const mentoringNotesCount = Array.isArray(mentoringState.notes) ? mentoringState.notes.length : 0;
      const isMentoringStep = requiredField === "mentoring_notes";
      const canCloseMentoring = !isMentoringStep || mentoringNotesCount > 0;
      const parts = [`<article class="request-workflow-card">`];
      if (processApproval) {
        parts.push(processApproval);
      }
      const actions = [];
      if (canTakeAction && isMentoringStep) {
        actions.push(`<button id="request-action-add-note" class="secondary" type="button">+ ADD NOTE</button>`);
      }
      if (canTakeAction && action && canCloseMentoring) {
        const requestCopy = copy();
        const actionLabel = isMentoringStep
          ? (requestCopy.mentoringActionLabel || "Finalize mentoring")
          : (text(workflow.pending_action_label) || actionText(action));
        actions.push(`<button id="request-action" class="primary" type="button">${window.appListPage.escapeHTML(actionLabel)}</button>`);
      }
      if (canUnarchive) {
        const unarchiveLabel = copy().unarchiveActionLabel || "Unarchive request";
        actions.push(`<button id="request-action-unarchive" class="secondary" type="button">${window.appListPage.escapeHTML(unarchiveLabel)}</button>`);
      }
      if (actions.length === 0 && !processApproval) {
        workflowNode.hidden = true;
        workflowNode.innerHTML = "";
        return;
      }
      if (actions.length > 0) {
        parts.push(`<div class="request-bottom-actions">`);
        parts.push(`<div class="action-row request-bottom-actions-row">`);
        if (actions.length > 0) {
          parts.push(actions.join(""));
        }
        parts.push(`</div>`);
        parts.push(`</div>`);
      }
      parts.push(`</article>`);
      workflowNode.innerHTML = parts.join("");
      workflowNode.hidden = false;

      workflowNode.querySelectorAll(".request-process-info").forEach((infoButton) => {
        infoButton.addEventListener("click", () => {
          const infoTitle = infoButton.dataset.infoTitle || "";
          const rawHTML = infoButton.dataset.infoHtml ? decodeURIComponent(infoButton.dataset.infoHtml) : "";
          const rawText = infoButton.dataset.infoText ? decodeURIComponent(infoButton.dataset.infoText) : "";
          const contentHTML = rawHTML
            || (rawText ? `<p>${escapeHTML(rawText).replaceAll("\n", "<br>")}</p>` : "");
          if (!contentHTML) {
            return;
          }
          window.appActionSheet.open({
            title: infoTitle,
            contentHTML,
          });
        });
      });

      const button = workflowNode.querySelector("#request-action");
      const addNoteButton = workflowNode.querySelector("#request-action-add-note");
      const unarchiveButton = workflowNode.querySelector("#request-action-unarchive");

      if (unarchiveButton) {
        unarchiveButton.addEventListener("click", async (event) => {
          try {
            window.appListPage.setStatus(statusNode, "");
            await window.appRequestActions.submitAndRedirect({
              actionURL: config.actionURL(id),
              body: { action: "unarchive" },
              button: event.currentTarget,
              redirectURL: config.requestsURL,
            });
          } catch (error) {
            window.appListPage.setStatus(
              statusNode,
              window.appRequestActions.errorMessage(error, "Unarchive unavailable."),
            );
          }
        });
      }

      if (addNoteButton) {
        addNoteButton.addEventListener("click", () => {
          window.appRequestNoteSheet.open({
            title: "Add mentoring note",
            submitLabel: "Add",
            emptyStatus: "Write a note.",
            statusNode,
            onSubmit: async (note, sheetButton) => {
              await window.appRequestActions.submitAndRedirect({
                actionURL: config.actionURL(id),
                body: {
                  action: "add_mentoring_note",
                  mentoring_notes: note,
                },
                button: sheetButton,
                redirectURL: config.requestsURL,
              });
            },
          });
        });
      }

      if (button) {
        const submitPrimaryAction = async (submitButton) => {
          try {
            const body = { action };
            window.appListPage.setStatus(statusNode, "");
            await window.appRequestActions.submitAndRedirect({
              actionURL: config.actionURL(id),
              body,
              button: submitButton,
              redirectURL: config.requestsURL,
            });
          } catch (error) {
            window.appListPage.setStatus(
              statusNode,
              window.appRequestActions.errorMessage(error, "Action unavailable."),
            );
            submitButton.disabled = false;
          }
        };

        button.addEventListener("click", () => {
          if (!action) {
            return;
          }
          if ((requiredField === "group" || requiredField === "guardian") && window.appRequestAssignmentSheet) {
            window.appRequestAssignmentSheet.open({
              requestID: id,
              payload,
              requestsURL: config.requestsURL,
              detailURL: config.detailURL,
              actionURL: config.actionURL,
            });
            return;
          }

          if (action === "set_mentoring_done") {
            const closeMentoringDialog = copy().closeMentoringDialog || {};
            window.appActionSheet.open({
              title: closeMentoringDialog.title || "Are you sure you want to close mentoring?",
              footerAction: {
                label: closeMentoringDialog.submitLabel || "Close mentoring",
                onSelect: (sheetButton) => submitPrimaryAction(sheetButton),
              },
            });
            return;
          }

          submitPrimaryAction(button).catch(() => {});
        });
      }
    }

    refreshDetail().catch(() => {
      window.appListPage.setStatus(statusNode, "Request unavailable.");
    });
  }

  return { init };
})();
