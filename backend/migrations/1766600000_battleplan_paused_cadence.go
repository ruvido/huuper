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
			"name = 'battleplan'",
			map[string]any{},
		)
		if err != nil || record == nil {
			return err
		}

		data := backendinternal.ParseJSONMap(record.Get("data"))
		data["cadences"] = []map[string]any{
			{"type": "paused", "label": "In pausa", "default": true},
			{"type": "daily", "label": "Ogni giorno"},
			{"type": "specific_days", "label": "Giorni specifici"},
			{"type": "times_per_week", "label": "Volte a settimana"},
		}
		record.Set("data", data)
		return app.Save(record)
	}, func(app core.App) error {
		return nil
	})
}
