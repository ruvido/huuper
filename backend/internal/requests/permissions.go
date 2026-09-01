package requests

import (
	"strings"

	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func VisibleRequestForActor(app *pocketbase.PocketBase, actor *core.Record, requestID string) (*core.Record, error) {
	if actor == nil {
		return nil, apis.NewUnauthorizedError("Unauthorized", nil)
	}

	id := strings.TrimSpace(requestID)
	if id == "" {
		return nil, apis.NewBadRequestError("invalid_request", nil)
	}

	record, err := app.FindRecordById("requests", id)
	if err != nil || record == nil {
		return nil, apis.NewNotFoundError("request_not_found", err)
	}

	if !CanViewRequest(app, actor, record) {
		return nil, apis.NewForbiddenError("forbidden_request", nil)
	}

	return record, nil
}

func CanViewRequest(app *pocketbase.PocketBase, actor *core.Record, record *core.Record) bool {
	if actor == nil || record == nil {
		return false
	}
	if actor.GetBool("admin") {
		return true
	}
	if IsAssistantForRequestGroup(app, actor, record) {
		return true
	}
	return IsGuardianForRequest(actor, record)
}

func CanTakeFlowStepAction(app *pocketbase.PocketBase, actor *core.Record, record *core.Record, step FlowStep) (bool, error) {
	if actor == nil || record == nil {
		return false, nil
	}
	if record.GetBool("archived") {
		return false, nil
	}
	if actor.GetBool("admin") {
		return true, nil
	}

	switch step.Action {
	case FlowActionAssignGroup, FlowActionAdminApproved:
		return false, nil
	case FlowActionAssignGuardian, FlowActionGroupApproved:
		return IsAssistantForRequestGroup(app, actor, record), nil
	case FlowActionMentoring:
		if IsAssistantForRequestGroup(app, actor, record) {
			return true, nil
		}
		return IsGuardianForRequest(actor, record), nil
	default:
		return false, nil
	}
}

func CanArchiveRequest(app *pocketbase.PocketBase, actor *core.Record, record *core.Record) (bool, error) {
	if actor == nil || record == nil {
		return false, nil
	}
	if record.GetBool("archived") {
		return false, nil
	}
	if actor.GetBool("admin") {
		return true, nil
	}
	if IsAssistantForRequestGroup(app, actor, record) {
		return true, nil
	}
	return IsGuardianForRequest(actor, record), nil
}

func CanUnarchiveRequest(app *pocketbase.PocketBase, actor *core.Record, record *core.Record) (bool, error) {
	if actor == nil || record == nil {
		return false, nil
	}
	if !record.GetBool("archived") {
		return false, nil
	}
	if actor.GetBool("admin") {
		return true, nil
	}
	if IsAssistantForRequestGroup(app, actor, record) {
		return true, nil
	}
	return IsGuardianForRequest(actor, record), nil
}

func IsPrimaryActorForFlowStep(app *pocketbase.PocketBase, actor *core.Record, record *core.Record, step FlowStep) (bool, error) {
	return backendinternal.IsActorAssignedForRole(app, actor, record, step.Role, RoleAdmin, RoleGuardian, RoleAssistant)
}

func IsAssistantForRequestGroup(app *pocketbase.PocketBase, actor *core.Record, record *core.Record) bool {
	if app == nil || actor == nil || record == nil {
		return false
	}

	groupID := strings.TrimSpace(record.GetString("group"))
	if groupID == "" {
		return false
	}

	group, err := app.FindRecordById("groups", groupID)
	if err != nil || group == nil {
		return false
	}
	return strings.TrimSpace(group.GetString("assistant")) == actor.Id
}

func IsGuardianForRequest(actor *core.Record, record *core.Record) bool {
	if actor == nil || record == nil {
		return false
	}
	return strings.TrimSpace(record.GetString("guardian")) == actor.Id
}
