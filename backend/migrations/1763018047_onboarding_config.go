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
			"start_page": map[string]any{
				"title":  "Benvenuto in realmen",
				"text":   "Ti faremo alcune domande per completare il tuo profilo.\nSei pronto?",
				"button": "INIZIA",
			},
			"steps": []map[string]any{
				{
					"field": "work",
					"title": "In che campo lavori?",
				},
				{
					"field": "skills",
					"title": "Le tue skill",
					"label": "Nel lavoro o tempo libero, cosa sai fare con le mani?",
				},
				{
					"field": "interests",
					"title": "I tuoi interessi",
					"label": "Cosa ti appassiona? Quali sono i tuoi hobby?",
				},
				{
					"field": "sports",
					"title": "I tuoi sport",
					"label": "Dove ti piace metterti alla prova?",
				},
				{
					"field": "avatar",
					"title": "La tua foto!",
					"label": "Fatti vedere, così possiamo riconoscerti!",
				},
			},
			"confirmation": map[string]any{
				"title":  "Tutto pronto!",
				"text":   "Hai completato il tuo profilo.\n\nClicca il pulsante per finalizzare l'accesso.",
				"button": "Completa",
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
