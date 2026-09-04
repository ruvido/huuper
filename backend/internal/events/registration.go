package events

import (
	"encoding/json"
	"fmt"
	"html"
	"net/mail"
	"sort"
	"strings"

	backendinternal "members/backend/internal"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/core/validators"
)

const (
	TemplateKindUserRegistrationReceived = "events.user.registration_received"
	TemplateKindUserRegistrationAccepted = "events.user.registration_accepted"
	TemplateKindUserPaymentLink          = "events.user.payment_link"
	TemplateKindAdminNewRegistration     = "events.admin.new_registration"
	maxRegistrationDataBytes             = 4000
)

func NormalizeRegistrationNames(data map[string]any) {
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

func IsRegistrationDataSizeOK(data map[string]any) bool {
	if data == nil {
		return true
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return false
	}
	return len(raw) <= maxRegistrationDataBytes
}

func SendRegistrationEmail(app *pocketbase.PocketBase, event *core.Record, recipient string) bool {
	eventID := ""
	if event != nil {
		eventID = event.Id
	}
	return SendUserTemplateEmailByKind(
		app,
		eventID,
		TemplateKindUserRegistrationReceived,
		recipient,
		"Failed to send registration email",
	)
}

func SendAdminNotification(app *pocketbase.PocketBase, event *core.Record, registrantEmail string, acceptToken string, data map[string]any) {
	template, adminAddress, ok := adminTemplateOrWarn(app, event, registrantEmail, TemplateKindAdminNewRegistration)
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

func IsUniqueRegistrationConstraintError(err error) bool {
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
		if parsed, ok := ParseAddress(email); ok {
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
	acceptURL := BuildAcceptURL(app, token)
	if acceptURL == "" {
		return ""
	}
	return formatter(acceptURL)
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

func adminTemplateOrWarn(app *pocketbase.PocketBase, event *core.Record, registrantEmail string, kind string) (TemplateData, mail.Address, bool) {
	eventID := ""
	if event != nil {
		eventID = event.Id
	}
	template, found, err := LoadTemplateDataByKind(app, eventID, kind)
	if err != nil {
		app.Logger().Warn("Failed to load admin template", "error", err)
		sendAdminTemplateMissing(app, event, registrantEmail, kind)
		return TemplateData{}, mail.Address{}, false
	}
	if !found || strings.TrimSpace(template.Subject) == "" || strings.TrimSpace(template.Body) == "" {
		sendAdminTemplateMissing(app, event, registrantEmail, kind)
		return TemplateData{}, mail.Address{}, false
	}
	adminAddress, ok := ParseAddress(template.To)
	if !ok {
		sendAdminTemplateMissing(app, event, registrantEmail, kind)
		return TemplateData{}, mail.Address{}, false
	}
	return template, adminAddress, true
}
