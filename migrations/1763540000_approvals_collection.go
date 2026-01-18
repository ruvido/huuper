package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		groups, err := app.FindCollectionByNameOrId("groups")
		if err != nil {
			return err
		}

		approvals := core.NewBaseCollection("approvals")
		approvals.ListRule = nil
		approvals.ViewRule = nil
		approvals.CreateRule = nil
		approvals.UpdateRule = nil
		approvals.DeleteRule = nil

		approvals.Fields.Add(
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
			&core.DateField{
				Name:     "leader_approved_at",
				Required: false,
			},
			&core.DateField{
				Name:     "admin_confirmed_at",
				Required: false,
			},
		)

		approvals.AddIndex("idx_approvals_user", true, "user", "")

		return app.Save(approvals)
	}, func(app core.App) error {
		approvals, err := app.FindCollectionByNameOrId("approvals")
		if err != nil {
			return err
		}
		return app.Delete(approvals)
	})
}
