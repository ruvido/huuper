(() => {
  const button = document.querySelector("[data-logout]");
  if (!button || !window.huuperAuth) {
    return;
  }

  button.addEventListener("click", () => {
    window.huuperAuth.clear();
    window.location.href = "/";
  });
})();
