package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		requests, err := app.FindCollectionByNameOrId("requests")
		if err != nil {
			return err
		}

		if requests.Fields.GetByName("rejected") == nil {
			requests.Fields.Add(&core.BoolField{
				Name:     "rejected",
				Required: false,
			})
			if err := app.Save(requests); err != nil {
				return err
			}
		}

		records, err := app.FindRecordsByFilter("requests", "", "", 0, 0)
		if err != nil {
			return err
		}

		for _, record := range records {
			data := map[string]any{}
			_ = record.UnmarshalJSONField("data", &data)

			rejected := false
			if raw, ok := data["rejected"]; ok {
				if value, ok := raw.(bool); ok {
					rejected = value
				}
				delete(data, "rejected")
			}

			record.Set("rejected", rejected)
			record.Set("data", data)
			if err := app.Save(record); err != nil {
				return err
			}
		}

		return nil
	}, func(app core.App) error {
		requests, err := app.FindCollectionByNameOrId("requests")
		if err != nil {
			return nil
		}

		if field := requests.Fields.GetByName("rejected"); field != nil {
			requests.Fields.RemoveById(field.GetId())
			return app.Save(requests)
		}

		return nil
	})
}
