package requests

import (
	"strings"
	"time"

	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func ApplyAdvanceAction(app *pocketbase.PocketBase, actor *core.Record, record *core.Record, data map[string]any, payload ActionPayload) error {
	if record.GetBool("rejected") {
		return apis.NewBadRequestError("request_rejected", nil)
	}

	flow, err := LoadFlowSettings(app)
	if err != nil {
		return apis.NewBadRequestError("invalid_requests_flow_settings", err)
	}

	stepIndex := EffectiveStepIndex(record, data, flow)
	nextStep, hasNext := FlowStepAt(flow, stepIndex)
	if !hasNext {
		return apis.NewBadRequestError("last_step_reached", nil)
	}

	ok, err := backendinternal.HasRoleForRequest(app, actor, record, nextStep.Role, RoleAdmin, RoleGuardian, RoleAssistant)
	if err != nil {
		return apis.NewBadRequestError("role_resolution_failed", err)
	}
	if !ok {
		return apis.NewForbiddenError("forbidden_transition", nil)
	}

	if nextStep.Action == FlowActionAssignGroup {
		if err := applyGroupAssignment(app, record, strings.TrimSpace(payload.GroupID)); err != nil {
			return err
		}
	}

	if nextStep.Action == FlowActionAssignGuardian {
		if err := applyGuardianAssignment(app, record, data, actor, strings.TrimSpace(payload.GuardianID)); err != nil {
			return err
		}
	}

	if nextStep.Action == FlowActionMentoring {
		note := strings.TrimSpace(payload.MentoringNotes)
		if note == "" {
			return apis.NewBadRequestError("missing_mentoring_notes", nil)
		}
		data["mentoring_notes"] = note
		data["mentoring_done_at"] = time.Now().UTC().Format(time.RFC3339)
		if actor != nil {
			data["mentoring_done_by"] = actorDisplayName(actor)
		}
	}

	nextStepIndex := stepIndex + 1
	data[FlowVersionDataKey] = flow.Version
	data[StepIndexDataKey] = nextStepIndex
	record.Set("data", data)
	return nil
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
	if record.GetBool("rejected") {
		return apis.NewBadRequestError("request_rejected", nil)
	}

	ok, err := backendinternal.HasRoleForRequest(app, actor, record, RoleAssistant, RoleAdmin, RoleGuardian, RoleAssistant)
	if err != nil {
		return apis.NewBadRequestError("role_resolution_failed", err)
	}
	if !ok {
		return apis.NewForbiddenError("forbidden_transition", nil)
	}

	if guardianID == "" {
		record.Set("guardian", "")
		delete(data, "guardian")
		record.Set("data", data)
		return nil
	}

	if err := applyGuardianAssignment(app, record, data, actor, guardianID); err != nil {
		return err
	}
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

	flow, err := LoadFlowSettings(app)
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

func applyGuardianAssignment(app *pocketbase.PocketBase, record *core.Record, data map[string]any, actor *core.Record, guardianID string) error {
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
