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

		field, ok := groups.Fields.GetByName("type").(*core.SelectField)
		if !ok || field == nil {
			return nil
		}

		for _, v := range field.Values {
			if v == "beta_testers" {
				return nil
			}
		}

		field.Values = append(field.Values, "beta_testers")
		return app.Save(groups)
	}, func(app core.App) error {
		groups, err := app.FindCollectionByNameOrId("groups")
		if err != nil {
			return err
		}

		field, ok := groups.Fields.GetByName("type").(*core.SelectField)
		if !ok || field == nil {
			return nil
		}

		filtered := make([]string, 0, len(field.Values))
		for _, v := range field.Values {
			if v == "beta_testers" {
				continue
			}
			filtered = append(filtered, v)
		}
		field.Values = filtered
		return app.Save(groups)
	})
}
