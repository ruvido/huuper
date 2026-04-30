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
		field := events.Fields.GetByName("title")
		if field == nil {
			return nil
		}
		text, ok := field.(*core.TextField)
		if !ok {
			return nil
		}
		if !text.Required {
			return nil
		}
		text.Required = false
		return app.Save(events)
	}, func(app core.App) error {
		events, err := app.FindCollectionByNameOrId("events")
		if err != nil {
			return err
		}
		field := events.Fields.GetByName("title")
		if field == nil {
			return nil
		}
		text, ok := field.(*core.TextField)
		if !ok {
			return nil
		}
		text.Required = true
		return app.Save(events)
	})
}
