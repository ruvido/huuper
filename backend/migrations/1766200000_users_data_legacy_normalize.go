package migrations

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		records, err := app.FindRecordsByFilter("users", "", "", 0, 0)
		if err != nil {
			return err
		}

		for _, record := range records {
			if record == nil {
				continue
			}

			data := map[string]any{}
			_ = record.UnmarshalJSONField("data", &data)
			normalized, changed := normalizeLegacyUserData(data)
			if !changed {
				continue
			}

			record.Set("data", normalized)
			if err := app.SaveNoValidate(record); err != nil {
				return err
			}
		}

		return nil
	}, func(app core.App) error {
		return nil
	})
}

func normalizeLegacyUserData(data map[string]any) (map[string]any, bool) {
	if data == nil {
		return nil, false
	}

	normalized := make(map[string]any, len(data))
	for key, value := range data {
		normalized[key] = value
	}

	changed := false

	// Some legacy payloads were nested as data.data.
	if nested, ok := normalized["data"].(map[string]any); ok && nested != nil {
		for key, value := range nested {
			if _, exists := normalized[key]; !exists {
				normalized[key] = value
			}
		}
		delete(normalized, "data")
		changed = true
	}

	// Fields consumed as scalar strings by backend/frontend.
	for _, key := range []string{
		"full_name",
		"birth_year",
		"region",
		"marital_status",
		"children",
		"work",
		"motivation",
		"why",
	} {
		next, ok := normalizeScalarStringValue(normalized[key])
		if !ok {
			continue
		}
		if !valuesEqual(normalized[key], next) {
			normalized[key] = next
			changed = true
		}
	}

	// Fields consumed as lists by user summary.
	for _, key := range []string{"skills", "interests", "sports"} {
		next, ok := normalizeStringListValue(normalized[key])
		if !ok {
			continue
		}
		if !valuesEqual(normalized[key], next) {
			normalized[key] = next
			changed = true
		}
	}

	// Backward-compat: legacy records may have motivation but no why.
	if strings.TrimSpace(toString(normalized["why"])) == "" {
		if motivation := strings.TrimSpace(toString(normalized["motivation"])); motivation != "" {
			normalized["why"] = motivation
			changed = true
		}
	}

	if fullName := strings.TrimSpace(toString(normalized["full_name"])); fullName != "" {
		normalizedFullName := normalizeDisplayName(fullName)
		if normalizedFullName != fullName {
			normalized["full_name"] = normalizedFullName
			changed = true
		}
	}

	return normalized, changed
}

func normalizeScalarStringValue(value any) (string, bool) {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed), true
	case []string:
		return firstNonEmptyString(typed), true
	case []any:
		items := make([]string, 0, len(typed))
		for _, item := range typed {
			if text, ok := item.(string); ok {
				items = append(items, text)
			}
		}
		return firstNonEmptyString(items), true
	default:
		return "", false
	}
}

func normalizeStringListValue(value any) ([]string, bool) {
	switch typed := value.(type) {
	case string:
		trimmed := strings.TrimSpace(typed)
		if trimmed == "" {
			return []string{}, true
		}
		return []string{trimmed}, true
	case []string:
		return compactUniqueStrings(typed), true
	case []any:
		items := make([]string, 0, len(typed))
		for _, item := range typed {
			if text, ok := item.(string); ok {
				items = append(items, text)
			}
		}
		return compactUniqueStrings(items), true
	default:
		return nil, false
	}
}

func firstNonEmptyString(values []string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func compactUniqueStrings(values []string) []string {
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func valuesEqual(left any, right any) bool {
	return fmt.Sprintf("%#v", left) == fmt.Sprintf("%#v", right)
}

func normalizeDisplayName(value string) string {
	parts := strings.Fields(strings.TrimSpace(value))
	if len(parts) == 0 {
		return ""
	}

	out := make([]string, 0, len(parts))
	for _, part := range parts {
		out = append(out, capitalizeWord(part))
	}
	return strings.Join(out, " ")
}

func capitalizeWord(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	first, size := utf8.DecodeRuneInString(trimmed)
	if first == utf8.RuneError {
		return trimmed
	}

	var b strings.Builder
	b.Grow(len(trimmed))
	b.WriteRune(unicode.ToUpper(first))
	for _, r := range trimmed[size:] {
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}
