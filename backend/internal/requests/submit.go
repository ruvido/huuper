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

	profileByKey := make(map[string]struct{}, len(profile.Fields))
	for _, field := range profile.Fields {
		key := strings.TrimSpace(field.Key)
		if key == "" {
			continue
		}
		profileByKey[key] = struct{}{}
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
			out[key] = value
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
	default:
		return true
	}
}
