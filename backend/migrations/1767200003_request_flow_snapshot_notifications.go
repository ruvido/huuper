package migrations

import (
	"strings"

	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		if err := normalizeRequestFlowSettingsNotifications(app); err != nil {
			return err
		}
		return normalizeOpenRequestFlowSnapshotNotifications(app)
	}, func(app core.App) error {
		return nil
	})
}

func normalizeRequestFlowSettingsNotifications(app core.App) error {
	record, err := app.FindFirstRecordByFilter(
		"settings",
		"name = 'requests_flow'",
		map[string]any{},
	)
	if err != nil || record == nil {
		return nil
	}

	data := backendinternal.ParseJSONMap(record.Get("data"))
	if !normalizeRequestFlowNotifications(data) {
		return nil
	}
	record.Set("data", data)
	return app.Save(record)
}

func normalizeOpenRequestFlowSnapshotNotifications(app core.App) error {
	records, err := app.FindRecordsByFilter("requests", "", "", 0, 0)
	if err != nil {
		return err
	}

	for _, record := range records {
		data := backendinternal.ParseJSONMap(record.Get("data"))
		rawSnapshot, ok := data["__request_flow"].(map[string]any)
		if !ok || rawSnapshot == nil {
			continue
		}
		if !normalizeRequestFlowNotifications(rawSnapshot) {
			continue
		}
		data["__request_flow"] = rawSnapshot
		record.Set("data", data)
		if err := app.Save(record); err != nil {
			return err
		}
	}
	return nil
}

func normalizeRequestFlowNotifications(data map[string]any) bool {
	steps, ok := data["steps"].([]any)
	if !ok || len(steps) == 0 {
		return false
	}

	changed := false
	for i, raw := range steps {
		step, ok := raw.(map[string]any)
		if !ok || step == nil {
			continue
		}

		emailTo, telegramMessage := defaultRequestFlowNotificationForSnapshot(backendinternal.AnyToString(step["action"]))
		if emailTo != "" && strings.TrimSpace(backendinternal.AnyToString(step["email_to"])) == "" {
			step["email_to"] = emailTo
			changed = true
		}
		if _, ok := step["telegram_message"].(bool); !ok {
			if strings.TrimSpace(backendinternal.AnyToString(step["telegram_message"])) == "" {
				step["telegram_message"] = telegramMessage
				changed = true
			}
		}
		steps[i] = step
	}
	if changed {
		data["steps"] = steps
	}
	return changed
}

func defaultRequestFlowNotificationForSnapshot(action string) (string, bool) {
	switch strings.TrimSpace(action) {
	case "assign_group":
		return "assistant", true
	case "assign_guardian":
		return "guardian", false
	case "mentoring":
		return "assistant", false
	case "group_approved":
		return "admin", false
	case "admin_approved":
		return "candidate", true
	default:
		return "", false
	}
}
