(() => {
  const button = document.querySelector("[data-logout]");
  if (!button || !window.appAuth) {
    return;
  }

  button.addEventListener("click", () => {
    window.appAuth.clear();
    window.location.href = "/";
  });
})();
