package api

import (
	"strings"
	"unicode"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterUsersNormalizationHooks normalizes selected user profile fields.
func RegisterUsersNormalizationHooks(app *pocketbase.PocketBase) {
	app.OnRecordValidate("users").BindFunc(func(e *core.RecordEvent) error {
		normalizeUserProfileData(e.Record)
		return e.Next()
	})
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
