(() => {
  const titleNode = document.querySelector("#event-payment-title");
  const messageNode = document.querySelector("#event-payment-message");
  if (!titleNode || !messageNode) {
    return;
  }

  const status = new URLSearchParams(window.location.search).get("status");
  if (status === "success") {
    titleNode.textContent = "Pagamento ricevuto";
    messageNode.textContent = "Grazie! La tua iscrizione è confermata, riceverai una email a breve.";
  } else {
    titleNode.textContent = "Pagamento annullato";
    messageNode.textContent = "Il pagamento non è stato completato. Puoi riprovare dalla pagina dell'evento.";
  }
})();
