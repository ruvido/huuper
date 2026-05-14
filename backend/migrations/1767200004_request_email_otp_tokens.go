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
		if field := tokens.Fields.GetByName("user"); field != nil {
			if rel, ok := field.(*core.RelationField); ok {
				rel.Required = false
			}
		}
		if tokens.Fields.GetByName("email") == nil {
			tokens.Fields.Add(&core.EmailField{
				Name:     "email",
				Required: false,
			})
		}
		if tokens.Fields.GetByName("code") == nil {
			tokens.Fields.Add(&core.TextField{
				Name:     "code",
				Required: false,
				Max:      20,
			})
		}
		if tokens.Fields.GetByName("expires") == nil {
			tokens.Fields.Add(&core.DateField{
				Name:     "expires",
				Required: false,
			})
		}
		if tokens.Fields.GetByName("attempts") == nil {
			tokens.Fields.Add(&core.NumberField{
				Name:     "attempts",
				Required: false,
				Min:      pointerToFloat(0),
				Max:      pointerToFloat(10),
			})
		}
		return app.Save(tokens)
	}, func(app core.App) error {
		tokens, err := app.FindCollectionByNameOrId("tokens")
		if err != nil {
			return err
		}
		for _, name := range []string{"email", "code", "expires", "attempts"} {
			if field := tokens.Fields.GetByName(name); field != nil {
				tokens.Fields.RemoveById(field.GetId())
			}
		}
		if field := tokens.Fields.GetByName("user"); field != nil {
			if rel, ok := field.(*core.RelationField); ok {
				rel.Required = true
			}
		}
		return app.Save(tokens)
	})
}
