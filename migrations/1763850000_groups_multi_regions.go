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

		field := groups.Fields.GetByName("region")
		relation, ok := field.(*core.RelationField)
		if !ok {
			return nil
		}

		relation.MaxSelect = 0

		return app.Save(groups)
	}, func(app core.App) error {
		groups, err := app.FindCollectionByNameOrId("groups")
		if err != nil {
			return err
		}

		field := groups.Fields.GetByName("region")
		relation, ok := field.(*core.RelationField)
		if !ok {
			return nil
		}

		relation.MaxSelect = 1

		return app.Save(groups)
	})
}
