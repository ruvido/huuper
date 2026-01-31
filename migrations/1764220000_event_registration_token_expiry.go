package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(app core.App) error {
		registrations, err := app.FindCollectionByNameOrId("event_registrations")
		if err != nil {
			return err
		}

		if registrations.Fields.GetByName("accept_expires_at") == nil {
			registrations.Fields.Add(&core.DateField{
				Name:     "accept_expires_at",
				Required: false,
			})
		}

		if err := app.Save(registrations); err != nil {
			return err
		}

		records, err := app.FindRecordsByFilter("event_registrations", "", "", 0, 0)
		if err != nil {
			return err
		}
		for _, record := range records {
			if record.GetDateTime("accept_expires_at").IsZero() {
				record.Set("accept_expires_at", types.NowDateTime().AddDate(0, 0, 7))
				if err := app.Save(record); err != nil {
					return err
				}
			}
		}

		return nil
	}, func(app core.App) error {
		registrations, err := app.FindCollectionByNameOrId("event_registrations")
		if err != nil {
			return err
		}

		if field := registrations.Fields.GetByName("accept_expires_at"); field != nil {
			registrations.Fields.RemoveById(field.GetId())
			return app.Save(registrations)
		}

		return nil
	})
}
