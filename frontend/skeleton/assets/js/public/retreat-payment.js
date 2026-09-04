(() => {
  const titleNode = document.querySelector("#retreat-payment-title");
  const messageNode = document.querySelector("#retreat-payment-message");
  if (!titleNode || !messageNode) {
    return;
  }

  const status = new URLSearchParams(window.location.search).get("status");
  if (status === "success") {
    titleNode.textContent = "Payment received";
    messageNode.textContent = "Thank you! Your registration is confirmed, you will receive an email shortly.";
  } else {
    titleNode.textContent = "Payment cancelled";
    messageNode.textContent = "The payment was not completed. You can try again from the retreat page.";
  }
})();
