package me

import (
	"net/http"
	"strconv"
	"strings"

	backendinternal "members/backend/internal"
	groupinternal "members/backend/internal/groups"
	backendrequests "members/backend/internal/requests"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func ListRequestsHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor, err := backendinternal.RequireAuthenticatedActor(e)
		if err != nil {
			return err
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
		visibilityFilter, visibilityParams, err := backendinternal.VisibleRequestsFilter(app, actor)
		if err != nil {
			return apis.NewBadRequestError("failed_requests_visibility", err)
		}
		if visibilityFilter != "" {
			filters = append(filters, visibilityFilter)
			for key, value := range visibilityParams {
				params[key] = value
			}
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
		filters = append(filters, "rejected = false")

		filter := strings.Join(filters, " && ")
		items, err := listRequestItems(app, actor, filter, sort, page, perPage, params, status)
		if err != nil {
			return err
		}

		return e.JSON(http.StatusOK, map[string]any{
			"items": items,
			"page":  page,
		})
	}
}

func listRequestItems(app *pocketbase.PocketBase, actor *core.Record, filter string, sort string, page int, perPage int, params map[string]any, status string) ([]backendrequests.ListItem, error) {
	if strings.TrimSpace(status) == "" {
		records, err := app.FindRecordsByFilter("requests", filter, sort, perPage, (page-1)*perPage, params)
		if err != nil {
			return nil, apis.NewBadRequestError("failed_requests", err)
		}
		return mapRequestItems(app, actor, records, "")
	}

	targetCount := page * perPage
	batchSize := perPage
	if batchSize < 100 {
		batchSize = 100
	}
	if batchSize > 500 {
		batchSize = 500
	}

	start := (page - 1) * perPage
	selected := make([]*core.Record, 0, perPage)
	matched := 0

	for offset := 0; ; offset += batchSize {
		records, err := app.FindRecordsByFilter("requests", filter, sort, batchSize, offset, params)
		if err != nil {
			return nil, apis.NewBadRequestError("failed_requests", err)
		}
		if len(records) == 0 {
			break
		}

		for _, record := range records {
			ok, err := backendrequests.RecordMatchesStatus(app, record, status)
			if err != nil {
				return nil, apis.NewBadRequestError("failed_requests_workflow", err)
			}
			if !ok {
				continue
			}
			if matched >= start && len(selected) < perPage {
				selected = append(selected, record)
			}
			matched++
			if len(selected) >= perPage && matched >= targetCount {
				break
			}
		}
		if len(selected) >= perPage && matched >= targetCount {
			break
		}
		if len(records) < batchSize {
			break
		}
	}

	if len(selected) == 0 {
		return []backendrequests.ListItem{}, nil
	}
	return mapRequestItems(app, actor, selected, "")
}

func mapRequestItems(app *pocketbase.PocketBase, actor *core.Record, records []*core.Record, status string) ([]backendrequests.ListItem, error) {
	items := make([]backendrequests.ListItem, 0, len(records))
	for _, record := range records {
		item, err := backendrequests.MapItemWithWorkflow(app, actor, record)
		if err != nil {
			return nil, apis.NewBadRequestError("failed_requests_workflow", err)
		}
		if strings.TrimSpace(status) != "" && !strings.EqualFold(strings.TrimSpace(item.Status), strings.TrimSpace(status)) {
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

func GetRequestHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor, err := backendinternal.RequireAuthenticatedActor(e)
		if err != nil {
			return err
		}
		record, err := backendinternal.VisibleRequestForActor(app, actor, e.Request.PathValue("id"))
		if err != nil {
			return err
		}

		item := backendrequests.MapItem(record)
		flow, err := backendrequests.LoadFlowForRequest(app, item.Data)
		if err != nil {
			return apis.NewBadRequestError("invalid_requests_flow_settings", err)
		}
		liveFlowData, err := backendrequests.FindSettingData(app, "requests_flow")
		if err != nil {
			return apis.NewBadRequestError("invalid_requests_flow_settings", err)
		}
		liveFlowData = requestFlowResponseData(liveFlowData)
		flowVersion := backendrequests.FlowVersionFromData(item.Data)
		state, err := backendrequests.BuildWorkflowState(app, actor, record, item.Data, item.Rejected, flow)
		if err != nil {
			return apis.NewBadRequestError("failed_requests_workflow", err)
		}
		statusLabel := requestStatusLabel(state.Status, state.StepIndex, flow)
		fullName := requestDisplayName(item.Data, item.Email, item.ID)
		groupName, err := requestGroupName(app, item.GroupID)
		if err != nil {
			return apis.NewBadRequestError("failed_group_lookup", err)
		}
		guardianName, err := requestGuardianName(app, item.Guardian)
		if err != nil {
			return apis.NewBadRequestError("failed_guardian_lookup", err)
		}

		options, err := requestWorkflowOptions(app, actor, record, state.NextStep, state.RequiredField, state.CanTakeAction)
		if err != nil {
			return apis.NewBadRequestError("workflow_options_failed", err)
		}

		workflow := backendrequests.BuildWorkflowPayload(state, flow)
		workflow["pending_action_notes_html"] = requestNotesHTML(state.NextStep.Notes)
		workflow["filter"] = state.NextStep.Filter
		workflow["options"] = options

		return e.JSON(http.StatusOK, map[string]any{
			"id":                   item.ID,
			"full_name":            fullName,
			"email":                item.Email,
			"status":               state.Status,
			"status_label":         statusLabel,
			"rejected":             item.Rejected,
			"group":                item.GroupID,
			"group_name":           groupName,
			"guardian":             item.Guardian,
			"guardian_name":        guardianName,
			"assigned_at":          requestAssignedAt(item.Data),
			"mentoring_notes":      requestMentoringNotes(item.Data),
			"mentoring_notes_html": requestMentoringNotesHTML(item.Data),
			"flow_version":         flowVersion,
			"request_flow":         map[string]any{"data": liveFlowData},
			"created":              item.Created,
			"updated":              item.Updated,
			"data":                 item.Data,
			"workflow":             workflow,
		})
	}
}

func requestStatusLabel(status string, stepIndex int, flow backendrequests.FlowConfig) string {
	nextStep, hasNext := backendrequests.FlowStepAt(flow, stepIndex)
	if hasNext && strings.TrimSpace(nextStep.Label) != "" {
		return strings.TrimSpace(nextStep.Label)
	}
	return strings.ReplaceAll(backendrequests.NormalizeStatus(status), "_", " ")
}

func requestGroupName(app *pocketbase.PocketBase, groupID string) (string, error) {
	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		return "", nil
	}
	group, err := app.FindRecordById("groups", groupID)
	if err != nil || group == nil {
		return "", err
	}
	return strings.TrimSpace(group.GetString("name")), nil
}

func requestGuardianName(app *pocketbase.PocketBase, userID string) (string, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return "", nil
	}
	user, err := app.FindRecordById("users", userID)
	if err != nil || user == nil {
		return "", err
	}
	return groupinternal.UserDisplayName(user), nil
}

func requestDisplayName(data map[string]any, email string, fallbackID string) string {
	return strings.TrimSpace(backendrequests.DisplayName(data, strings.TrimSpace(email), fallbackID))
}

func requestAssignedAt(data map[string]any) string {
	guardian, ok := data["guardian"].(map[string]any)
	if !ok {
		return ""
	}
	assignedAt, _ := guardian["assigned_at"].(string)
	return strings.TrimSpace(assignedAt)
}

func requestMentoringNotes(data map[string]any) string {
	value, _ := data["mentoring_notes"].(string)
	return strings.TrimSpace(value)
}

func requestMentoringNotesHTML(data map[string]any) string {
	html, ok := backendinternal.RenderMarkdownHTML(requestMentoringNotes(data))
	if !ok {
		return ""
	}
	return html
}

func requestNotesHTML(raw string) string {
	html, ok := backendinternal.RenderMarkdownHTML(raw)
	if !ok {
		return ""
	}
	return html
}

func requestFlowResponseData(data map[string]any) map[string]any {
	if data == nil {
		return map[string]any{}
	}

	steps, _ := data["steps"].([]any)
	if len(steps) == 0 {
		return data
	}

	out := backendinternal.ParseJSONMap(data)
	if out == nil {
		out = map[string]any{}
	}

	renderedSteps := make([]any, 0, len(steps))
	for _, rawStep := range steps {
		step, ok := rawStep.(map[string]any)
		if !ok || step == nil {
			continue
		}

		renderedStep := backendinternal.ParseJSONMap(step)
		if renderedStep == nil {
			renderedStep = map[string]any{}
		}
		if notes, ok := renderedStep["notes"].(string); ok {
			if html, rendered := backendinternal.RenderMarkdownHTML(notes); rendered {
				renderedStep["notes_html"] = html
			}
		}
		if info, ok := renderedStep["info"].(string); ok {
			renderedStep["info"] = strings.TrimSpace(info)
		}
		renderedSteps = append(renderedSteps, renderedStep)
	}

	out["steps"] = renderedSteps
	return out
}

func requestWorkflowOptions(app *pocketbase.PocketBase, actor *core.Record, record *core.Record, step backendrequests.FlowStep, requiredField string, canAdvance bool) (map[string]any, error) {
	options := map[string]any{}
	if !canAdvance {
		return options, nil
	}

	switch requiredField {
	case "group":
		filter := ""
		params := map[string]any{}
		if step.Filter == backendrequests.FilterLocal {
			filter = "type = {:type}"
			params["type"] = backendrequests.FilterLocal
		}
		groups, err := app.FindRecordsByFilter("groups", filter, "name", 500, 0, params)
		if err != nil {
			return nil, err
		}
		items := make([]map[string]any, 0, len(groups))
		for _, group := range groups {
			if group == nil {
				continue
			}
			items = append(items, map[string]any{
				"id":   group.Id,
				"name": strings.TrimSpace(group.GetString("name")),
				"type": strings.TrimSpace(group.GetString("type")),
			})
		}
		options["groups"] = items
	case "guardian":
		groupID := strings.TrimSpace(record.GetString("group"))
		if groupID == "" {
			return options, nil
		}
		response, err := groupinternal.MembersResponseForGroup(app, groupID)
		if err != nil {
			return nil, err
		}
		options["guardians"] = response.Items
	}

	return options, nil
}

func RequestActionHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor, err := backendinternal.RequireAuthenticatedActor(e)
		if err != nil {
			return err
		}

		id := e.Request.PathValue("id")
		if id == "" {
			return apis.NewBadRequestError("invalid_request", nil)
		}

		record, err := backendinternal.VisibleRequestForActor(app, actor, id)
		if err != nil {
			return err
		}

		var payload backendrequests.ActionPayload
		if err := e.BindBody(&payload); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}

		action := strings.TrimSpace(payload.Action)
		if action == "" {
			return apis.NewBadRequestError("missing_action", nil)
		}

		data := backendinternal.ParseJSONMap(record.Get("data"))

		deleteRequest := false
		promotedUserID := ""
		promotedUserCreated := false
		switch action {
		case backendrequests.ActionSetGroup:
			if err := backendrequests.ApplySetGroupAction(app, actor, record, data, strings.TrimSpace(payload.GroupID)); err != nil {
				return err
			}
		case backendrequests.ActionSetGuardian:
			if err := backendrequests.ApplySetGuardianAction(app, actor, record, data, strings.TrimSpace(payload.GuardianID)); err != nil {
				return err
			}
		case backendrequests.ActionSetMentoring:
			if err := backendrequests.ApplySetMentoringAction(app, actor, record, data, payload); err != nil {
				return err
			}
		case backendrequests.ActionSetGroupApprove:
			if err := backendrequests.ApplySetGroupApprovedAction(app, actor, record, data, payload); err != nil {
				return err
			}
		case backendrequests.ActionSetAdminApprove:
			if err := backendrequests.ApplySetAdminApprovedAction(app, actor, record, data, payload); err != nil {
				return err
			}
			flow, err := backendrequests.LoadFlowForRequest(app, data)
			if err != nil {
				return apis.NewBadRequestError("invalid_requests_flow_settings", err)
			}
			if backendrequests.EffectiveStepIndex(record, data, flow) >= len(flow.Steps) {
				promoteResult, err := backendrequests.ApplyPromoteAction(app, actor, record, data)
				if err != nil {
					notifyPromoteFailure(app, record, "auto-promote after admin approval", err)
					return err
				}
				deleteRequest = true
				promotedUserID = promoteResult.UserID
				promotedUserCreated = promoteResult.Created
			}
		case backendrequests.ActionReject:
			if err := backendrequests.ApplyRejectAction(actor, record, data, strings.TrimSpace(payload.Reason)); err != nil {
				return err
			}
		case backendrequests.ActionPromote:
			promoteResult, err := backendrequests.ApplyPromoteAction(app, actor, record, data)
			if err != nil {
				notifyPromoteFailure(app, record, "manual promote", err)
				return err
			}
			deleteRequest = true
			promotedUserID = promoteResult.UserID
			promotedUserCreated = promoteResult.Created
		default:
			return apis.NewBadRequestError("unsupported_action", nil)
		}

		if deleteRequest {
			if err := app.Delete(record); err != nil {
				if promotedUserCreated && strings.TrimSpace(promotedUserID) != "" {
					backendrequests.RollbackPromotedUser(app, promotedUserID, record)
				}
				return apis.NewBadRequestError("failed_to_delete_request", err)
			}
		} else if err := app.Save(record); err != nil {
			return apis.NewBadRequestError("failed_to_update_request", err)
		}

		flow, err := backendrequests.LoadFlowForRequest(app, data)
		if err != nil {
			return apis.NewBadRequestError("invalid_requests_flow_settings", err)
		}

		var notificationStep backendrequests.FlowStep
		switch action {
		case backendrequests.ActionPromote:
			if step, ok := backendrequests.FlowStepForAction(flow, backendrequests.FlowActionAdminApproved); ok {
				notificationStep = step
			}
		default:
			if flowAction := backendrequests.FlowActionForAction(action); flowAction != "" {
				if step, ok := backendrequests.FlowStepForAction(flow, flowAction); ok {
					notificationStep = step
				}
			}
		}
		if strings.TrimSpace(notificationStep.Action) != "" {
			backendrequests.NotifyRequestStep(app, actor, record, data, notificationStep)
		}

		if deleteRequest {
			return e.JSON(http.StatusOK, map[string]any{
				"id":       id,
				"promoted": true,
				"user_id":  promotedUserID,
			})
		}

		state, err := backendrequests.BuildWorkflowState(app, actor, record, data, record.GetBool("rejected"), flow)
		if err != nil {
			return apis.NewBadRequestError("failed_requests_workflow", err)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"id":       record.Id,
			"status":   state.Status,
			"rejected": record.GetBool("rejected"),
		})
	}
}

func notifyPromoteFailure(app *pocketbase.PocketBase, record *core.Record, stage string, cause error) {
	if app == nil || record == nil || cause == nil {
		return
	}

	data := backendinternal.ParseJSONMap(record.Get("data"))
	fullName := strings.TrimSpace(backendinternal.AnyToString(data["full_name"]))
	email := strings.TrimSpace(record.GetString("email"))
	if fullName == "" {
		fullName = email
	}
	if fullName == "" {
		fullName = strings.TrimSpace(record.Id)
	}

	subject := "Request promote failed"
	if email != "" {
		subject += " for " + email
	}

	body := strings.Join([]string{
		"Request promote failed.",
		"",
		"Stage: " + strings.TrimSpace(stage),
		"Reason: " + strings.TrimSpace(cause.Error()),
		"User: " + fullName,
		"Email: " + email,
		"Request ID: " + strings.TrimSpace(record.Id),
		"Group ID: " + strings.TrimSpace(record.GetString("group")),
	}, "\n")

	if !backendinternal.SendAdminFailureEmail(app, subject, body) {
		app.Logger().Warn("Failed to send promote failure email", "request", record.Id, "stage", stage, "error", cause)
	}
}
