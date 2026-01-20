package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		groups, err := app.FindCollectionByNameOrId("groups")
		if err != nil {
			return err
		}

		regions, err := app.FindCollectionByNameOrId("regions")
		if err != nil {
			return err
		}

		if field := groups.Fields.GetByName("regions"); field != nil {
			if relation, ok := field.(*core.RelationField); ok {
				relation.CollectionId = regions.Id
				relation.MaxSelect = 99
			}
		} else {
			groups.Fields.Add(&core.RelationField{
				Name:         "regions",
				Required:     false,
				CollectionId: regions.Id,
				MaxSelect:    99,
			})
		}

		if err := app.Save(groups); err != nil {
			return err
		}

		existingGroups, err := app.FindRecordsByFilter("groups", "", "", 0, 0)
		if err != nil {
			return err
		}

		for _, group := range existingGroups {
			regionID := group.GetString("region")
			if regionID == "" {
				continue
			}

			group.Set("regions", []string{regionID})
			if err := app.Save(group); err != nil {
				return err
			}
		}

		if field := groups.Fields.GetByName("region"); field != nil {
			groups.Fields.RemoveById(field.GetId())
			if err := app.Save(groups); err != nil {
				return err
			}
		}

		return nil
	}, func(app core.App) error {
		groups, err := app.FindCollectionByNameOrId("groups")
		if err != nil {
			return err
		}

		regions, err := app.FindCollectionByNameOrId("regions")
		if err != nil {
			return err
		}

		if field := groups.Fields.GetByName("region"); field != nil {
			if relation, ok := field.(*core.RelationField); ok {
				relation.CollectionId = regions.Id
				relation.MaxSelect = 1
			}
		} else {
			groups.Fields.Add(&core.RelationField{
				Name:         "region",
				Required:     false,
				CollectionId: regions.Id,
				MaxSelect:    1,
			})
		}

		if err := app.Save(groups); err != nil {
			return err
		}

		existingGroups, err := app.FindRecordsByFilter("groups", "", "", 0, 0)
		if err != nil {
			return err
		}

		for _, group := range existingGroups {
			regionIDs := group.GetStringSlice("regions")
			if len(regionIDs) == 0 {
				continue
			}

			group.Set("region", regionIDs[0])
			if err := app.Save(group); err != nil {
				return err
			}
		}

		if field := groups.Fields.GetByName("regions"); field != nil {
			groups.Fields.RemoveById(field.GetId())
			if err := app.Save(groups); err != nil {
				return err
			}
		}

		return nil
	})
}
