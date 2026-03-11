package migrations

import (
	"fmt"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		return updateUserStatusValuesWithRejected(app, []string{"pending", "assigned", "visitor", "approved", "active", "rejected"})
	}, func(app core.App) error {
		allowed := map[string]bool{
			"pending":  true,
			"assigned": true,
			"visitor":  true,
			"approved": true,
			"active":   true,
		}

		if err := normalizeUserStatusesWithRejected(app, allowed, "pending"); err != nil {
			return err
		}

		return updateUserStatusValuesWithRejected(app, []string{"pending", "assigned", "visitor", "approved", "active"})
	})
}

func updateUserStatusValuesWithRejected(app core.App, values []string) error {
	users, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		return err
	}

	statusField := users.Fields.GetByName("status")
	if statusField == nil {
		return fmt.Errorf("missing users.status field")
	}

	selectField, ok := statusField.(*core.SelectField)
	if !ok {
		return fmt.Errorf("users.status is not a select field")
	}

	selectField.Values = values

	return app.Save(users)
}

func normalizeUserStatusesWithRejected(app core.App, allowed map[string]bool, fallback string) error {
	records, err := app.FindRecordsByFilter("users", "", "", 0, 0)
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
