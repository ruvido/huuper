package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// The retreat module's own settings record. Only the deposit chase for now: how
// long to wait before reminding someone who has not paid, and how many times to
// do it. They are settings and not constants because how insistent to be with
// people is a decision that gets revised, and revising it should not need a
// deploy.
func init() {
	m.Register(func(app core.App) error {
		if _, err := app.FindFirstRecordByFilter("settings", "name = 'retreats'"); err == nil {
			return nil // already there
		}
		collection, err := app.FindCollectionByNameOrId("settings")
		if err != nil {
			return err
		}
		record := core.NewRecord(collection)
		record.Set("name", "retreats")
		data, err := json.Marshal(map[string]any{
			"payment_reminder_days": 3,
			"payment_reminder_max":  2,
		})
		if err != nil {
			return err
		}
		record.Set("data", json.RawMessage(data))
		return app.Save(record)
	}, nil)
}
