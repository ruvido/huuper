package events

import (
	"net/mail"
	"strings"

	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// DepositCentsForEvent returns the configured deposit amount (in cents) for
// the event's type, or 0 if registration for that type has no deposit.
func DepositCentsForEvent(app *pocketbase.PocketBase, event *core.Record) int {
	if event == nil {
		return 0
	}
	cfg, err := LoadConfig(app)
	if err != nil {
		return 0
	}
	typeDef, ok := cfg.TypeDef(event.GetString("type"))
	if !ok {
		return 0
	}
	return typeDef.Registration.DepositCents
}

func PaymentSuccessURL(app *pocketbase.PocketBase) string {
	base := strings.TrimRight(app.Settings().Meta.AppURL, "/")
	return base + "/event-payment/?status=success"
}

func PaymentCancelURL(app *pocketbase.PocketBase) string {
	base := strings.TrimRight(app.Settings().Meta.AppURL, "/")
	return base + "/event-payment/?status=cancelled"
}

// MarkAwaitingPayment moves a registration to awaiting_payment and emails the
// registrant the Stripe checkout link. Used instead of ActivateRegistration
// whenever the event type has a deposit configured.
func MarkAwaitingPayment(app *pocketbase.PocketBase, record *core.Record, paymentURL string) error {
	record.Set("status", "awaiting_payment")
	data := backendinternal.ParseJSONMap(record.Get("data"))
	data["payment_url"] = paymentURL
	record.Set("data", data)
	if err := app.Save(record); err != nil {
		return err
	}
	sendPaymentLinkEmail(app, record, paymentURL)
	return nil
}

// ConfirmPayment activates a registration once its deposit has been paid.
// Registered with payments.RegisterConfirmHandler for purpose_type
// "event_registration".
func ConfirmPayment(app *pocketbase.PocketBase, registrationID string) error {
	record, err := app.FindRecordById("event_registrations", registrationID)
	if err != nil || record == nil {
		return err
	}
	return ActivateRegistration(app, record, TemplateKindUserRegistrationAccepted)
}

func sendPaymentLinkEmail(app *pocketbase.PocketBase, record *core.Record, paymentURL string) bool {
	recipientAddress, ok := ParseAddress(record.GetString("email"))
	if !ok {
		return false
	}

	template, found, err := LoadTemplateDataByKind(app, record.GetString("event"), TemplateKindUserPaymentLink)
	if err != nil {
		app.Logger().Warn("Failed to load payment link template", "error", err)
		return false
	}
	if !found || !templateHasContent(template) {
		app.Logger().Warn("Missing template kind", "kind", TemplateKindUserPaymentLink)
		return false
	}

	subject := renderTemplate(template.Subject, []string{"[payment_url]", paymentURL})
	body := renderTemplate(template.Body, []string{"[payment_url]", paymentURL})

	var replyToAddr *mail.Address
	if replyTo := strings.TrimSpace(template.ReplyTo); replyTo != "" {
		if parsed, ok := ParseAddress(replyTo); ok {
			replyToAddr = &parsed
		}
	}

	return renderAndSendEmail(
		app,
		[]mail.Address{recipientAddress},
		subject,
		body,
		replyToAddr,
		"Failed to send payment link email",
	)
}
