package migrations

import (
	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// One-shot correction of the recipient seeded before 1767500005 stopped
// hardcoding one. These two literals are history, not configuration: from here
// on the address lives only in the template record's `data.to`, and nothing in
// the code reads or re-applies them.
func init() {
	m.Register(func(app core.App) error {
		return setRetreatsAdminRecipient(app, "info@realmen.it", "scrivi@realmen.it")
	}, func(app core.App) error {
		return setRetreatsAdminRecipient(app, "scrivi@realmen.it", "info@realmen.it")
	})
}

func setRetreatsAdminRecipient(app core.App, from string, to string) error {
	record, err := app.FindFirstRecordByFilter(
		"templates",
		"kind = {:kind}",
		map[string]any{"kind": "retreats.admin.new_registration"},
	)
	if err != nil || record == nil {
		return nil
	}
	data := backendinternal.ParseJSONMap(record.Get("data"))
	if current, _ := data["to"].(string); current != from {
		return nil
	}
	data["to"] = to
	record.Set("data", data)
	return app.Save(record)
}
