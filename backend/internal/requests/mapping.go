package requests

import (
	"strings"

	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func MapItem(record *core.Record) ListItem {
	return ListItem{
		ID:       record.Id,
		Email:    record.GetString("email"),
		Status:   "",
		Rejected: record.GetBool("rejected"),
		GroupID:  record.GetString("group"),
		Guardian: record.GetString("guardian"),
		Created:  record.GetString("created"),
		Updated:  record.GetString("updated"),
		Data:     backendinternal.ParseJSONMap(record.Get("data")),
	}
}

func MapItemWithWorkflow(app *pocketbase.PocketBase, actor *core.Record, record *core.Record, flow FlowConfig) (ListItem, error) {
	item := MapItem(record)
	flowVersion := FlowVersionFromData(item.Data)
	stepIndex := EffectiveStepIndex(record, item.Data, flow)

	nextStep, hasNext := FlowStepAt(flow, stepIndex)
	canAdvance := false
	requiredField := ""
	if hasNext && !item.Rejected {
		requiredField = RequiredFieldForAction(nextStep.Action)
		ok, err := backendinternal.HasRoleForRequest(app, actor, record, nextStep.Role, RoleAdmin, RoleGuardian, RoleAssistant)
		if err != nil {
			return ListItem{}, err
		}
		canAdvance = ok
	}

	item.FlowVersion = flowVersion
	item.Status = StatusForItem(item.Rejected, stepIndex, flow.Steps)
	currentAction := ""
	currentActionLabel := ""
	if hasNext {
		currentAction = nextStep.Action
		currentActionLabel = nextStep.Label
	}
	if currentActionLabel == "" {
		currentActionLabel = strings.ReplaceAll(NormalizeStatus(item.Status), "_", " ")
	}
	item.Workflow = map[string]any{
		"total_steps":          len(flow.Steps),
		"has_next_step":        hasNext,
		"current_action":       currentAction,
		"current_action_label": currentActionLabel,
		"next_role":            nextStep.Role,
		"next_action":          nextStep.Action,
		"next_action_label":    nextStep.Label,
		"next_action_notes":    nextStep.Notes,
		"required_field":       requiredField,
		"can_advance":          canAdvance,
		"current_version":      flow.Version,
	}
	return item, nil
}

func EffectiveStepIndex(record *core.Record, data map[string]any, flow FlowConfig) int {
	if record == nil {
		return 0
	}

	stepIndex := 0
	for stepIndex < len(flow.Steps) {
		if stepSatisfied(record, data, flow.Steps[stepIndex].Action) {
			stepIndex++
			continue
		}
		break
	}
	return stepIndex
}

func stepSatisfied(record *core.Record, data map[string]any, action string) bool {
	switch action {
	case FlowActionAssignGroup:
		return strings.TrimSpace(record.GetString("group")) != ""
	case FlowActionAssignGuardian:
		return strings.TrimSpace(record.GetString("guardian")) != ""
	case FlowActionMentoring:
		value, _ := data["mentoring_done_at"].(string)
		return strings.TrimSpace(value) != ""
	case FlowActionGroupApproved:
		value, _ := data["group_approved_at"].(string)
		return strings.TrimSpace(value) != ""
	case FlowActionAdminApproved:
		value, _ := data["admin_approved_at"].(string)
		return strings.TrimSpace(value) != ""
	default:
		return false
	}
}
