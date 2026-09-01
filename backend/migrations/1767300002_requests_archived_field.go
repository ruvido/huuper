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

		if requests.Fields.GetByName("archived") == nil {
			requests.Fields.Add(&core.BoolField{
				Name:     "archived",
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

			if raw, ok := data["rejected"]; ok {
				if block, ok := raw.(map[string]any); ok {
					archived := map[string]any{
						"reason": block["reason"],
					}
					if at, ok := block["rejected_at"]; ok {
						archived["archived_at"] = at
					}
					if by, ok := block["rejected_by"]; ok {
						archived["archived_by"] = by
					}
					data["archived"] = archived
				}
				delete(data, "rejected")
			}

			record.Set("archived", record.GetBool("rejected"))
			record.Set("data", data)
			if err := app.Save(record); err != nil {
				return err
			}
		}

		if field := requests.Fields.GetByName("rejected"); field != nil {
			requests.Fields.RemoveById(field.GetId())
			if err := app.Save(requests); err != nil {
				return err
			}
		}

		return nil
	}, func(app core.App) error {
		requests, err := app.FindCollectionByNameOrId("requests")
		if err != nil {
			return nil
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

			if raw, ok := data["archived"]; ok {
				if block, ok := raw.(map[string]any); ok {
					rejected := map[string]any{
						"reason": block["reason"],
					}
					if at, ok := block["archived_at"]; ok {
						rejected["rejected_at"] = at
					}
					if by, ok := block["archived_by"]; ok {
						rejected["rejected_by"] = by
					}
					data["rejected"] = rejected
				}
				delete(data, "archived")
			}

			record.Set("rejected", record.GetBool("archived"))
			record.Set("data", data)
			if err := app.Save(record); err != nil {
				return err
			}
		}

		if field := requests.Fields.GetByName("archived"); field != nil {
			requests.Fields.RemoveById(field.GetId())
			return app.Save(requests)
		}

		return nil
	})
}
