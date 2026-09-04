package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		return updateEventRegistrationStatusValues(app, []string{"pending", "awaiting_payment", "active", "cancelled", "rejected"})
	}, func(app core.App) error {
		allowed := map[string]bool{
			"pending":   true,
			"active":    true,
			"cancelled": true,
			"rejected":  true,
		}
		if err := normalizeEventRegistrationStatuses(app, allowed, "pending"); err != nil {
			return err
		}
		return updateEventRegistrationStatusValues(app, []string{"pending", "active", "cancelled", "rejected"})
	})
}
