package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		registrations, err := app.FindCollectionByNameOrId("event_registrations")
		if err != nil {
			return err
		}

		if registrations.Fields.GetByName("status") == nil {
			registrations.Fields.Add(&core.SelectField{
				Name:     "status",
				Required: true,
				Values:   []string{"pending", "active", "cancelled"},
			})
			if err := app.Save(registrations); err != nil {
				return err
			}
		}

		records, err := app.FindRecordsByFilter("event_registrations", "", "", 0, 0)
		if err != nil {
			return err
		}

		for _, record := range records {
			status := record.GetString("status")
			if status != "" {
				continue
			}

			if record.GetString("user") != "" {
				status = "active"
			} else {
				status = "pending"
			}

			record.Set("status", status)
			if err := app.Save(record); err != nil {
				return err
			}
		}

		return nil
	}, func(app core.App) error {
		registrations, err := app.FindCollectionByNameOrId("event_registrations")
		if err != nil {
			return err
		}

		if field := registrations.Fields.GetByName("status"); field != nil {
			registrations.Fields.RemoveById(field.GetId())
			return app.Save(registrations)
		}

		return nil
	})
}
