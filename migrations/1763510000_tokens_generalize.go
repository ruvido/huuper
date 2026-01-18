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

		tokens.Fields.Add(
			&core.DateField{
				Name:     "expires_at",
				Required: true,
			},
			&core.DateField{
				Name:     "used_at",
				Required: false,
			},
			&core.RelationField{
				Name:         "group",
				Required:     false,
				CollectionId: groups.Id,
				MaxSelect:    1,
			},
			&core.JSONField{
				Name:     "meta",
				Required: false,
			},
		)

		return app.Save(tokens)
	}, func(app core.App) error {
		tokens, err := app.FindCollectionByNameOrId("tokens")
		if err != nil {
			return err
		}

		for _, fieldName := range []string{"expires_at", "used_at", "group", "meta"} {
			field := tokens.Fields.GetByName(fieldName)
			if field != nil {
				tokens.Fields.RemoveById(field.GetId())
			}
		}

		return app.Save(tokens)
	})
}
