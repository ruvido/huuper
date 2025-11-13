package migrations

import (
	"encoding/json"

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

		// Delete existing signup config if it exists (cleanup old config)
		existingSignup, err := app.FindFirstRecordByFilter(
			"settings",
			"name = 'signup'",
			map[string]any{},
		)
		if err == nil && existingSignup != nil {
			if err := app.Delete(existingSignup); err != nil {
				return err
			}
		}

		// Delete existing onboarding config if it exists
		existingOnboarding, err := app.FindFirstRecordByFilter(
			"settings",
			"name = 'onboarding'",
			map[string]any{},
		)
		if err == nil && existingOnboarding != nil {
			if err := app.Delete(existingOnboarding); err != nil {
				return err
			}
		}

		// Create onboarding multistep config
		onboardingConfig := map[string]any{
			"steps": []map[string]any{
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
					"type":    "checkboxes",
					"field":   "hobbies",
					"label":   "Seleziona i tuoi interessi",
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
			},
		}

		onboardingJSON, _ := json.Marshal(onboardingConfig)

		onboardingRecord := core.NewRecord(settings)
		onboardingRecord.Set("name", "onboarding")
		onboardingRecord.Set("data", string(onboardingJSON))
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
