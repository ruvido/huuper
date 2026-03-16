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
		records, err := app.FindRecordsByFilter("requests", filter, sort, perPage, (page-1)*perPage, params)
		if err != nil {
			return apis.NewBadRequestError("failed_requests", err)
		}

		assistantGroups, err := backendinternal.AssistantGroupIDsForUser(app, actor)
		if err != nil {
			return apis.NewBadRequestError("failed_groups_lookup", err)
		}
		flow, err := backendrequests.LoadFlowSettings(app)
		if err != nil {
			return apis.NewBadRequestError("invalid_requests_flow_settings", err)
		}

		items := make([]backendrequests.ListItem, 0, len(records))
		for _, record := range records {
			if !backendinternal.CanViewRequest(actor, record, assistantGroups) {
				continue
			}
			item, err := backendrequests.MapItemWithWorkflow(app, actor, record, flow)
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
		flow, err := backendrequests.LoadFlowSettings(app)
		if err != nil {
			return apis.NewBadRequestError("invalid_requests_flow_settings", err)
		}
		flowVersion, _ := backendrequests.ProgressFromData(item.Data)
		stepIndex := backendrequests.EffectiveStepIndex(record, item.Data, flow)
		computedStatus := backendrequests.StatusForItem(item.Rejected, stepIndex, flow.Steps)
		statusLabel := requestStatusLabel(computedStatus, stepIndex, flow)
		fullName := requestDisplayName(item.Data, item.Email, item.ID)
		groupName, err := requestGroupName(app, item.GroupID)
		if err != nil {
			return apis.NewBadRequestError("failed_group_lookup", err)
		}
		guardianName, err := requestGuardianName(app, item.Guardian)
		if err != nil {
			return apis.NewBadRequestError("failed_guardian_lookup", err)
		}

		nextStep, hasNext := backendrequests.FlowStepAt(flow, stepIndex)
		canAdvance := false
		requiredField := ""
		if hasNext && !item.Rejected {
			requiredField = backendrequests.RequiredFieldForAction(nextStep.Action)
			canAdvance, err = backendinternal.HasRoleForRequest(app, actor, record, nextStep.Role, backendrequests.RoleAdmin, backendrequests.RoleGuardian, backendrequests.RoleAssistant)
			if err != nil {
				return apis.NewBadRequestError("role_resolution_failed", err)
			}
		}
		options, err := requestWorkflowOptions(app, actor, record, requiredField, canAdvance)
		if err != nil {
			return apis.NewBadRequestError("workflow_options_failed", err)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"id":                   item.ID,
			"full_name":            fullName,
			"email":                item.Email,
			"status":               computedStatus,
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
			"step_index":           stepIndex,
			"created":              item.Created,
			"updated":              item.Updated,
			"data":                 item.Data,
			"workflow": map[string]any{
				"total_steps":            len(flow.Steps),
				"has_next_step":          hasNext,
				"next_role":              nextStep.Role,
				"next_action":            nextStep.Action,
				"next_action_label":      nextStep.Label,
				"next_action_notes":      nextStep.Notes,
				"next_action_notes_html": requestNotesHTML(nextStep.Notes),
				"required_field":         requiredField,
				"can_advance":            canAdvance,
				"options":                options,
				"current_version":        flow.Version,
			},
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

func requestWorkflowOptions(app *pocketbase.PocketBase, actor *core.Record, record *core.Record, requiredField string, canAdvance bool) (map[string]any, error) {
	options := map[string]any{}
	if !canAdvance {
		return options, nil
	}

	switch requiredField {
	case "group":
		groups, err := app.FindRecordsByFilter("groups", "", "name", 500, 0, nil)
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
		items := make([]map[string]any, 0, len(response.Items))
		for _, member := range response.Items {
			items = append(items, map[string]any{
				"id":        member.ID,
				"full_name": member.FullName,
				"email":     member.Email,
			})
		}
		options["guardians"] = items
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
		switch action {
		case backendrequests.ActionAdvance, "transition":
			if err := backendrequests.ApplyAdvanceAction(app, actor, record, data, payload); err != nil {
				return err
			}
		case backendrequests.ActionSetGuardian:
			if err := backendrequests.ApplySetGuardianAction(app, actor, record, data, strings.TrimSpace(payload.GuardianID)); err != nil {
				return err
			}
		case backendrequests.ActionReject:
			if err := backendrequests.ApplyRejectAction(actor, record, data, strings.TrimSpace(payload.Reason)); err != nil {
				return err
			}
		case backendrequests.ActionPromote:
			userID, err := backendrequests.ApplyPromoteAction(app, actor, record, data)
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

		flow, err := backendrequests.LoadFlowSettings(app)
		if err != nil {
			return apis.NewBadRequestError("invalid_requests_flow_settings", err)
		}
		step := backendrequests.ParseVersion(data[backendrequests.StepIndexDataKey])
		status := backendrequests.StatusForItem(record.GetBool("rejected"), step, flow.Steps)

		return e.JSON(http.StatusOK, map[string]any{
			"id":       record.Id,
			"status":   status,
			"step":     step,
			"rejected": record.GetBool("rejected"),
			"data":     record.Get("data"),
		})
	}
}
