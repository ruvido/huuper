(() => {
  const button = document.querySelector("[data-back]");
  if (!button) {
    return;
  }

  button.addEventListener("click", () => {
    window.history.back();
  });
})();
