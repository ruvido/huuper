package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// Two organiser-facing emails the module was missing.
//
// A member registering and paying produced no notification at all: the only
// admin email was the guest pre-registration, so a seat could be taken without
// anyone hearing about it. And there was no way to see where a retreat stood
// without opening the panel.
//
// Neither template gets a `to`: the organiser's address lives on
// "retreats.admin.new_registration" and retreats.adminRecipient falls back to
// it, so it is set once for the whole module.
//
// Bodies are English like the rest of the code; the real wording is record
// content and is edited in the database.
//
// Placeholders: everything retreatPlaceholders provides, plus [email] [name]
// [phone] on the completion one, and on both the figures from
// statsPlaceholders: [active] [members] [guests] [awaiting_payment] [pending]
// [remaining] [capacity].
var retreatsAdminEmailKinds = []string{
	"retreats.admin.registration_completed",
	"retreats.admin.daily_stats",
}

func init() {
	m.Register(func(app core.App) error {
		templates, err := app.FindCollectionByNameOrId("templates")
		if err != nil {
			return err
		}

		if field, ok := templates.Fields.GetByName("kind").(*core.SelectField); ok {
			existing := map[string]bool{}
			for _, value := range field.Values {
				existing[value] = true
			}
			for _, value := range retreatsAdminEmailKinds {
				if !existing[value] {
					field.Values = append(field.Values, value)
				}
			}
			if err := app.Save(templates); err != nil {
				return err
			}
		}

		seeds := []struct {
			kind    string
			subject string
			body    string
		}{
			{
				kind:    "retreats.admin.registration_completed",
				subject: "[retreat] — a place has been taken: [email]",
				body: `A registration for [retreat] is confirmed and paid.

[name]
[email]
[phone]

Places taken: [active] of [capacity] — [members] members, [guests] from outside.
Still free: [remaining].`,
			},
			{
				kind:    "retreats.admin.daily_stats",
				subject: "[retreat] — registrations today",
				body: `[retreat] — [dates]

Confirmed: [active]
  members: [members]
  from outside: [guests]

Waiting on their deposit: [awaiting_payment]
Pre-registrations to call back: [pending]

Places still free: [remaining] of [capacity].`,
			},
		}

		for _, seed := range seeds {
			existing, _ := app.FindFirstRecordByFilter("templates", "kind = {:kind}", map[string]any{"kind": seed.kind})
			if existing != nil {
				continue
			}
			record := core.NewRecord(templates)
			record.Set("kind", seed.kind)
			record.Set("data", map[string]any{"subject": seed.subject, "body": seed.body})
			if err := app.Save(record); err != nil {
				return err
			}
		}

		return nil
	}, func(app core.App) error {
		for _, kind := range retreatsAdminEmailKinds {
			if record, _ := app.FindFirstRecordByFilter("templates", "kind = {:kind}", map[string]any{"kind": kind}); record != nil {
				if err := app.Delete(record); err != nil {
					return err
				}
			}
		}

		templates, err := app.FindCollectionByNameOrId("templates")
		if err != nil {
			return nil
		}
		field, ok := templates.Fields.GetByName("kind").(*core.SelectField)
		if !ok {
			return nil
		}
		remove := map[string]bool{}
		for _, value := range retreatsAdminEmailKinds {
			remove[value] = true
		}
		kept := make([]string, 0, len(field.Values))
		for _, value := range field.Values {
			if !remove[value] {
				kept = append(kept, value)
			}
		}
		field.Values = kept

		return app.Save(templates)
	})
}
