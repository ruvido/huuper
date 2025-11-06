package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Get users collection for relation
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		// Create tokens collection
		tokens := core.NewBaseCollection("tokens")
		tokens.ListRule = nil   // Admin only
		tokens.ViewRule = nil   // Admin only
		tokens.CreateRule = nil // Admin only
		tokens.UpdateRule = nil // Admin only
		tokens.DeleteRule = nil // Admin only

		tokens.Fields.Add(
			&core.AutodateField{
				Name:     "created",
				OnCreate: true,
			},
			&core.TextField{
				Name:     "token",
				Required: true,
				Max:      255,
			},
			&core.RelationField{
				Name:         "user",
				Required:     true,
				CollectionId: users.Id,
				MaxSelect:    1,
			},
			&core.TextField{
				Name:     "service",
				Required: true,
				Max:      50,
			},
		)

		// Add unique index on token field
		tokens.AddIndex("idx_token", true, "token", "")

		if err := app.Save(tokens); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		// Downgrade: delete tokens collection
		tokens, err := app.FindCollectionByNameOrId("tokens")
		if err != nil {
			return err
		}
		return app.Delete(tokens)
	})
}
