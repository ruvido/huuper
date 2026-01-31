package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(app core.App) error {
		templates, err := app.FindCollectionByNameOrId("templates")
		if err != nil {
			return err
		}

		events := core.NewBaseCollection("events")
		events.ListRule = types.Pointer("") // Public read
		events.ViewRule = types.Pointer("") // Public read
		events.CreateRule = nil             // Admin only
		events.UpdateRule = nil             // Admin only
		events.DeleteRule = nil             // Admin only

		events.Fields.Add(
			&core.AutodateField{
				Name:     "created",
				OnCreate: true,
			},
			&core.AutodateField{
				Name:     "updated",
				OnCreate: true,
				OnUpdate: true,
			},
			&core.DateField{
				Name:     "event_date",
				Required: true,
			},
			&core.TextField{
				Name:     "title",
				Required: true,
				Max:      200,
			},
			&core.TextField{
				Name:     "slug",
				Required: true,
				Max:      200,
			},
			&core.BoolField{
				Name:     "active",
				Required: false,
			},
			&core.RelationField{
				Name:         "reply_template",
				Required:     false,
				CollectionId: templates.Id,
				MaxSelect:    1,
			},
			&core.JSONField{
				Name:     "data",
				Required: false,
			},
		)

		if err := app.Save(events); err != nil {
			return err
		}

		registrations := core.NewBaseCollection("event_registrations")
		registrations.ListRule = nil   // Admin only
		registrations.ViewRule = nil   // Admin only
		registrations.CreateRule = nil // Admin only
		registrations.UpdateRule = nil // Admin only
		registrations.DeleteRule = nil // Admin only

		registrations.Fields.Add(
			&core.AutodateField{
				Name:     "created",
				OnCreate: true,
			},
			&core.RelationField{
				Name:         "event",
				Required:     true,
				CollectionId: events.Id,
				MaxSelect:    1,
			},
			&core.JSONField{
				Name:     "data",
				Required: true,
			},
		)

		if err := app.Save(registrations); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		registrations, err := app.FindCollectionByNameOrId("event_registrations")
		if err == nil {
			if err := app.Delete(registrations); err != nil {
				return err
			}
		}

		events, err := app.FindCollectionByNameOrId("events")
		if err != nil {
			return err
		}
		return app.Delete(events)
	})
}
