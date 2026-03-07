package api

import (
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func mapRequestItem(record *core.Record) requestListItem {
	return requestListItem{
		ID:       record.Id,
		Email:    record.GetString("email"),
		Status:   "",
		Rejected: record.GetBool("rejected"),
		GroupID:  record.GetString("group"),
		Guardian: record.GetString("guardian"),
		Created:  record.GetString("created"),
		Updated:  record.GetString("updated"),
		Data:     parseJSONMap(record.Get("data")),
	}
}

func mapRequestItemWithWorkflow(app *pocketbase.PocketBase, actor *core.Record, record *core.Record, flow requestsFlowConfig) (requestListItem, error) {
	item := mapRequestItem(record)
	flowVersion, _ := requestProgressFromData(item.Data)
	stepIndex := effectiveRequestStepIndex(record, item.Data, flow)

	nextStep, hasNext := flowStepAt(flow, stepIndex)
	canAdvance := false
	requiredField := ""
	if hasNext && !item.Rejected {
		requiredField = requiredFieldForAction(nextStep.Action)
		ok, err := hasRoleForRequest(app, actor, record, nextStep.Role)
		if err != nil {
			return requestListItem{}, err
		}
		canAdvance = ok
	}

	item.FlowVersion = flowVersion
	item.StepIndex = stepIndex
	item.Status = requestStatusForItem(item.Rejected, stepIndex, flow.Steps)
	currentStep, hasCurrent := flowStepAt(flow, stepIndex-1)
	currentAction := ""
	currentActionLabel := ""
	if hasCurrent {
		currentAction = currentStep.Action
		currentActionLabel = currentStep.Label
	}
	if currentActionLabel == "" {
		normalized := normalizeStatusValue(item.Status)
		currentActionLabel = strings.ReplaceAll(normalized, "_", " ")
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

func effectiveRequestStepIndex(record *core.Record, data map[string]any, flow requestsFlowConfig) int {
	_, stepIndex := requestProgressFromData(data)
	if record == nil {
		return stepIndex
	}

	for stepIndex < len(flow.Steps) {
		required := requiredFieldForAction(flow.Steps[stepIndex].Action)
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
