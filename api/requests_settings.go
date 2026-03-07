package api

import (
	"fmt"
	"maps"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
)

func loadSignupSettings(app *pocketbase.PocketBase) (signupSettingsConfig, error) {
	var cfg signupSettingsConfig
	raw, err := findSettingData(app, "signup")
	if err != nil {
		return signupSettingsConfig{}, err
	}
	if rawSteps, ok := raw["steps"].([]any); ok {
		cfg.Steps = make([]signupFieldConfig, 0, len(rawSteps))
		for _, item := range rawSteps {
			step, ok := item.(map[string]any)
			if !ok {
				continue
			}
			field := strings.TrimSpace(anyToString(step["field"]))
			if field == "" {
				continue
			}
			cfg.Steps = append(cfg.Steps, signupFieldConfig{Field: field})
		}
	}
	return cfg, nil
}

func loadProfileSchemaSettings(app *pocketbase.PocketBase) (profileSchemaConfig, error) {
	raw, err := findSettingData(app, "profile_schema")
	if err != nil {
		return profileSchemaConfig{}, err
	}
	rawFields, ok := raw["fields"].([]any)
	if !ok {
		return profileSchemaConfig{}, fmt.Errorf("profile_schema fields missing")
	}

	cfg := profileSchemaConfig{
		Fields: make([]profileFieldConfig, 0, len(rawFields)),
	}
	for _, item := range rawFields {
		field, ok := item.(map[string]any)
		if !ok {
			continue
		}
		key := strings.TrimSpace(anyToString(field["key"]))
		if key == "" {
			continue
		}
		cfg.Fields = append(cfg.Fields, profileFieldConfig{
			Key: key,
		})
	}
	return cfg, nil
}

func loadRequestsFlowSettings(app *pocketbase.PocketBase) (requestsFlowConfig, error) {
	raw, err := findSettingData(app, "requests_flow")
	if err != nil {
		return requestsFlowConfig{}, err
	}
	return parseRequestsFlowConfig(raw)
}

func normalizeSubmitInput(raw map[string]any) map[string]any {
	if raw == nil {
		return map[string]any{}
	}
	if nested, ok := raw["data"].(map[string]any); ok && nested != nil {
		return nested
	}
	return raw
}

func validateAndBuildRequestData(input map[string]any, signup signupSettingsConfig, profile profileSchemaConfig) (map[string]any, string, error) {
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
		_, ok := profileByKey[key]
		if !ok {
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
	normalizedEmail, err := normalizeEmail(email)
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

func ensureSubmitEmailAvailable(app *pocketbase.PocketBase, email string) error {
	existingUser, err := app.FindFirstRecordByFilter(
		"users",
		"email = {:email}",
		map[string]any{"email": email},
	)
	if err == nil && existingUser != nil {
		return apis.NewBadRequestError("email_exists_user", nil)
	}

	existingRequest, err := app.FindFirstRecordByFilter(
		"requests",
		"email = {:email}",
		map[string]any{"email": email},
	)
	if err == nil && existingRequest != nil {
		return apis.NewBadRequestError("email_exists_request", nil)
	}

	return nil
}

func findSettingData(app *pocketbase.PocketBase, name string) (map[string]any, error) {
	record, err := app.FindFirstRecordByFilter(
		"settings",
		"name = {:name}",
		map[string]any{"name": name},
	)
	if err != nil || record == nil {
		return nil, fmt.Errorf("%s settings not found", name)
	}

	return unwrapSettingData(record.Get("data")), nil
}

func buildUserDataFromRequest(data map[string]any) map[string]any {
	if data == nil {
		return map[string]any{}
	}
	out := maps.Clone(data)
	delete(out, requestFlowVersionDataKey)
	delete(out, requestStepIndexDataKey)
	return out
}

func flowStepAt(flow requestsFlowConfig, stepIndex int) (requestsFlowStep, bool) {
	if stepIndex < 0 || stepIndex >= len(flow.Steps) {
		return requestsFlowStep{}, false
	}
	return flow.Steps[stepIndex], true
}

func parseBoolQuery(raw string, fallback bool) (bool, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return fallback, nil
	}
	switch strings.ToLower(value) {
	case "1", "true", "yes":
		return true, nil
	case "0", "false", "no":
		return false, nil
	default:
		return false, fmt.Errorf("invalid bool")
	}
}
