package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		templates, err := app.FindCollectionByNameOrId("templates")
		if err != nil {
			return err
		}

		if templates.Fields.GetByName("slug") == nil {
			templates.Fields.Add(&core.TextField{
				Name:     "slug",
				Required: false,
				Max:      200,
			})
		}

		return app.Save(templates)
	}, func(app core.App) error {
		templates, err := app.FindCollectionByNameOrId("templates")
		if err != nil {
			return err
		}

		if field := templates.Fields.GetByName("slug"); field != nil {
			templates.Fields.RemoveById(field.GetId())
			return app.Save(templates)
		}

		return nil
	})
}
