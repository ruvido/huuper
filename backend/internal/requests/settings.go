package requests

import (
	"fmt"
	"maps"
	"strconv"
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

		options := []string{}
		if rawOptions, ok := field["options"].([]any); ok {
			options = make([]string, 0, len(rawOptions))
			for _, rawOption := range rawOptions {
				option := strings.TrimSpace(backendinternal.AnyToString(rawOption))
				if option == "" {
					continue
				}
				options = append(options, option)
			}
		}

		cfg.Fields = append(cfg.Fields, ProfileFieldConfig{
			Key:      key,
			Type:     strings.ToLower(strings.TrimSpace(backendinternal.AnyToString(field["type"]))),
			Required: parseRequiredFlag(field["required"]),
			Unique:   parseOptionalBool(field["unique"]),
			Options:  options,
			Min:      parseOptionalInt(field["min"]),
			Max:      parseOptionalInt(field["max"]),
		})
	}
	return cfg, nil
}

func LoadOnboardingSettings(app *pocketbase.PocketBase) (OnboardingSettingsConfig, error) {
	raw, err := FindSettingData(app, "onboarding")
	if err != nil {
		return OnboardingSettingsConfig{}, err
	}

	cfg := OnboardingSettingsConfig{}
	if rawStartPage, ok := raw["start_page"].(map[string]any); ok && rawStartPage != nil {
		cfg.StartPage = &OnboardingPageConfig{
			Title:  strings.TrimSpace(backendinternal.AnyToString(rawStartPage["title"])),
			Text:   strings.TrimSpace(backendinternal.AnyToString(rawStartPage["text"])),
			Button: strings.TrimSpace(backendinternal.AnyToString(rawStartPage["button"])),
		}
	}

	if rawConfirmation, ok := raw["confirmation"].(map[string]any); ok && rawConfirmation != nil {
		cfg.Confirmation = &OnboardingPageConfig{
			Title:  strings.TrimSpace(backendinternal.AnyToString(rawConfirmation["title"])),
			Text:   strings.TrimSpace(backendinternal.AnyToString(rawConfirmation["text"])),
			Button: strings.TrimSpace(backendinternal.AnyToString(rawConfirmation["button"])),
		}
	}

	if rawSteps, ok := raw["steps"].([]any); ok {
		cfg.Steps = make([]OnboardingStepConfig, 0, len(rawSteps))
		for _, item := range rawSteps {
			step, ok := item.(map[string]any)
			if !ok {
				continue
			}
			field := strings.TrimSpace(backendinternal.AnyToString(step["field"]))
			if field == "" {
				continue
			}
			cfg.Steps = append(cfg.Steps, OnboardingStepConfig{
				Field: field,
				Title: strings.TrimSpace(backendinternal.AnyToString(step["title"])),
				Label: strings.TrimSpace(backendinternal.AnyToString(step["label"])),
				Unique: parseOptionalBool(step["unique"]),
			})
		}
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

	if _, ok := RequestFlowSnapshotFromData(data); ok {
		return data
	}

	steps := make([]any, 0, len(flow.Steps))
	for _, step := range flow.Steps {
		steps = append(steps, map[string]any{
			"role":             step.Role,
			"action":           step.Action,
			"label":            step.Label,
			"cta":              step.Cta,
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

func SetRequestFlowSnapshotData(data map[string]any, flowData map[string]any) map[string]any {
	if data == nil {
		data = map[string]any{}
	}

	cloned := backendinternal.DeepCopyJSONMap(flowData)
	if version := ParseVersion(cloned["version"]); version > 0 {
		data[FlowVersionDataKey] = version
	}
	data[RequestFlowDataKey] = cloned
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

func MergeUserData(base map[string]any, updates map[string]any) map[string]any {
	out := BuildUserData(base)
	for key, value := range BuildUserData(updates) {
		out[key] = value
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

func parseOptionalInt(raw any) int {
	value := strings.TrimSpace(backendinternal.AnyToString(raw))
	if value == "" {
		return 0
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
}

func parseRequiredFlag(raw any) bool {
	switch value := raw.(type) {
	case bool:
		return value
	default:
		parsed, err := ParseBoolQuery(backendinternal.AnyToString(raw), false)
		if err != nil {
			return false
		}
		return parsed
	}
}

func parseOptionalBool(raw any) *bool {
	switch value := raw.(type) {
	case nil:
		return nil
	case bool:
		v := value
		return &v
	default:
		parsed, err := ParseBoolQuery(backendinternal.AnyToString(raw), false)
		if err != nil {
			return nil
		}
		v := parsed
		return &v
	}
}
