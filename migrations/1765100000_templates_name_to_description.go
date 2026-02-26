package migrations

import (
	"strings"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		templates, err := app.FindCollectionByNameOrId("templates")
		if err != nil {
			return err
		}

		if templates.Fields.GetByName("description") == nil {
			templates.Fields.Add(&core.TextField{
				Name:     "description",
				Required: false,
				Max:      500,
			})
			if err := app.Save(templates); err != nil {
				return err
			}
		}

		records, err := app.FindRecordsByFilter("templates", "", "", 0, 0)
		if err != nil {
			return err
		}
		for _, record := range records {
			description := strings.TrimSpace(record.GetString("description"))
			if description != "" {
				continue
			}
			record.Set("description", strings.TrimSpace(record.GetString("name")))
			if err := app.Save(record); err != nil {
				return err
			}
		}

		templates, err = app.FindCollectionByNameOrId("templates")
		if err != nil {
			return err
		}
		if field := templates.Fields.GetByName("name"); field != nil {
			templates.Fields.RemoveById(field.GetId())
			return app.Save(templates)
		}

		return nil
	}, func(app core.App) error {
		templates, err := app.FindCollectionByNameOrId("templates")
		if err != nil {
			return err
		}

		if templates.Fields.GetByName("name") == nil {
			templates.Fields.Add(&core.TextField{
				Name:     "name",
				Required: true,
				Max:      200,
			})
			if err := app.Save(templates); err != nil {
				return err
			}
		}

		records, err := app.FindRecordsByFilter("templates", "", "", 0, 0)
		if err != nil {
			return err
		}
		for _, record := range records {
			name := strings.TrimSpace(record.GetString("name"))
			if name != "" {
				continue
			}
			record.Set("name", strings.TrimSpace(record.GetString("description")))
			if err := app.Save(record); err != nil {
				return err
			}
		}

		templates, err = app.FindCollectionByNameOrId("templates")
		if err != nil {
			return err
		}
		if field := templates.Fields.GetByName("description"); field != nil {
			templates.Fields.RemoveById(field.GetId())
			return app.Save(templates)
		}

		return nil
	})
}
