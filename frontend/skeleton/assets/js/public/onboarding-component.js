window.appOnboardingComponent = (() => {
  function text(value) {
    return window.appPublicCommon.text(value);
  }

  function renderStartScreen(options) {
    const stepNode = options.stepNode;
    const startButton = options.startButton;
    const page = options.page || {};
    if (!stepNode || !startButton || !window.appPublicCommon) {
      return;
    }

    const title = text(page.title);
    const copy = text(page.text);

    stepNode.className = "public-request-step public-request-step-start";
    stepNode.innerHTML = "";

    const wrapper = document.createElement("div");
    wrapper.className = "public-request-start-stack";

    const titleNode = document.createElement("h2");
    titleNode.className = "public-request-title public-request-title-start public-request-title-start-lg";
    titleNode.textContent = title;
    wrapper.appendChild(titleNode);

    const copyNode = document.createElement("p");
    copyNode.className = "public-request-copy";
    copyNode.innerHTML = window.appPublicCommon.escapeHTML(copy).replace(/\n/g, "<br>");
    wrapper.appendChild(copyNode);

    const actionRow = document.createElement("div");
    actionRow.className = "action-row";
    actionRow.appendChild(startButton);
    wrapper.appendChild(actionRow);
    stepNode.appendChild(wrapper);
  }

  function renderCenteredScreen(options) {
    const stepNode = options.stepNode;
    const page = options.page || {};
    const titleClass = text(options.titleClass) || "public-request-title public-request-title-start";
    if (!stepNode || !window.appPublicCommon) {
      return;
    }

    const title = text(page.title);
    const copy = text(page.text);

    stepNode.className = "public-request-step public-request-step-start";
    stepNode.innerHTML = "";

    const wrapper = document.createElement("div");
    wrapper.className = "public-request-start-stack";

    const titleNode = document.createElement("h2");
    titleNode.className = titleClass;
    titleNode.textContent = title;
    wrapper.appendChild(titleNode);

    const copyNode = document.createElement("p");
    copyNode.className = "public-request-copy";
    copyNode.innerHTML = window.appPublicCommon.escapeHTML(copy).replace(/\n/g, "<br>");
    wrapper.appendChild(copyNode);

    stepNode.appendChild(wrapper);
    return wrapper;
  }

  return {
    renderStartScreen,
    renderCenteredScreen,
  };
})();
