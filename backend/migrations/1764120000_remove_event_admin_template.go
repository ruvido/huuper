package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		events, err := app.FindCollectionByNameOrId("events")
		if err != nil {
			return err
		}

		if field := events.Fields.GetByName("admin_template"); field != nil {
			events.Fields.RemoveById(field.GetId())
			return app.Save(events)
		}

		return nil
	}, func(app core.App) error {
		// No downgrade
		return nil
	})
}
