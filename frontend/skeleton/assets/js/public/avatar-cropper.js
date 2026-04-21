window.huuperAvatarCropper = (() => {
  const OUTPUT_SIZE = 512;
  const OUTPUT_TYPE = "image/webp";
  const OUTPUT_QUALITY = 0.85;
  const OUTPUT_EXT = "webp";
  const TEMPLATE = [
    '<cropper-canvas background>',
    '<cropper-image rotatable scalable translatable></cropper-image>',
    '<cropper-handle action="move" plain></cropper-handle>',
    '<cropper-selection initial-coverage="1" aspect-ratio="1">',
    '</cropper-selection>',
    '</cropper-canvas>',
  ].join("");

  let activeSession = null;

  function text(value) {
    return String(value || "").trim();
  }

  function createElement(tag, className, content) {
    const node = document.createElement(tag);
    if (className) {
      node.className = className;
    }
    if (typeof content === "string") {
      node.textContent = content;
    }
    return node;
  }

  function canvasToBlob(canvas, type, quality) {
    return new Promise((resolve, reject) => {
      canvas.toBlob((blob) => {
        if (blob) {
          resolve(blob);
          return;
        }
        reject(new Error("failed_to_encode_image"));
      }, type, quality);
    });
  }

  function buildModal() {
    const overlay = createElement("div", "avatar-cropper-overlay");
    const panel = createElement("div", "avatar-cropper-panel");
    const header = createElement("div", "avatar-cropper-header");
    const title = createElement("h3", "avatar-cropper-title", "Ritaglia la foto");
    const subtitle = createElement(
      "p",
      "avatar-cropper-subtitle",
      "Sposta l'immagine per inquadrare il quadrato."
    );
    const stage = createElement("div", "avatar-cropper-stage");

    const errorNode = createElement("p", "avatar-cropper-error meta-text is-error");
    errorNode.hidden = true;

    const footer = createElement("div", "avatar-cropper-footer action-row");
    const cancelButton = document.createElement("button");
    cancelButton.type = "button";
    cancelButton.className = "secondary";
    cancelButton.textContent = "Annulla";
    const confirmButton = document.createElement("button");
    confirmButton.type = "button";
    confirmButton.className = "primary";
    confirmButton.textContent = "Conferma";
    confirmButton.disabled = true;

    header.appendChild(title);
    header.appendChild(subtitle);
    footer.appendChild(cancelButton);
    footer.appendChild(confirmButton);
    panel.appendChild(header);
    panel.appendChild(stage);
    panel.appendChild(errorNode);
    panel.appendChild(footer);
    overlay.appendChild(panel);

    return {
      overlay,
      stage,
      errorNode,
      cancelButton,
      confirmButton,
    };
  }

  function buildFilename(originalFile) {
    const raw = text(originalFile && originalFile.name) || "avatar";
    return `${raw.replace(/\.[^/.]+$/, "")}.${OUTPUT_EXT}`;
  }

  function destroySession(session) {
    if (!session) {
      return;
    }
    if (session.objectUrl) {
      URL.revokeObjectURL(session.objectUrl);
    }
    if (session.overlay && session.overlay.parentNode) {
      session.overlay.parentNode.removeChild(session.overlay);
    }
    document.removeEventListener("keydown", session.handleKeyDown);
    if (activeSession === session) {
      activeSession = null;
    }
  }

  function settleSession(session, result, isError) {
    if (!session || session.settled) {
      return;
    }
    session.settled = true;
    destroySession(session);
    if (isError) {
      session.reject(result);
      return;
    }
    session.resolve(result);
  }

  async function open(file) {
    if (!(file instanceof File)) {
      return null;
    }
    if (!file.type || !file.type.startsWith("image/")) {
      throw new Error("unsupported_image_type");
    }
    if (!window.Cropper || !window.Cropper.DEFAULT_TEMPLATE) {
      throw new Error("cropper_library_missing");
    }
    if (activeSession) {
      destroySession(activeSession);
    }

    const modal = buildModal();
    const objectUrl = URL.createObjectURL(file);

    const session = {
      overlay: modal.overlay,
      stage: modal.stage,
      errorNode: modal.errorNode,
      cancelButton: modal.cancelButton,
      confirmButton: modal.confirmButton,
      objectUrl,
      cropperCanvas: null,
      cropperImage: null,
      cropperSelection: null,
      settled: false,
      resolve: null,
      reject: null,
      handleKeyDown: null,
    };
    activeSession = session;

    const promise = new Promise((resolve, reject) => {
      session.resolve = resolve;
      session.reject = reject;
    });

    document.body.appendChild(modal.overlay);

    modal.stage.innerHTML = TEMPLATE;
    session.cropperCanvas = modal.stage.querySelector("cropper-canvas");
    session.cropperImage = modal.stage.querySelector("cropper-image");
    session.cropperSelection = modal.stage.querySelector("cropper-selection");

    if (session.cropperImage) {
      session.cropperImage.setAttribute("src", objectUrl);
      session.cropperImage.alt = "Foto da ritagliare";
      session.cropperImage.loading = "eager";
    }

    session.handleKeyDown = (event) => {
      if (event.key === "Escape") {
        event.preventDefault();
        settleSession(session, null, false);
      }
    };
    document.addEventListener("keydown", session.handleKeyDown);

    modal.cancelButton.addEventListener("click", () => settleSession(session, null, false));
    modal.confirmButton.addEventListener("click", async () => {
      const selection = session.cropperSelection;
      if (!selection) {
        modal.errorNode.textContent = "Canvas immagine non disponibile.";
        modal.errorNode.hidden = false;
        return;
      }

      modal.confirmButton.disabled = true;
      modal.cancelButton.disabled = true;
      modal.confirmButton.classList.add("request-action-loading");
      modal.confirmButton.setAttribute("aria-busy", "true");
      modal.errorNode.hidden = true;
      modal.errorNode.textContent = "";

      try {
        const output = await selection.$toCanvas({
          width: OUTPUT_SIZE,
          height: OUTPUT_SIZE,
        });
        const blob = await canvasToBlob(output, OUTPUT_TYPE, OUTPUT_QUALITY);
        const cropped = new File([blob], buildFilename(file), { type: OUTPUT_TYPE });
        settleSession(session, cropped, false);
      } catch (error) {
        modal.errorNode.textContent = "Impossibile creare l'immagine ritagliata.";
        modal.errorNode.hidden = false;
        modal.confirmButton.disabled = false;
        modal.cancelButton.disabled = false;
        modal.confirmButton.classList.remove("request-action-loading");
        modal.confirmButton.removeAttribute("aria-busy");
      }
    });

    try {
      if (session.settled) {
        return promise;
      }
      if (session.cropperImage) {
        await session.cropperImage.$ready();
        if (session.settled) {
          return promise;
        }
        session.cropperImage.$center("cover");
        session.cropperImage.$render?.();
      }
      if (session.cropperCanvas) {
        session.cropperCanvas.style.touchAction = "none";
      }
      if (!session.settled) {
        modal.confirmButton.disabled = false;
        modal.overlay.classList.add("is-ready");
      }
    } catch (error) {
      if (session.settled) {
        return promise;
      }
      modal.errorNode.textContent = "Impossibile caricare l'immagine.";
      modal.errorNode.hidden = false;
    }

    return promise;
  }

  return {
    open,
  };
})();
