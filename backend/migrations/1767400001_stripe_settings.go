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
			"name = 'stripe'",
			map[string]any{},
		)
		if existing != nil {
			return nil
		}

		record := core.NewRecord(settings)
		record.Set("name", "stripe")
		record.Set("data", map[string]any{
			"secret_key":      "",
			"publishable_key": "",
			"webhook_secret":  "",
		})
		return app.Save(record)
	}, func(app core.App) error {
		record, _ := app.FindFirstRecordByFilter(
			"settings",
			"name = 'stripe'",
			map[string]any{},
		)
		if record != nil {
			return app.Delete(record)
		}
		return nil
	})
}
