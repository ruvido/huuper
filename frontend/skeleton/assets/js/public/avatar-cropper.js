window.appAvatarCropper = (() => {
  const OUTPUT_SIZE = 512;
  const OUTPUT_TYPE = "image/webp";
  const OUTPUT_QUALITY = 0.85;
  const OUTPUT_EXT = "webp";
  // A photo taken with a phone comes in at 48 megapixels: 8000x6000 is 192 MB
  // of decoded pixels, which cropper-image keeps in memory for as long as the
  // editor is open, before the canvases the crop allocates on top. On an
  // entry-level phone the system kills the renderer and the tab dies with no
  // error. For a 512px output twice the side is enough: 1024px is 4 MB and
  // leaves room for whoever zooms in. The actual resizing is done by the
  // browser with resizeQuality "high", which filters properly.
  const MAX_SHORT_SIDE = OUTPUT_SIZE * 2;
  const SOURCE_TYPE = "image/jpeg";
  const SOURCE_QUALITY = 0.92;
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

  // Returns the file to hand the editor: the original if it is already small
  // enough, otherwise a reduced copy. If the browser cannot decode the file we
  // fall back to the original, and the error surfaces where there already is a
  // message for the user instead of here.
  async function downscaleSource(file) {
    if (typeof createImageBitmap !== "function") {
      return file;
    }

    let width = 0;
    let height = 0;
    try {
      const probe = await createImageBitmap(file);
      width = probe.width;
      height = probe.height;
      // Without close() the full-resolution decode stays allocated until the GC
      // gets to it, which is exactly when the memory is needed for the copy.
      probe.close?.();
    } catch (error) {
      return file;
    }

    const shortSide = Math.min(width, height);
    if (shortSide <= MAX_SHORT_SIDE) {
      return file;
    }

    const scale = MAX_SHORT_SIDE / shortSide;
    const targetWidth = Math.max(1, Math.round(width * scale));
    const targetHeight = Math.max(1, Math.round(height * scale));

    let bitmap = null;
    try {
      bitmap = await createImageBitmap(file, {
        resizeWidth: targetWidth,
        resizeHeight: targetHeight,
        resizeQuality: "high",
      });
    } catch (error) {
      return file;
    }

    try {
      const canvas = document.createElement("canvas");
      canvas.width = targetWidth;
      canvas.height = targetHeight;
      const context = canvas.getContext("2d");
      if (!context) {
        return file;
      }
      context.drawImage(bitmap, 0, 0);
      const blob = await canvasToBlob(canvas, SOURCE_TYPE, SOURCE_QUALITY);
      return new File([blob], file.name, { type: blob.type || SOURCE_TYPE });
    } catch (error) {
      return file;
    } finally {
      bitmap.close?.();
    }
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

    const source = await downscaleSource(file);

    const modal = buildModal();
    const objectUrl = URL.createObjectURL(source);

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
