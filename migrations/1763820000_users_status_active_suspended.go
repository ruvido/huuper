package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		allowed := map[string]bool{
			"active":    true,
			"suspended": true,
		}

		if err := normalizeUserStatuses(app, allowed, "suspended"); err != nil {
			return err
		}

		return updateUserStatusValues(app, []string{"active", "suspended"})
	}, func(app core.App) error {
		allowed := map[string]bool{
			"active":  true,
			"pending": true,
		}

		if err := normalizeUserStatuses(app, allowed, "pending"); err != nil {
			return err
		}

		return updateUserStatusValues(app, []string{"active", "pending"})
	})
}
