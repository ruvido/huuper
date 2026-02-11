package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		if field := users.Fields.GetByName("name"); field != nil {
			users.Fields.RemoveById(field.GetId())
			return app.Save(users)
		}

		return nil
	}, func(app core.App) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		if users.Fields.GetByName("name") == nil {
			users.Fields.Add(&core.TextField{
				Name:     "name",
				Required: false,
				Max:      200,
			})
			return app.Save(users)
		}

		return nil
	})
}
