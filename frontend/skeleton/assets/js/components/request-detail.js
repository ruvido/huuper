window.huuperRequestDetail = (() => {
  function text(value) {
    return window.huuperListPage.text(value);
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
    return `<p class="request-row"><span>${window.huuperListPage.escapeHTML(label)}:</span> <strong>${window.huuperListPage.escapeHTML(rendered)}</strong></p>`;
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
      if (!canTakeAction && !canReject) {
        workflowNode.hidden = true;
        workflowNode.innerHTML = "";
        return;
      }

      const requiredField = workflow.required_field || "";
      const action = text(workflow.pending_action);
      const parts = [`<article class="request-workflow-card">`];
      if (canTakeAction && requiredField === "mentoring_notes") {
        parts.push(`<label class="form-field"><span>Mentoring notes</span><textarea id="request-mentoring-notes"></textarea></label>`);
      }
      const actions = [];
      if (canTakeAction && action) {
        const actionLabel = text(workflow.pending_action_label) || actionText(action);
        actions.push(`<button id="request-action" class="primary" type="button">${window.huuperListPage.escapeHTML(actionLabel)}</button>`);
      }
      if (canReject) {
        actions.push(`<button id="request-reject" class="request-reject-button" type="button">Reject</button>`);
      }
      if (actions.length === 0) {
        workflowNode.hidden = true;
        workflowNode.innerHTML = "";
        return;
      }
      parts.push(`<div class="action-row request-page-actions">${actions.join("")}</div>`);
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
