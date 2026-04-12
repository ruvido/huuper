package migrations

import (
	"strings"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		groups, err := app.FindCollectionByNameOrId("groups")
		if err != nil {
			return err
		}

		switch field := groups.Fields.GetByName("type").(type) {
		case *core.SelectField:
			field.Values = []string{"default", "local", "public", "private"}
			field.Required = false
		case nil:
			groups.Fields.Add(&core.SelectField{
				Name:     "type",
				Required: false,
				Values:   []string{"default", "local", "public", "private"},
			})
		default:
			groups.Fields.RemoveById(field.GetId())
			groups.Fields.Add(&core.SelectField{
				Name:     "type",
				Required: false,
				Values:   []string{"default", "local", "public", "private"},
			})
		}

		if err := app.Save(groups); err != nil {
			return err
		}

		records, err := app.FindRecordsByFilter("groups", "", "", 0, 0)
		if err != nil {
			return err
		}

		allowed := map[string]struct{}{
			"default": {},
			"local":   {},
			"public":  {},
			"private": {},
		}

		for _, record := range records {
			value := strings.TrimSpace(record.GetString("type"))
			if value == "general" || value == "default" {
				record.Set("type", "local")
			} else {
				if _, ok := allowed[value]; !ok {
					record.Set("type", "private")
				}
			}
			if err := app.Save(record); err != nil {
				return err
			}
		}

		return nil
	}, func(app core.App) error {
		groups, err := app.FindCollectionByNameOrId("groups")
		if err != nil {
			return err
		}

		switch field := groups.Fields.GetByName("type").(type) {
		case *core.SelectField:
			field.Values = []string{"general", "local", "public", "private"}
			field.Required = false
		case nil:
			groups.Fields.Add(&core.SelectField{
				Name:     "type",
				Required: false,
				Values:   []string{"general", "local", "public", "private"},
			})
		default:
			groups.Fields.RemoveById(field.GetId())
			groups.Fields.Add(&core.SelectField{
				Name:     "type",
				Required: false,
				Values:   []string{"general", "local", "public", "private"},
			})
		}

		if err := app.Save(groups); err != nil {
			return err
		}

		records, err := app.FindRecordsByFilter("groups", "", "", 0, 0)
		if err != nil {
			return err
		}

		allowed := map[string]struct{}{
			"general": {},
			"local":   {},
			"public":  {},
			"private": {},
		}

		for _, record := range records {
			value := strings.TrimSpace(record.GetString("type"))
			if value == "default" {
				record.Set("type", "")
			} else {
				if _, ok := allowed[value]; !ok {
					record.Set("type", "private")
				}
			}
			if err := app.Save(record); err != nil {
				return err
			}
		}

		return nil
	})
}
