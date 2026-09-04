package retreats

import (
	"net/mail"
	"strings"

	eventinternal "members/backend/internal/events"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// Template kinds reuse the exact same templates collection + email sending
// mechanism as events (eventinternal.LoadTemplateDataByKind /
// SendPlainEmailToRecipients), just with their own `kind` values, always
// looked up without an event scope (templates.event only relates to the
// `events` collection, so retreats always resolve the global-fallback
// template for a kind — see LoadTemplateDataByKind's fallback behavior).
const (
	TemplateKindUserRegistrationReceived = "retreats.user.registration_received"
	TemplateKindUserRegistrationAccepted = "retreats.user.registration_accepted"
	TemplateKindUserPaymentLink          = "retreats.user.payment_link"
	TemplateKindAdminNewRegistration     = "retreats.admin.new_registration"
)

// SendRegistrationEmail notifies the registrant their submission was
// received (pending review for guests, or "check your email" for members).
func SendRegistrationEmail(app *pocketbase.PocketBase, recipient string) bool {
	return sendTemplateEmail(app, TemplateKindUserRegistrationReceived, recipient, nil)
}

// SendAcceptedEmail notifies the registrant their registration is active,
// optionally embedding a Telegram invite link via the [invite_link]
// placeholder.
func SendAcceptedEmail(app *pocketbase.PocketBase, recipient string, inviteLink string) bool {
	return sendTemplateEmail(app, TemplateKindUserRegistrationAccepted, recipient, []string{
		"[invite_link]", inviteLink,
	})
}

// SendPaymentLinkEmail notifies the registrant of the Stripe checkout URL
// for their deposit.
func SendPaymentLinkEmail(app *pocketbase.PocketBase, recipient string, paymentURL string) bool {
	return sendTemplateEmail(app, TemplateKindUserPaymentLink, recipient, []string{
		"[payment_url]", paymentURL,
	})
}

// SendAdminNewRegistrationNotification notifies admins a new (pending)
// guest registration needs review.
func SendAdminNewRegistrationNotification(app *pocketbase.PocketBase, retreat *core.Record, registrantEmail string) {
	template, found, err := eventinternal.LoadTemplateDataByKind(app, "", TemplateKindAdminNewRegistration)
	if err != nil || !found || strings.TrimSpace(template.Subject) == "" || strings.TrimSpace(template.Body) == "" {
		return
	}
	adminAddress, ok := eventinternal.ParseAddress(template.To)
	if !ok {
		return
	}

	title := ""
	if retreat != nil {
		title = strings.TrimSpace(retreat.GetString("title"))
	}
	replacements := []string{
		"[retreat]", title,
		"[email]", strings.TrimSpace(registrantEmail),
	}
	subject := replaceAll(template.Subject, replacements)
	body := replaceAll(template.Body, replacements)

	eventinternal.SendPlainEmailToRecipients(app, []mail.Address{adminAddress}, subject, body)
}

// BroadcastEmail sends a one-off email (admin-authored subject/body, not a
// template) to every retreat_registrations record with status=active for
// the given retreat. Returns (sent, failed) counts.
func BroadcastEmail(app *pocketbase.PocketBase, retreat *core.Record, subject string, body string) (int, int, error) {
	if retreat == nil {
		return 0, 0, nil
	}
	registrations, err := RegistrationsByStatus(app, retreat.Id, "active")
	if err != nil {
		return 0, 0, err
	}

	recipients := make([]mail.Address, 0, len(registrations))
	for _, registration := range registrations {
		if parsed, ok := eventinternal.ParseAddress(registration.GetString("email")); ok {
			recipients = append(recipients, parsed)
		}
	}

	sent, failed := eventinternal.SendPlainEmailToRecipients(app, recipients, subject, body)
	return sent, failed, nil
}

func sendTemplateEmail(app *pocketbase.PocketBase, kind string, recipient string, replacements []string) bool {
	recipientAddress, ok := eventinternal.ParseAddress(recipient)
	if !ok {
		return false
	}
	template, found, err := eventinternal.LoadTemplateDataByKind(app, "", kind)
	if err != nil || !found || strings.TrimSpace(template.Subject) == "" || strings.TrimSpace(template.Body) == "" {
		return false
	}
	subject := replaceAll(template.Subject, replacements)
	body := replaceAll(template.Body, replacements)

	sent, failed := eventinternal.SendPlainEmailToRecipients(app, []mail.Address{recipientAddress}, subject, body)
	return sent == 1 && failed == 0
}

func replaceAll(raw string, replacements []string) string {
	out := raw
	for i := 0; i+1 < len(replacements); i += 2 {
		out = strings.ReplaceAll(out, replacements[i], replacements[i+1])
	}
	return out
}
