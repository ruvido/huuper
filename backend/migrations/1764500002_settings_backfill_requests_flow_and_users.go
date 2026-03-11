package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		settings, err := app.FindCollectionByNameOrId("settings")
		if err != nil {
			return err
		}

		if err := upsertRequestsFlowSetting(app, settings); err != nil {
			return err
		}

		if err := upsertUsersSetting(app, settings); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		return nil
	})
}

func upsertRequestsFlowSetting(app core.App, settings *core.Collection) error {
	record, _ := app.FindFirstRecordByFilter(
		"settings",
		"name = 'requests_flow'",
		map[string]any{},
	)
	if record == nil {
		record = core.NewRecord(settings)
		record.Set("name", "requests_flow")
	}

	record.Set("data", map[string]any{
		"statuses": []string{
			"1-submitted",
			"2-group_assigned",
			"3-guardian_assigned",
			"4-mentoring",
			"5-group_approved",
			"6-admin_approved",
		},
		"set_status_by": map[string]string{
			"2-group_assigned":    "admin",
			"3-guardian_assigned": "assistant",
			"4-mentoring":         "guardian",
			"5-group_approved":    "assistant",
			"6-admin_approved":    "admin",
		},
	})

	return app.Save(record)
}

func upsertUsersSetting(app core.App, settings *core.Collection) error {
	record, _ := app.FindFirstRecordByFilter(
		"settings",
		"name = 'users'",
		map[string]any{},
	)
	if record == nil {
		record = core.NewRecord(settings)
		record.Set("name", "users")
	}

	record.Set("data", map[string]any{
		"pact_required": true,
	})

	return app.Save(record)
}
