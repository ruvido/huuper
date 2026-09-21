package retreats

import (
	"fmt"
	"strings"
	"time"

	backendinternal "members/backend/internal"
	paymentsinternal "members/backend/internal/payments"

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

// AcceptPageURL is the link in the organiser's email: a page that shows the
// request and asks before approving. Only the token travels in the URL — never
// the registrant's details, which would end up in browser history and
// referrers — and opening it changes nothing, so a mail client that prefetches
// links cannot approve anyone by itself.
func AcceptPageURL(app *pocketbase.PocketBase, token string) string {
	base := strings.TrimRight(app.Settings().Meta.AppURL, "/")
	return base + "/retreat-accept/?token=" + strings.TrimSpace(token)
}

// ArchivePageURL opens the same review page ready to put the request aside:
// the organiser who has already decided on the phone should not have to go
// through "confirm?" to reach "archive".
func ArchivePageURL(app *pocketbase.PocketBase, token string) string {
	return AcceptPageURL(app, token) + "&action=archive"
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

// MarkAwaitingPayment moves a registration to awaiting_payment, and emails the
// registrant the Stripe checkout link when they are not already on their way
// to it (see resolveActivation).
func MarkAwaitingPayment(app *pocketbase.PocketBase, record *core.Record, paymentURL string, emailPaymentLink bool) error {
	record.Set("status", "awaiting_payment")
	data := backendinternal.ParseJSONMap(record.Get("data"))
	data["payment_url"] = paymentURL
	record.Set("data", data)
	if err := app.Save(record); err != nil {
		return err
	}
	if !emailPaymentLink {
		return nil
	}
	retreat, _ := app.FindRecordById("retreats", record.GetString("retreat"))
	if SendPaymentLinkEmail(app, retreat, record.GetString("email"), paymentURL) {
		_ = countPaymentNotice(app, record)
	}
	return nil
}

// countPaymentNotice records that the deposit has been asked for once more. The
// figure the organiser needs is how many times this person has been told to
// pay — the first email included — not how many times the reminder job ran.
//
// It only ever adds one to what is written down. Deriving a starting point
// from the reminder counter assumed a first email that, on the very first
// send, had not happened yet, and recorded 2 for a person written to once.
func countPaymentNotice(app *pocketbase.PocketBase, record *core.Record) error {
	data := backendinternal.ParseJSONMap(record.Get("data"))
	data["payment_notices_sent"] = DataInt(data, "payment_notices_sent") + 1
	data["payment_notice_at"] = time.Now().UTC().Format(time.RFC3339)
	record.Set("data", data)
	return app.Save(record)
}

// PaymentRetries is how many times we went back to them after that first email.
func PaymentRetries(record *core.Record) int {
	data := backendinternal.ParseJSONMap(record.Get("data"))
	return DataInt(data, "payment_reminders_sent")
}

// ResumeCheckout issues a fresh Stripe session for a registration that is
// still waiting for its deposit, and returns the new checkout URL.
//
// Closing the Stripe page is not a mistake and not a withdrawal: the person
// signed up and simply did not finish paying. Without this they are stuck —
// the unique index on (retreat, email) refuses a second registration, so they
// can neither pay nor sign up again.
//
// The abandoned session is left alone rather than deleted: its `payments`
// record is how the webhook recognises a payment, and the visitor may still
// have that tab open. Stripe expires unused sessions on its own.
func ResumeCheckout(app *pocketbase.PocketBase, retreat *core.Record, registration *core.Record) (string, error) {
	depositCents := DepositCentsForRetreat(retreat)
	if depositCents <= 0 {
		return "", fmt.Errorf("retreat has no deposit to pay")
	}

	_, url, err := paymentsinternal.CreateCheckoutSession(app, paymentsinternal.CheckoutInput{
		PurposeType: "retreat_registration",
		PurposeID:   registration.Id,
		Email:       registration.GetString("email"),
		AmountCents: int64(depositCents),
		Currency:    "eur",
		ProductName: strings.TrimSpace(retreat.GetString("title")) + " - deposit",
		SuccessURL:  PaymentSuccessURL(app, retreat),
		CancelURL:   PaymentCancelURL(app, retreat),
	})
	if err != nil {
		return "", err
	}

	// No email here, unlike MarkAwaitingPayment: the visitor asked for this
	// from the page and is about to be redirected to it.
	data := backendinternal.ParseJSONMap(registration.Get("data"))
	data["payment_url"] = url
	registration.Set("data", data)
	if err := app.Save(registration); err != nil {
		return "", err
	}
	return url, nil
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
