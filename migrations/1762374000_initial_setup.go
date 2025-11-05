package migrations

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(app core.App) error {
		// Load .env for admin credentials
		godotenv.Load()

		// Update users collection with telegram and data fields
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		users.Fields.Add(
			&core.JSONField{
				Name:     "telegram",
				Required: false,
			},
			&core.JSONField{
				Name:     "data",
				Required: false,
			},
		)

		if err := app.Save(users); err != nil {
			return err
		}

		// Create groups collection
		groups := core.NewBaseCollection("groups")
		groups.ListRule = types.Pointer("") // Public read
		groups.ViewRule = types.Pointer("") // Public read
		groups.CreateRule = nil              // Admin only
		groups.UpdateRule = nil              // Admin only
		groups.DeleteRule = nil              // Admin only

		groups.Fields.Add(
			&core.AutodateField{
				Name:     "created",
				OnCreate: true,
			},
			&core.AutodateField{
				Name:     "updated",
				OnCreate: true,
				OnUpdate: true,
			},
			&core.TextField{
				Name:     "name",
				Required: true,
				Max:      200,
			},
			&core.SelectField{
				Name:     "type",
				Required: true,
				Values:   []string{"telegram", "discord"},
			},
			&core.URLField{
				Name:     "invite_link",
				Required: false,
			},
			&core.TextField{
				Name:     "description",
				Required: false,
				Max:      1000,
			},
			&core.JSONField{
				Name:     "telegram",
				Required: false,
			},
		)

		if err := app.Save(groups); err != nil {
			return err
		}

		// Create user_groups collection
		userGroups := core.NewBaseCollection("user_groups")
		userGroups.ListRule = types.Pointer("@request.auth.id != ''") // Authenticated users
		userGroups.ViewRule = types.Pointer("@request.auth.id != ''") // Authenticated users
		userGroups.CreateRule = nil                                    // Admin only
		userGroups.UpdateRule = nil                                    // Admin only
		userGroups.DeleteRule = nil                                    // Admin only

		userGroups.Fields.Add(
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
				Name:         "user",
				Required:     true,
				CollectionId: users.Id,
				MaxSelect:    1,
			},
			&core.RelationField{
				Name:         "group",
				Required:     true,
				CollectionId: groups.Id,
				MaxSelect:    1,
			},
			&core.SelectField{
				Name:     "role",
				Required: true,
				Values:   []string{"member", "admin"},
			},
		)

		if err := app.Save(userGroups); err != nil {
			return err
		}

		// Create admin superuser
		adminEmail := os.Getenv("ADMIN_EMAIL")
		adminPassword := os.Getenv("ADMIN_PASSWORD")

		if adminEmail != "" && adminPassword != "" {
			superusers, err := app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
			if err != nil {
				return err
			}

			// Check if admin already exists
			existingAdmin, _ := app.FindFirstRecordByFilter(
				core.CollectionNameSuperusers,
				"email = {:email}",
				map[string]any{"email": adminEmail},
			)

			if existingAdmin == nil {
				record := core.NewRecord(superusers)
				record.Set("email", adminEmail)
				record.Set("password", adminPassword)
				if err := app.Save(record); err != nil {
					return err
				}
			}
		}

		return nil
	}, func(app core.App) error {
		// Downgrade: delete user_groups collection
		userGroups, err := app.FindCollectionByNameOrId("user_groups")
		if err == nil {
			if err := app.Delete(userGroups); err != nil {
				return err
			}
		}

		// Downgrade: delete groups collection
		groups, err := app.FindCollectionByNameOrId("groups")
		if err != nil {
			return err
		}
		return app.Delete(groups)
	})
}
