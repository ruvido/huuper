package public

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/mail"
	"sort"
	"strings"
	"time"

	backendinternal "members/internal"
	eventinternal "members/internal/events"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/core/validators"
	"github.com/pocketbase/pocketbase/tools/mailer"
)

type registrationPayload struct {
	Email string         `json:"email"`
	Data  map[string]any `json:"data"`
}

type templateData struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
	ReplyTo string `json:"reply_to"`
	To      string `json:"to"`
}

const (
	errInvalidEvent     = "invalid_event"
	errInvalidEmail     = "invalid_email"
	errEventClosed      = "event_closed"
	errAlreadySubmitted = "already_submitted"
	errGeneric          = "error_generic"
)

const (
	templateKindUserRegistrationReceived = "events.user.registration_received"
	templateKindUserRegistrationAccepted = "events.user.registration_accepted"
	templateKindAdminNewRegistration     = "events.admin.new_registration"
)

// RegisterEventHandler creates a registration for an active event by slug.
func RegisterEventHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		slug := e.Request.PathValue("slug")
		if slug == "" {
			return apis.NewBadRequestError(errInvalidEvent, nil)
		}

		event, err := eventinternal.FindBySlug(app, slug)
		if err != nil {
			return apis.NewNotFoundError(errInvalidEvent, err)
		}

		if !event.GetBool("active") {
			return apis.NewForbiddenError(errEventClosed, nil)
		}

		eventDate := event.GetDateTime("event_date")
		if eventDate.IsZero() {
			return apis.NewBadRequestError(errInvalidEvent, nil)
		}

		now := time.Now().In(time.Local)
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		eventDay := eventDate.Time().In(now.Location())
		eventDay = time.Date(eventDay.Year(), eventDay.Month(), eventDay.Day(), 0, 0, 0, 0, eventDay.Location())

		if !eventDay.After(today) {
			return apis.NewForbiddenError(errEventClosed, nil)
		}

		var payload registrationPayload
		if err := e.BindBody(&payload); err != nil {
			return apis.NewBadRequestError(errGeneric, err)
		}
		if payload.Data == nil {
			payload.Data = map[string]any{}
		}
		normalizeRegistrationNames(payload.Data)
		recipient, err := normalizeEmail(payload.Email)
		if err != nil {
			return apis.NewBadRequestError(errInvalidEmail, nil)
		}

		linkedUser := e.Auth
		if linkedUser == nil {
			user, err := app.FindFirstRecordByFilter(
				"users",
				"email = {:email}",
				map[string]any{"email": recipient},
			)
			if err == nil && user != nil && strings.TrimSpace(user.GetString("status")) == "active" {
				linkedUser = user
			}
		}

		registrationData := payload.Data
		if linkedUser != nil {
			// Keep required field non-empty without duplicating profile data.
			registrationData = map[string]any{"linked_user": true}
		}
		if !isDataSizeOk(registrationData) {
			return apis.NewBadRequestError(errGeneric, nil)
		}

		registrations, err := app.FindCollectionByNameOrId("event_registrations")
		if err != nil {
			return apis.NewNotFoundError(errGeneric, err)
		}

		existing, err := eventinternal.FindRegistrationByEmail(app, event.Id, recipient, false)
		if err == nil && existing != nil {
			return apis.NewBadRequestError(errAlreadySubmitted, nil)
		}

		acceptToken, err := eventinternal.GenerateAcceptToken(app)
		if err != nil {
			return apis.NewBadRequestError(errGeneric, err)
		}

		record := core.NewRecord(registrations)
		record.Set("accept_token", acceptToken)
		record.Set("accept_expires_at", eventinternal.AcceptTokenExpiryForEvent(event))
		record.Set("event", event.Id)
		if linkedUser != nil {
			record.Set("user", linkedUser.Id)
		}
		record.Set("email", recipient)
		record.Set("status", "pending")
		record.Set("data", registrationData)

		if err := app.Save(record); err != nil {
			if isUniqueConstraintError(err) {
				return apis.NewBadRequestError(errAlreadySubmitted, err)
			}
			return apis.NewBadRequestError(errGeneric, err)
		}

		emailSent := false
		effectiveData := registrationData
		if linkedUser != nil {
			effectiveData = backendinternal.ParseJSONMap(linkedUser.Get("data"))
		}
		if linkedUser != nil {
			if err := eventinternal.ActivateRegistration(app, record, templateKindUserRegistrationAccepted); err != nil {
				return apis.NewBadRequestError(errGeneric, err)
			}
		}
		emailSent = sendRegistrationEmail(app, event, recipient)
		sendAdminNotification(app, event, recipient, record.GetString("accept_token"), effectiveData)

		return e.JSON(http.StatusCreated, map[string]any{
			"id":         record.Id,
			"email_sent": emailSent,
		})
	}
}

func sendRegistrationEmail(app *pocketbase.PocketBase, event *core.Record, recipient string) bool {
	eventID := ""
	if event != nil {
		eventID = event.Id
	}
	return sendUserTemplateEmailByKind(
		app,
		eventID,
		templateKindUserRegistrationReceived,
		recipient,
		"Failed to send registration email",
	)
}

func sendUserTemplateEmailByKind(
	app *pocketbase.PocketBase,
	eventID string,
	kind string,
	recipient string,
	logMessage string,
) bool {
	return eventinternal.SendUserTemplateEmailByKind(app, eventID, kind, recipient, logMessage)
}

func sendAdminNotification(app *pocketbase.PocketBase, event *core.Record, registrantEmail string, acceptToken string, data map[string]any) {
	template, adminAddress, ok := adminTemplateOrWarn(app, event, registrantEmail, templateKindAdminNewRegistration)
	if !ok {
		return
	}

	subject := renderAdminTemplateText(template.Subject, event, registrantEmail, data, acceptToken, app)
	textTemplate := renderAdminTemplateText(template.Body, event, registrantEmail, data, acceptToken, app)
	htmlTemplate := renderAdminTemplateHTML(template.Body, event, registrantEmail, data, acceptToken, app)
	_ = renderAndSendEmailTemplates(
		app,
		[]mail.Address{adminAddress},
		subject,
		textTemplate,
		htmlTemplate,
		nil,
		"Failed to send admin notification",
	)
}

func sendAdminTemplateMissing(app *pocketbase.PocketBase, event *core.Record, registrantEmail string, kind string) {
	adminAddresses := findSuperuserEmails(app)
	if len(adminAddresses) == 0 {
		return
	}

	eventTitle := ""
	if event != nil {
		eventTitle = strings.TrimSpace(event.GetString("title"))
	}
	body := "Missing admin email template kind: " + kind + "\n" +
		"Create it in collection \"templates\" with matching field \"kind\".\n" +
		"Used by /api/events/{slug}/register to notify admins.\n\n" +
		"Suggested content:\n" +
		"Subject: New registration for [event]\n" +
		"Body: New registration for [event]. Email: [email].\n\n" +
		"Placeholders: [event] [name] [email] [data] [accept_button]\n\n" +
		"Event: " + eventTitle + "\n" +
		"Registrant: " + strings.TrimSpace(registrantEmail)

	_ = renderAndSendEmail(
		app,
		adminAddresses,
		"Missing admin template kind",
		body,
		nil,
		"Failed to send admin template warning",
	)
}

func findSuperuserEmails(app *pocketbase.PocketBase) []mail.Address {
	records, err := app.FindRecordsByFilter("_superusers", "", "", 0, 0)
	if err != nil {
		app.Logger().Warn("Failed to load superusers", "error", err)
		return nil
	}

	addresses := make([]mail.Address, 0, len(records))
	for _, record := range records {
		email := strings.TrimSpace(record.GetString("email"))
		if email == "" {
			continue
		}
		if parsed, err := mail.ParseAddress(email); err == nil {
			addresses = append(addresses, mail.Address{Address: parsed.Address, Name: parsed.Name})
		}
	}
	return addresses
}

func renderAdminTemplateText(raw string, event *core.Record, registrantEmail string, data map[string]any, token string, app *pocketbase.PocketBase) string {
	return renderAdminTemplate(
		raw,
		event,
		registrantEmail,
		data,
		token,
		app,
		func(value string) string { return value },
		renderDataText,
		renderAcceptButtonText,
	)
}

func renderAdminTemplateHTML(raw string, event *core.Record, registrantEmail string, data map[string]any, token string, app *pocketbase.PocketBase) string {
	return renderAdminTemplate(
		raw,
		event,
		registrantEmail,
		data,
		token,
		app,
		safeHTML,
		renderDataHTML,
		renderAcceptButtonHTML,
	)
}

func renderAdminTemplate(
	raw string,
	event *core.Record,
	registrantEmail string,
	data map[string]any,
	token string,
	app *pocketbase.PocketBase,
	esc func(string) string,
	dataRenderer func(map[string]any) string,
	acceptRenderer func(*pocketbase.PocketBase, string) string,
) string {
	if raw == "" {
		return ""
	}
	eventTitle, name, email := templateVars(event, registrantEmail, data)
	return renderTemplate(raw, []string{
		"[event]", esc(eventTitle),
		"[name]", esc(name),
		"[email]", esc(email),
		"[data]", dataRenderer(data),
		"[accept_button]", acceptRenderer(app, token),
	})
}

func renderAcceptButtonHTML(app *pocketbase.PocketBase, token string) string {
	return renderAcceptButton(app, token, func(url string) string {
		return `<a href="` + html.EscapeString(url) + `" style="display:inline-block;padding:10px 16px;background:#000;color:#fff;text-decoration:none;border-radius:6px">Accetta</a>`
	})
}

func renderAcceptButtonText(app *pocketbase.PocketBase, token string) string {
	return renderAcceptButton(app, token, func(url string) string {
		return url
	})
}

func renderAcceptButton(app *pocketbase.PocketBase, token string, formatter func(string) string) string {
	acceptURL := eventinternal.BuildAcceptURL(app, token)
	if acceptURL == "" {
		return ""
	}
	return formatter(acceptURL)
}

func normalizeEmail(raw string) (string, error) {
	return backendinternal.NormalizeEmail(raw)
}

func parseAddress(raw string) (mail.Address, bool) {
	return eventinternal.ParseAddress(raw)
}

func senderFromEvents(app *pocketbase.PocketBase) (mail.Address, bool) {
	return eventinternal.SenderFromEvents(app)
}

func buildMessage(from mail.Address, to []mail.Address, subject string, text string, html string, replyTo *mail.Address) *mailer.Message {
	message := &mailer.Message{
		From:    from,
		To:      to,
		Subject: subject,
		Text:    text,
		HTML:    html,
	}
	if replyTo != nil {
		message.Headers = map[string]string{
			"Reply-To": replyTo.String(),
		}
	}
	return message
}

func sendEmailBodies(
	app *pocketbase.PocketBase,
	to []mail.Address,
	subject string,
	textBody string,
	htmlBody string,
	replyTo *mail.Address,
	logMessage string,
) bool {
	sender, ok := senderFromEvents(app)
	if !ok {
		return false
	}

	message := buildMessage(sender, to, subject, textBody, htmlBody, replyTo)
	if err := app.NewMailClient().Send(message); err != nil {
		app.Logger().Warn(logMessage, "error", err)
		return false
	}

	return true
}

func renderDataHTML(data map[string]any) string {
	return renderData(data, "<br>", func(key string, value any) string {
		return safeHTML(key) + ": " + safeHTML(stringify(value))
	})
}

func renderDataText(data map[string]any) string {
	return renderData(data, "\n", func(key string, value any) string {
		return key + ": " + stringify(value)
	})
}

func renderData(data map[string]any, sep string, renderPair func(string, any) string) string {
	if data == nil {
		return ""
	}
	var b strings.Builder
	for i, key := range sortedKeys(data) {
		if i > 0 {
			b.WriteString(sep)
		}
		b.WriteString(renderPair(key, data[key]))
	}
	return b.String()
}

func sortedKeys(data map[string]any) []string {
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func stringify(value any) string {
	switch v := value.(type) {
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}


func renderEmailBody(body string) (string, string) {
	return eventinternal.RenderEmailBody(body)
}

func renderAndSendEmail(
	app *pocketbase.PocketBase,
	to []mail.Address,
	subject string,
	body string,
	replyTo *mail.Address,
	logMessage string,
) bool {
	textBody, htmlBody := renderEmailBody(body)
	return sendEmailBodies(app, to, subject, textBody, htmlBody, replyTo, logMessage)
}

func renderAndSendEmailTemplates(
	app *pocketbase.PocketBase,
	to []mail.Address,
	subject string,
	textTemplate string,
	htmlTemplate string,
	replyTo *mail.Address,
	logMessage string,
) bool {
	textBody, _ := renderEmailBody(textTemplate)
	_, htmlBody := renderEmailBody(htmlTemplate)
	return sendEmailBodies(app, to, subject, textBody, htmlBody, replyTo, logMessage)
}

const maxRegistrationDataBytes = 4000

func isDataSizeOk(data map[string]any) bool {
	if data == nil {
		return true
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return false
	}
	return len(raw) <= maxRegistrationDataBytes
}

func templateVars(event *core.Record, registrantEmail string, data map[string]any) (string, string, string) {
	eventTitle := ""
	if event != nil {
		eventTitle = strings.TrimSpace(event.GetString("title"))
	}
	name := ""
	if data != nil {
		if value, ok := data["full_name"].(string); ok {
			name = backendinternal.NormalizePersonName(value)
		}
	}
	email := strings.TrimSpace(registrantEmail)
	return eventTitle, name, email
}


func normalizeRegistrationNames(data map[string]any) {
	if data == nil {
		return
	}
	if value, ok := data["full_name"].(string); ok {
		normalized := backendinternal.NormalizePersonName(value)
		if normalized != "" {
			data["full_name"] = normalized
		}
	}
}

func safeHTML(value string) string {
	return html.EscapeString(value)
}

func renderTemplate(raw string, replacements []string) string {
	out := raw
	for i := 0; i+1 < len(replacements); i += 2 {
		out = strings.ReplaceAll(out, replacements[i], replacements[i+1])
	}
	return out
}

func adminTemplateOrWarn(app *pocketbase.PocketBase, event *core.Record, registrantEmail string, kind string) (templateData, mail.Address, bool) {
	eventID := ""
	if event != nil {
		eventID = event.Id
	}
	template, found, err := loadTemplateDataByKind(app, eventID, kind)
	if err != nil {
		app.Logger().Warn("Failed to load admin template", "error", err)
		sendAdminTemplateMissing(app, event, registrantEmail, kind)
		return templateData{}, mail.Address{}, false
	}
	if !found || strings.TrimSpace(template.Subject) == "" || strings.TrimSpace(template.Body) == "" {
		sendAdminTemplateMissing(app, event, registrantEmail, kind)
		return templateData{}, mail.Address{}, false
	}
	adminAddress, ok := parseAddress(template.To)
	if !ok {
		sendAdminTemplateMissing(app, event, registrantEmail, kind)
		return templateData{}, mail.Address{}, false
	}
	return template, adminAddress, true
}

func loadTemplateDataByKind(app *pocketbase.PocketBase, eventID string, kind string) (templateData, bool, error) {
	template, found, err := eventinternal.LoadTemplateDataByKind(app, eventID, kind)
	return templateData(template), found, err
}

func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	normalized := validators.NormalizeUniqueIndexError(
		err,
		"event_registrations",
		[]string{"event", "email"},
	)
	fieldErrors, ok := normalized.(validation.Errors)
	if !ok {
		return false
	}
	_, hasEvent := fieldErrors["event"]
	_, hasEmail := fieldErrors["email"]
	return hasEvent || hasEmail
}
