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

	record.Set("data", defaultRequestsFlowSettingsData())

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
