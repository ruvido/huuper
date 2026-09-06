package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// Seeds the retreat emails. The previous migration only widened templates.kind,
// so no record existed and sendTemplateEmail bailed out silently: registrants
// were never written to. Bodies are English like the rest of the code; the
// real, localized wording is record content and is edited in the database.
//
// The admin notification is seeded WITHOUT a recipient: `data.to` on the
// record is the only place that address lives, editable from the admin panel
// like the wording. Until someone sets it, SendAdminNewRegistrationNotification
// warns the superusers instead of guessing an address here.
//
// Placeholders filled in by retreats.retreatPlaceholders: [retreat] [tagline]
// [dates] [start_date] [end_date] [location] [place] [region] [meeting_time]
// [arrival_car] [arrival_train] [arrival_carpooling] [price] [deposit]
// [balance] [contact_email], plus [payment_url] and [invite_link] on their
// own templates and [email] on the admin one.
func init() {
	m.Register(func(app core.App) error {
		templates, err := app.FindCollectionByNameOrId("templates")
		if err != nil {
			return err
		}

		seeds := []struct {
			kind    string
			subject string
			body    string
		}{
			{
				kind:    "retreats.user.registration_received",
				subject: "[retreat] — we received your request",
				body: `Hi,

we received your registration request for **[retreat]** ([dates], [location]).

Places are limited: we will get back to you to confirm.

See you soon`,
			},
			{
				kind:    "retreats.user.payment_link",
				subject: "[retreat] — confirm your place with the deposit",
				body: `Hi,

your registration for **[retreat]** has been accepted.

To confirm your place, pay the [deposit] deposit:

[PAY THE DEPOSIT]([payment_url])

The remaining [balance] is paid on arrival. Full fee: [price].

See you soon`,
			},
			{
				kind:    "retreats.user.registration_accepted",
				subject: "[retreat] — you are registered",
				body: `Hi,

you are registered for **[retreat]**. Here is everything you need.

**When**
[dates] — [meeting_time]

**Where**
[place]
[location]

**Getting there**
By car: [arrival_car]
Car sharing: [arrival_carpooling]

**Fee**
Deposit paid: [deposit]. To pay on arrival: [balance].

[JOIN THE TELEGRAM GROUP]([invite_link])

See you there`,
			},
			{
				kind:    "retreats.admin.new_registration",
				subject: "[retreat] — new registration request",
				body: `New registration request for [retreat] ([dates]).

Email: [email]

Approve it from the admin panel to send out the payment link.`,
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
		for _, kind := range retreatsTemplateKindValues {
			if record, _ := app.FindFirstRecordByFilter("templates", "kind = {:kind}", map[string]any{"kind": kind}); record != nil {
				if err := app.Delete(record); err != nil {
					return err
				}
			}
		}
		return nil
	})
}
