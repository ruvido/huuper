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

// The slug travels in the return URL so the result page can show wording the
// organiser wrote for this retreat instead of one generic sentence.
func PaymentSuccessURL(app *pocketbase.PocketBase, retreat *core.Record) string {
	return paymentReturnURL(app, retreat, "success")
}

func PaymentCancelURL(app *pocketbase.PocketBase, retreat *core.Record) string {
	return paymentReturnURL(app, retreat, "cancelled")
}

// AcceptResultURL is where the approve-from-email link lands the organiser.
// The click has to answer a human on a phone, not a script: it returns a page
// instead of JSON. Only the outcome and the slug travel in the URL — never the
// registrant's email, which would end up in browser history and referrers.
func AcceptResultURL(app *pocketbase.PocketBase, retreat *core.Record, status string) string {
	return returnURL(app, retreat, "/retreat-accept/", status)
}

func paymentReturnURL(app *pocketbase.PocketBase, retreat *core.Record, status string) string {
	return returnURL(app, retreat, "/retreat-payment/", status)
}

func returnURL(app *pocketbase.PocketBase, retreat *core.Record, page string, status string) string {
	base := strings.TrimRight(app.Settings().Meta.AppURL, "/")
	url := base + page + "?status=" + status
	if retreat != nil {
		if slug := strings.TrimSpace(retreat.GetString("slug")); slug != "" {
			url += "&retreat=" + slug
		}
	}
	return url
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
	retreat, _ := app.FindRecordById("retreats", record.GetString("retreat"))
	SendPaymentLinkEmail(app, retreat, record.GetString("email"), paymentURL)
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
