package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		templates, err := app.FindCollectionByNameOrId("templates")
		if err != nil {
			return err
		}

		const newKind = "requests.request_submitted"

		switch field := templates.Fields.GetByName("kind").(type) {
		case *core.SelectField:
			field.Values = appendUniqueString(field.Values, newKind)
		case nil:
			templates.Fields.Add(&core.SelectField{
				Name:     "kind",
				Required: false,
				Values:   []string{newKind},
			})
		default:
			templates.Fields.RemoveById(field.GetId())
			templates.Fields.Add(&core.SelectField{
				Name:     "kind",
				Required: false,
				Values:   []string{newKind},
			})
		}

		return app.Save(templates)
	}, func(app core.App) error {
		return nil
	})
}

func appendUniqueString(values []string, value string) []string {
	for _, current := range values {
		if current == value {
			return values
		}
	}
	return append(values, value)
}
