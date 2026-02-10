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
	Field string `json:"field"`
}

type signupSettingsConfig struct {
	Steps []signupFieldConfig `json:"steps"`
}

type profileFieldConfig struct {
	Key      string `json:"key"`
	Required bool   `json:"required"`
}

type profileSchemaConfig struct {
	Fields []profileFieldConfig `json:"fields"`
}

type requestsFlowConfig struct {
	Statuses    []string          `json:"statuses"`
	SetStatusBy map[string]string `json:"set_status_by"`
}

type requestActionPayload struct {
	Action       string `json:"action"`
	TargetStatus string `json:"target_status"`
	Reason       string `json:"reason"`
	GroupID      string `json:"group"`
	GuardianID   string `json:"guardian"`
}

// SubmitRequestHandler creates a request from public signup config.
func SubmitRequestHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		signup, err := loadSignupSettings(app)
		if err != nil {
			return apis.NewBadRequestError("invalid_signup_settings", err)
		}
		profileSchema, err := loadProfileSchemaSettings(app)
		if err != nil {
			return apis.NewBadRequestError("invalid_profile_schema", err)
		}
		flow, err := loadRequestsFlowSettings(app)
		if err != nil {
			return apis.NewBadRequestError("invalid_requests_flow_settings", err)
		}
		initialStatus := strings.TrimSpace(flow.Statuses[0])
		if initialStatus == "" {
			return apis.NewBadRequestError("invalid_requests_flow_settings", nil)
		}

		var raw map[string]any
		if err := e.BindBody(&raw); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}

		input := normalizeSubmitInput(raw)
		data, email, err := validateAndBuildRequestData(input, signup, profileSchema, initialStatus)
		if err != nil {
			return apis.NewBadRequestError("invalid_request_data", err)
		}
		if err := ensureSubmitEmailAvailable(app, email); err != nil {
			return err
		}

		requests, err := app.FindCollectionByNameOrId("requests")
		if err != nil {
			return apis.NewNotFoundError("requests_collection_not_found", err)
		}

		record := core.NewRecord(requests)
		record.Set("email", email)
		record.Set("data", data)
		record.Set("rejected", false)
		if err := app.Save(record); err != nil {
			return apis.NewBadRequestError("failed_to_create_request", err)
		}

		return e.JSON(http.StatusCreated, map[string]any{
			"id":       record.Id,
			"email":    record.GetString("email"),
			"rejected": false,
			"data":     data,
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

		deleteRequest := false
		promotedUserID := ""
		switch action {
		case "transition":
			if err := applyTransitionAction(app, e.Auth, record, data, payload, strings.TrimSpace(payload.TargetStatus)); err != nil {
				return err
			}
		case "reject":
			if err := applyRejectAction(app, e.Auth, record, data, strings.TrimSpace(payload.Reason)); err != nil {
				return err
			}
		case "promote":
			userID, err := applyPromoteAction(app, e.Auth, record, data)
			if err != nil {
				return err
			}
			deleteRequest = true
			promotedUserID = userID
		default:
			return apis.NewBadRequestError("unsupported_action", nil)
		}

		if deleteRequest {
			if err := app.Delete(record); err != nil {
				return apis.NewBadRequestError("failed_to_delete_request", err)
			}
			return e.JSON(http.StatusOK, map[string]any{
				"id":       id,
				"promoted": true,
				"user_id":  promotedUserID,
			})
		}

		if err := app.Save(record); err != nil {
			return apis.NewBadRequestError("failed_to_update_request", err)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"id":       record.Id,
			"rejected": record.GetBool("rejected"),
			"data":     record.Get("data"),
		})
	}
}

func applyTransitionAction(app *pocketbase.PocketBase, actor *core.Record, record *core.Record, data map[string]any, payload requestActionPayload, target string) error {
	if target == "" {
		return apis.NewBadRequestError("missing_target_status", nil)
	}

	if record.GetBool("rejected") {
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

	if !statusInFlow(flow.Statuses, target) {
		return apis.NewBadRequestError("invalid_transition", nil)
	}

	if !actor.GetBool("admin") {
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
	}

	if target == "2-group_assigned" {
		groupID := strings.TrimSpace(payload.GroupID)
		if groupID == "" {
			return apis.NewBadRequestError("missing_group", nil)
		}
		group, err := app.FindRecordById("groups", groupID)
		if err != nil || group == nil {
			return apis.NewBadRequestError("invalid_group", err)
		}
		record.Set("group", groupID)
	}

	if target == "3-guardian_assigned" {
		groupID := strings.TrimSpace(record.GetString("group"))
		if groupID == "" {
			return apis.NewBadRequestError("missing_group_assignment", nil)
		}

		guardianID := strings.TrimSpace(payload.GuardianID)
		if guardianID == "" {
			return apis.NewBadRequestError("missing_guardian", nil)
		}

		guardian, err := app.FindRecordById("users", guardianID)
		if err != nil || guardian == nil {
			return apis.NewBadRequestError("invalid_guardian", err)
		}

		guardianMemberships, err := app.FindRecordsByFilter(
			"user_groups",
			"user = {:user} && group = {:group}",
			"",
			1,
			0,
			map[string]any{
				"user":  guardianID,
				"group": groupID,
			},
		)
		if err != nil {
			return apis.NewBadRequestError("guardian_membership_check_failed", err)
		}
		if len(guardianMemberships) == 0 {
			return apis.NewBadRequestError("guardian_not_in_group", nil)
		}

		record.Set("guardian", guardianID)
	}

	data["status"] = target
	record.Set("data", data)
	return nil
}

func applyRejectAction(app *pocketbase.PocketBase, actor *core.Record, record *core.Record, data map[string]any, reason string) error {
	if reason == "" {
		return apis.NewBadRequestError("missing_reject_reason", nil)
	}

	if !actor.GetBool("admin") {
		return apis.NewForbiddenError("forbidden_reject", nil)
	}

	data["reject_reason"] = reason
	data["rejected_at"] = time.Now().UTC().Format(time.RFC3339)
	data["rejected_by"] = adminDisplayName(actor)
	record.Set("rejected", true)
	record.Set("data", data)
	return nil
}

func adminDisplayName(actor *core.Record) string {
	if actor == nil {
		return ""
	}

	data := parseJSONMap(actor.Get("data"))
	if fullName, ok := data["full_name"].(string); ok && strings.TrimSpace(fullName) != "" {
		return strings.TrimSpace(fullName)
	}
	if name, ok := data["name"].(string); ok && strings.TrimSpace(name) != "" {
		return strings.TrimSpace(name)
	}

	email := strings.TrimSpace(actor.GetString("email"))
	if email != "" {
		return email
	}

	return actor.Id
}

func applyPromoteAction(app *pocketbase.PocketBase, actor *core.Record, record *core.Record, data map[string]any) (string, error) {
	if !actor.GetBool("admin") {
		return "", apis.NewForbiddenError("forbidden_promote", nil)
	}
	if record.GetBool("rejected") {
		return "", apis.NewBadRequestError("request_rejected", nil)
	}

	current := asString(data["status"])
	if current != "6-admin_approved" {
		return "", apis.NewBadRequestError("invalid_promote_status", nil)
	}

	email, err := normalizeEmail(record.GetString("email"))
	if err != nil {
		return "", apis.NewBadRequestError("invalid_email", nil)
	}

	existing, err := app.FindFirstRecordByFilter(
		"users",
		"email = {:email}",
		map[string]any{"email": email},
	)
	if err == nil && existing != nil {
		return "", apis.NewBadRequestError("user_already_exists", nil)
	}

	users, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		return "", apis.NewNotFoundError("users_collection_not_found", err)
	}

	user := core.NewRecord(users)
	tempPassword := randomToken()
	if tempPassword == "" {
		return "", apis.NewBadRequestError("failed_to_generate_password", nil)
	}
	user.Set("email", email)
	user.Set("password", tempPassword)
	user.Set("passwordConfirm", tempPassword)
	user.Set("status", "active")

	userData := buildUserDataFromRequest(data)
	if len(userData) > 0 {
		user.Set("data", userData)
	}

	if err := app.Save(user); err != nil {
		return "", apis.NewBadRequestError("failed_to_create_user", err)
	}

	return user.Id, nil
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
	raw := unwrapSettingData(record.Get("data"))
	if rawSteps, ok := raw["steps"].([]any); ok {
		cfg.Steps = make([]signupFieldConfig, 0, len(rawSteps))
		for _, item := range rawSteps {
			step, ok := item.(map[string]any)
			if !ok {
				continue
			}
			field := strings.TrimSpace(asString(step["field"]))
			if field == "" {
				continue
			}
			cfg.Steps = append(cfg.Steps, signupFieldConfig{Field: field})
		}
	}
	return cfg, nil
}

func loadProfileSchemaSettings(app *pocketbase.PocketBase) (profileSchemaConfig, error) {
	record, err := app.FindFirstRecordByFilter(
		"settings",
		"name = 'profile_schema'",
		map[string]any{},
	)
	if err != nil || record == nil {
		return profileSchemaConfig{}, fmt.Errorf("profile_schema settings not found")
	}

	raw := unwrapSettingData(record.Get("data"))
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
		key := strings.TrimSpace(asString(field["key"]))
		if key == "" {
			continue
		}
		required := true
		if value, ok := field["required"].(bool); ok {
			required = value
		}
		cfg.Fields = append(cfg.Fields, profileFieldConfig{
			Key:      key,
			Required: required,
		})
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

	raw := unwrapSettingData(record.Get("data"))
	var cfg requestsFlowConfig
	if rawStatuses, ok := raw["statuses"].([]any); ok {
		cfg.Statuses = make([]string, 0, len(rawStatuses))
		for _, value := range rawStatuses {
			status := strings.TrimSpace(asString(value))
			if status == "" {
				continue
			}
			cfg.Statuses = append(cfg.Statuses, status)
		}
	}
	if rawSetStatusBy, ok := raw["set_status_by"].(map[string]any); ok {
		cfg.SetStatusBy = make(map[string]string, len(rawSetStatusBy))
		for status, roleValue := range rawSetStatusBy {
			role := strings.TrimSpace(asString(roleValue))
			if role == "" {
				continue
			}
			cfg.SetStatusBy[strings.TrimSpace(status)] = role
		}
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

func validateAndBuildRequestData(input map[string]any, signup signupSettingsConfig, profile profileSchemaConfig, initialStatus string) (map[string]any, string, error) {
	if len(signup.Steps) == 0 {
		return nil, "", fmt.Errorf("signup steps not configured")
	}

	profileByKey := make(map[string]profileFieldConfig, len(profile.Fields))
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

	allowed := make(map[string]profileFieldConfig, len(signup.Steps))
	for _, step := range signup.Steps {
		key := strings.TrimSpace(step.Field)
		if key == "" {
			continue
		}
		fieldCfg, ok := profileByKey[key]
		if !ok {
			return nil, "", fmt.Errorf("signup step field not in profile_schema: %s", key)
		}
		allowed[key] = fieldCfg
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

	out["status"] = initialStatus
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
		groupID := strings.TrimSpace(record.GetString("group"))
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

func buildUserDataFromRequest(data map[string]any) map[string]any {
	if data == nil {
		return map[string]any{}
	}

	out := map[string]any{}
	for key, value := range data {
		out[key] = value
	}

	return out
}

func statusInFlow(statuses []string, target string) bool {
	for _, status := range statuses {
		if status == target {
			return true
		}
	}
	return false
}

func asString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return ""
	}
}
