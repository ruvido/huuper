package api

import (
	"strings"
	"unicode"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterUsersNormalizationHooks normalizes selected user profile fields.
func RegisterUsersNormalizationHooks(app *pocketbase.PocketBase) {
	app.OnRecordValidate("users").BindFunc(func(e *core.RecordEvent) error {
		if err := normalizeUserEmailAndValidateUnique(e.App, e.Record); err != nil {
			return err
		}
		normalizeUserProfileData(e.Record)
		return e.Next()
	})
}

func normalizeUserEmailAndValidateUnique(app core.App, record *core.Record) error {
	if record == nil {
		return nil
	}

	email := strings.ToLower(strings.TrimSpace(record.GetString("email")))
	if email == "" {
		return nil
	}
	record.Set("email", email)

	records, err := app.FindRecordsByFilter("users", "", "", 0, 0)
	if err != nil {
		return apis.NewBadRequestError("failed_to_validate_email_uniqueness", err)
	}

	for _, existing := range records {
		if existing == nil || existing.Id == record.Id {
			continue
		}
		existingEmail := strings.TrimSpace(existing.GetString("email"))
		if existingEmail == "" {
			continue
		}
		if strings.EqualFold(existingEmail, email) {
			return apis.NewBadRequestError("email_already_exists", nil)
		}
	}

	return nil
}

func normalizeUserProfileData(record *core.Record) {
	if record == nil {
		return
	}

	data := parseJSONMap(record.Get("data"))
	value, ok := data["full_name"].(string)
	if !ok {
		return
	}

	normalized := normalizePersonName(value)
	if normalized == "" {
		return
	}

	data["full_name"] = normalized
	record.Set("data", data)
}

func normalizePersonName(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	parts := strings.Fields(strings.ToLower(trimmed))
	for i, part := range parts {
		runes := []rune(part)
		if len(runes) == 0 {
			continue
		}
		runes[0] = unicode.ToUpper(runes[0])
		for j := 1; j < len(runes); j++ {
			runes[j] = unicode.ToLower(runes[j])
		}
		parts[i] = string(runes)
	}
	return strings.Join(parts, " ")
}
