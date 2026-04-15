package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		tokens, err := app.FindCollectionByNameOrId("tokens")
		if err != nil {
			return err
		}
		groups, err := app.FindCollectionByNameOrId("groups")
		if err != nil {
			return err
		}

		if tokens.Fields.GetByName("group") != nil {
			return nil
		}

		tokens.Fields.Add(&core.RelationField{
			Name:         "group",
			CollectionId: groups.Id,
			MaxSelect:    1,
		})
		return app.Save(tokens)
	}, func(app core.App) error {
		tokens, err := app.FindCollectionByNameOrId("tokens")
		if err != nil {
			return err
		}
		if field := tokens.Fields.GetByName("group"); field != nil {
			tokens.Fields.RemoveById(field.GetId())
			return app.Save(tokens)
		}
		return nil
	})
}
