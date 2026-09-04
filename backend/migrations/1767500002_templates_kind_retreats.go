package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// retreatsTemplateKindValues are appended to templates.kind's allowed select
// values so the retreats module can reuse the existing template + email
// sending mechanism (no new collection, no new send path). Templates for
// these kinds are looked up without an `event` scope (global fallback),
// since templates.event only relates to the `events` collection.
var retreatsTemplateKindValues = []string{
	"retreats.user.registration_received",
	"retreats.user.registration_accepted",
	"retreats.user.payment_link",
	"retreats.admin.new_registration",
}

func init() {
	m.Register(func(app core.App) error {
		templates, err := app.FindCollectionByNameOrId("templates")
		if err != nil {
			return err
		}

		field, ok := templates.Fields.GetByName("kind").(*core.SelectField)
		if !ok {
			return nil
		}

		existing := map[string]bool{}
		for _, v := range field.Values {
			existing[v] = true
		}
		for _, v := range retreatsTemplateKindValues {
			if !existing[v] {
				field.Values = append(field.Values, v)
			}
		}

		return app.Save(templates)
	}, func(app core.App) error {
		templates, err := app.FindCollectionByNameOrId("templates")
		if err != nil {
			return nil
		}

		field, ok := templates.Fields.GetByName("kind").(*core.SelectField)
		if !ok {
			return nil
		}

		remove := map[string]bool{}
		for _, v := range retreatsTemplateKindValues {
			remove[v] = true
		}
		kept := make([]string, 0, len(field.Values))
		for _, v := range field.Values {
			if !remove[v] {
				kept = append(kept, v)
			}
		}
		field.Values = kept

		return app.Save(templates)
	})
}
