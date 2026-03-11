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

		if registrations.Fields.GetByName("email") == nil {
			registrations.Fields.Add(&core.TextField{
				Name:     "email",
				Required: false,
				Max:      255,
			})
		}

		return app.Save(registrations)
	}, func(app core.App) error {
		registrations, err := app.FindCollectionByNameOrId("event_registrations")
		if err != nil {
			return err
		}

		if field := registrations.Fields.GetByName("email"); field != nil {
			registrations.Fields.RemoveById(field.GetId())
			return app.Save(registrations)
		}

		return nil
	})
}
