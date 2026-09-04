package migrations

import (
	"encoding/json"

	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// The previous migration created the hidden `capacity` field but failed to
// copy the value across: it type-asserted record.Get("data") to a map, while
// PocketBase returns raw JSON there, so every record was skipped and the
// capacity limit silently stopped applying.
func init() {
	m.Register(func(app core.App) error {
		records, err := app.FindRecordsByFilter("retreats", "", "", 0, 0)
		if err != nil {
			return err
		}
		for _, record := range records {
			data := backendinternal.ParseJSONMap(record.Get("data"))
			if data == nil {
				continue
			}
			raw, ok := data["capacity"]
			if !ok {
				continue
			}
			capacity := 0
			switch value := raw.(type) {
			case float64:
				capacity = int(value)
			case int:
				capacity = value
			case json.Number:
				if n, err := value.Int64(); err == nil {
					capacity = int(n)
				}
			}
			if capacity > 0 && record.GetInt("capacity") == 0 {
				record.Set("capacity", capacity)
			}
			delete(data, "capacity")
			record.Set("data", data)
			if err := app.Save(record); err != nil {
				return err
			}
		}
		return nil
	}, func(app core.App) error {
		return nil
	})
}
