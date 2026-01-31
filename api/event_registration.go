package api

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/mail"
	"sort"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/mailer"
	"github.com/yuin/goldmark"
	htmlrender "github.com/yuin/goldmark/renderer/html"
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

// RegisterEventHandler creates a registration for an active event by slug.
func RegisterEventHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		slug := e.Request.PathValue("slug")
		if slug == "" {
			return apis.NewBadRequestError(errInvalidEvent, nil)
		}

		event, err := app.FindFirstRecordByFilter(
			"events",
			"slug = {:slug}",
			map[string]any{"slug": slug},
		)
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
		if !isDataSizeOk(payload.Data) {
			return apis.NewBadRequestError(errGeneric, nil)
		}
		recipient, err := normalizeEmail(payload.Email)
		if err != nil {
			return apis.NewBadRequestError(errInvalidEmail, nil)
		}

		registrations, err := app.FindCollectionByNameOrId("event_registrations")
		if err != nil {
			return apis.NewNotFoundError(errGeneric, err)
		}

		existing, err := app.FindFirstRecordByFilter(
			"event_registrations",
			"event = {:event} && email = {:email}",
			map[string]any{
				"event": event.Id,
				"email": recipient,
			},
		)
		if err == nil && existing != nil {
			return apis.NewBadRequestError(errAlreadySubmitted, nil)
		}

		acceptToken, err := generateAcceptToken(app)
		if err != nil {
			return apis.NewBadRequestError(errGeneric, err)
		}

		record := core.NewRecord(registrations)
		record.Set("accept_token", acceptToken)
		record.Set("accept_expires_at", time.Now().UTC().Add(7*24*time.Hour))
		record.Set("event", event.Id)
		record.Set("email", recipient)
		record.Set("accepted", false)
		record.Set("data", payload.Data)

		if err := app.Save(record); err != nil {
			if isUniqueConstraintError(err) {
				return apis.NewBadRequestError(errAlreadySubmitted, err)
			}
			return apis.NewBadRequestError(errGeneric, err)
		}

		emailSent := false
		if templateId := getRelationId(event, "reply_template"); templateId != "" {
			emailSent = sendRegistrationEmail(app, templateId, recipient, payload.Data)
		}
		sendAdminNotification(app, event, recipient, record.GetString("accept_token"), payload.Data)

		return e.JSON(http.StatusCreated, map[string]any{
			"id":         record.Id,
			"email_sent": emailSent,
		})
	}
}

func getRelationId(record *core.Record, field string) string {
	raw := record.Get(field)
	switch value := raw.(type) {
	case string:
		return value
	case []string:
		if len(value) > 0 {
			return value[0]
		}
	case []any:
		if len(value) > 0 {
			if id, ok := value[0].(string); ok {
				return id
			}
		}
	}
	return ""
}

func sendRegistrationEmail(app *pocketbase.PocketBase, templateId string, recipient string, data map[string]any) bool {
	recipientAddress, ok := parseAddress(recipient)
	if !ok {
		return false
	}

	template, err := loadTemplateDataById(app, templateId)
	if err != nil {
		app.Logger().Warn("Failed to load reply template", "error", err)
		return false
	}

	if !templateHasContent(template) {
		return false
	}

	var replyToAddr *mail.Address
	if replyTo := strings.TrimSpace(template.ReplyTo); replyTo != "" {
		if parsed, ok := parseAddress(replyTo); ok {
			replyToAddr = &parsed
		}
	}

	return renderAndSendEmail(
		app,
		[]mail.Address{recipientAddress},
		template.Subject,
		template.Body,
		replyToAddr,
		"Failed to send registration email",
	)
}

func sendAdminNotification(app *pocketbase.PocketBase, event *core.Record, registrantEmail string, acceptToken string, data map[string]any) {
	template, adminAddress, ok := adminTemplateOrWarn(app, event, registrantEmail)
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

func sendAdminTemplateMissing(app *pocketbase.PocketBase, event *core.Record, registrantEmail string) {
	adminAddresses := findSuperuserEmails(app)
	if len(adminAddresses) == 0 {
		return
	}

	eventTitle := ""
	if event != nil {
		eventTitle = strings.TrimSpace(event.GetString("title"))
	}
	body := "Missing admin email template: admin-email-event\n" +
		"Create it in collection \"templates\" with field \"slug\" = \"admin-email-event\".\n" +
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
		"Missing admin-email-event template",
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
	acceptURL := buildAcceptURL(app, token)
	if acceptURL == "" {
		return ""
	}
	return formatter(acceptURL)
}

func normalizeEmail(raw string) (string, error) {
	_, normalized, ok := parseNormalizedEmail(raw)
	if !ok {
		return "", fmt.Errorf("missing email")
	}
	return normalized, nil
}

func parseAddress(raw string) (mail.Address, bool) {
	parsed, _, ok := parseNormalizedEmail(raw)
	return parsed, ok
}

func senderFrom(app *pocketbase.PocketBase) (mail.Address, bool) {
	senderAddress := strings.TrimSpace(app.Settings().Meta.SenderAddress)
	if senderAddress == "" {
		return mail.Address{}, false
	}
	parsed, ok := parseAddress(senderAddress)
	if !ok {
		return mail.Address{}, false
	}
	parsed.Name = app.Settings().Meta.SenderName
	return parsed, true
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
	sender, ok := senderFrom(app)
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

func buildAcceptURL(app *pocketbase.PocketBase, token string) string {
	base := strings.TrimRight(app.Settings().Meta.AppURL, "/")
	if base == "" {
		return ""
	}
	if token == "" {
		return ""
	}
	return base + "/#/event-accept?token=" + token
}

func randomToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func generateAcceptToken(app *pocketbase.PocketBase) (string, error) {
	const attempts = 5
	for i := 0; i < attempts; i++ {
		token := randomToken()
		if token == "" {
			continue
		}
		unique, err := isTokenUnique(app, token)
		if err != nil {
			return "", err
		}
		if unique {
			return token, nil
		}
	}

	return "", fmt.Errorf("unable to generate unique accept token")
}

func isTokenUnique(app *pocketbase.PocketBase, token string) (bool, error) {
	records, err := app.FindRecordsByFilter(
		"event_registrations",
		"accept_token = {:token}",
		"",
		1,
		0,
		map[string]any{"token": token},
	)
	if err != nil {
		return false, err
	}
	return len(records) == 0, nil
}

func renderEmailBody(body string) (string, string) {
	clean := strings.TrimSpace(body)
	if strings.HasPrefix(clean, "md:") {
		clean = strings.TrimSpace(strings.TrimPrefix(clean, "md:"))
	}
	if clean == "" {
		return "", ""
	}
	htmlBody, ok := markdownToHTML(clean)
	if !ok {
		htmlBody = `<div style="white-space:pre-wrap">` + html.EscapeString(clean) + `</div>`
	}
	return clean, htmlBody
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

func parseNormalizedEmail(raw string) (mail.Address, string, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return mail.Address{}, "", false
	}
	parsed, err := mail.ParseAddress(trimmed)
	if err != nil {
		return mail.Address{}, "", false
	}
	normalized := strings.ToLower(strings.TrimSpace(parsed.Address))
	if normalized == "" {
		return mail.Address{}, "", false
	}
	return mail.Address{Name: parsed.Name, Address: normalized}, normalized, true
}

func markdownToHTML(input string) (string, bool) {
	md := goldmark.New(
		goldmark.WithRendererOptions(
			htmlrender.WithUnsafe(),
		),
	)
	var out bytes.Buffer
	if err := md.Convert([]byte(input), &out); err != nil {
		return "", false
	}
	return out.String(), true
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
		if value, ok := data["name"].(string); ok {
			name = strings.TrimSpace(value)
		}
	}
	email := strings.TrimSpace(registrantEmail)
	return eventTitle, name, email
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

func adminTemplateOrWarn(app *pocketbase.PocketBase, event *core.Record, registrantEmail string) (templateData, mail.Address, bool) {
	template, err := loadTemplateDataBySlug(app, "admin-email-event")
	if err != nil {
		app.Logger().Warn("Failed to load admin template", "error", err)
		sendAdminTemplateMissing(app, event, registrantEmail)
		return templateData{}, mail.Address{}, false
	}
	if !templateHasContent(template) {
		sendAdminTemplateMissing(app, event, registrantEmail)
		return templateData{}, mail.Address{}, false
	}
	adminAddress, ok := parseAddress(template.To)
	if !ok {
		sendAdminTemplateMissing(app, event, registrantEmail)
		return templateData{}, mail.Address{}, false
	}
	return template, adminAddress, true
}

func loadTemplateDataById(app *pocketbase.PocketBase, templateId string) (templateData, error) {
	record, err := app.FindRecordById("templates", templateId)
	if err != nil {
		return templateData{}, err
	}
	return parseTemplateData(record)
}

func loadTemplateDataBySlug(app *pocketbase.PocketBase, slug string) (templateData, error) {
	record, err := app.FindFirstRecordByFilter(
		"templates",
		"slug = {:slug}",
		map[string]any{"slug": slug},
	)
	if err != nil {
		return templateData{}, err
	}
	return parseTemplateData(record)
}

func templateHasContent(template templateData) bool {
	return strings.TrimSpace(template.Subject) != "" && strings.TrimSpace(template.Body) != ""
}

func parseTemplateData(record *core.Record) (templateData, error) {
	var template templateData
	if err := record.UnmarshalJSONField("data", &template); err != nil {
		return templateData{}, err
	}
	return template, nil
}

func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "unique")
}
