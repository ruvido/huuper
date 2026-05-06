package migrations

import (
	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		settings, err := app.FindCollectionByNameOrId("settings")
		if err != nil {
			return err
		}
		record, _ := app.FindFirstRecordByFilter(
			"settings",
			"name = 'eventflow'",
			map[string]any{},
		)
		if record == nil {
			record = core.NewRecord(settings)
			record.Set("name", "eventflow")
			record.Set("data", defaultEventflowV2())
			return app.Save(record)
		}
		data := backendinternal.ParseJSONMap(record.Get("data"))
		types, _ := data["types"].([]any)
		for _, rawType := range types {
			typeDef, ok := rawType.(map[string]any)
			if !ok {
				continue
			}
			required, _ := typeDef["required"].(map[string]any)
			if required == nil {
				required = map[string]any{}
				typeDef["required"] = required
			}
			value, _ := typeDef["value"].(string)
			required["time"] = value == "call"
			if _, ok := required["description"]; !ok {
				required["description"] = false
			}
		}
		record.Set("data", data)
		return app.Save(record)
	}, func(app core.App) error {
		return nil
	})
}
