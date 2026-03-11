package events

import (
	"net/mail"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/mailer"
)

type TemplateData struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
	ReplyTo string `json:"reply_to"`
	To      string `json:"to"`
}

func SendUserTemplateEmailByKind(
	app *pocketbase.PocketBase,
	eventID string,
	kind string,
	recipient string,
	logMessage string,
) bool {
	recipientAddress, ok := ParseAddress(recipient)
	if !ok {
		return false
	}

	template, found, err := LoadTemplateDataByKind(app, eventID, kind)
	if err != nil {
		app.Logger().Warn("Failed to load template", "kind", kind, "error", err)
		return false
	}
	if !found || !templateHasContent(template) {
		return false
	}

	var replyToAddr *mail.Address
	if replyTo := strings.TrimSpace(template.ReplyTo); replyTo != "" {
		if parsed, ok := ParseAddress(replyTo); ok {
			replyToAddr = &parsed
		}
	}

	return renderAndSendEmail(
		app,
		[]mail.Address{recipientAddress},
		template.Subject,
		template.Body,
		replyToAddr,
		logMessage,
	)
}

func LoadTemplateDataByKind(app *pocketbase.PocketBase, eventID string, kind string) (TemplateData, bool, error) {
	record, err := findTemplateByKindAndEvent(app, kind, eventID)
	if err != nil {
		return TemplateData{}, false, err
	}
	if record != nil {
		template, err := parseTemplateData(record)
		if err != nil {
			return TemplateData{}, false, err
		}
		return template, true, nil
	}

	if eventID == "" {
		return TemplateData{}, false, nil
	}

	record, err = findTemplateByKindAndEvent(app, kind, "")
	if err != nil {
		return TemplateData{}, false, err
	}
	if record == nil {
		return TemplateData{}, false, nil
	}

	template, err := parseTemplateData(record)
	if err != nil {
		return TemplateData{}, false, err
	}
	return template, true, nil
}

func renderAndSendEmail(
	app *pocketbase.PocketBase,
	to []mail.Address,
	subject string,
	body string,
	replyTo *mail.Address,
	logMessage string,
) bool {
	textBody, htmlBody := RenderEmailBody(body)
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
	textBody, _ := RenderEmailBody(textTemplate)
	_, htmlBody := RenderEmailBody(htmlTemplate)
	return sendEmailBodies(app, to, subject, textBody, htmlBody, replyTo, logMessage)
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
	sender, ok := SenderFromEvents(app)
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

func findTemplateByKindAndEvent(app *pocketbase.PocketBase, kind string, eventID string) (*core.Record, error) {
	if strings.TrimSpace(kind) == "" {
		return nil, nil
	}

	filter := "kind = {:kind} && event = ''"
	params := map[string]any{"kind": kind}
	if strings.TrimSpace(eventID) != "" {
		filter = "kind = {:kind} && event = {:event}"
		params["event"] = strings.TrimSpace(eventID)
	}

	records, err := app.FindRecordsByFilter("templates", filter, "", 1, 0, params)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, nil
	}
	return records[0], nil
}

func templateHasContent(template TemplateData) bool {
	return strings.TrimSpace(template.Subject) != "" && strings.TrimSpace(template.Body) != ""
}

func parseTemplateData(record *core.Record) (TemplateData, error) {
	var template TemplateData
	if err := record.UnmarshalJSONField("data", &template); err != nil {
		return TemplateData{}, err
	}
	return template, nil
}

func ActivateRegistration(app *pocketbase.PocketBase, record *core.Record, acceptedTemplateKind string) error {
	record.Set("status", "active")
	if err := app.Save(record); err != nil {
		return err
	}

	eventID := strings.TrimSpace(record.GetString("event"))
	recipient := strings.TrimSpace(record.GetString("email"))
	_ = SendUserTemplateEmailByKind(
		app,
		eventID,
		acceptedTemplateKind,
		recipient,
		"Failed to send accepted registration email",
	)

	return nil
}
