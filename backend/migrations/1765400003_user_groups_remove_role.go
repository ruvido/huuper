package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		userGroups, err := app.FindCollectionByNameOrId("user_groups")
		if err != nil {
			return err
		}

		if field := userGroups.Fields.GetByName("role"); field != nil {
			userGroups.Fields.RemoveById(field.GetId())
			return app.Save(userGroups)
		}

		return nil
	}, func(app core.App) error {
		userGroups, err := app.FindCollectionByNameOrId("user_groups")
		if err != nil {
			return err
		}

		if userGroups.Fields.GetByName("role") != nil {
			return nil
		}

		userGroups.Fields.Add(&core.SelectField{
			Name:     "role",
			Required: true,
			Values:   []string{"member", "admin"},
		})

		return app.Save(userGroups)
	})
}
