package requests

import (
	"html"
	"strings"
)

type EmailTemplate struct {
	Kind    string
	Subject string
	Body    string
}

func emailLayout(preheader string, title string, intro string, rows [][2]string) string {
	var b strings.Builder
	b.WriteString("html:")
	b.WriteString(`<div style="margin:0;padding:24px 0;background:#f3f4f6;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;color:#111827;">`)
	b.WriteString(`<div style="display:none;max-height:0;overflow:hidden;opacity:0;">` + escape(preheader) + `</div>`)
	b.WriteString(`<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0"><tr><td align="center">`)
	b.WriteString(`<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0" style="max-width:640px;background:#ffffff;border:1px solid #e5e7eb;border-radius:16px;overflow:hidden;">`)
	b.WriteString(`<tr><td style="padding:32px 32px 16px 32px;">`)
	b.WriteString(`<div style="font-size:12px;line-height:18px;font-weight:700;letter-spacing:0.08em;text-transform:uppercase;color:#6b7280;">Realmen</div>`)
	b.WriteString(`<h1 style="margin:12px 0 0 0;font-size:28px;line-height:34px;font-weight:700;color:#111827;">` + escape(title) + `</h1>`)
	b.WriteString(`</td></tr>`)
	if strings.TrimSpace(intro) != "" {
		b.WriteString(`<tr><td style="padding:0 32px 20px 32px;"><p style="margin:0;font-size:16px;line-height:24px;color:#374151;">` + escape(intro) + `</p></td></tr>`)
	}
	if len(rows) > 0 {
		b.WriteString(`<tr><td style="padding:0 32px 8px 32px;">`)
		b.WriteString(`<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0" style="border-collapse:collapse;">`)
		for _, row := range rows {
			if strings.TrimSpace(row[1]) == "" {
				continue
			}
			b.WriteString(`<tr>`)
			b.WriteString(`<td style="padding:10px 0;border-top:1px solid #e5e7eb;font-size:13px;line-height:18px;font-weight:600;color:#6b7280;width:140px;vertical-align:top;">` + escape(row[0]) + `</td>`)
			b.WriteString(`<td style="padding:10px 0;border-top:1px solid #e5e7eb;font-size:15px;line-height:22px;color:#111827;vertical-align:top;">` + escape(row[1]) + `</td>`)
			b.WriteString(`</tr>`)
		}
		b.WriteString(`</table></td></tr>`)
	}
	b.WriteString(`<tr><td style="padding:24px 32px 32px 32px;">`)
	b.WriteString(`<a href="https://branco.realmen.it" style="display:inline-block;padding:12px 18px;background:#111827;color:#ffffff;text-decoration:none;border-radius:10px;font-size:15px;line-height:20px;font-weight:600;">Take action</a>`)
	b.WriteString(`</td></tr>`)
	b.WriteString(`</table></td></tr></table></div>`)
	return b.String()
}

func escape(value string) string {
	return html.EscapeString(strings.TrimSpace(value))
}

func row(label string, value string) [2]string {
	return [2]string{label, value}
}

func titleFor(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "Request update"
	}
	return name
}

func ByKind(kind string) (EmailTemplate, bool) {
	switch kind {
	case newRequestTemplate.Kind:
		return newRequestTemplate, true
	case assignGroupTemplate.Kind:
		return assignGroupTemplate, true
	case assignGuardianTemplate.Kind:
		return assignGuardianTemplate, true
	case mentoringTemplate.Kind:
		return mentoringTemplate, true
	case groupApprovedTemplate.Kind:
		return groupApprovedTemplate, true
	case adminApprovedTemplate.Kind:
		return adminApprovedTemplate, true
	default:
		return EmailTemplate{}, false
	}
}
