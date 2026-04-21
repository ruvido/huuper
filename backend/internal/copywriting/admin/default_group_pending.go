package admin

import (
	"fmt"
	"html"
	"strings"
)

type PendingDefaultGroupUser struct {
	FullName   string
	LocalGroup string
	DaysSince  int
}

type PendingDefaultGroupEmail struct {
	Subject string
	Body    string
}

func BuildPendingDefaultGroupEmail(groupName string, users []PendingDefaultGroupUser) PendingDefaultGroupEmail {
	displayGroup := strings.TrimSpace(groupName)
	if displayGroup == "" {
		displayGroup = "default group"
	}

	subject := fmt.Sprintf("[Huuper] %d utenti non ancora in %s", len(users), displayGroup)

	var body strings.Builder
	body.WriteString("html:")
	body.WriteString(`<div style="margin:0;padding:24px 0;background:#f3f4f6;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;color:#111827;">`)
	body.WriteString(`<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0"><tr><td align="center">`)
	body.WriteString(`<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0" style="max-width:720px;background:#ffffff;border:1px solid #e5e7eb;border-radius:16px;overflow:hidden;">`)

	body.WriteString(`<tr><td style="padding:32px 32px 8px 32px;">`)
	body.WriteString(`<h1 style="margin:0;font-size:24px;line-height:30px;font-weight:700;color:#111827;">Utenti non ancora nel ` + escape(displayGroup) + `</h1>`)
	body.WriteString(`</td></tr>`)

	body.WriteString(`<tr><td style="padding:0 32px 16px 32px;">`)
	body.WriteString(fmt.Sprintf(`<p style="margin:0;font-size:15px;line-height:22px;color:#374151;">Scan al boot: <strong>%d</strong> utenti approvati non sono ancora membri del gruppo Telegram di default. Ogni utente ha (o ha appena ricevuto) un invite link personale.</p>`, len(users)))
	body.WriteString(`</td></tr>`)

	body.WriteString(`<tr><td style="padding:8px 32px 32px 32px;">`)
	body.WriteString(`<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0" style="border-collapse:collapse;">`)
	body.WriteString(`<tr>`)
	body.WriteString(`<th style="text-align:left;padding:10px 8px;border-bottom:2px solid #e5e7eb;font-size:13px;color:#6b7280;text-transform:uppercase;letter-spacing:0.04em;">Nome e cognome</th>`)
	body.WriteString(`<th style="text-align:left;padding:10px 8px;border-bottom:2px solid #e5e7eb;font-size:13px;color:#6b7280;text-transform:uppercase;letter-spacing:0.04em;">Gruppo local</th>`)
	body.WriteString(`<th style="text-align:right;padding:10px 8px;border-bottom:2px solid #e5e7eb;font-size:13px;color:#6b7280;text-transform:uppercase;letter-spacing:0.04em;">Giorni</th>`)
	body.WriteString(`</tr>`)

	for _, u := range users {
		name := strings.TrimSpace(u.FullName)
		if name == "" {
			name = "—"
		}
		local := strings.TrimSpace(u.LocalGroup)
		if local == "" {
			local = "—"
		}
		body.WriteString(`<tr>`)
		body.WriteString(`<td style="padding:10px 8px;border-bottom:1px solid #f3f4f6;font-size:14px;color:#111827;">` + escape(name) + `</td>`)
		body.WriteString(`<td style="padding:10px 8px;border-bottom:1px solid #f3f4f6;font-size:14px;color:#374151;">` + escape(local) + `</td>`)
		body.WriteString(fmt.Sprintf(`<td style="padding:10px 8px;border-bottom:1px solid #f3f4f6;font-size:14px;color:#374151;text-align:right;">%d</td>`, u.DaysSince))
		body.WriteString(`</tr>`)
	}

	body.WriteString(`</table></td></tr>`)
	body.WriteString(`</table></td></tr></table></div>`)

	return PendingDefaultGroupEmail{
		Subject: subject,
		Body:    body.String(),
	}
}

func escape(value string) string {
	return html.EscapeString(strings.TrimSpace(value))
}
