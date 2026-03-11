package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(app core.App) error {
		settings, err := app.FindCollectionByNameOrId("settings")
		if err != nil {
			return err
		}

		// Admin-only access; settings are served via the custom API.
		settings.ListRule = nil
		settings.ViewRule = nil
		settings.CreateRule = nil
		settings.UpdateRule = nil
		settings.DeleteRule = nil

		return app.Save(settings)
	}, func(app core.App) error {
		settings, err := app.FindCollectionByNameOrId("settings")
		if err != nil {
			return err
		}

		// Revert to public read rules.
		settings.ListRule = types.Pointer("")
		settings.ViewRule = types.Pointer("")
		settings.CreateRule = nil
		settings.UpdateRule = nil
		settings.DeleteRule = nil

		return app.Save(settings)
	})
}
