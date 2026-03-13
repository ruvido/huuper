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
	flowVersion, _ := ProgressFromData(item.Data)
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
	item.StepIndex = stepIndex
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
	_, stepIndex := ProgressFromData(data)
	if record == nil {
		return stepIndex
	}

	for stepIndex < len(flow.Steps) {
		required := RequiredFieldForAction(flow.Steps[stepIndex].Action)
		if required == "group" && strings.TrimSpace(record.GetString("group")) != "" {
			stepIndex++
			continue
		}
		if required == "guardian" && strings.TrimSpace(record.GetString("guardian")) != "" {
			stepIndex++
			continue
		}
		break
	}
	return stepIndex
}
