package api

import (
	"fmt"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterSettingsValidationHooks enforces cross-setting integrity for
// profile_schema, signup and onboarding.
func RegisterSettingsValidationHooks(app *pocketbase.PocketBase) {
	app.OnRecordValidate("settings").BindFunc(func(e *core.RecordEvent) error {
		name := strings.TrimSpace(e.Record.GetString("name"))
		if name == "" {
			return e.Next()
		}

		data := unwrapSettingData(e.Record.Get("data"))
		if len(data) == 0 {
			return apis.NewBadRequestError("invalid_settings_data", fmt.Errorf("settings.%s data is empty", name))
		}

		switch name {
		case "profile_schema":
			keys, err := extractProfileSchemaKeys(data)
			if err != nil {
				return apis.NewBadRequestError("invalid_profile_schema", err)
			}

			if err := validateExistingFlowSetting(e.App, "signup", keys); err != nil {
				return apis.NewBadRequestError("invalid_signup_settings", err)
			}
			if err := validateExistingFlowSetting(e.App, "onboarding", keys); err != nil {
				return apis.NewBadRequestError("invalid_onboarding_settings", err)
			}
		case "signup", "onboarding":
			keys, err := loadProfileSchemaKeys(e.App)
			if err != nil {
				// During fresh bootstrap migrations, signup/onboarding can be created
				// before profile_schema exists. Skip cross-validation in that phase.
				return e.Next()
			}
			if err := validateFlowSettingData(name, data, keys); err != nil {
				return apis.NewBadRequestError("invalid_"+name+"_settings", err)
			}
		case "requests_flow":
			if err := validateRequestsFlowSettingData(data); err != nil {
				return apis.NewBadRequestError("invalid_requests_flow_settings", err)
			}
		}

		return e.Next()
	})
}

func validateRequestsFlowSettingData(data map[string]any) error {
	if _, err := parseRequestsFlowConfig(data); err == nil {
		return nil
	}
	return validateLegacyRequestsFlowSettingData(data)
}

func validateLegacyRequestsFlowSettingData(data map[string]any) error {
	rawStatuses, ok := data["statuses"]
	if !ok {
		return fmt.Errorf("settings.requests_flow missing statuses")
	}
	statuses, ok := rawStatuses.([]any)
	if !ok || len(statuses) == 0 {
		return fmt.Errorf("settings.requests_flow statuses must be a non-empty array")
	}
	for i, raw := range statuses {
		status := strings.TrimSpace(anyToString(raw))
		if status == "" {
			return fmt.Errorf("settings.requests_flow statuses[%d] is empty", i)
		}
	}
	return nil
}

func loadProfileSchemaKeys(app core.App) (map[string]struct{}, error) {
	record, err := app.FindFirstRecordByFilter(
		"settings",
		"name = 'profile_schema'",
		map[string]any{},
	)
	if err != nil || record == nil {
		return nil, fmt.Errorf("settings.profile_schema not found")
	}

	return extractProfileSchemaKeys(unwrapSettingData(record.Get("data")))
}

func extractProfileSchemaKeys(data map[string]any) (map[string]struct{}, error) {
	rawFields, ok := data["fields"]
	if !ok {
		return nil, fmt.Errorf("missing fields")
	}

	fields, ok := rawFields.([]any)
	if !ok || len(fields) == 0 {
		return nil, fmt.Errorf("fields must be a non-empty array")
	}

	keys := make(map[string]struct{}, len(fields))
	for _, raw := range fields {
		field, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid field entry")
		}
		key := strings.TrimSpace(anyToString(field["key"]))
		if key == "" {
			return nil, fmt.Errorf("field key is required")
		}
		if _, exists := keys[key]; exists {
			return nil, fmt.Errorf("duplicate field key: %s", key)
		}
		keys[key] = struct{}{}
	}

	return keys, nil
}

func validateExistingFlowSetting(app core.App, name string, keys map[string]struct{}) error {
	record, err := app.FindFirstRecordByFilter(
		"settings",
		"name = {:name}",
		map[string]any{"name": name},
	)
	if err != nil || record == nil {
		return nil
	}

	data := unwrapSettingData(record.Get("data"))
	return validateFlowSettingData(name, data, keys)
}

func validateFlowSettingData(name string, data map[string]any, keys map[string]struct{}) error {
	rawSteps, ok := data["steps"]
	if !ok {
		return fmt.Errorf("settings.%s missing steps", name)
	}

	steps, ok := rawSteps.([]any)
	if !ok || len(steps) == 0 {
		return fmt.Errorf("settings.%s steps must be a non-empty array", name)
	}

	for _, raw := range steps {
		step, ok := raw.(map[string]any)
		if !ok {
			return fmt.Errorf("settings.%s has invalid step entry", name)
		}

		field := strings.TrimSpace(anyToString(step["field"]))
		if field == "" {
			return fmt.Errorf("settings.%s step field is required", name)
		}
		if _, exists := keys[field]; !exists {
			return fmt.Errorf("settings.%s step field not found in profile_schema: %s", name, field)
		}
	}

	return nil
}
