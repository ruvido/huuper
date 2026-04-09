package migrations

import (
	"strings"

	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		record, err := app.FindFirstRecordByFilter(
			"settings",
			"name = 'requests_flow'",
			map[string]any{},
		)
		if err != nil || record == nil {
			return nil
		}

		data := map[string]any{}
		if err := record.UnmarshalJSONField("data", &data); err != nil {
			return err
		}

		steps, ok := data["steps"].([]any)
		if !ok || len(steps) == 0 {
			return nil
		}

		changed := false
		for i, raw := range steps {
			step, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			action := strings.TrimSpace(backendinternal.AnyToString(step["action"]))
			emailTo, telegramMessage := defaultRequestFlowNotification(action)
			if emailTo == "" && !telegramMessage {
				continue
			}

			if current, ok := step["email_to"]; !ok || strings.TrimSpace(backendinternal.AnyToString(current)) == "" {
				if emailTo != "" {
					step["email_to"] = emailTo
					changed = true
				}
			}

			if current, ok := step["telegram_message"]; !ok {
				step["telegram_message"] = telegramMessage
				changed = true
			} else if _, isBool := current.(bool); !isBool && strings.TrimSpace(backendinternal.AnyToString(current)) == "" {
				step["telegram_message"] = telegramMessage
				changed = true
			}

			steps[i] = step
		}

		if !changed {
			return nil
		}

		data["steps"] = steps
		record.Set("data", data)
		return app.Save(record)
	}, func(app core.App) error {
		return nil
	})
}

func defaultRequestFlowNotification(action string) (string, bool) {
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
