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

		if field := groups.Fields.GetByName("invite_link"); field != nil {
			groups.Fields.RemoveById(field.GetId())
		}

		return app.Save(groups)
	}, func(app core.App) error {
		groups, err := app.FindCollectionByNameOrId("groups")
		if err != nil {
			return err
		}

		if groups.Fields.GetByName("invite_link") == nil {
			groups.Fields.Add(&core.URLField{
				Name:     "invite_link",
				Required: false,
			})
		}

		return app.Save(groups)
	})
}
