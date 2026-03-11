package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		groups, err := app.FindCollectionByNameOrId("groups")
		if err != nil {
			return err
		}

		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		if groups.Fields.GetByName("assistant") != nil {
			return nil
		}

		groups.Fields.Add(&core.RelationField{
			Name:         "assistant",
			Required:     false,
			CollectionId: users.Id,
			MaxSelect:    1,
		})

		return app.Save(groups)
	}, func(app core.App) error {
		groups, err := app.FindCollectionByNameOrId("groups")
		if err != nil {
			return nil
		}

		if field := groups.Fields.GetByName("assistant"); field != nil {
			groups.Fields.RemoveById(field.GetId())
			return app.Save(groups)
		}

		return nil
	})
}
