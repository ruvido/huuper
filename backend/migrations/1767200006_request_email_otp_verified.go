package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		tokens, err := app.FindCollectionByNameOrId("tokens")
		if err != nil {
			return err
		}
		if tokens.Fields.GetByName("verified") == nil {
			tokens.Fields.Add(&core.BoolField{Name: "verified"})
		}
		return app.Save(tokens)
	}, func(app core.App) error {
		tokens, err := app.FindCollectionByNameOrId("tokens")
		if err != nil {
			return err
		}
		if field := tokens.Fields.GetByName("verified"); field != nil {
			tokens.Fields.RemoveById(field.GetId())
		}
		return app.Save(tokens)
	})
}
