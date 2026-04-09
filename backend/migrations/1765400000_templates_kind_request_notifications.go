package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

var requestTemplateKindAllowedValues = []string{
	"events.user.registration_received",
	"events.user.registration_accepted",
	"events.admin.new_registration",
	"requests.new_request",
	"requests.assign_group",
	"requests.assign_guardian",
	"requests.mentoring",
	"requests.group_approved",
	"requests.admin_approved",
}

func init() {
	m.Register(func(app core.App) error {
		templates, err := app.FindCollectionByNameOrId("templates")
		if err != nil {
			return err
		}

		switch field := templates.Fields.GetByName("kind").(type) {
		case *core.SelectField:
			field.Values = requestTemplateKindAllowedValues
		case nil:
			templates.Fields.Add(&core.SelectField{
				Name:     "kind",
				Required: false,
				Values:   requestTemplateKindAllowedValues,
			})
		default:
			templates.Fields.RemoveById(field.GetId())
			templates.Fields.Add(&core.SelectField{
				Name:     "kind",
				Required: false,
				Values:   requestTemplateKindAllowedValues,
			})
		}

		return app.Save(templates)
	}, func(app core.App) error {
		return nil
	})
}
