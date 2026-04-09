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

func MapItemWithWorkflow(app *pocketbase.PocketBase, actor *core.Record, record *core.Record) (ListItem, error) {
	item := MapItem(record)
	flow, err := LoadFlowForRequest(app, item.Data)
	if err != nil {
		return ListItem{}, err
	}
	state, err := BuildWorkflowState(app, actor, record, item.Data, item.Rejected, flow)
	if err != nil {
		return ListItem{}, err
	}

	item.FlowVersion = FlowVersionFromData(item.Data)
	item.Status = state.Status
	item.Workflow = BuildWorkflowPayload(state, flow)
	return item, nil
}

func BuildWorkflowPayload(state WorkflowState, flow FlowConfig) map[string]any {
	action := state.CurrentAction
	return map[string]any{
		"total_steps":          len(flow.Steps),
		"has_next_step":        state.HasNext,
		"current_action":       action,
		"current_action_label": state.CurrentActionLabel,
		"next_role":            state.NextStep.Role,
		"next_action":          action,
		"next_action_label":    state.NextStep.Label,
		"next_action_notes":    state.NextStep.Notes,
		"required_field":       state.RequiredField,
		"can_take_action":      state.CanTakeAction,
		"can_reject":           state.CanReject,
		"current_version":      flow.Version,
	}
}

func EffectiveStepIndex(record *core.Record, data map[string]any, flow FlowConfig) int {
	if record == nil {
		return 0
	}

	stepIndex := 0
	for stepIndex < len(flow.Steps) {
		if StepSatisfied(record, data, flow.Steps[stepIndex].Action) {
			stepIndex++
			continue
		}
		break
	}
	return stepIndex
}

type WorkflowState struct {
	StepIndex          int
	Status             string
	HasNext            bool
	NextStep           FlowStep
	RequiredField      string
	CanTakeAction      bool
	CanReject          bool
	CurrentAction      string
	CurrentActionLabel string
}

func BuildWorkflowState(app *pocketbase.PocketBase, actor *core.Record, record *core.Record, data map[string]any, rejected bool, flow FlowConfig) (WorkflowState, error) {
	stepIndex := EffectiveStepIndex(record, data, flow)
	nextStep, hasNext := FlowStepAt(flow, stepIndex)
	status := StatusForItem(rejected, stepIndex, flow.Steps)

	state := WorkflowState{
		StepIndex: stepIndex,
		Status:    status,
		HasNext:   hasNext,
		NextStep:  nextStep,
	}
	if hasNext {
		state.CurrentAction = ActionForFlowAction(nextStep.Action)
		state.CurrentActionLabel = nextStep.Label
	}
	if state.CurrentActionLabel == "" {
		state.CurrentActionLabel = strings.ReplaceAll(NormalizeStatus(status), "_", " ")
	}
	if hasNext && !rejected {
		state.RequiredField = RequiredFieldForAction(nextStep.Action)
		ok, err := backendinternal.HasRoleForRequest(app, actor, record, nextStep.Role, RoleAdmin, RoleGuardian, RoleAssistant)
		if err != nil {
			return WorkflowState{}, err
		}
		state.CanTakeAction = ok
	}
	state.CanReject = actor != nil && actor.GetBool("admin") && !rejected
	return state, nil
}
