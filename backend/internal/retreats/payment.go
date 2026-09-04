package retreats

import (
	"strings"

	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// DepositCentsForRetreat reads data.deposit_cents — deliberately not a
// dedicated column (see plan rule 3: flexible content lives in `data`,
// dedicated columns only for what recurrence/gating logic needs, and this
// value already lives behind Data{} accessors here, not new schema).
func DepositCentsForRetreat(retreat *core.Record) int {
	if retreat == nil {
		return 0
	}
	data := backendinternal.ParseJSONMap(retreat.Get("data"))
	return DataInt(data, "deposit_cents")
}

func PaymentSuccessURL(app *pocketbase.PocketBase) string {
	base := strings.TrimRight(app.Settings().Meta.AppURL, "/")
	return base + "/retreat-payment/?status=success"
}

func PaymentCancelURL(app *pocketbase.PocketBase) string {
	base := strings.TrimRight(app.Settings().Meta.AppURL, "/")
	return base + "/retreat-payment/?status=cancelled"
}

// MarkAwaitingPayment moves a registration to awaiting_payment and emails
// the registrant the Stripe checkout link.
func MarkAwaitingPayment(app *pocketbase.PocketBase, record *core.Record, paymentURL string) error {
	record.Set("status", "awaiting_payment")
	data := backendinternal.ParseJSONMap(record.Get("data"))
	data["payment_url"] = paymentURL
	record.Set("data", data)
	if err := app.Save(record); err != nil {
		return err
	}
	SendPaymentLinkEmail(app, record.GetString("email"), paymentURL)
	return nil
}

// ConfirmPayment activates a registration once its deposit has been paid.
// Registered with payments.RegisterConfirmHandler for purpose_type
// "retreat_registration".
func ConfirmPayment(app *pocketbase.PocketBase, registrationID string) error {
	record, err := app.FindRecordById("retreat_registrations", registrationID)
	if err != nil || record == nil {
		return err
	}
	return Activate(app, record)
}
