package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		payments := core.NewBaseCollection("payments")
		payments.ListRule = nil   // Admin only
		payments.ViewRule = nil   // Admin only
		payments.CreateRule = nil // Admin only (created server-side)
		payments.UpdateRule = nil // Admin only (updated server-side)
		payments.DeleteRule = nil // Admin only

		payments.Fields.Add(
			&core.AutodateField{
				Name:     "created",
				OnCreate: true,
			},
			&core.AutodateField{
				Name:     "updated",
				OnCreate: true,
				OnUpdate: true,
			},
			&core.TextField{
				Name:     "purpose_type",
				Required: true,
				Max:      100,
			},
			&core.TextField{
				Name:     "purpose_id",
				Required: true,
				Max:      100,
			},
			&core.TextField{
				Name:     "email",
				Required: true,
				Max:      200,
			},
			&core.NumberField{
				Name:     "amount",
				Required: true,
			},
			&core.TextField{
				Name:     "currency",
				Required: true,
				Max:      10,
			},
			&core.SelectField{
				Name:     "status",
				Required: true,
				Values:   []string{"pending", "paid", "failed", "cancelled", "refunded"},
			},
			&core.TextField{
				Name:     "stripe_session_id",
				Required: false,
				Max:      200,
			},
			&core.TextField{
				Name:     "stripe_payment_intent",
				Required: false,
				Max:      200,
			},
			&core.JSONField{
				Name:     "data",
				Required: false,
			},
		)

		payments.AddIndex("idx_payments_stripe_session", true, "stripe_session_id", "")

		return app.Save(payments)
	}, func(app core.App) error {
		payments, err := app.FindCollectionByNameOrId("payments")
		if err != nil {
			return nil
		}
		return app.Delete(payments)
	})
}
