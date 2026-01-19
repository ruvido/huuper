package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(app core.App) error {
		regions, err := app.FindCollectionByNameOrId("regions")
		if err != nil {
			return err
		}

		regions.ListRule = types.Pointer("")
		regions.ViewRule = types.Pointer("")

		return app.Save(regions)
	}, func(app core.App) error {
		regions, err := app.FindCollectionByNameOrId("regions")
		if err != nil {
			return err
		}

		regions.ListRule = types.Pointer("@request.auth.id != ''")
		regions.ViewRule = types.Pointer("@request.auth.id != ''")

		return app.Save(regions)
	})
}
