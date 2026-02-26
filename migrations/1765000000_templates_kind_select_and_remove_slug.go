package migrations

import (
	"strings"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

var templateKindAllowedValues = []string{
	"events.user.registration_received",
	"events.user.registration_accepted",
	"events.admin.new_registration",
}

func init() {
	m.Register(func(app core.App) error {
		templates, err := app.FindCollectionByNameOrId("templates")
		if err != nil {
			return err
		}

		switch field := templates.Fields.GetByName("kind").(type) {
		case *core.SelectField:
			field.Values = templateKindAllowedValues
			field.Required = false
		case nil:
			templates.Fields.Add(&core.SelectField{
				Name:     "kind",
				Required: false,
				Values:   templateKindAllowedValues,
			})
		default:
			templates.Fields.RemoveById(field.GetId())
			templates.Fields.Add(&core.SelectField{
				Name:     "kind",
				Required: false,
				Values:   templateKindAllowedValues,
			})
		}

		templates.RemoveIndex("idx_templates_slug")
		if field := templates.Fields.GetByName("slug"); field != nil {
			templates.Fields.RemoveById(field.GetId())
		}

		if err := app.Save(templates); err != nil {
			return err
		}

		records, err := app.FindRecordsByFilter("templates", "", "", 0, 0)
		if err != nil {
			return err
		}

		allowed := make(map[string]struct{}, len(templateKindAllowedValues))
		for _, value := range templateKindAllowedValues {
			allowed[value] = struct{}{}
		}

		for _, record := range records {
			kind := strings.TrimSpace(record.GetString("kind"))
			if kind == "" {
				continue
			}
			if _, ok := allowed[kind]; !ok {
				kind = ""
			}
			record.Set("kind", kind)
			if err := app.Save(record); err != nil {
				return err
			}
		}

		return nil
	}, func(app core.App) error {
		templates, err := app.FindCollectionByNameOrId("templates")
		if err != nil {
			return err
		}

		records, err := app.FindRecordsByFilter("templates", "", "", 0, 0)
		if err != nil {
			return err
		}

		existingKinds := make(map[string]string, len(records))
		for _, record := range records {
			existingKinds[record.Id] = strings.TrimSpace(record.GetString("kind"))
		}

		switch field := templates.Fields.GetByName("kind").(type) {
		case *core.TextField:
			field.Required = false
			field.Max = 200
		case nil:
			templates.Fields.Add(&core.TextField{
				Name:     "kind",
				Required: false,
				Max:      200,
			})
		default:
			templates.Fields.RemoveById(field.GetId())
			templates.Fields.Add(&core.TextField{
				Name:     "kind",
				Required: false,
				Max:      200,
			})
		}

		if templates.Fields.GetByName("slug") == nil {
			templates.Fields.Add(&core.TextField{
				Name:     "slug",
				Required: false,
				Max:      200,
			})
		}

		templates.AddIndex("idx_templates_slug", true, "slug", "slug != ''")
		if err := app.Save(templates); err != nil {
			return err
		}

		for _, record := range records {
			kind := existingKinds[record.Id]
			if kind == "" {
				continue
			}
			record.Set("kind", kind)
			if err := app.Save(record); err != nil {
				return err
			}
		}

		return nil
	})
}
