package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		col, err := app.FindCollectionByNameOrId("battleplans")
		if err != nil {
			return err
		}
		field, ok := col.Fields.GetByName("status").(*core.SelectField)
		if !ok {
			return nil
		}
		values := make([]string, 0, len(field.Values))
		for _, v := range field.Values {
			if v == "abandoned" {
				values = append(values, "archived")
			} else {
				values = append(values, v)
			}
		}
		field.Values = values
		return app.Save(col)
	}, func(app core.App) error {
		col, err := app.FindCollectionByNameOrId("battleplans")
		if err != nil {
			return err
		}
		field, ok := col.Fields.GetByName("status").(*core.SelectField)
		if !ok {
			return nil
		}
		values := make([]string, 0, len(field.Values))
		for _, v := range field.Values {
			if v == "archived" {
				values = append(values, "abandoned")
			} else {
				values = append(values, v)
			}
		}
		field.Values = values
		return app.Save(col)
	})
}
