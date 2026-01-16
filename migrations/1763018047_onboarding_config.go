package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Get settings collection
		settings, err := app.FindCollectionByNameOrId("settings")
		if err != nil {
			return err
		}

		// Create onboarding multistep config if missing
		existingOnboarding, err := app.FindFirstRecordByFilter(
			"settings",
			"name = 'onboarding'",
			map[string]any{},
		)
		if err == nil && existingOnboarding != nil {
			return nil
		}

		onboardingConfig := map[string]any{
			"steps": []map[string]any{
				{
					"id":     "start",
					"title":  "Benvenuto in realmen",
					"type":   "start",
					"text":   "Per iniziare a usare i gruppi<br />completa il tuo profilo",
					"button": "INIZIA",
				},
				{
					"id":    "why",
					"title": "Perché vuoi unirti?",
					"type":  "textarea",
					"field": "why",
					"label": "Raccontaci perché vuoi far parte della community",
				},
				{
					"id":      "hobbies",
					"title":   "I tuoi interessi",
					"type":    "select",
					"field":   "hobbies",
					"label":   "Seleziona i tuoi interessi",
					"min":     1,
					"options": []string{
						"Musica",
						"Sport",
						"Tecnologia",
						"Arte",
						"Viaggi",
						"Lettura",
					},
				},
				{
					"id":    "avatar",
					"title": "Foto profilo",
					"type":  "file",
					"field": "avatar",
					"label": "Carica la tua foto profilo",
				},
				{
					"id":     "confirmation",
					"type":   "confirmation",
					"title":  "Tutto pronto!",
					"text":   "Hai completato il tuo profilo.\n\nClicca il pulsante per inviare la tua richiesta di iscrizione.",
					"button": "Iscriviti!",
				},
			},
		}

		onboardingRecord := core.NewRecord(settings)
		onboardingRecord.Set("name", "onboarding")
		onboardingRecord.Set("data", onboardingConfig)
		if err := app.Save(onboardingRecord); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		// Downgrade: delete onboarding config
		onboardingRecord, err := app.FindFirstRecordByFilter(
			"settings",
			"name = 'onboarding'",
			map[string]any{},
		)
		if err == nil && onboardingRecord != nil {
			app.Delete(onboardingRecord)
		}
		return nil
	})
}
