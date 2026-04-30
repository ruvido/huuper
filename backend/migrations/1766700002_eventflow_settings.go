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

		existing, _ := app.FindFirstRecordByFilter(
			"settings",
			"name = 'eventflow'",
			map[string]any{},
		)
		if existing != nil {
			return nil
		}

		config := map[string]any{
			"types": []map[string]any{
				{
					"value":                 "rally",
					"creator":               "admin",
					"requires_group":        false,
					"registration_approval": true,
					"registration_scope":    "occurrence",
					"requires_title":        true,
				},
				{
					"value":                 "call",
					"creator":               "admin_or_assistant",
					"requires_group":        false,
					"registration_approval": false,
					"registration_scope":    "series",
					"requires_title":        false,
				},
				{
					"value":                 "meetup",
					"creator":               "admin_or_assistant",
					"requires_group":        true,
					"registration_approval": false,
					"registration_scope":    "occurrence",
					"requires_title":        false,
				},
			},
			"recurrence": map[string]any{
				"max_occurrences": 52,
				"rules": []map[string]any{
					{"type": "weekly"},
					{"type": "biweekly"},
					{"type": "monthly"},
				},
			},
			"list": map[string]any{
				"default_window":  "future",
				"collapse_series": true,
			},
		}

		record := core.NewRecord(settings)
		record.Set("name", "eventflow")
		record.Set("data", config)
		return app.Save(record)
	}, func(app core.App) error {
		record, _ := app.FindFirstRecordByFilter(
			"settings",
			"name = 'eventflow'",
			map[string]any{},
		)
		if record != nil {
			return app.Delete(record)
		}
		return nil
	})
}
