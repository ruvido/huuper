package requests

import (
	"fmt"
	"maps"
	"strings"

	backendinternal "members/backend/internal"
	backendsettings "members/backend/internal/settings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func LoadSignupSettings(app *pocketbase.PocketBase) (SignupSettingsConfig, error) {
	var cfg SignupSettingsConfig
	raw, err := FindSettingData(app, "signup")
	if err != nil {
		return SignupSettingsConfig{}, err
	}
	if rawSteps, ok := raw["steps"].([]any); ok {
		cfg.Steps = make([]SignupFieldConfig, 0, len(rawSteps))
		for _, item := range rawSteps {
			step, ok := item.(map[string]any)
			if !ok {
				continue
			}
			field := strings.TrimSpace(backendinternal.AnyToString(step["field"]))
			if field == "" {
				continue
			}
			cfg.Steps = append(cfg.Steps, SignupFieldConfig{Field: field})
		}
	}
	return cfg, nil
}

func LoadProfileSchemaSettings(app *pocketbase.PocketBase) (ProfileSchemaConfig, error) {
	raw, err := FindSettingData(app, "profile_schema")
	if err != nil {
		return ProfileSchemaConfig{}, err
	}
	rawFields, ok := raw["fields"].([]any)
	if !ok {
		return ProfileSchemaConfig{}, fmt.Errorf("profile_schema fields missing")
	}

	cfg := ProfileSchemaConfig{
		Fields: make([]ProfileFieldConfig, 0, len(rawFields)),
	}
	for _, item := range rawFields {
		field, ok := item.(map[string]any)
		if !ok {
			continue
		}
		key := strings.TrimSpace(backendinternal.AnyToString(field["key"]))
		if key == "" {
			continue
		}
		cfg.Fields = append(cfg.Fields, ProfileFieldConfig{Key: key})
	}
	return cfg, nil
}

func LoadFlowSettings(app *pocketbase.PocketBase) (FlowConfig, error) {
	raw, err := FindSettingData(app, "requests_flow")
	if err != nil {
		return FlowConfig{}, err
	}
	return ParseFlowConfig(raw)
}

func LoadFlowForRequest(app *pocketbase.PocketBase, data map[string]any) (FlowConfig, error) {
	if snapshot, ok := RequestFlowSnapshotFromData(data); ok {
		return snapshot, nil
	}
	return LoadFlowSettings(app)
}

func RequestFlowSnapshotFromData(data map[string]any) (FlowConfig, bool) {
	if data == nil {
		return FlowConfig{}, false
	}

	raw, ok := data[RequestFlowDataKey]
	if !ok || raw == nil {
		return FlowConfig{}, false
	}

	snapshotData, ok := raw.(map[string]any)
	if !ok || snapshotData == nil {
		return FlowConfig{}, false
	}

	snapshot, err := ParseFlowConfig(snapshotData)
	if err != nil {
		return FlowConfig{}, false
	}

	return snapshot, true
}

func SetRequestFlowSnapshot(data map[string]any, flow FlowConfig) map[string]any {
	if data == nil {
		data = map[string]any{}
	}

	steps := make([]any, 0, len(flow.Steps))
	for _, step := range flow.Steps {
		steps = append(steps, map[string]any{
			"role":             step.Role,
			"action":           step.Action,
			"label":            step.Label,
			"notes":            step.Notes,
			"filter":           step.Filter,
			"email_to":         step.EmailTo,
			"telegram_message": step.TelegramMessage,
		})
	}

	data[FlowVersionDataKey] = flow.Version
	data[RequestFlowDataKey] = map[string]any{
		"version": flow.Version,
		"steps":   steps,
	}
	return data
}

func NormalizeSubmitInput(raw map[string]any) map[string]any {
	if raw == nil {
		return map[string]any{}
	}
	if nested, ok := raw["data"].(map[string]any); ok && nested != nil {
		return nested
	}
	return raw
}

func EnsureSubmitEmailAvailable(app *pocketbase.PocketBase, email string) error {
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

func FindSettingData(app core.App, name string) (map[string]any, error) {
	raw, err := backendsettings.FindSettingData(app, name)
	if err != nil {
		return nil, fmt.Errorf("%s settings not found", name)
	}
	if raw == nil {
		return nil, fmt.Errorf("%s settings not found", name)
	}
	return raw, nil
}

func BuildUserData(data map[string]any) map[string]any {
	if data == nil {
		return map[string]any{}
	}
	out := maps.Clone(data)
	for key := range out {
		if strings.HasPrefix(strings.TrimSpace(key), "__") {
			delete(out, key)
		}
	}
	return out
}

func FlowStepAt(flow FlowConfig, stepIndex int) (FlowStep, bool) {
	if stepIndex < 0 || stepIndex >= len(flow.Steps) {
		return FlowStep{}, false
	}
	return flow.Steps[stepIndex], true
}

func ParseBoolQuery(raw string, fallback bool) (bool, error) {
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
