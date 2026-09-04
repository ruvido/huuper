package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		retreats, err := app.FindCollectionByNameOrId("retreats")
		if err != nil {
			return err
		}

		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		registrations := core.NewBaseCollection("retreat_registrations")
		registrations.ListRule = nil   // Admin only
		registrations.ViewRule = nil   // Admin only
		registrations.CreateRule = nil // Admin only (created server-side)
		registrations.UpdateRule = nil // Admin only (updated server-side)
		registrations.DeleteRule = nil // Admin only

		registrations.Fields.Add(
			&core.AutodateField{
				Name:     "created",
				OnCreate: true,
			},
			&core.AutodateField{
				Name:     "updated",
				OnCreate: true,
				OnUpdate: true,
			},
			&core.RelationField{
				Name:         "retreat",
				Required:     true,
				CollectionId: retreats.Id,
				MaxSelect:    1,
			},
			&core.TextField{
				Name:     "email",
				Required: true,
				Max:      200,
			},
			&core.RelationField{
				Name:         "user",
				Required:     false,
				CollectionId: users.Id,
				MaxSelect:    1,
			},
			&core.SelectField{
				Name:     "status",
				Required: true,
				Values:   []string{"pending", "awaiting_payment", "active", "rejected", "cancelled"},
			},
			&core.TextField{
				Name:     "accept_token",
				Required: false,
				Max:      255,
			},
			&core.DateField{
				Name:     "accept_expires_at",
				Required: false,
			},
			&core.JSONField{
				Name:     "data",
				Required: false,
			},
		)

		registrations.AddIndex("idx_retreat_registrations_retreat_email", true, "retreat, email", "email != ''")
		registrations.AddIndex("idx_retreat_registrations_accept_token", true, "accept_token", "accept_token != ''")

		return app.Save(registrations)
	}, func(app core.App) error {
		registrations, err := app.FindCollectionByNameOrId("retreat_registrations")
		if err != nil {
			return nil
		}
		return app.Delete(registrations)
	})
}
