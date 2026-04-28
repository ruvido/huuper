package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		if existing, err := app.FindCollectionByNameOrId("battleplans"); err == nil && existing != nil {
			return nil
		}

		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		battleplans := core.NewBaseCollection("battleplans")
		battleplans.ListRule = nil
		battleplans.ViewRule = nil
		battleplans.CreateRule = nil
		battleplans.UpdateRule = nil
		battleplans.DeleteRule = nil

		battleplans.Fields.Add(
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
			&core.RelationField{
				Name:          "user",
				Required:      true,
				CollectionId:  users.Id,
				MaxSelect:     1,
				CascadeDelete: true,
			},
			&core.DateField{Name: "start_date", Required: true},
			&core.DateField{Name: "end_date", Required: true},
			&core.SelectField{
				Name:      "status",
				Required:  true,
				MaxSelect: 1,
				Values:    []string{"active", "completed", "abandoned"},
			},
			&core.SelectField{
				Name:      "visibility",
				Required:  true,
				MaxSelect: 1,
				Values:    []string{"group", "public"},
			},
			&core.JSONField{Name: "data", Required: true},
		)

		battleplans.AddIndex("idx_battleplans_user_active", true, "user", "status = 'active'")
		battleplans.AddIndex("idx_battleplans_end_date", false, "end_date", "")
		battleplans.AddIndex("idx_battleplans_visibility_end", false, "visibility, end_date", "")

		return app.Save(battleplans)
	}, func(app core.App) error {
		bp, err := app.FindCollectionByNameOrId("battleplans")
		if err != nil || bp == nil {
			return nil
		}
		return app.Delete(bp)
	})
}
