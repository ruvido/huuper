window.appRequestNoteSheet = (() => {
  function text(value) {
    return window.appListPage.text(value);
  }

  function escapeHTML(value) {
    return window.appListPage.escapeHTML(value);
  }

  function open(config) {
    if (!window.appActionSheet || !window.appListPage) {
      return;
    }

    const title = text(config && config.title);
    const placeholder = text(config && config.placeholder) || "Write";
    const submitLabel = text(config && config.submitLabel) || "Submit";
    const submitTone = text(config && config.submitTone);
    const emptyStatus = text(config && config.emptyStatus) || "Write notes.";
    const statusNode = config && config.statusNode ? config.statusNode : null;
    const inputID = "request-note-sheet-input";
    let inputNode = null;

    window.appActionSheet.open({
      title,
      contentHTML: `
        <div class="request-note-sheet-field">
          <textarea id="${inputID}" placeholder="${escapeHTML(placeholder)}"></textarea>
        </div>
      `,
      footerAction: {
        label: submitLabel,
        tone: submitTone,
        onSelect: async (sheetButton) => {
          if (typeof config.onSubmit !== "function") {
            return;
          }
          const value = text(inputNode && inputNode.value);
          if (!value) {
            if (inputNode) {
              inputNode.focus();
            }
            if (statusNode) {
              window.appListPage.setStatus(statusNode, emptyStatus);
            }
            return;
          }
          try {
            sheetButton.disabled = true;
            await config.onSubmit(value, sheetButton, inputNode);
            window.appActionSheet.close();
          } catch (error) {
            sheetButton.disabled = false;
            if (statusNode) {
              window.appListPage.setStatus(
                statusNode,
                window.appRequestActions.errorMessage(error, "Action unavailable."),
              );
            }
            throw error;
          }
        },
      },
      onOpen: (root) => {
        inputNode = root.querySelector(`#${inputID}`);
        if (inputNode) {
          inputNode.focus();
        }
      },
    });
  }

  return { open };
})();
