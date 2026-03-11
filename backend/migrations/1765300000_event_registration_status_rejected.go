package migrations

import (
	"fmt"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		return updateEventRegistrationStatusValues(app, []string{"pending", "active", "cancelled", "rejected"})
	}, func(app core.App) error {
		allowed := map[string]bool{
			"pending":   true,
			"active":    true,
			"cancelled": true,
		}
		if err := normalizeEventRegistrationStatuses(app, allowed, "pending"); err != nil {
			return err
		}
		return updateEventRegistrationStatusValues(app, []string{"pending", "active", "cancelled"})
	})
}

func updateEventRegistrationStatusValues(app core.App, values []string) error {
	registrations, err := app.FindCollectionByNameOrId("event_registrations")
	if err != nil {
		return err
	}

	statusField := registrations.Fields.GetByName("status")
	if statusField == nil {
		return fmt.Errorf("missing event_registrations.status field")
	}

	selectField, ok := statusField.(*core.SelectField)
	if !ok {
		return fmt.Errorf("event_registrations.status is not a select field")
	}

	selectField.Values = values
	return app.Save(registrations)
}

func normalizeEventRegistrationStatuses(app core.App, allowed map[string]bool, fallback string) error {
	records, err := app.FindRecordsByFilter("event_registrations", "", "", 0, 0)
	if err != nil {
		return err
	}

	for _, record := range records {
		status := record.GetString("status")
		if allowed[status] {
			continue
		}

		record.Set("status", fallback)
		if err := app.Save(record); err != nil {
			return err
		}
	}

	return nil
}
