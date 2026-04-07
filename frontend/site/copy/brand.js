window.huuperCopy = {
  brandName: "Il Branco Realmen",
  shortName: "Il Branco Realmen",
};

(() => {
  const titleNode = document.querySelector("title");
  if (!titleNode || !window.huuperCopy) {
    return;
  }

  titleNode.textContent = window.huuperCopy.brandName;
})();
