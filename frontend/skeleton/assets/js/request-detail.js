window.huuperRequestDetail = (() => {
  function text(value) {
    return window.huuperListPage.text(value);
  }

  function renderOptions(node, items, placeholder, formatLabel) {
    const options = [`<option value="">${placeholder}</option>`];
    for (const item of items) {
      options.push(`<option value="${window.huuperListPage.escapeHTML(item.id)}">${window.huuperListPage.escapeHTML(formatLabel(item))}</option>`);
    }
    node.innerHTML = options.join("");
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
    if (!statusNode || !summaryNode || !workflowNode || !window.huuperAuth || !window.huuperListPage || !window.huuperRequestCard) {
      return;
    }

    const id = window.huuperListPage.queryParam("id");
    if (!id) {
      window.huuperListPage.setStatus(statusNode, "Missing request.");
      return;
    }

    function renderWorkflow(payload) {
      const workflow = payload && typeof payload.workflow === "object" ? payload.workflow : {};
      if (!workflow.can_advance) {
        workflowNode.hidden = true;
        workflowNode.innerHTML = "";
        return;
      }

      const requiredField = workflow.required_field || "";
      const actionLabel = workflow.next_action_label || workflow.next_action || "Continue";
      const actionNotesHTML = text(workflow.next_action_notes_html);

      const parts = [`<article class="request-card">`];
      if (actionNotesHTML) {
        parts.push(`<div class="request-notes"><span>Notes:</span><div>${actionNotesHTML}</div></div>`);
      }
      if (requiredField === "group") {
        parts.push(`<label class="field"><span>Group</span><select id="request-group"></select></label>`);
      } else if (requiredField === "guardian") {
        parts.push(`<label class="field"><span>Guardian</span><select id="request-guardian"></select></label>`);
      } else if (requiredField === "mentoring_notes") {
        parts.push(`<label class="field"><span>Notes</span><textarea id="request-mentoring-notes"></textarea></label>`);
      }
      parts.push(`<div class="actions"><button id="request-advance" type="button">${window.huuperListPage.escapeHTML(actionLabel)}</button></div>`);
      parts.push(`</article>`);
      workflowNode.innerHTML = parts.join("");
      workflowNode.hidden = false;

      if (requiredField === "group") {
        const select = workflowNode.querySelector("#request-group");
        renderOptions(select, (workflow.options && workflow.options.groups) || [], "Select group", (item) => item.name || item.id);
      }
      if (requiredField === "guardian") {
        const select = workflowNode.querySelector("#request-guardian");
        renderOptions(select, (workflow.options && workflow.options.guardians) || [], "Select guardian", (item) => item.full_name || item.email || item.id);
      }

      const button = workflowNode.querySelector("#request-advance");
      button.addEventListener("click", async () => {
        const body = { action: "advance" };
        if (requiredField === "group") {
          body.group = text(workflowNode.querySelector("#request-group").value);
          if (!body.group) {
            window.huuperListPage.setStatus(statusNode, "Select a group.");
            return;
          }
        }
        if (requiredField === "guardian") {
          body.guardian = text(workflowNode.querySelector("#request-guardian").value);
          if (!body.guardian) {
            window.huuperListPage.setStatus(statusNode, "Select a guardian.");
            return;
          }
        }
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
          await window.huuperAuth.apiFetch(config.actionURL(id), {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(body),
          });
          window.location.reload();
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

    window.huuperAuth.apiFetch(config.detailURL(id)).then((payload) => {
      summaryNode.hidden = false;
      summaryNode.innerHTML = window.huuperRequestCard.renderDetail(payload);
      renderWorkflow(payload);
      window.huuperListPage.setStatus(statusNode, "");
    }).catch(() => {
      window.huuperListPage.setStatus(statusNode, "Request unavailable.");
    });
  }

  return { init };
})();
