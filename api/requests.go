package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
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

		params := map[string]any{}
		filters := []string{}

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

		// Rejected requests are never listed.
		filters = append(filters, "rejected = false")

		filter := strings.Join(filters, " && ")
		records, err := app.FindRecordsByFilter("requests", filter, sort, perPage, (page-1)*perPage, params)
		if err != nil {
			return apis.NewBadRequestError("failed_requests", err)
		}

		assistantGroups, err := assistantGroupIDsForUser(app, actor)
		if err != nil {
			return apis.NewBadRequestError("failed_groups_lookup", err)
		}
		flow, err := loadRequestsFlowSettings(app)
		if err != nil {
			return apis.NewBadRequestError("invalid_requests_flow_settings", err)
		}

		items := make([]requestListItem, 0, len(records))
		for _, record := range records {
			if !canViewRequest(actor, record, assistantGroups) {
				continue
			}
			item, err := mapRequestItemWithWorkflow(app, actor, record, flow)
			if err != nil {
				return apis.NewBadRequestError("failed_requests_workflow", err)
			}
			if status != "" && !strings.EqualFold(strings.TrimSpace(item.Status), strings.TrimSpace(status)) {
				continue
			}
			items = append(items, item)
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

		item := mapRequestItem(record)
		flow, err := loadRequestsFlowSettings(app)
		if err != nil {
			return apis.NewBadRequestError("invalid_requests_flow_settings", err)
		}
		flowVersion, _ := requestProgressFromData(item.Data)
		stepIndex := effectiveRequestStepIndex(record, item.Data, flow)
		computedStatus := requestStatusForItem(item.Rejected, stepIndex, flow.Steps)

		nextStep, hasNext := flowStepAt(flow, stepIndex)
		canAdvance := false
		requiredField := ""
		if hasNext && !item.Rejected {
			requiredField = requiredFieldForAction(nextStep.Action)
			canAdvance, err = hasRoleForRequest(app, actor, record, nextStep.Role)
			if err != nil {
				return apis.NewBadRequestError("role_resolution_failed", err)
			}
		}

		return e.JSON(http.StatusOK, map[string]any{
			"id":           item.ID,
			"email":        item.Email,
			"status":       computedStatus,
			"rejected":     item.Rejected,
			"group":        item.GroupID,
			"guardian":     item.Guardian,
			"flow_version": flowVersion,
			"step_index":   stepIndex,
			"created":      item.Created,
			"updated":      item.Updated,
			"data":         item.Data,
			"workflow": map[string]any{
				"total_steps":       len(flow.Steps),
				"has_next_step":     hasNext,
				"next_role":         nextStep.Role,
				"next_action":       nextStep.Action,
				"next_action_label": nextStep.Label,
				"next_action_notes": nextStep.Notes,
				"required_field":    requiredField,
				"can_advance":       canAdvance,
				"current_version":   flow.Version,
			},
		})
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
		data[requestFlowVersionDataKey] = flow.Version
		data[requestStepIndexDataKey] = 0
		record.Set("data", data)
		record.Set("rejected", false)
		if err := app.Save(record); err != nil {
			return apis.NewBadRequestError("failed_to_create_request", err)
		}

		return e.JSON(http.StatusCreated, map[string]any{
			"id":       record.Id,
			"email":    record.GetString("email"),
			"status":   requestStatusSubmitted,
			"step":     0,
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
		case requestActionAdvance, "transition":
			if err := applyAdvanceAction(app, e.Auth, record, data, payload); err != nil {
				return err
			}
		case requestActionSetGuardian:
			if err := applySetGuardianAction(app, e.Auth, record, data, strings.TrimSpace(payload.GuardianID)); err != nil {
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

		flow, err := loadRequestsFlowSettings(app)
		if err != nil {
			return apis.NewBadRequestError("invalid_requests_flow_settings", err)
		}
		step := parseFlowVersion(data[requestStepIndexDataKey])
		status := requestStatusForItem(record.GetBool("rejected"), step, flow.Steps)

		return e.JSON(http.StatusOK, map[string]any{
			"id":       record.Id,
			"status":   status,
			"step":     step,
			"rejected": record.GetBool("rejected"),
			"data":     record.Get("data"),
		})
	}
}
