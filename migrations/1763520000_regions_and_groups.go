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

		groups, err := app.FindCollectionByNameOrId("groups")
		if err != nil {
			return err
		}

		// Regions collection (readable by authenticated users)
		regions := core.NewBaseCollection("regions")
		regions.ListRule = types.Pointer("@request.auth.id != ''")
		regions.ViewRule = types.Pointer("@request.auth.id != ''")
		regions.CreateRule = nil
		regions.UpdateRule = nil
		regions.DeleteRule = nil

		regions.Fields.Add(
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
		)

		regions.AddIndex("idx_regions_name", true, "name", "")

		if err := app.Save(regions); err != nil {
			return err
		}

		// Extend groups with region + leader + is_open
		groups.Fields.Add(
			&core.RelationField{
				Name:         "region",
				Required:     false,
				CollectionId: regions.Id,
				MaxSelect:    1,
			},
			&core.RelationField{
				Name:         "leader",
				Required:     false,
				CollectionId: users.Id,
				MaxSelect:    1,
			},
			&core.BoolField{
				Name:     "is_open",
				Required: true,
			},
		)

		if err := app.Save(groups); err != nil {
			return err
		}

		// Default existing groups to open.
		existingGroups, err := app.FindRecordsByFilter("groups", "", "", 0, 0)
		if err != nil {
			return err
		}

		for _, group := range existingGroups {
			if group.GetBool("is_open") {
				continue
			}
			group.Set("is_open", true)
			if err := app.Save(group); err != nil {
				return err
			}
		}

		return nil
	}, func(app core.App) error {
		groups, err := app.FindCollectionByNameOrId("groups")
		if err != nil {
			return err
		}

		for _, fieldName := range []string{"region", "leader", "is_open"} {
			field := groups.Fields.GetByName(fieldName)
			if field != nil {
				groups.Fields.RemoveById(field.GetId())
			}
		}

		if err := app.Save(groups); err != nil {
			return err
		}

		regions, err := app.FindCollectionByNameOrId("regions")
		if err != nil {
			return err
		}

		return app.Delete(regions)
	})
}
