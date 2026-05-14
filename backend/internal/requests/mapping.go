package requests

import (
	"strings"

	backendinternal "members/backend/internal"
	copywritingui "members/backend/internal/copywriting"

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

func StatusForRecord(app *pocketbase.PocketBase, record *core.Record) (string, error) {
	item := MapItem(record)
	flow, err := LoadFlowForRequest(app, item.Data)
	if err != nil {
		return "", err
	}
	stepIndex := EffectiveStepIndex(record, item.Data, flow)
	return StatusForItem(item.Rejected, stepIndex, flow.Steps), nil
}

func RecordMatchesStatus(app *pocketbase.PocketBase, record *core.Record, status string) (bool, error) {
	expected := NormalizeStatus(status)
	if expected == "" {
		return true, nil
	}

	current, err := StatusForRecord(app, record)
	if err != nil {
		return false, err
	}
	return strings.EqualFold(NormalizeStatus(current), expected), nil
}

func BuildWorkflowPayload(state WorkflowState, flow FlowConfig) map[string]any {
	return map[string]any{
		"total_steps":             len(flow.Steps),
		"has_pending_action":      state.HasNext,
		"pending_role":            state.NextStep.Role,
		"pending_action":          state.CurrentAction,
		"pending_flow_action":     state.NextStep.Action,
		"pending_action_label":    state.CurrentActionLabel,
		"pending_action_notes":    state.NextStep.Notes,
		"required_field":          state.RequiredField,
		"can_take_pending_action": state.CanTakeAction,
		"actor_is_assigned":       state.ActorIsAssigned,
		"can_reject":              state.CanReject,
		"flow_version":            flow.Version,
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
	ActorIsAssigned    bool
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
		state.CurrentActionLabel = copywritingui.RequestButtonLabel(nextStep.Cta, nextStep.Label)
	}
	if state.CurrentActionLabel == "" {
		state.CurrentActionLabel = strings.ReplaceAll(NormalizeStatus(status), "_", " ")
	}
	assigned := false
	if hasNext && !rejected {
		state.RequiredField = RequiredFieldForAction(nextStep.Action)
		ok, err := backendinternal.HasRoleForRequest(app, actor, record, nextStep.Role, RoleAdmin, RoleGuardian, RoleAssistant)
		if err != nil {
			return WorkflowState{}, err
		}
		state.CanTakeAction = ok
		assigned, err = backendinternal.IsActorAssignedForRole(app, actor, record, nextStep.Role, RoleAdmin, RoleGuardian, RoleAssistant)
		if err != nil {
			return WorkflowState{}, err
		}
		state.ActorIsAssigned = assigned
	}
	if !rejected {
		canReject, err := canRejectRequestForFlow(app, actor, record, data, flow)
		if err != nil {
			return WorkflowState{}, err
		}
		state.CanReject = canReject
	}
	return state, nil
}
