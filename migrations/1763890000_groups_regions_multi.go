package migrations

import (
	"fmt"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("groups")
		if err != nil {
			return err
		}

		field := collection.Fields.GetByName("regions")
		if field == nil {
			return fmt.Errorf("missing groups.regions field")
		}

		relation, ok := field.(*core.RelationField)
		if !ok {
			return fmt.Errorf("groups.regions is not a relation field")
		}

		relation.MaxSelect = 99

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("groups")
		if err != nil {
			return err
		}

		field := collection.Fields.GetByName("regions")
		if field == nil {
			return fmt.Errorf("missing groups.regions field")
		}

		relation, ok := field.(*core.RelationField)
		if !ok {
			return fmt.Errorf("groups.regions is not a relation field")
		}

		relation.MaxSelect = 1

		return app.Save(collection)
	})
}
