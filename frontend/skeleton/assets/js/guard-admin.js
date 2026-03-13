(() => {
  if (!window.huuperAuth) {
    return;
  }

  window.huuperAuth.requireScope("admin");
})();
