window.huuperRequestDetail = (() => {
  function text(value) {
    return window.huuperListPage.text(value);
  }

  function escapeHTML(value) {
    return window.huuperListPage.escapeHTML(value);
  }

  function requestAgeDaysLabel(value) {
    const raw = text(value);
    if (!raw) {
      return "";
    }
    const parsed = new Date(raw);
    if (Number.isNaN(parsed.getTime())) {
      return "";
    }
    const diffDays = Math.floor((Date.now() - parsed.getTime()) / 86400000);
    if (diffDays <= 0) {
      return "";
    }
    return `${diffDays} days ago`;
  }

  function actionText(value) {
    return window.huuperRequestItem.actionText(value);
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

  function processSignature(prefix, who, at) {
    const cleanPrefix = text(prefix);
    const cleanWho = text(who);
    const parts = [];
    if (cleanPrefix) {
      parts.push(cleanPrefix);
    }
    if (cleanWho) {
      parts.push(`by ${cleanWho}`);
    }
    const rendered = parts.join(" ").trim();
    if (!rendered) {
      return "";
    }
    return rendered;
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

  function processStep(options) {
    const title = text(options.title);
    const info = text(options.info);
    const date = dateOnly(options.date);
    const valueHTML = typeof options.valueHTML === "string" ? options.valueHTML : "";
    const value = text(options.valueText);
    const subtitle = text(options.subtitle);
    const signature = processSignature(options.signaturePrefix, options.signatureWho, options.signatureAt);
    const note = processNote(options.noteHTML, options.noteText);
    const flowNoteHTML = text(options.flowNoteHTML);
    const flowNoteText = text(options.flowNoteText);
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
                <div class="request-process-label-wrap">
                  <p class="request-detail-term request-process-label">${escapeHTML(title)}</p>
                  ${info ? `<a class="request-process-info" href="${escapeHTML(info)}" target="_blank" rel="noreferrer" aria-label="${escapeHTML(`${title} info`)}">i</a>` : ""}
                </div>
                ${(value || date || subtitle || (completed && signature && options.showSignature !== false)) ? `
                  <div class="request-process-meta">
                    ${date ? `<p class="request-process-date">${escapeHTML(date)}</p>` : ""}
                    ${valueHTML ? `<div class="request-process-value">${valueHTML}</div>` : value ? `<div class="request-process-value">${escapeHTML(value)}</div>` : ""}
                    ${subtitle ? `<p class="request-process-note request-process-subtitle">${escapeHTML(subtitle)}</p>` : ""}
                    ${(completed && signature && options.showSignature !== false) ? `<p class="request-process-signature">${escapeHTML(signature)}</p>` : ""}
                  </div>
                ` : ""}
              </div>
            ` : ""}
          </div>
          ${flowNoteHTML ? `<div class="request-process-flow-note">${flowNoteHTML}</div>` : flowNoteText ? `<p class="request-process-flow-note">${escapeHTML(flowNoteText)}</p>` : ""}
          ${note ? `<div class="request-process-note">${note}</div>` : ""}
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

  function renderProcessApproval(payload) {
    const data = payload && typeof payload.data === "object" ? payload.data : {};
    const flow = requestFlowFromPayload(payload);
    const flowSteps = Array.isArray(flow.steps) ? flow.steps : [];
    const flowTitle = text(flow.title);
    const mentoringNoteHTML = text(payload.mentoring_notes_html);
    const mentoringNoteText = text(payload.mentoring_notes || data.mentoring_notes);
    const guardianData = data.guardian && typeof data.guardian === "object" ? data.guardian : {};
    const groupName = titleCase(payload.group_name);
    const groupAssignedAt = text(data.group_assigned_at);
    const groupAssignedBy = text(data.group_assigned_by);
    const stepCompleted = (action) => {
      switch (action) {
        case "assign_group":
          return groupAssignedAt !== "" || text(data.group) !== "";
        case "assign_guardian":
          return text(guardianData.assigned_at) !== "";
        case "mentoring":
          return text(data.mentoring_done_at) !== "";
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
            return {
              completed: groupAssignedAt !== "" || text(data.group) !== "",
              current: isCurrent,
              date: groupAssignedAt,
              subtitle: "",
              signaturePrefix: groupAssignedBy ? "Assigned" : "",
              signatureWho: groupAssignedBy,
              signatureAt: groupAssignedAt,
              flowNoteHTML: isCurrent ? flowNote.html : "",
              flowNoteText: isCurrent ? flowNote.text : "",
              variantClass: "request-process-step-group",
              valueHTML: groupName ? `<span class="request-process-value-inline"><svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-arrow-return-right request-process-value-icon" viewBox="0 0 16 16" aria-hidden="true"><path fill-rule="evenodd" d="M1.5 1.5A.5.5 0 0 0 1 2v4.8a2.5 2.5 0 0 0 2.5 2.5h9.793l-3.347 3.346a.5.5 0 0 0 .708.708l4.2-4.2a.5.5 0 0 0 0-.708l-4-4a.5.5 0 0 0-.708.708L13.293 8.3H3.5A1.5 1.5 0 0 1 2 6.8V2a.5.5 0 0 0-.5-.5"></path></svg><strong class="request-process-value-text">${escapeHTML(groupName)}</strong></span>` : "",
            };
          }
        case "assign_guardian":
          {
            const guardianName = titleCase(guardianData.name);
            return {
              completed: text(guardianData.assigned_at) !== "",
              current: isCurrent,
              date: guardianData.assigned_at,
              subtitle: "",
              signaturePrefix: text(guardianData.assigned_by) ? "Assigned" : "",
              signatureWho: guardianData.assigned_by || payload.guardian_name,
              signatureAt: guardianData.assigned_at,
              flowNoteHTML: isCurrent ? flowNote.html : "",
              flowNoteText: isCurrent ? flowNote.text : "",
              valueHTML: guardianName ? `<span class="request-process-value-inline"><svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-arrow-return-right request-process-value-icon" viewBox="0 0 16 16" aria-hidden="true"><path fill-rule="evenodd" d="M1.5 1.5A.5.5 0 0 0 1 2v4.8a2.5 2.5 0 0 0 2.5 2.5h9.793l-3.347 3.346a.5.5 0 0 0 .708.708l4.2-4.2a.5.5 0 0 0 0-.708l-4-4a.5.5 0 0 0-.708.708L13.293 8.3H3.5A1.5 1.5 0 0 1 2 6.8V2a.5.5 0 0 0-.5-.5"></path></svg><strong class="request-process-value-text">${escapeHTML(guardianName)}</strong></span>` : "",
            };
          }
        case "mentoring":
          return {
            completed: text(data.mentoring_done_at) !== "",
            current: isCurrent,
            date: data.mentoring_done_at,
            signaturePrefix: "Completed",
            signatureWho: data.mentoring_done_by,
            signatureAt: data.mentoring_done_at,
            flowNoteHTML: mentoringNoteHTML ? "" : undefined,
            flowNoteText: mentoringNoteHTML ? "" : undefined,
            noteHTML: mentoringNoteHTML ? mentoringNoteHTML : "",
            noteText: "",
          };
        case "group_approved":
          return {
            completed: text(data.group_approved_at) !== "",
            current: isCurrent,
            date: data.group_approved_at,
            subtitle: text(data.group_approved_at) !== "" ? "Approved" : "",
            signaturePrefix: "Completed",
            signatureWho: data.group_approved_by,
            signatureAt: data.group_approved_at,
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
            signaturePrefix: "Completed",
            signatureWho: data.admin_approved_by,
            signatureAt: data.admin_approved_at,
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
      return processStep({
        title: displayTitle,
        info: text(step && step.info),
        flowNoteHTML: stepFlowNoteHTML,
        flowNoteText: stepFlowNoteText,
        completed: state.completed === true,
        current: state.current === true,
        date: state.date,
        subtitle: state.subtitle,
        signaturePrefix: state.signaturePrefix,
        signatureWho: state.signatureWho,
        signatureAt: state.signatureAt,
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
    const topbarMetaNode = document.querySelector("#request-topbar-meta");
    if (!statusNode || !summaryNode || !workflowNode || !window.huuperAuth || !window.huuperListPage || !window.huuperRequestItem || !window.huuperActionSheet || !window.huuperRequestActions) {
      return;
    }

    document.body.classList.add("request-page");
    const id = window.huuperListPage.queryParam("id");
    if (!id) {
      window.huuperListPage.setStatus(statusNode, "Missing request.");
      return;
    }
    function renderWorkflow(payload) {
      const workflow = payload && typeof payload.workflow === "object" ? payload.workflow : {};
      const canTakeAction = workflow.can_take_pending_action === true;
      const canReject = workflow.can_reject === true;
      const processApproval = renderProcessApproval(payload);
      if (!canTakeAction && !canReject && !processApproval) {
        workflowNode.hidden = true;
        workflowNode.innerHTML = "";
        return;
      }

      const requiredField = workflow.required_field || "";
      const action = text(workflow.pending_action);
      const parts = [`<article class="request-workflow-card">`];
      if (processApproval) {
        parts.push(processApproval);
      }
      if (canTakeAction && requiredField === "mentoring_notes") {
        parts.push(`<label class="form-field"><span>Mentoring notes</span><textarea id="request-mentoring-notes"></textarea></label>`);
      }
      const actions = [];
      if (canTakeAction && action) {
        const actionLabel = text(workflow.pending_action_label) || actionText(action);
        actions.push(`<button id="request-action" class="primary" type="button">${window.huuperListPage.escapeHTML(actionLabel)}</button>`);
      }
      const hasReject = canReject;
      if (actions.length === 0 && !hasReject && !processApproval) {
        workflowNode.hidden = true;
        workflowNode.innerHTML = "";
        return;
      }
      if (actions.length > 0 || hasReject) {
        parts.push(`<div class="request-bottom-actions">`);
        parts.push(`<div class="action-row request-bottom-actions-row">`);
        if (actions.length > 0) {
          parts.push(actions.join(""));
        }
        if (hasReject) {
          parts.push(`<button id="request-reject" class="request-reject-button" type="button">Reject</button>`);
        }
        parts.push(`</div>`);
        parts.push(`</div>`);
      }
      parts.push(`</article>`);
      workflowNode.innerHTML = parts.join("");
      workflowNode.hidden = false;

      const button = workflowNode.querySelector("#request-action");
      const rejectButton = workflowNode.querySelector("#request-reject");

      if (rejectButton) {
        rejectButton.addEventListener("click", () => {
          window.huuperActionSheet.open({
            contentHTML: `
              <label class="form-field request-reject-field">
                <span>Reason</span>
                <textarea id="request-reject-reason" placeholder="Write a reason"></textarea>
              </label>
            `,
            footerAction: {
              label: "Reject",
              tone: "danger",
              onSelect: async (sheetButton) => {
                const reasonNode = document.querySelector("#request-reject-reason");
                const fieldNode = document.querySelector(".request-reject-field");
                const reason = text(reasonNode && reasonNode.value);
                if (!reason) {
                  if (fieldNode) {
                    fieldNode.classList.add("request-reject-field-error");
                  }
                  if (reasonNode) {
                    reasonNode.focus();
                  }
                  window.huuperListPage.setStatus(statusNode, "Write a reason.");
                  return;
                }
                if (fieldNode) {
                  fieldNode.classList.remove("request-reject-field-error");
                }
                try {
                  sheetButton.disabled = true;
                  window.huuperActionSheet.close();
                  await window.huuperRequestActions.submitAndRedirect({
                    actionURL: config.actionURL(id),
                    body: {
                      action: "reject",
                      reason,
                    },
                    redirectURL: config.requestsURL,
                  });
                } catch (_) {
                  sheetButton.disabled = false;
                  window.huuperListPage.setStatus(statusNode, "Reject unavailable.");
                }
              },
            },
          });
        });
      }

      if (button) {
        button.addEventListener("click", async () => {
          if (!action) {
            return;
          }
          if ((requiredField === "group" || requiredField === "guardian") && window.huuperRequestAssignmentSheet) {
            window.huuperRequestAssignmentSheet.open({
              requestID: id,
              payload,
              requestsURL: config.requestsURL,
              detailURL: config.detailURL,
              actionURL: config.actionURL,
            });
            return;
          }

          const body = { action };
          if (requiredField === "mentoring_notes") {
            body.mentoring_notes = text(workflowNode.querySelector("#request-mentoring-notes").value);
            if (!body.mentoring_notes) {
              window.huuperListPage.setStatus(statusNode, "Write notes.");
              return;
            }
          }

          try {
            button.disabled = true;
            window.huuperListPage.setStatus(statusNode, "");
            await window.huuperRequestActions.submitAndRedirect({
              actionURL: config.actionURL(id),
              body,
              redirectURL: config.requestsURL,
            });
          } catch (error) {
            if (error && error.message === "missing_mentoring_notes") {
              window.huuperListPage.setStatus(statusNode, "Write notes.");
            } else {
              window.huuperListPage.setStatus(statusNode, "Action unavailable.");
            }
            button.disabled = false;
          }
        });
      }
    }

    window.huuperAuth.apiFetch(config.detailURL(id)).then((payload) => {
      if (topbarMetaNode) {
        const ageLabel = requestAgeDaysLabel(payload && payload.created);
        topbarMetaNode.textContent = ageLabel;
        topbarMetaNode.hidden = !ageLabel;
      }
      summaryNode.hidden = false;
      summaryNode.innerHTML = window.huuperRequestItem.renderDetail(payload);
      renderWorkflow(payload);
      window.huuperListPage.setStatus(statusNode, "");
    }).catch(() => {
      window.huuperListPage.setStatus(statusNode, "Request unavailable.");
    });
  }

  return { init };
})();
