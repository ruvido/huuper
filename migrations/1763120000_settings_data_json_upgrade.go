package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		settings, err := app.FindCollectionByNameOrId("settings")
		if err != nil {
			return err
		}

		dataField := settings.Fields.GetByName("data")
		if dataField != nil {
			if _, isTextField := dataField.(*core.TextField); isTextField {
				settings.Fields.RemoveById(dataField.GetId())
				settings.Fields.Add(&core.JSONField{
					Id:       dataField.GetId(),
					Name:     "data",
					Required: true,
				})
				if err := app.Save(settings); err != nil {
					return err
				}
			}
		}

		records, err := app.FindRecordsByFilter("settings", "", "", 0, 0)
		if err != nil {
			return err
		}

		for _, record := range records {
			raw := record.Get("data")
			value, ok := raw.(string)
			if !ok || value == "" {
				continue
			}

			var parsed any
			if err := json.Unmarshal([]byte(value), &parsed); err != nil {
				continue
			}

			record.Set("data", parsed)
			if err := app.Save(record); err != nil {
				return err
			}
		}

		return nil
	}, func(app core.App) error {
		return nil
	})
}
