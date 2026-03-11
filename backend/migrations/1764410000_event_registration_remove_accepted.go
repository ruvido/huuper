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

		if field := registrations.Fields.GetByName("accepted"); field != nil {
			registrations.Fields.RemoveById(field.GetId())
			return app.Save(registrations)
		}

		return nil
	}, func(app core.App) error {
		registrations, err := app.FindCollectionByNameOrId("event_registrations")
		if err != nil {
			return err
		}

		if registrations.Fields.GetByName("accepted") == nil {
			registrations.Fields.Add(&core.BoolField{
				Name:     "accepted",
				Required: false,
			})
			return app.Save(registrations)
		}

		return nil
	})
}
