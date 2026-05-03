(() => {
  if (!window.appAuth) {
    return;
  }

  window.appAuth.requireScope("admin");
})();
