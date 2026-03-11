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
		if err != nil || requests == nil {
			return err
		}

		requests.RemoveIndex("idx_requests_status")
		if field := requests.Fields.GetByName("status"); field != nil {
			requests.Fields.RemoveById(field.GetId())
		}

		return app.Save(requests)
	}, func(app core.App) error {
		requests, err := app.FindCollectionByNameOrId("requests")
		if err != nil || requests == nil {
			return err
		}

		statuses, err := loadRequestsFlowStatusesForStatusRollback(app)
		if err != nil {
			statuses = []string{
				"1-submitted",
				"2-group_assigned",
				"3-guardian_assigned",
				"4-mentoring",
				"5-group_approved",
				"6-admin_approved",
			}
		}

		if requests.Fields.GetByName("status") == nil {
			requests.Fields.Add(&core.SelectField{
				Name:   "status",
				Values: statuses,
			})
		}

		if err := app.Save(requests); err != nil {
			return err
		}

		records, err := app.FindRecordsByFilter("requests", "", "", 0, 0)
		if err != nil {
			return err
		}

		for _, record := range records {
			data := map[string]any{}
			_ = record.UnmarshalJSONField("data", &data)
			step := parseStepIndexForStatusRollback(data)
			status := statusForStepRollback(step, statuses)
			record.Set("status", status)
			if err := app.Save(record); err != nil {
				return err
			}
		}

		requests.AddIndex("idx_requests_status", false, "status", "status != ''")
		return app.Save(requests)
	})
}

func loadRequestsFlowStatusesForStatusRollback(app core.App) ([]string, error) {
	record, err := app.FindFirstRecordByFilter("settings", "name = 'requests_flow'", map[string]any{})
	if err != nil || record == nil {
		return nil, fmt.Errorf("settings.requests_flow not found")
	}

	data := map[string]any{}
	_ = record.UnmarshalJSONField("data", &data)
	if nested, ok := data["data"].(map[string]any); ok && nested != nil {
		data = nested
	}

	rawSteps, ok := data["steps"].([]any)
	if !ok || len(rawSteps) == 0 {
		return nil, fmt.Errorf("settings.requests_flow steps missing")
	}

	out := []string{"1-submitted"}
	seen := map[string]struct{}{"1-submitted": {}}
	for _, raw := range rawSteps {
		step, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		action, _ := step["action"].(string)
		status := statusForActionRollback(strings.TrimSpace(action))
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
		return nil, fmt.Errorf("requests flow statuses empty")
	}
	return out, nil
}

func parseStepIndexForStatusRollback(data map[string]any) int {
	if data == nil {
		return 0
	}
	raw, ok := data["__step_index"]
	if !ok {
		return 0
	}
	switch v := raw.(type) {
	case float64:
		if v < 0 {
			return 0
		}
		return int(v)
	case int:
		if v < 0 {
			return 0
		}
		return v
	default:
		return 0
	}
}

func statusForStepRollback(step int, statuses []string) string {
	if len(statuses) == 0 {
		return "1-submitted"
	}
	if step < 0 {
		step = 0
	}
	if step >= len(statuses) {
		step = len(statuses) - 1
	}
	return statuses[step]
}

func statusForActionRollback(action string) string {
	switch action {
	case "assign_group":
		return "2-group_assigned"
	case "assign_guardian":
		return "3-guardian_assigned"
	case "mentoring":
		return "4-mentoring"
	case "group_approved":
		return "5-group_approved"
	case "admin_approved":
		return "6-admin_approved"
	default:
		return ""
	}
}
