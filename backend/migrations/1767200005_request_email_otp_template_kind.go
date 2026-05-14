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

		const kind = "requests.email_otp"
		switch field := templates.Fields.GetByName("kind").(type) {
		case *core.SelectField:
			field.Values = appendUniqueString(field.Values, kind)
		case nil:
			templates.Fields.Add(&core.SelectField{
				Name:     "kind",
				Required: false,
				Values:   []string{kind},
			})
		default:
			templates.Fields.RemoveById(field.GetId())
			templates.Fields.Add(&core.SelectField{
				Name:     "kind",
				Required: false,
				Values:   []string{kind},
			})
		}
		return app.Save(templates)
	}, func(app core.App) error {
		return nil
	})
}
