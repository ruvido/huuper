package internal

import (
	"html"
	"net/mail"
	"regexp"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/tools/mailer"
)

const (
	EmailSenderKeyGeneral  = "general"
	EmailSenderKeyEvents   = "events"
	EmailRecipientKeyAdmin = "admin"
)

func ParseAddress(raw string) (mail.Address, bool) {
	parsed, _, ok := ParseNormalizedEmail(raw)
	return parsed, ok
}

func RenderEmailBody(body string) (string, string) {
	clean := strings.TrimSpace(body)
	if strings.HasPrefix(clean, "html:") {
		htmlBody := strings.TrimSpace(strings.TrimPrefix(clean, "html:"))
		return htmlToText(htmlBody), htmlBody
	}
	if strings.HasPrefix(clean, "md:") {
		clean = strings.TrimSpace(strings.TrimPrefix(clean, "md:"))
	}
	if clean == "" {
		return "", ""
	}
	htmlBody, ok := RenderMarkdownHTML(clean)
	if !ok {
		htmlBody = `<div style="white-space:pre-wrap">` + html.EscapeString(clean) + `</div>`
	}
	return clean, htmlBody
}

var htmlTagPattern = regexp.MustCompile(`<[^>]+>`)

func htmlToText(raw string) string {
	text := raw
	replacements := []struct {
		old string
		new string
	}{
		{"<br>", "\n"},
		{"<br/>", "\n"},
		{"<br />", "\n"},
		{"</p>", "\n\n"},
		{"</div>", "\n"},
		{"</tr>", "\n"},
		{"</table>", "\n"},
		{"</h1>", "\n\n"},
		{"</h2>", "\n\n"},
	}
	for _, replacement := range replacements {
		text = strings.ReplaceAll(text, replacement.old, replacement.new)
		text = strings.ReplaceAll(text, strings.ToUpper(replacement.old), replacement.new)
	}
	text = htmlTagPattern.ReplaceAllString(text, "")
	text = html.UnescapeString(text)

	lines := strings.Split(text, "\n")
	out := make([]string, 0, len(lines))
	empty := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if empty {
				continue
			}
			empty = true
			out = append(out, "")
			continue
		}
		empty = false
		out = append(out, trimmed)
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

func SendPlainEmailToRecipients(app *pocketbase.PocketBase, recipients []mail.Address, subject string, body string, senderKey string) (int, int) {
	from, ok := SenderFromEmailSettings(app, senderKey)
	if !ok {
		return 0, len(recipients)
	}

	textBody, htmlBody := RenderEmailBody(body)

	sent := 0
	failed := 0
	for _, recipient := range recipients {
		message := &mailer.Message{
			From:    from,
			To:      []mail.Address{recipient},
			Subject: subject,
			Text:    textBody,
			HTML:    htmlBody,
		}
		if err := app.NewMailClient().Send(message); err == nil {
			sent++
		} else {
			failed++
		}
	}
	return sent, failed
}

func SenderFromEmailSettings(app *pocketbase.PocketBase, preferredKey string) (mail.Address, bool) {
	raw := ""
	if settingsData, err := FindSettingData(app, "email"); err == nil {
		if preferredKey != "" {
			if value, ok := settingsData[preferredKey].(string); ok {
				raw = strings.TrimSpace(value)
			}
		}
		if raw == "" {
			if value, ok := settingsData[EmailSenderKeyGeneral].(string); ok {
				raw = strings.TrimSpace(value)
			}
		}
	}
	if raw == "" {
		raw = strings.TrimSpace(app.Settings().Meta.SenderAddress)
	}
	if raw == "" {
		return mail.Address{}, false
	}

	parsed, ok := ParseAddress(raw)
	if !ok {
		return mail.Address{}, false
	}
	if strings.TrimSpace(parsed.Name) == "" {
		parsed.Name = app.Settings().Meta.SenderName
	}
	return parsed, true
}

func RecipientFromEmailSettings(app *pocketbase.PocketBase, key string) (mail.Address, bool) {
	raw := ""
	if settingsData, err := FindSettingData(app, "email"); err == nil {
		if value, ok := settingsData[key].(string); ok {
			raw = strings.TrimSpace(value)
		}
	}
	if raw == "" {
		return mail.Address{}, false
	}
	return ParseAddress(raw)
}

func AdminRecipientFromEmailSettings(app *pocketbase.PocketBase) (mail.Address, bool) {
	return RecipientFromEmailSettings(app, EmailRecipientKeyAdmin)
}

func SendAdminFailureEmail(app *pocketbase.PocketBase, subject string, body string) bool {
	recipient, ok := AdminRecipientFromEmailSettings(app)
	if !ok {
		return false
	}
	sent, _ := SendPlainEmailToRecipients(app, []mail.Address{recipient}, subject, body, EmailSenderKeyGeneral)
	return sent > 0
}
