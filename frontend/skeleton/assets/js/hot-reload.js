(() => {
  const source = new EventSource("/_dev/live-reload");
  source.addEventListener("reload", () => {
    window.location.reload();
  });
})();
