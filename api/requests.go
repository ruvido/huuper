package api

import (
	"fmt"
	"maps"
	"net/http"
	"strconv"
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
	Key string `json:"key"`
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

type requestListItem struct {
	ID       string         `json:"id"`
	Email    string         `json:"email"`
	Status   string         `json:"status"`
	Rejected bool           `json:"rejected"`
	GroupID  string         `json:"group"`
	Guardian string         `json:"guardian"`
	Created  string         `json:"created"`
	Updated  string         `json:"updated"`
	Data     map[string]any `json:"data"`
}

const (
	requestActionTransition = "transition"
	requestActionReject     = "reject"
	requestActionPromote    = "promote"

	requestStatusGroupAssigned    = "2-group_assigned"
	requestStatusGuardianAssigned = "3-guardian_assigned"
)

// ListRequestsHandler returns requests visible to the authenticated user.
func ListRequestsHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor := e.Auth
		if actor == nil {
			return apis.NewUnauthorizedError("Unauthorized", nil)
		}

		query := e.Request.URL.Query()
		status := strings.TrimSpace(query.Get("status"))
		groupID := strings.TrimSpace(query.Get("group_id"))
		guardianID := strings.TrimSpace(query.Get("guardian"))
		search := strings.TrimSpace(query.Get("q"))
		sort := strings.TrimSpace(query.Get("sort"))
		if sort == "" {
			sort = "-updated"
		}

		page := 1
		if pageRaw := strings.TrimSpace(query.Get("page")); pageRaw != "" {
			parsed, err := strconv.Atoi(pageRaw)
			if err != nil || parsed < 1 {
				return apis.NewBadRequestError("invalid_page", nil)
			}
			page = parsed
		}

		perPage := 200
		if perPageRaw := strings.TrimSpace(query.Get("per_page")); perPageRaw != "" {
			parsed, err := strconv.Atoi(perPageRaw)
			if err != nil || parsed < 1 {
				return apis.NewBadRequestError("invalid_per_page", nil)
			}
			if parsed > 500 {
				parsed = 500
			}
			perPage = parsed
		}

		includeRejected, err := parseBoolQuery(query.Get("include_rejected"), false)
		if err != nil {
			return apis.NewBadRequestError("invalid_include_rejected", nil)
		}

		params := map[string]any{}
		filters := []string{}

		if status != "" {
			filters = append(filters, "status = {:status}")
			params["status"] = status
		}
		if groupID != "" {
			filters = append(filters, "group = {:group}")
			params["group"] = groupID
		}
		if guardianID != "" {
			filters = append(filters, "guardian = {:guardian}")
			params["guardian"] = guardianID
		}
		if search != "" {
			filters = append(filters, "email ~ {:q}")
			params["q"] = search
		}

		// Keep rejected hidden by default. Admin can explicitly include them.
		if actor.GetBool("admin") {
			if !includeRejected {
				filters = append(filters, "rejected = false")
			}
		} else {
			filters = append(filters, "rejected = false")
		}

		filter := strings.Join(filters, " && ")
		records, err := app.FindRecordsByFilter("requests", filter, sort, perPage, (page-1)*perPage, params)
		if err != nil {
			return apis.NewBadRequestError("failed_requests", err)
		}

		assistantGroups, err := assistantGroupIDsForUser(app, actor)
		if err != nil {
			return apis.NewBadRequestError("failed_groups_lookup", err)
		}

		items := make([]requestListItem, 0, len(records))
		for _, record := range records {
			if !canViewRequest(actor, record, assistantGroups) {
				continue
			}
			items = append(items, mapRequestItem(record))
		}

		return e.JSON(http.StatusOK, map[string]any{
			"items": items,
			"page":  page,
		})
	}
}

// GetRequestHandler returns one request if visible to the authenticated user.
func GetRequestHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor := e.Auth
		if actor == nil {
			return apis.NewUnauthorizedError("Unauthorized", nil)
		}

		id := strings.TrimSpace(e.Request.PathValue("id"))
		if id == "" {
			return apis.NewBadRequestError("invalid_request", nil)
		}

		record, err := app.FindRecordById("requests", id)
		if err != nil || record == nil {
			return apis.NewNotFoundError("request_not_found", err)
		}

		if record.GetBool("rejected") && !actor.GetBool("admin") {
			return apis.NewForbiddenError("forbidden_request", nil)
		}

		assistantGroups, err := assistantGroupIDsForUser(app, actor)
		if err != nil {
			return apis.NewBadRequestError("failed_groups_lookup", err)
		}
		if !canViewRequest(actor, record, assistantGroups) {
			return apis.NewForbiddenError("forbidden_request", nil)
		}

		return e.JSON(http.StatusOK, mapRequestItem(record))
	}
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
		data, email, err := validateAndBuildRequestData(input, signup, profileSchema)
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
		record.Set("status", initialStatus)
		record.Set("data", data)
		record.Set("rejected", false)
		if err := app.Save(record); err != nil {
			return apis.NewBadRequestError("failed_to_create_request", err)
		}

		return e.JSON(http.StatusCreated, map[string]any{
			"id":       record.Id,
			"email":    record.GetString("email"),
			"status":   record.GetString("status"),
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
		case requestActionTransition:
			if err := applyTransitionAction(app, e.Auth, record, data, payload, strings.TrimSpace(payload.TargetStatus)); err != nil {
				return err
			}
		case requestActionReject:
			if err := applyRejectAction(app, e.Auth, record, data, strings.TrimSpace(payload.Reason)); err != nil {
				return err
			}
		case requestActionPromote:
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
			"status":   record.GetString("status"),
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

	current := strings.TrimSpace(record.GetString("status"))
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

		ok, err := hasRoleForRequest(app, actor, record, requiredRole)
		if err != nil {
			return apis.NewBadRequestError("role_resolution_failed", err)
		}
		if !ok {
			return apis.NewForbiddenError("forbidden_transition", nil)
		}
	}

	if target == requestStatusGroupAssigned {
		if err := applyGroupAssignment(app, record, strings.TrimSpace(payload.GroupID)); err != nil {
			return err
		}
	}

	if target == requestStatusGuardianAssigned {
		if err := applyGuardianAssignment(app, record, strings.TrimSpace(payload.GuardianID)); err != nil {
			return err
		}
	}

	record.Set("status", target)
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

func applyGroupAssignment(app *pocketbase.PocketBase, record *core.Record, groupID string) error {
	if groupID == "" {
		return apis.NewBadRequestError("missing_group", nil)
	}
	group, err := app.FindRecordById("groups", groupID)
	if err != nil || group == nil {
		return apis.NewBadRequestError("invalid_group", err)
	}
	record.Set("group", groupID)
	return nil
}

func applyGuardianAssignment(app *pocketbase.PocketBase, record *core.Record, guardianID string) error {
	groupID := strings.TrimSpace(record.GetString("group"))
	if groupID == "" {
		return apis.NewBadRequestError("missing_group_assignment", nil)
	}
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

	current := strings.TrimSpace(record.GetString("status"))
	if current == "" {
		return "", apis.NewBadRequestError("missing_current_status", nil)
	}

	flow, err := loadRequestsFlowSettings(app)
	if err != nil {
		return "", apis.NewBadRequestError("invalid_requests_flow_settings", err)
	}
	finalStatus := strings.TrimSpace(flow.Statuses[len(flow.Statuses)-1])
	if finalStatus == "" {
		return "", apis.NewBadRequestError("invalid_requests_flow_settings", nil)
	}
	if current != finalStatus {
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
		key := strings.TrimSpace(asString(field["key"]))
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

func hasRoleForRequest(app *pocketbase.PocketBase, actor *core.Record, record *core.Record, role string) (bool, error) {
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
	return maps.Clone(data)
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

func assistantGroupIDsForUser(app *pocketbase.PocketBase, user *core.Record) (map[string]struct{}, error) {
	out := map[string]struct{}{}
	if user == nil || user.GetBool("admin") {
		return out, nil
	}

	groups, err := app.FindRecordsByFilter(
		"groups",
		"assistant = {:assistant}",
		"",
		500,
		0,
		map[string]any{"assistant": user.Id},
	)
	if err != nil {
		return nil, err
	}

	for _, group := range groups {
		out[group.Id] = struct{}{}
	}
	return out, nil
}

func canViewRequest(actor *core.Record, request *core.Record, assistantGroups map[string]struct{}) bool {
	if actor == nil || request == nil {
		return false
	}
	if actor.GetBool("admin") {
		return true
	}

	if strings.TrimSpace(request.GetString("guardian")) == actor.Id {
		return true
	}

	groupID := strings.TrimSpace(request.GetString("group"))
	if groupID == "" {
		return false
	}
	_, ok := assistantGroups[groupID]
	return ok
}

func mapRequestItem(record *core.Record) requestListItem {
	return requestListItem{
		ID:       record.Id,
		Email:    record.GetString("email"),
		Status:   record.GetString("status"),
		Rejected: record.GetBool("rejected"),
		GroupID:  record.GetString("group"),
		Guardian: record.GetString("guardian"),
		Created:  record.GetString("created"),
		Updated:  record.GetString("updated"),
		Data:     parseJSONMap(record.Get("data")),
	}
}
