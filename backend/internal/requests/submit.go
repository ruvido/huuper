package requests

import (
	"fmt"
	"strings"

	backendinternal "members/backend/internal"
)

func ValidateAndBuildData(input map[string]any, signup SignupSettingsConfig, profile ProfileSchemaConfig) (map[string]any, string, error) {
	if len(signup.Steps) == 0 {
		return nil, "", fmt.Errorf("signup steps not configured")
	}

	profileByKey := make(map[string]ProfileFieldConfig, len(profile.Fields))
	for _, field := range profile.Fields {
		key := strings.TrimSpace(field.Key)
		if key == "" {
			continue
		}
		profileByKey[key] = field
	}
	if len(profileByKey) == 0 {
		return nil, "", fmt.Errorf("profile fields not configured")
	}

	allowed := make(map[string]struct{}, len(signup.Steps))
	for _, step := range signup.Steps {
		key := strings.TrimSpace(step.Field)
		if key == "" {
			continue
		}
		if _, ok := profileByKey[key]; !ok {
			return nil, "", fmt.Errorf("signup step field not in profile_schema: %s", key)
		}
		allowed[key] = struct{}{}
	}
	if len(allowed) == 0 {
		return nil, "", fmt.Errorf("signup fields not configured")
	}

	for key := range input {
		if _, ok := allowed[key]; !ok {
			return nil, "", fmt.Errorf("unknown field: %s", key)
		}
	}

	out := map[string]any{}
	for key := range allowed {
		if value, ok := input[key]; ok {
			profileField := profileByKey[key]
			normalized, err := normalizeRequestFieldValue(profileField, value)
			if err != nil {
				return nil, "", fmt.Errorf("invalid field %s: %w", key, err)
			}
			out[key] = normalized
		}
	}

	for key := range allowed {
		if !hasNonEmptyValue(out[key]) {
			return nil, "", fmt.Errorf("missing required field: %s", key)
		}
	}

	email, ok := out["email"].(string)
	if !ok || strings.TrimSpace(email) == "" {
		return nil, "", fmt.Errorf("missing required field: email")
	}
	normalizedEmail, err := backendinternal.NormalizeEmail(email)
	if err != nil {
		return nil, "", fmt.Errorf("invalid email")
	}
	delete(out, "email")

	return out, normalizedEmail, nil
}

func hasNonEmptyValue(value any) bool {
	if value == nil {
		return false
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed) != ""
	case []string:
		return len(typed) > 0
	case []any:
		return len(typed) > 0
	default:
		return true
	}
}

func normalizeRequestFieldValue(field ProfileFieldConfig, value any) (any, error) {
	fieldType := strings.ToLower(strings.TrimSpace(field.Type))
	switch fieldType {
	case "select":
		return normalizeSelectFieldValue(field, value)
	case "textarea", "text", "phone", "email", "":
		return normalizeStringFieldValue(value)
	default:
		return normalizeStringFieldValue(value)
	}
}

func normalizeStringFieldValue(value any) (string, error) {
	raw, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("expected string")
	}
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("empty value")
	}
	return trimmed, nil
}

func normalizeSelectFieldValue(field ProfileFieldConfig, value any) (any, error) {
	switch typed := value.(type) {
	case string:
		normalized, err := normalizeSelectOptionValue(field, typed)
		if err != nil {
			return nil, err
		}
		return normalized, nil
	case []string:
		return normalizeSelectOptionValues(field, typed)
	case []any:
		items := make([]string, 0, len(typed))
		for _, item := range typed {
			raw, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("expected string options")
			}
			items = append(items, raw)
		}
		return normalizeSelectOptionValues(field, items)
	default:
		return nil, fmt.Errorf("expected string or string array")
	}
}

func normalizeSelectOptionValues(field ProfileFieldConfig, values []string) ([]string, error) {
	if len(values) == 0 {
		return nil, fmt.Errorf("empty value")
	}

	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		item, err := normalizeSelectOptionValue(field, value)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		normalized = append(normalized, item)
	}

	if field.Min > 0 && len(normalized) < field.Min {
		return nil, fmt.Errorf("expected at least %d options", field.Min)
	}
	if field.Max > 0 && len(normalized) > field.Max {
		return nil, fmt.Errorf("expected at most %d options", field.Max)
	}
	return normalized, nil
}

func normalizeSelectOptionValue(field ProfileFieldConfig, value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("empty value")
	}
	if len(field.Options) == 0 {
		return trimmed, nil
	}

	allowCustom := false
	for _, option := range field.Options {
		normalizedOption := strings.TrimSpace(option)
		if normalizedOption == "" {
			continue
		}
		if strings.EqualFold(normalizedOption, trimmed) {
			return normalizedOption, nil
		}
		if strings.HasSuffix(strings.ToLower(normalizedOption), ":input") {
			allowCustom = true
		}
	}
	if allowCustom {
		return trimmed, nil
	}
	return "", fmt.Errorf("invalid option")
}
