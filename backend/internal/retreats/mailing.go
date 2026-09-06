package retreats

import (
	"fmt"
	"net/mail"
	"strings"
	"time"

	backendinternal "members/backend/internal"
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
	TemplateKindAdminRegistrationDone    = "retreats.admin.registration_completed"
	TemplateKindAdminDailyStats          = "retreats.admin.daily_stats"
)

var italianMonths = [...]string{
	"gennaio", "febbraio", "marzo", "aprile", "maggio", "giugno",
	"luglio", "agosto", "settembre", "ottobre", "novembre", "dicembre",
}

func formatDay(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return fmt.Sprintf("%d %s %d", t.Day(), italianMonths[int(t.Month())-1], t.Year())
}

// formatDateRange keeps a same-month range compact: "8 – 11 ottobre 2026".
func formatDateRange(start, end time.Time) string {
	if start.IsZero() {
		return ""
	}
	if end.IsZero() || start.Equal(end) {
		return formatDay(start)
	}
	if start.Month() == end.Month() && start.Year() == end.Year() {
		return fmt.Sprintf("%d – %d %s %d", start.Day(), end.Day(), italianMonths[int(start.Month())-1], start.Year())
	}
	return formatDay(start) + " – " + formatDay(end)
}

func formatMoney(cents int) string {
	if cents <= 0 {
		return ""
	}
	return fmt.Sprintf("%d,%02d €", cents/100, cents%100)
}

// retreatPlaceholders exposes the retreat's own content to the templates, so
// an email can carry the dates, the place and how to get there without any of
// it being written into the code. Every value comes from the record.
func retreatPlaceholders(retreat *core.Record) []string {
	if retreat == nil {
		return nil
	}
	data := parseData(retreat)
	arrival, _ := data["arrival"].(map[string]any)
	arrivalText := func(key string) string {
		if arrival == nil {
			return ""
		}
		value, _ := arrival[key].(string)
		return value
	}

	price := DataInt(data, "price_cents")
	deposit := DataInt(data, "deposit_cents")
	balance := 0
	if price > deposit {
		balance = price - deposit
	}

	dataString := func(key string) string {
		value, _ := data[key].(string)
		return value
	}

	return []string{
		"[retreat]", strings.TrimSpace(retreat.GetString("title")),
		"[tagline]", strings.TrimSpace(retreat.GetString("tagline")),
		"[dates]", formatDateRange(retreat.GetDateTime("start_date").Time(), retreat.GetDateTime("end_date").Time()),
		"[start_date]", formatDay(retreat.GetDateTime("start_date").Time()),
		"[end_date]", formatDay(retreat.GetDateTime("end_date").Time()),
		"[location]", strings.TrimSpace(retreat.GetString("location")),
		"[place]", dataString("place_name"),
		"[region]", dataString("region"),
		"[meeting_time]", dataString("meeting_time_note"),
		"[arrival_car]", arrivalText("auto"),
		"[arrival_train]", arrivalText("train"),
		"[arrival_carpooling]", arrivalText("carpooling"),
		"[price]", formatMoney(price),
		"[deposit]", formatMoney(deposit),
		"[balance]", formatMoney(balance),
		"[contact_email]", dataString("contact_email"),
	}
}

// SendRegistrationEmail notifies the registrant their submission was
// received (pending review for guests, or "check your email" for members).
func SendRegistrationEmail(app *pocketbase.PocketBase, retreat *core.Record, recipient string) bool {
	return sendTemplateEmail(app, TemplateKindUserRegistrationReceived, recipient, retreatPlaceholders(retreat))
}

// SendAcceptedEmail notifies the registrant their registration is active,
// optionally embedding a Telegram invite link via the [invite_link]
// placeholder.
func SendAcceptedEmail(app *pocketbase.PocketBase, retreat *core.Record, recipient string, inviteLink string) bool {
	return sendTemplateEmail(app, TemplateKindUserRegistrationAccepted, recipient,
		append(retreatPlaceholders(retreat), "[invite_link]", inviteLink))
}

// SendPaymentLinkEmail notifies the registrant of the Stripe checkout URL
// for their deposit.
func SendPaymentLinkEmail(app *pocketbase.PocketBase, retreat *core.Record, recipient string, paymentURL string) bool {
	return sendTemplateEmail(app, TemplateKindUserPaymentLink, recipient,
		append(retreatPlaceholders(retreat), "[payment_url]", paymentURL))
}

// SendAdminNewRegistrationNotification notifies admins a new (pending)
// guest registration needs review.
func SendAdminNewRegistrationNotification(app *pocketbase.PocketBase, retreat *core.Record, registration *core.Record, registrantEmail string) {
	template, found, err := eventinternal.LoadTemplateDataByKind(app, "", TemplateKindAdminNewRegistration)
	if err != nil || !found || strings.TrimSpace(template.Subject) == "" || strings.TrimSpace(template.Body) == "" {
		return
	}
	adminAddress, ok := adminRecipient(app, TemplateKindAdminNewRegistration)
	if !ok {
		// The recipient lives only on the template record, so a missing one
		// would drop a real person's registration on the floor. Tell whoever
		// can fix it rather than returning quietly.
		warnAdminRecipientMissing(app, TemplateKindAdminNewRegistration, "registration from "+strings.TrimSpace(registrantEmail))
		return
	}

	title := ""
	if retreat != nil {
		title = strings.TrimSpace(retreat.GetString("title"))
	}
	// Everything the organiser needs to call this person back, straight from
	// what they filled in — plus a one-click link that runs the same approval
	// as the admin API, so it can be done from a phone.
	registrationData := map[string]any{}
	acceptURL := ""
	if registration != nil {
		registrationData = backendinternal.ParseJSONMap(registration.Get("data"))
		if token := strings.TrimSpace(registration.GetString("accept_token")); token != "" {
			base := strings.TrimRight(app.Settings().Meta.AppURL, "/")
			acceptURL = base + "/api/public/retreats/accept?token=" + token
		}
	}
	field := func(key string) string {
		value, _ := registrationData[key].(string)
		return strings.TrimSpace(value)
	}

	replacements := append(retreatPlaceholders(retreat),
		"[retreat]", title,
		"[email]", strings.TrimSpace(registrantEmail),
		"[name]", field("full_name"),
		"[phone]", field("mobile"),
		"[birth_year]", field("birth_year"),
		"[provenance]", field("provenance"),
		"[accept_url]", acceptURL,
	)
	subject := replaceAll(template.Subject, replacements)
	body := replaceAll(template.Body, replacements)

	eventinternal.SendPlainEmailToRecipients(app, []mail.Address{adminAddress}, subject, body)
}

// adminRecipient resolves where a retreat notification to the organiser goes.
// The address lives on templates."retreats.admin.new_registration".data.to —
// one place for the whole module, so adding another organiser-facing email
// does not mean setting the same address again somewhere else. A template may
// still carry its own `to` to divert just that one.
func adminRecipient(app *pocketbase.PocketBase, kind string) (mail.Address, bool) {
	if template, found, err := eventinternal.LoadTemplateDataByKind(app, "", kind); err == nil && found {
		if address, ok := eventinternal.ParseAddress(template.To); ok {
			return address, true
		}
	}
	if kind == TemplateKindAdminNewRegistration {
		return mail.Address{}, false
	}
	template, found, err := eventinternal.LoadTemplateDataByKind(app, "", TemplateKindAdminNewRegistration)
	if err != nil || !found {
		return mail.Address{}, false
	}
	return eventinternal.ParseAddress(template.To)
}

// sendAdminTemplateEmail renders one of the organiser-facing templates and
// sends it to whoever adminRecipient resolves.
func sendAdminTemplateEmail(app *pocketbase.PocketBase, kind string, replacements []string) bool {
	template, found, err := eventinternal.LoadTemplateDataByKind(app, "", kind)
	if err != nil || !found || strings.TrimSpace(template.Subject) == "" || strings.TrimSpace(template.Body) == "" {
		return false
	}
	address, ok := adminRecipient(app, kind)
	if !ok {
		warnAdminRecipientMissing(app, kind, "")
		return false
	}
	subject := replaceAll(template.Subject, replacements)
	body := replaceAll(template.Body, replacements)
	sent, failed := eventinternal.SendPlainEmailToRecipients(app, []mail.Address{address}, subject, body)
	return sent == 1 && failed == 0
}

// SendAdminRegistrationCompletedNotification tells the organiser a place has
// actually been taken.
//
// Nothing used to reach them when a member registered and paid: the only
// organiser-facing email was the guest pre-registration, so a member could
// sign up, pay the deposit and turn up in October without anyone being told.
func SendAdminRegistrationCompletedNotification(app *pocketbase.PocketBase, retreat *core.Record, registration *core.Record) {
	if registration == nil {
		return
	}
	data := backendinternal.ParseJSONMap(registration.Get("data"))
	field := func(key string) string {
		value, _ := data[key].(string)
		return strings.TrimSpace(value)
	}

	replacements := append(retreatPlaceholders(retreat),
		"[email]", strings.TrimSpace(registration.GetString("email")),
		"[name]", field("full_name"),
		"[phone]", field("mobile"),
	)
	// The running totals travel with it, so the organiser sees where the
	// retreat stands without opening anything.
	if stats, err := CountRegistrations(app, retreat); err == nil {
		replacements = append(replacements, statsPlaceholders(stats)...)
	}

	sendAdminTemplateEmail(app, TemplateKindAdminRegistrationDone, replacements)
}

// warnAdminRecipientMissing emails the superusers when an organiser-facing
// template has no usable `to`, naming what could not be announced so it can be
// picked up by hand.
func warnAdminRecipientMissing(app *pocketbase.PocketBase, kind string, detail string) {
	app.Logger().Warn(
		"retreats: admin notification has no recipient",
		"kind", kind,
		"detail", detail,
	)
	recipients := eventinternal.SuperuserAddresses(app)
	if len(recipients) == 0 {
		return
	}
	body := "A retreat notification could not be delivered: the template \"" + kind +
		"\" has no valid \"to\" address, and neither does \"" +
		TemplateKindAdminNewRegistration + "\".\n\n" +
		"Set it in collection \"templates\" on that record."
	if strings.TrimSpace(detail) != "" {
		body += "\n\nWhat went unannounced: " + strings.TrimSpace(detail)
	}
	eventinternal.SendPlainEmailToRecipients(app, recipients, "Retreat notification not delivered", body)
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
