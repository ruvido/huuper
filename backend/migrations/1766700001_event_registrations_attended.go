package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		registrations, err := app.FindCollectionByNameOrId("event_registrations")
		if err != nil {
			return err
		}

		if registrations.Fields.GetByName("attended") == nil {
			registrations.Fields.Add(&core.BoolField{
				Name:     "attended",
				Required: false,
			})
			if err := app.Save(registrations); err != nil {
				return err
			}
		}

		return nil
	}, func(app core.App) error {
		registrations, err := app.FindCollectionByNameOrId("event_registrations")
		if err != nil {
			return err
		}
		if field := registrations.Fields.GetByName("attended"); field != nil {
			registrations.Fields.RemoveById(field.GetId())
			return app.Save(registrations)
		}
		return nil
	})
}
