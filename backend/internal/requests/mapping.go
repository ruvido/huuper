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
	state, err := BuildWorkflowState(app, actor, record, item.Data, item.Rejected, flow)
	if err != nil {
		return ListItem{}, err
	}

	item.FlowVersion = FlowVersionFromData(item.Data)
	item.Status = state.Status
	item.Workflow = map[string]any{
		"total_steps":          len(flow.Steps),
		"has_next_step":        state.HasNext,
		"current_action":       state.CurrentAction,
		"current_action_label": state.CurrentActionLabel,
		"next_role":            state.NextStep.Role,
		"next_action":          state.NextStep.Action,
		"next_action_label":    state.NextStep.Label,
		"next_action_notes":    state.NextStep.Notes,
		"required_field":       state.RequiredField,
		"can_advance":          state.CanAdvance,
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
	CanAdvance         bool
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
		state.CurrentAction = nextStep.Action
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
		state.CanAdvance = ok
	}
	return state, nil
}
