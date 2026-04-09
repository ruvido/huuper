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

		if err := ensureRequestsFlowSettings(app, settings); err != nil {
			return err
		}

		if err := ensureUsersSettings(app, settings); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		return nil
	})
}

func ensureRequestsFlowSettings(app core.App, settings *core.Collection) error {
	record, err := app.FindFirstRecordByFilter(
		"settings",
		"name = 'requests_flow'",
		map[string]any{},
	)
	if err != nil || record == nil {
		record = core.NewRecord(settings)
		record.Set("name", "requests_flow")
		record.Set("data", defaultRequestsFlowSettingsData())
		return app.Save(record)
	}

	record.Set("data", defaultRequestsFlowSettingsData())
	return app.Save(record)
}

func defaultRequestsFlowSettingsData() map[string]any {
	return map[string]any{
		"version": 1,
		"steps": []map[string]any{
			{
				"role":             "admin",
				"action":           "assign_group",
				"label":            "Assegna gruppo",
				"filter":           "local",
				"email_to":         "assistant",
				"telegram_message": true,
			},
			{
				"role":             "assistant",
				"action":           "assign_guardian",
				"label":            "Assegna guardian",
				"filter":           "group_members",
				"email_to":         "guardian",
				"telegram_message": false,
			},
			{
				"role":             "guardian",
				"action":           "mentoring",
				"label":            "Mentoring",
				"notes":            "Consulta note su come fare l'angelo custode",
				"email_to":         "assistant",
				"telegram_message": false,
			},
			{
				"role":             "assistant",
				"action":           "group_approved",
				"label":            "Approvazione gruppo",
				"notes":            "Per approvare una richiesta è necessaria una votazione del gruppo",
				"email_to":         "admin",
				"telegram_message": false,
			},
			{
				"role":             "admin",
				"action":           "admin_approved",
				"label":            "In verifica",
				"notes":            "In attesa di approvazione finale.",
				"email_to":         "candidate",
				"telegram_message": true,
			},
		},
	}
}

func ensureUsersSettings(app core.App, settings *core.Collection) error {
	record, err := app.FindFirstRecordByFilter(
		"settings",
		"name = 'users'",
		map[string]any{},
	)

	if err != nil || record == nil {
		record = core.NewRecord(settings)
		record.Set("name", "users")
		record.Set("data", map[string]any{
			"pact_required": true,
		})
		return app.Save(record)
	}

	data := map[string]any{}
	_ = record.UnmarshalJSONField("data", &data)
	data["pact_required"] = true
	record.Set("data", data)
	return app.Save(record)
}
