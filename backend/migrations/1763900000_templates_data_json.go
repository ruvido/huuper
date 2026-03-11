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

		if templates.Fields.GetByName("data") == nil {
			templates.Fields.Add(&core.JSONField{
				Name:     "data",
				Required: true,
			})
		}

		if bodyField := templates.Fields.GetByName("body"); bodyField != nil {
			templates.Fields.RemoveById(bodyField.GetId())
		}

		return app.Save(templates)
	}, func(app core.App) error {
		templates, err := app.FindCollectionByNameOrId("templates")
		if err != nil {
			return err
		}

		if dataField := templates.Fields.GetByName("data"); dataField != nil {
			templates.Fields.RemoveById(dataField.GetId())
		}

		if templates.Fields.GetByName("body") == nil {
			templates.Fields.Add(&core.TextField{
				Name:     "body",
				Required: true,
				Max:      20000,
			})
		}

		return app.Save(templates)
	})
}
