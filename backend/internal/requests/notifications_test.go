package requests

import (
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase/core"
)

func TestDefaultNotificationTemplateIncludesAssignGroupTelegramCopy(t *testing.T) {
	template, found, err := defaultNotificationTemplate(templateKindAssignGroup)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Fatalf("expected default template to be found")
	}
	if got := strings.TrimSpace(template.Telegram.Body); got == "" {
		t.Fatalf("expected telegram body to be populated")
	}
	if !strings.Contains(template.Telegram.Body, "{{request_url}}") {
		t.Fatalf("expected request_url placeholder in telegram body, got %q", template.Telegram.Body)
	}
}

func TestMergeNotificationTemplateUsesDefaultTelegramBody(t *testing.T) {
	defaultTemplate := NotificationTemplate{}
	defaultTemplate.Email.Subject = "Default subject"
	defaultTemplate.Email.Body = "Default body"
	defaultTemplate.Telegram.Body = "Default telegram"

	template := NotificationTemplate{}
	template.Email.Subject = "Custom subject"
	merged := mergeNotificationTemplate(defaultTemplate, template)

	if merged.Email.Subject != "Custom subject" {
		t.Fatalf("expected custom subject to be preserved, got %q", merged.Email.Subject)
	}
	if merged.Email.Body != "Default body" {
		t.Fatalf("expected default email body to backfill, got %q", merged.Email.Body)
	}
	if merged.Telegram.Body != "Default telegram" {
		t.Fatalf("expected default telegram body to backfill, got %q", merged.Telegram.Body)
	}
}

func TestRequestNotificationValuesIncludeRequestURL(t *testing.T) {
	collection := core.NewBaseCollection("requests")
	collection.Fields.Add(
		&core.TextField{Name: "email"},
		&core.JSONField{Name: "data"},
	)
	record := core.NewRecord(collection)
	record.Set("id", "req123")
	record.Set("email", "candidate@example.com")

	values := requestNotificationValues(nil, record, map[string]any{}, nil, "", "", nil, nil, nil)

	if got := values["request_url"]; got != "https://branco.realmen.it/me/request/?id=req123" {
		t.Fatalf("unexpected request_url: %q", got)
	}
}
