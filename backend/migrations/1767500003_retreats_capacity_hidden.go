package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// Capacity is an organiser-side limit: it gates registrations but must never
// be readable by the public. `data` is served verbatim on a publicly readable
// collection, so the number moves to a hidden field, which PocketBase keeps
// out of API responses.
func init() {
	m.Register(func(app core.App) error {
		retreats, err := app.FindCollectionByNameOrId("retreats")
		if err != nil {
			return err
		}

		if retreats.Fields.GetByName("capacity") == nil {
			field := &core.NumberField{Name: "capacity", Required: false}
			field.SetHidden(true)
			retreats.Fields.Add(field)
			if err := app.Save(retreats); err != nil {
				return err
			}
		}

		records, err := app.FindRecordsByFilter("retreats", "", "", 0, 0)
		if err != nil {
			return err
		}
		for _, record := range records {
			data, _ := record.Get("data").(map[string]any)
			if data == nil {
				continue
			}
			raw, ok := data["capacity"]
			if !ok {
				continue
			}
			switch value := raw.(type) {
			case float64:
				record.Set("capacity", int(value))
			case int:
				record.Set("capacity", value)
			}
			delete(data, "capacity")
			record.Set("data", data)
			if err := app.Save(record); err != nil {
				return err
			}
		}
		return nil
	}, func(app core.App) error {
		retreats, err := app.FindCollectionByNameOrId("retreats")
		if err != nil {
			return nil
		}
		field := retreats.Fields.GetByName("capacity")
		if field == nil {
			return nil
		}
		retreats.Fields.RemoveById(field.GetId())
		return app.Save(retreats)
	})
}
