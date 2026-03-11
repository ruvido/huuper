package events

import (
	"bytes"
	"html"
	"net/mail"
	"strings"

	requestinternal "members/internal/requests"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/tools/mailer"
	"github.com/yuin/goldmark"
	htmlrender "github.com/yuin/goldmark/renderer/html"
)

const (
	senderScopeGeneral = "general"
	senderScopeEvents  = "events"
)

func ParseAddress(raw string) (mail.Address, bool) {
	parsed, _, ok := parseNormalizedEmail(raw)
	return parsed, ok
}

func SenderFromEvents(app *pocketbase.PocketBase) (mail.Address, bool) {
	return senderFromScope(app, senderScopeEvents)
}

func RenderEmailBody(body string) (string, string) {
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

func SendPlainEmailToRecipients(app *pocketbase.PocketBase, recipients []mail.Address, subject string, body string) (int, int) {
	from, ok := SenderFromEvents(app)
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

func senderFromScope(app *pocketbase.PocketBase, scope string) (mail.Address, bool) {
	raw := ""
	if settingsData, err := requestinternal.FindSettingData(app, "email"); err == nil {
		if scope == senderScopeEvents {
			if eventsFrom, ok := settingsData[senderScopeEvents].(string); ok {
				raw = strings.TrimSpace(eventsFrom)
			}
		}
		if raw == "" {
			if general, ok := settingsData[senderScopeGeneral].(string); ok {
				raw = strings.TrimSpace(general)
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
