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

		field := collection.Fields.GetByName("region")
		if field == nil {
			return fmt.Errorf("missing groups.region field")
		}

		relation, ok := field.(*core.RelationField)
		if !ok {
			return fmt.Errorf("groups.region is not a relation field")
		}

		relation.MaxSelect = 0

		return app.SaveNoValidate(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("groups")
		if err != nil {
			return err
		}

		field := collection.Fields.GetByName("region")
		if field == nil {
			return fmt.Errorf("missing groups.region field")
		}

		relation, ok := field.(*core.RelationField)
		if !ok {
			return fmt.Errorf("groups.region is not a relation field")
		}

		relation.MaxSelect = 1

		return app.SaveNoValidate(collection)
	})
}
