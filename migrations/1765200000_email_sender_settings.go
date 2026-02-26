package migrations

import (
	"strings"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		settings, err := app.FindCollectionByNameOrId("settings")
		if err != nil {
			return err
		}

		record, err := app.FindFirstRecordByFilter(
			"settings",
			"name = {:name}",
			map[string]any{"name": "email"},
		)
		if err != nil || record == nil {
			record = core.NewRecord(settings)
			record.Set("name", "email")
		}

		data := map[string]any{}
		if existing, ok := record.Get("data").(map[string]any); ok && existing != nil {
			data = existing
		}

		defaultSender := defaultSenderValue(app)
		if strings.TrimSpace(stringValue(data["general"])) == "" {
			data["general"] = defaultSender
		}
		if strings.TrimSpace(stringValue(data["events"])) == "" {
			data["events"] = defaultSender
		}

		record.Set("data", data)
		return app.Save(record)
	}, func(app core.App) error {
		return nil
	})
}

func defaultSenderValue(app core.App) string {
	settings := app.Settings()
	address := strings.TrimSpace(settings.Meta.SenderAddress)
	if address == "" {
		return ""
	}
	name := strings.TrimSpace(settings.Meta.SenderName)
	if name == "" {
		return address
	}
	return name + " <" + address + ">"
}

func stringValue(raw any) string {
	value, _ := raw.(string)
	return value
}
