package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

type signupFieldConfig struct {
	Key      string `json:"key"`
	Required bool   `json:"required"`
}

type signupSettingsConfig struct {
	Fields          []signupFieldConfig `json:"fields"`
	RequestDefaults map[string]any      `json:"request_defaults"`
}

type requestsFlowConfig struct {
	Statuses    []string          `json:"statuses"`
	SetStatusBy map[string]string `json:"set_status_by"`
}

type requestActionPayload struct {
	Action       string `json:"action"`
	TargetStatus string `json:"target_status"`
	Reason       string `json:"reason"`
}

// SubmitRequestHandler creates a request from public signup config.
func SubmitRequestHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		signup, err := loadSignupSettings(app)
		if err != nil {
			return apis.NewBadRequestError("invalid_signup_settings", err)
		}

		var raw map[string]any
		if err := e.BindBody(&raw); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}

		input := normalizeSubmitInput(raw)
		data, err := validateAndBuildRequestData(input, signup)
		if err != nil {
			return apis.NewBadRequestError("invalid_request_data", err)
		}

		requests, err := app.FindCollectionByNameOrId("requests")
		if err != nil {
			return apis.NewNotFoundError("requests_collection_not_found", err)
		}

		record := core.NewRecord(requests)
		record.Set("data", data)
		if err := app.Save(record); err != nil {
			return apis.NewBadRequestError("failed_to_create_request", err)
		}

		return e.JSON(http.StatusCreated, map[string]any{
			"id":   record.Id,
			"data": data,
		})
	}
}

// RequestActionHandler executes request actions (transition/reject).
func RequestActionHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if e.Auth == nil {
			return apis.NewUnauthorizedError("Unauthorized", nil)
		}

		id := e.Request.PathValue("id")
		if id == "" {
			return apis.NewBadRequestError("invalid_request", nil)
		}

		record, err := app.FindRecordById("requests", id)
		if err != nil || record == nil {
			return apis.NewNotFoundError("request_not_found", err)
		}

		var payload requestActionPayload
		if err := e.BindBody(&payload); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}

		action := strings.TrimSpace(payload.Action)
		if action == "" {
			return apis.NewBadRequestError("missing_action", nil)
		}

		data := parseJSONMap(record.Get("data"))

		switch action {
		case "transition":
			if err := applyTransitionAction(app, e.Auth, record, data, strings.TrimSpace(payload.TargetStatus)); err != nil {
				return err
			}
		case "reject":
			if err := applyRejectAction(app, e.Auth, record, data, strings.TrimSpace(payload.Reason)); err != nil {
				return err
			}
		default:
			return apis.NewBadRequestError("unsupported_action", nil)
		}

		if err := app.Save(record); err != nil {
			return apis.NewBadRequestError("failed_to_update_request", err)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"id":   record.Id,
			"data": record.Get("data"),
		})
	}
}

func applyTransitionAction(app *pocketbase.PocketBase, actor *core.Record, record *core.Record, data map[string]any, target string) error {
	if target == "" {
		return apis.NewBadRequestError("missing_target_status", nil)
	}

	if toBool(data["rejected"]) {
		return apis.NewBadRequestError("request_rejected", nil)
	}

	flow, err := loadRequestsFlowSettings(app)
	if err != nil {
		return apis.NewBadRequestError("invalid_requests_flow_settings", err)
	}

	current := asString(data["status"])
	if current == "" {
		return apis.NewBadRequestError("missing_current_status", nil)
	}

	next, err := nextStatus(flow.Statuses, current)
	if err != nil {
		return apis.NewBadRequestError("invalid_current_status", err)
	}
	if target != next {
		return apis.NewBadRequestError("invalid_transition", nil)
	}

	requiredRole := strings.TrimSpace(flow.SetStatusBy[target])
	if requiredRole == "" {
		return apis.NewBadRequestError("missing_role_for_target_status", nil)
	}

	ok, err := hasRoleForRequest(app, actor, record, data, requiredRole)
	if err != nil {
		return apis.NewBadRequestError("role_resolution_failed", err)
	}
	if !ok {
		return apis.NewForbiddenError("forbidden_transition", nil)
	}

	data["status"] = target
	record.Set("data", data)
	return nil
}

func applyRejectAction(app *pocketbase.PocketBase, actor *core.Record, record *core.Record, data map[string]any, reason string) error {
	if reason == "" {
		return apis.NewBadRequestError("missing_reject_reason", nil)
	}

	isAdmin := actor.GetBool("admin")
	isAssistant, err := hasRoleForRequest(app, actor, record, data, "assistant")
	if err != nil {
		return apis.NewBadRequestError("role_resolution_failed", err)
	}
	if !isAdmin && !isAssistant {
		return apis.NewForbiddenError("forbidden_reject", nil)
	}

	data["rejected"] = true
	data["reject_reason"] = reason
	data["rejected_at"] = time.Now().UTC().Format(time.RFC3339)
	data["rejected_by"] = actor.Id
	record.Set("data", data)
	return nil
}

func loadSignupSettings(app *pocketbase.PocketBase) (signupSettingsConfig, error) {
	record, err := app.FindFirstRecordByFilter(
		"settings",
		"name = 'signup'",
		map[string]any{},
	)
	if err != nil || record == nil {
		return signupSettingsConfig{}, fmt.Errorf("signup settings not found")
	}

	var cfg signupSettingsConfig
	if err := record.UnmarshalJSONField("data", &cfg); err != nil {
		return signupSettingsConfig{}, err
	}
	return cfg, nil
}

func loadRequestsFlowSettings(app *pocketbase.PocketBase) (requestsFlowConfig, error) {
	record, err := app.FindFirstRecordByFilter(
		"settings",
		"name = 'requests_flow'",
		map[string]any{},
	)
	if err != nil || record == nil {
		return requestsFlowConfig{}, fmt.Errorf("requests_flow settings not found")
	}

	var cfg requestsFlowConfig
	if err := record.UnmarshalJSONField("data", &cfg); err != nil {
		return requestsFlowConfig{}, err
	}
	if len(cfg.Statuses) == 0 {
		return requestsFlowConfig{}, fmt.Errorf("empty statuses")
	}
	return cfg, nil
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

func validateAndBuildRequestData(input map[string]any, signup signupSettingsConfig) (map[string]any, error) {
	allowed := make(map[string]signupFieldConfig, len(signup.Fields))
	for _, field := range signup.Fields {
		key := strings.TrimSpace(field.Key)
		if key == "" {
			continue
		}
		allowed[key] = field
	}
	if len(allowed) == 0 {
		return nil, fmt.Errorf("signup fields not configured")
	}

	for key := range input {
		if _, ok := allowed[key]; !ok {
			return nil, fmt.Errorf("unknown field: %s", key)
		}
	}

	out := map[string]any{}
	for key := range allowed {
		if value, ok := input[key]; ok {
			out[key] = value
		}
	}

	for key, field := range allowed {
		if !field.Required {
			continue
		}
		if !hasNonEmptyValue(out[key]) {
			return nil, fmt.Errorf("missing required field: %s", key)
		}
	}

	status := "1-submitted"
	rejected := false
	if signup.RequestDefaults != nil {
		if value, ok := signup.RequestDefaults["status"].(string); ok && strings.TrimSpace(value) != "" {
			status = strings.TrimSpace(value)
		}
		if value, ok := signup.RequestDefaults["rejected"].(bool); ok {
			rejected = value
		}
	}

	out["status"] = status
	out["rejected"] = rejected
	return out, nil
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

func nextStatus(statuses []string, current string) (string, error) {
	for i, status := range statuses {
		if status != current {
			continue
		}
		if i+1 >= len(statuses) {
			return "", fmt.Errorf("last status reached")
		}
		return statuses[i+1], nil
	}
	return "", fmt.Errorf("current status not found in flow")
}

func hasRoleForRequest(app *pocketbase.PocketBase, actor *core.Record, record *core.Record, data map[string]any, role string) (bool, error) {
	switch role {
	case "admin":
		return actor.GetBool("admin"), nil
	case "guardian":
		return strings.TrimSpace(record.GetString("guardian")) == actor.Id, nil
	case "assistant":
		groupID := strings.TrimSpace(asString(data["group_id"]))
		if groupID == "" {
			return false, nil
		}
		group, err := app.FindRecordById("groups", groupID)
		if err != nil || group == nil {
			return false, err
		}
		return strings.TrimSpace(group.GetString("assistant")) == actor.Id, nil
	default:
		return false, nil
	}
}

func asString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return ""
	}
}

func toBool(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	default:
		return false
	}
}
