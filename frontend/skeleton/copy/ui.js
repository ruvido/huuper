window.huuperCopy = Object.assign({}, window.huuperCopy || {}, {
  ui: {
    admin: {
      dashboard: {
        heroTemplate: "Il Branco sono {users},<br>attivi in {groups} e {regions}",
        metrics: {
          users: "Users",
          requests: "Requests",
          guardians: "Guardians",
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
