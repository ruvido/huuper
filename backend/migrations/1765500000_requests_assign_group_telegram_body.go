package migrations

import (
	"strings"

	backendinternal "members/backend/internal"
	copywritingrequests "members/backend/internal/copywriting/requests"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		template, err := app.FindFirstRecordByFilter(
			"templates",
			"kind = {:kind}",
			map[string]any{"kind": "requests.assign_group"},
		)
		if err != nil || template == nil {
			return nil
		}

		copy, ok := copywritingrequests.ByKind("requests.assign_group")
		if !ok || strings.TrimSpace(copy.TelegramBody) == "" {
			return nil
		}

		data := map[string]any{}
		if err := template.UnmarshalJSONField("data", &data); err != nil {
			return err
		}

		telegram, _ := data["telegram"].(map[string]any)
		if telegram == nil {
			telegram = map[string]any{}
		}
		if strings.TrimSpace(backendinternal.AnyToString(telegram["body"])) != "" {
			return nil
		}

		telegram["body"] = copy.TelegramBody
		data["telegram"] = telegram
		template.Set("data", data)
		return app.Save(template)
	}, func(app core.App) error {
		return nil
	})
}
