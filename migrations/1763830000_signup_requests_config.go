package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		settings, err := app.FindCollectionByNameOrId("settings")
		if err != nil {
			return err
		}

		signupConfig := map[string]any{
			"steps": []map[string]any{
				{
					"id":     "start",
					"title":  "Request access",
					"type":   "start",
					"text":   "Share a few details. We review every request.",
					"button": "Start",
				},
				{
					"id":    "name",
					"title": "Your name",
					"type":  "text",
					"field": "name",
					"label": "Full name",
				},
				{
					"id":    "email",
					"title": "Your email",
					"type":  "text",
					"field": "email",
					"label": "Email",
					"check_unique": true,
					"error": "Registration request from this email already sent.",
					"error_invalid": "Please enter a valid email address.",
					"error_unavailable": "Unable to verify email right now. Please try again.",
				},
				{
					"id":    "birth_year",
					"title": "Year of birth",
					"type":  "text",
					"field": "birth_year",
					"label": "Year",
				},
				{
					"id":    "region",
					"title": "Your region",
					"type":  "select",
					"field": "region",
					"label": "Region",
					"options_source": map[string]any{
						"collection":  "regions",
						"field":       "name",
						"value_field": "id",
					},
				},
				{
					"id":      "civil_status",
					"title":   "Civil status",
					"type":    "select",
					"field":   "civil_status",
					"label":   "Status",
					"options": []string{"single", "married", "other"},
				},
				{
					"id":    "motivation",
					"title": "Why join?",
					"type":  "textarea",
					"field": "motivation",
					"label": "Tell us why you want to join",
				},
				{
					"id":     "confirm",
					"title":  "Request ready",
					"type":   "confirmation",
					"text":   "We will review your request and get back to you.",
					"button": "Send request",
				},
			},
		}

		existingRecord, err := app.FindFirstRecordByFilter(
			"settings",
			"name = 'signup'",
			map[string]any{},
		)
		if err == nil && existingRecord != nil {
			existingRecord.Set("data", signupConfig)
			return app.Save(existingRecord)
		}

		signupRecord := core.NewRecord(settings)
		signupRecord.Set("name", "signup")
		signupRecord.Set("data", signupConfig)
		return app.Save(signupRecord)
	}, func(app core.App) error {
		return nil
	})
}
