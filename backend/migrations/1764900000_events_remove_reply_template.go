package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		events, err := app.FindCollectionByNameOrId("events")
		if err != nil {
			return err
		}

		if field := events.Fields.GetByName("reply_template"); field != nil {
			events.Fields.RemoveById(field.GetId())
			return app.Save(events)
		}

		return nil
	}, func(app core.App) error {
		events, err := app.FindCollectionByNameOrId("events")
		if err != nil {
			return err
		}

		templates, err := app.FindCollectionByNameOrId("templates")
		if err != nil {
			return err
		}

		if events.Fields.GetByName("reply_template") == nil {
			events.Fields.Add(&core.RelationField{
				Name:         "reply_template",
				Required:     false,
				CollectionId: templates.Id,
				MaxSelect:    1,
			})
			return app.Save(events)
		}

		return nil
	})
}
