package migrations

import (
	"fmt"
	"strings"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		requests, err := app.FindCollectionByNameOrId("requests")
		if err != nil {
			return err
		}

		statuses, err := loadRequestsFlowStatusesForMigration(app)
		if err != nil {
			return err
		}
		defaultStatus := statuses[0]

		statusField := requests.Fields.GetByName("status")
		if statusField == nil {
			requests.Fields.Add(&core.SelectField{
				Name:     "status",
				Required: false,
				Values:   statuses,
			})
		} else if selectField, ok := statusField.(*core.SelectField); ok {
			selectField.Values = statuses
		}

		if err := app.Save(requests); err != nil {
			return err
		}

		records, err := app.FindRecordsByFilter("requests", "", "", 0, 0)
		if err != nil {
			return err
		}

		allowed := make(map[string]struct{}, len(statuses))
		for _, status := range statuses {
			allowed[status] = struct{}{}
		}

		for _, record := range records {
			data := map[string]any{}
			_ = record.UnmarshalJSONField("data", &data)

			status := strings.TrimSpace(record.GetString("status"))
			if status == "" {
				if raw, ok := data["status"].(string); ok {
					status = strings.TrimSpace(raw)
				}
			}

			if _, ok := allowed[status]; !ok {
				status = defaultStatus
			}

			delete(data, "status")
			record.Set("status", status)
			record.Set("data", data)
			if err := app.Save(record); err != nil {
				return err
			}
		}

		requests.AddIndex(
			"idx_requests_status",
			false,
			"status",
			"status != ''",
		)

		return app.Save(requests)
	}, func(app core.App) error {
		requests, err := app.FindCollectionByNameOrId("requests")
		if err != nil || requests == nil {
			return nil
		}

		records, err := app.FindRecordsByFilter("requests", "", "", 0, 0)
		if err != nil {
			return err
		}

		for _, record := range records {
			status := strings.TrimSpace(record.GetString("status"))
			if status == "" {
				continue
			}

			data := map[string]any{}
			_ = record.UnmarshalJSONField("data", &data)
			data["status"] = status
			record.Set("data", data)
			if err := app.Save(record); err != nil {
				return err
			}
		}

		requests.RemoveIndex("idx_requests_status")
		if field := requests.Fields.GetByName("status"); field != nil {
			requests.Fields.RemoveById(field.GetId())
		}

		return app.Save(requests)
	})
}

func loadRequestsFlowStatusesForMigration(app core.App) ([]string, error) {
	record, err := app.FindFirstRecordByFilter(
		"settings",
		"name = 'requests_flow'",
		map[string]any{},
	)
	if err != nil || record == nil {
		return nil, fmt.Errorf("settings.requests_flow not found")
	}

	data := map[string]any{}
	_ = record.UnmarshalJSONField("data", &data)
	if nested, ok := data["data"].(map[string]any); ok && nested != nil {
		data = nested
	}

	rawStatuses, ok := data["statuses"].([]any)
	if !ok || len(rawStatuses) == 0 {
		return nil, fmt.Errorf("settings.requests_flow statuses missing")
	}

	out := make([]string, 0, len(rawStatuses))
	seen := map[string]struct{}{}
	for _, raw := range rawStatuses {
		value, ok := raw.(string)
		if !ok {
			continue
		}
		status := strings.TrimSpace(value)
		if status == "" {
			continue
		}
		if _, exists := seen[status]; exists {
			continue
		}
		seen[status] = struct{}{}
		out = append(out, status)
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("settings.requests_flow statuses empty")
	}

	return out, nil
}
