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

		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		if registrations.Fields.GetByName("user") != nil {
			return nil
		}

		registrations.Fields.Add(&core.RelationField{
			Name:         "user",
			Required:     false,
			CollectionId: users.Id,
			MaxSelect:    1,
		})

		return app.Save(registrations)
	}, func(app core.App) error {
		registrations, err := app.FindCollectionByNameOrId("event_registrations")
		if err != nil {
			return err
		}

		if field := registrations.Fields.GetByName("user"); field != nil {
			registrations.Fields.RemoveById(field.GetId())
			return app.Save(registrations)
		}

		return nil
	})
}
