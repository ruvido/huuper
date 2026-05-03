window.appCopy = {
  brandName: "Il Branco Realmen",
  shortName: "Il Branco Realmen",
};

(() => {
  const titleNode = document.querySelector("title");
  if (!titleNode || !window.appCopy) {
    return;
  }

  titleNode.textContent = window.appCopy.brandName;
})();
