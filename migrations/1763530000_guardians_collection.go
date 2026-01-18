package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(app core.App) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		guardians := core.NewBaseCollection("guardians")
		guardians.ListRule = types.Pointer("@request.auth.admin = true || leader = @request.auth.id")
		guardians.ViewRule = types.Pointer("@request.auth.admin = true || leader = @request.auth.id")
		guardians.CreateRule = types.Pointer("@request.auth.admin = true || @request.data.leader = @request.auth.id")
		guardians.UpdateRule = types.Pointer("@request.auth.admin = true || leader = @request.auth.id")
		guardians.DeleteRule = types.Pointer("@request.auth.admin = true || leader = @request.auth.id")

		guardians.Fields.Add(
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
				Name:         "guardian",
				Required:     true,
				CollectionId: users.Id,
				MaxSelect:    1,
			},
			&core.RelationField{
				Name:         "leader",
				Required:     true,
				CollectionId: users.Id,
				MaxSelect:    1,
			},
			// steps format: { "step1": { "done": true, "at": "2026-01-12T10:00:00Z" }, ... }
			&core.JSONField{
				Name:     "steps",
				Required: false,
			},
			&core.TextField{
				Name:     "notes",
				Required: false,
				Max:      2000,
			},
		)

		guardians.AddIndex("idx_guardians_user", true, "user", "")

		return app.Save(guardians)
	}, func(app core.App) error {
		guardians, err := app.FindCollectionByNameOrId("guardians")
		if err != nil {
			return err
		}
		return app.Delete(guardians)
	})
}
