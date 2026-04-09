package requests

import (
	"strings"
	"time"

	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func expectedCurrentStep(app *pocketbase.PocketBase, actor *core.Record, record *core.Record, data map[string]any, expectedAction string) (FlowConfig, FlowStep, error) {
	if record.GetBool("rejected") {
		return FlowConfig{}, FlowStep{}, apis.NewBadRequestError("request_rejected", nil)
	}

	flow, err := LoadFlowForRequest(app, data)
	if err != nil {
		return FlowConfig{}, FlowStep{}, apis.NewBadRequestError("invalid_requests_flow_settings", err)
	}

	stepIndex := EffectiveStepIndex(record, data, flow)
	nextStep, hasNext := FlowStepAt(flow, stepIndex)
	if !hasNext {
		return FlowConfig{}, FlowStep{}, apis.NewBadRequestError("last_step_reached", nil)
	}
	if nextStep.Action != expectedAction {
		return FlowConfig{}, FlowStep{}, apis.NewBadRequestError("invalid_current_step", nil)
	}

	ok, err := backendinternal.HasRoleForRequest(app, actor, record, nextStep.Role, RoleAdmin, RoleGuardian, RoleAssistant)
	if err != nil {
		return FlowConfig{}, FlowStep{}, apis.NewBadRequestError("role_resolution_failed", err)
	}
	if !ok {
		return FlowConfig{}, FlowStep{}, apis.NewForbiddenError("forbidden_transition", nil)
	}

	return flow, nextStep, nil
}

func applyExpectedStepAction(app *pocketbase.PocketBase, actor *core.Record, record *core.Record, data map[string]any, payload ActionPayload, expectedAction string) error {
	flow, nextStep, err := expectedCurrentStep(app, actor, record, data, expectedAction)
	if err != nil {
		return err
	}

	if err := ApplyStepAction(app, actor, record, data, payload, nextStep); err != nil {
		return err
	}

	data = SetRequestFlowSnapshot(data, flow)
	record.Set("data", data)
	return nil
}

func ApplySetMentoringAction(app *pocketbase.PocketBase, actor *core.Record, record *core.Record, data map[string]any, payload ActionPayload) error {
	return applyExpectedStepAction(app, actor, record, data, payload, FlowActionMentoring)
}

func ApplySetGroupApprovedAction(app *pocketbase.PocketBase, actor *core.Record, record *core.Record, data map[string]any, payload ActionPayload) error {
	return applyExpectedStepAction(app, actor, record, data, payload, FlowActionGroupApproved)
}

func ApplySetAdminApprovedAction(app *pocketbase.PocketBase, actor *core.Record, record *core.Record, data map[string]any, payload ActionPayload) error {
	return applyExpectedStepAction(app, actor, record, data, payload, FlowActionAdminApproved)
}

func ApplyRejectAction(actor *core.Record, record *core.Record, data map[string]any, reason string) error {
	if reason == "" {
		return apis.NewBadRequestError("missing_reject_reason", nil)
	}

	if !actor.GetBool("admin") {
		return apis.NewForbiddenError("forbidden_reject", nil)
	}

	data["rejected"] = map[string]any{
		"reason":      reason,
		"rejected_at": time.Now().UTC().Format(time.RFC3339),
		"rejected_by": actorDisplayName(actor),
	}
	record.Set("rejected", true)
	record.Set("data", data)
	return nil
}

func ApplySetGuardianAction(app *pocketbase.PocketBase, actor *core.Record, record *core.Record, data map[string]any, guardianID string) error {
	flow, step, err := expectedCurrentStep(app, actor, record, data, FlowActionAssignGuardian)
	if err != nil {
		return err
	}

	if guardianID == "" {
		record.Set("guardian", "")
		delete(data, "guardian")
		ResetStepsAfterAction(flow, record, data, FlowActionAssignGuardian)
		data = SetRequestFlowSnapshot(data, flow)
		record.Set("data", data)
		return nil
	}

	if err := applyGuardianAssignment(app, record, data, actor, guardianID, step.Filter); err != nil {
		return err
	}
	ResetStepsAfterAction(flow, record, data, FlowActionAssignGuardian)
	data = SetRequestFlowSnapshot(data, flow)
	record.Set("data", data)
	return nil
}

func ApplySetGroupAction(app *pocketbase.PocketBase, actor *core.Record, record *core.Record, data map[string]any, groupID string) error {
	flow, step, err := expectedCurrentStep(app, actor, record, data, FlowActionAssignGroup)
	if err != nil {
		return err
	}

	if err := applyGroupAssignment(app, record, strings.TrimSpace(groupID), step.Filter); err != nil {
		return err
	}

	ResetStepsAfterAction(flow, record, data, FlowActionAssignGroup)
	data = SetRequestFlowSnapshot(data, flow)
	record.Set("data", data)
	return nil
}

func ApplyPromoteAction(app *pocketbase.PocketBase, actor *core.Record, record *core.Record, data map[string]any) (string, error) {
	if !actor.GetBool("admin") {
		return "", apis.NewForbiddenError("forbidden_promote", nil)
	}
	if record.GetBool("rejected") {
		return "", apis.NewBadRequestError("request_rejected", nil)
	}

	flow, err := LoadFlowForRequest(app, data)
	if err != nil {
		return "", apis.NewBadRequestError("invalid_requests_flow_settings", err)
	}
	stepIndex := EffectiveStepIndex(record, data, flow)
	if stepIndex < len(flow.Steps) {
		nextStep, hasNext := FlowStepAt(flow, stepIndex)
		if !hasNext || nextStep.Action != FlowActionAdminApproved {
			return "", apis.NewBadRequestError("invalid_promote_status", nil)
		}
		ok, err := backendinternal.HasRoleForRequest(app, actor, record, nextStep.Role, RoleAdmin, RoleGuardian, RoleAssistant)
		if err != nil {
			return "", apis.NewBadRequestError("role_resolution_failed", err)
		}
		if !ok {
			return "", apis.NewForbiddenError("forbidden_promote", nil)
		}
	}

	email, err := backendinternal.NormalizeEmail(record.GetString("email"))
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
	tempPassword := backendinternal.RandomToken()
	if tempPassword == "" {
		return "", apis.NewBadRequestError("failed_to_generate_password", nil)
	}
	user.Set("email", email)
	user.Set("password", tempPassword)
	user.Set("passwordConfirm", tempPassword)
	user.Set("status", "approved")

	userData := BuildUserData(data)
	if len(userData) > 0 {
		user.Set("data", userData)
	}

	if err := app.Save(user); err != nil {
		return "", apis.NewBadRequestError("failed_to_create_user", err)
	}

	return user.Id, nil
}

func applyGroupAssignment(app *pocketbase.PocketBase, record *core.Record, groupID string, filter string) error {
	if groupID == "" {
		return apis.NewBadRequestError("missing_group", nil)
	}
	group, err := app.FindRecordById("groups", groupID)
	if err != nil || group == nil {
		return apis.NewBadRequestError("invalid_group", err)
	}
	if filter == FilterLocal && strings.TrimSpace(group.GetString("type")) != FilterLocal {
		return apis.NewBadRequestError("invalid_group_filter", nil)
	}
	record.Set("group", groupID)
	return nil
}

func applyGuardianAssignment(app *pocketbase.PocketBase, record *core.Record, data map[string]any, actor *core.Record, guardianID string, filter string) error {
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

	if filter == FilterGroupMembers {
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
	}

	record.Set("guardian", guardianID)
	guardianPayload := map[string]any{
		"name":        actorDisplayName(guardian),
		"assigned_at": time.Now().UTC().Format(time.RFC3339),
	}
	if actor != nil {
		guardianPayload["assigned_by"] = actorDisplayName(actor)
	}
	data["guardian"] = guardianPayload
	return nil
}

func actorDisplayName(actor *core.Record) string {
	if actor == nil {
		return ""
	}

	data := backendinternal.ParseJSONMap(actor.Get("data"))
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
