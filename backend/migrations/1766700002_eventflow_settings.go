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
					"value":       "call",
					"label":       "Call",
					"description": "Online call.",
					"creators":    []string{"admin", "assistant"},
					"required": map[string]bool{
						"title":    false,
						"url":      true,
						"location": false,
						"group":    false,
						"end_date": false,
					},
					"registration": map[string]bool{
						"enabled":  true,
						"approval": false,
					},
				},
				{
					"value":       "meetup",
					"label":       "Meetup",
					"description": "In-person meetup.",
					"creators":    []string{"admin", "assistant"},
					"required": map[string]bool{
						"title":    false,
						"url":      false,
						"location": true,
						"group":    false,
						"end_date": false,
					},
					"registration": map[string]bool{
						"enabled":  true,
						"approval": false,
					},
				},
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
