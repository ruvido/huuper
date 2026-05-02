package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		bp, err := app.FindCollectionByNameOrId("battleplans")
		if err != nil {
			return err
		}
		field, ok := bp.Fields.GetByName("status").(*core.SelectField)
		if !ok || field == nil {
			return nil
		}
		for _, v := range field.Values {
			if v == "draft" {
				return nil
			}
		}
		field.Values = append(field.Values, "draft")
		return app.Save(bp)
	}, func(app core.App) error {
		bp, err := app.FindCollectionByNameOrId("battleplans")
		if err != nil {
			return err
		}
		field, ok := bp.Fields.GetByName("status").(*core.SelectField)
		if !ok || field == nil {
			return nil
		}
		filtered := make([]string, 0, len(field.Values))
		for _, v := range field.Values {
			if v == "draft" {
				continue
			}
			filtered = append(filtered, v)
		}
		field.Values = filtered
		return app.Save(bp)
	})
}
