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
			if strings.TrimSpace(backendinternal.AnyToString(step["action"])) != "admin_approved" {
				continue
			}
			if strings.TrimSpace(backendinternal.AnyToString(step["cta"])) != "" {
				continue
			}
			step["cta"] = "Accetta"
			steps[i] = step
			changed = true
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
