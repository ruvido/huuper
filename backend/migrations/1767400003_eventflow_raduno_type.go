package migrations

import (
	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		record, err := app.FindFirstRecordByFilter(
			"settings",
			"name = 'eventflow'",
			map[string]any{},
		)
		if err != nil || record == nil {
			return err
		}

		config := backendinternal.ParseJSONMap(record.Get("data"))
		types, _ := config["types"].([]any)
		for _, raw := range types {
			if t, ok := raw.(map[string]any); ok && t["value"] == "raduno" {
				return nil
			}
		}

		types = append(types, map[string]any{
			"value":       "raduno",
			"label":       "Raduno",
			"description": "Multi-day gathering with deposit-gated registration.",
			"creators":    []string{"admin", "assistant"},
			"required": map[string]bool{
				"title":    true,
				"location": true,
				"end_date": true,
			},
			"registration": map[string]any{
				"enabled":       true,
				"approval":      true,
				"deposit_cents": 5000,
			},
		})
		config["types"] = types
		record.Set("data", config)
		return app.Save(record)
	}, func(app core.App) error {
		record, err := app.FindFirstRecordByFilter(
			"settings",
			"name = 'eventflow'",
			map[string]any{},
		)
		if err != nil || record == nil {
			return err
		}

		config := backendinternal.ParseJSONMap(record.Get("data"))
		types, _ := config["types"].([]any)
		filtered := make([]any, 0, len(types))
		for _, raw := range types {
			if t, ok := raw.(map[string]any); ok && t["value"] == "raduno" {
				continue
			}
			filtered = append(filtered, raw)
		}
		config["types"] = filtered
		record.Set("data", config)
		return app.Save(record)
	})
}
