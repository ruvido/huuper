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

		// Delete existing signup config if it exists
		existingRecord, err := app.FindFirstRecordByFilter(
			"settings",
			"name = 'signup'",
			map[string]any{},
		)
		if err == nil && existingRecord != nil {
			if err := app.Delete(existingRecord); err != nil {
				return err
			}
		}

		// Create signup multistep config
		signupConfig := map[string]any{
			"steps": []map[string]any{
				{
					"id":    "credentials",
					"title": "Account",
					"type":  "form",
					"fields": []map[string]string{
						{"name": "email", "type": "email", "label": "Email"},
						{"name": "password", "type": "password", "label": "Password"},
						{"name": "passwordConfirm", "type": "password", "label": "Confirm Password"},
					},
				},
				{
					"id":    "why",
					"title": "Why join?",
					"type":  "textarea",
					"field": "data.why",
					"label": "Tell us why you want to join",
				},
				{
					"id":    "avatar",
					"title": "Avatar",
					"type":  "file",
					"field": "avatar",
					"label": "Upload your profile picture",
				},
				{
					"id":    "hobbies",
					"title": "Interests",
					"type":  "checkboxes",
					"field": "data.hobbies",
					"label": "Select your interests",
					"options": []string{
						"Music",
						"Sports",
						"Technology",
						"Art",
						"Travel",
						"Reading",
					},
				},
			},
		}

		signupJSON, _ := json.Marshal(signupConfig)

		signupRecord := core.NewRecord(settings)
		signupRecord.Set("name", "signup")
		signupRecord.Set("data", string(signupJSON))
		if err := app.Save(signupRecord); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		// Downgrade: delete signup config
		signupRecord, err := app.FindFirstRecordByFilter(
			"settings",
			"name = 'signup'",
			map[string]any{},
		)
		if err == nil && signupRecord != nil {
			app.Delete(signupRecord)
		}
		return nil
	})
}
