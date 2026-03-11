package migrations

import (
	"strings"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

var groupTypeAllowedValues = []string{
	"default",
	"local",
	"public",
	"private",
}

var legacyGroupTypeValues = []string{
	"telegram",
	"discord",
}

func init() {
	m.Register(func(app core.App) error {
		groups, err := app.FindCollectionByNameOrId("groups")
		if err != nil {
			return err
		}

		switch field := groups.Fields.GetByName("type").(type) {
		case *core.SelectField:
			field.Values = groupTypeAllowedValues
			field.Required = true
		case nil:
			groups.Fields.Add(&core.SelectField{
				Name:     "type",
				Required: true,
				Values:   groupTypeAllowedValues,
			})
		default:
			groups.Fields.RemoveById(field.GetId())
			groups.Fields.Add(&core.SelectField{
				Name:     "type",
				Required: true,
				Values:   groupTypeAllowedValues,
			})
		}

		if err := app.Save(groups); err != nil {
			return err
		}

		allowed := make(map[string]struct{}, len(groupTypeAllowedValues))
		for _, value := range groupTypeAllowedValues {
			allowed[value] = struct{}{}
		}

		records, err := app.FindRecordsByFilter("groups", "", "", 0, 0)
		if err != nil {
			return err
		}

		for _, record := range records {
			value := strings.TrimSpace(record.GetString("type"))
			if value == "general" {
				record.Set("type", "default")
				if err := app.Save(record); err != nil {
					return err
				}
				continue
			}
			if _, ok := allowed[value]; !ok {
				record.Set("type", "private")
				if err := app.Save(record); err != nil {
					return err
				}
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
			field.Values = legacyGroupTypeValues
			field.Required = true
		case nil:
			groups.Fields.Add(&core.SelectField{
				Name:     "type",
				Required: true,
				Values:   legacyGroupTypeValues,
			})
		default:
			groups.Fields.RemoveById(field.GetId())
			groups.Fields.Add(&core.SelectField{
				Name:     "type",
				Required: true,
				Values:   legacyGroupTypeValues,
			})
		}

		if err := app.Save(groups); err != nil {
			return err
		}

		allowed := make(map[string]struct{}, len(legacyGroupTypeValues))
		for _, value := range legacyGroupTypeValues {
			allowed[value] = struct{}{}
		}

		records, err := app.FindRecordsByFilter("groups", "", "", 0, 0)
		if err != nil {
			return err
		}

		for _, record := range records {
			value := strings.TrimSpace(record.GetString("type"))
			if value == "default" {
				record.Set("type", "telegram")
				if err := app.Save(record); err != nil {
					return err
				}
				continue
			}
			if _, ok := allowed[value]; !ok {
				record.Set("type", "telegram")
				if err := app.Save(record); err != nil {
					return err
				}
			}
		}

		return nil
	})
}
