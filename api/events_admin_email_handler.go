package api

import (
	"net/http"
	"net/mail"
	"strings"

	backendinternal "members/internal"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/mailer"
)

type adminEventEmailPayload struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
	Target  string `json:"target"`
	DryRun  bool   `json:"dry_run"`
}

// AdminEventEmailHandler sends an email to event registrants by target status.
// When dry_run=true it sends the same email only to the authenticated admin.
func AdminEventEmailHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if _, err := backendinternal.RequireAdmin(e); err != nil {
			return err
		}

		eventID := strings.TrimSpace(e.Request.PathValue("id"))
		if eventID == "" {
			return apis.NewBadRequestError("invalid_event", nil)
		}

		if _, err := app.FindRecordById("events", eventID); err != nil {
			return apis.NewNotFoundError("invalid_event", err)
		}

		var payload adminEventEmailPayload
		if err := e.BindBody(&payload); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}

		payload.Subject = strings.TrimSpace(payload.Subject)
		payload.Body = strings.TrimSpace(payload.Body)
		payload.Target = strings.TrimSpace(payload.Target)
		if payload.Subject == "" || payload.Body == "" {
			return apis.NewBadRequestError("missing_subject_or_body", nil)
		}

		if payload.Target != "active" && payload.Target != "pending" && payload.Target != "all" {
			return apis.NewBadRequestError("invalid_target", nil)
		}

		recipientEmails, err := eventRecipientEmails(app, eventID, payload.Target)
		if err != nil {
			return apis.NewBadRequestError("failed_recipients", err)
		}

		sendTo := recipientEmails
		mode := "live"
		if payload.DryRun {
			adminAddress, ok := parseAddress(e.Auth.GetString("email"))
			if !ok {
				return apis.NewBadRequestError("invalid_admin_email", nil)
			}
			sendTo = []mail.Address{adminAddress}
			mode = "dry_run"
		}

		sent, failed := sendPlainEmailToRecipients(app, sendTo, payload.Subject, payload.Body)

		return e.JSON(http.StatusOK, map[string]any{
			"mode":              mode,
			"target":            payload.Target,
			"event_recipients":  len(recipientEmails),
			"actual_recipients": len(sendTo),
			"sent":              sent,
			"failed":            failed,
		})
	}
}

func eventRecipientEmails(app *pocketbase.PocketBase, eventID string, target string) ([]mail.Address, error) {
	filter := "event = {:event} && status = 'active'"
	if target == "pending" {
		filter = "event = {:event} && status = 'pending'"
	}
	if target == "all" {
		filter = "event = {:event} && status != 'cancelled' && status != 'rejected'"
	}

	records, err := app.FindRecordsByFilter(
		"event_registrations",
		filter,
		"",
		0,
		0,
		map[string]any{"event": eventID},
	)
	if err != nil {
		return nil, err
	}

	return uniqueAddressesFromRecords(records), nil
}

func uniqueAddressesFromRecords(records []*core.Record) []mail.Address {
	out := make([]mail.Address, 0, len(records))
	seen := make(map[string]struct{}, len(records))
	for _, record := range records {
		addr, ok := parseAddress(record.GetString("email"))
		if !ok {
			continue
		}
		if _, exists := seen[addr.Address]; exists {
			continue
		}
		seen[addr.Address] = struct{}{}
		out = append(out, addr)
	}
	return out
}

func sendPlainEmailToRecipients(app *pocketbase.PocketBase, recipients []mail.Address, subject string, body string) (int, int) {
	from, ok := senderFromEvents(app)
	if !ok {
		return 0, len(recipients)
	}

	textBody, htmlBody := renderEmailBody(body)

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
