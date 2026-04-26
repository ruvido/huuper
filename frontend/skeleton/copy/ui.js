window.huuperCopy = Object.assign({}, window.huuperCopy || {}, {
  ui: {
    admin: {
      dashboard: {
        heroTemplate: "Il Branco è di {users},<br>attivi in {regions} regioni",
        metrics: {
          utenti: "Utenti",
          richieste: "Richieste",
          angeli: "Angeli",
        },
      },
      user: {
        cancelDialog: {
          title: "Set user as cancelled?",
          description: "The user will not be visible anymore in groups.",
          confirmLabel: "Cancel user",
          fallbackError: "Cancel unavailable.",
        },
      },
    },
  },
});

