package api

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	requestRoleAdmin     = "admin"
	requestRoleAssistant = "assistant"
	requestRoleGuardian  = "guardian"
)

const (
	requestFlowActionAssignGroup    = "assign_group"
	requestFlowActionAssignGuardian = "assign_guardian"
	requestFlowActionMentoring      = "mentoring"
	requestFlowActionGroupApproved  = "group_approved"
	requestFlowActionAdminApproved  = "admin_approved"
)

const (
	requestStatusSubmitted        = "1-submitted"
	requestStatusGroupAssigned    = "2-group_assigned"
	requestStatusGuardianAssigned = "3-guardian_assigned"
	requestStatusMentoring        = "4-mentoring"
	requestStatusGroupApproved    = "5-group_approved"
	requestStatusAdminApproved    = "6-admin_approved"
)

type requestsFlowStep struct {
	Role   string `json:"role"`
	Action string `json:"action"`
	Label  string `json:"label"`
}

type requestsFlowConfig struct {
	Version int                `json:"version"`
	Steps   []requestsFlowStep `json:"steps"`
}

const (
	requestFlowVersionDataKey = "__flow_version"
	requestStepIndexDataKey   = "__step_index"
)

var requestAllowedRoles = map[string]struct{}{
	requestRoleAdmin:     {},
	requestRoleAssistant: {},
	requestRoleGuardian:  {},
}

var requestAllowedActions = map[string]struct{}{
	requestFlowActionAssignGroup:    {},
	requestFlowActionAssignGuardian: {},
	requestFlowActionMentoring:      {},
	requestFlowActionGroupApproved:  {},
	requestFlowActionAdminApproved:  {},
}

var requestActionToStatus = map[string]string{
	requestFlowActionAssignGroup:    requestStatusGroupAssigned,
	requestFlowActionAssignGuardian: requestStatusGuardianAssigned,
	requestFlowActionMentoring:      requestStatusMentoring,
	requestFlowActionGroupApproved:  requestStatusGroupApproved,
	requestFlowActionAdminApproved:  requestStatusAdminApproved,
}

func parseRequestsFlowConfig(data map[string]any) (requestsFlowConfig, error) {
	version := parseFlowVersion(data["version"])
	if version < 1 {
		return requestsFlowConfig{}, fmt.Errorf("settings.requests_flow version must be >= 1")
	}

	rawSteps, ok := data["steps"].([]any)
	if !ok || len(rawSteps) == 0 {
		return requestsFlowConfig{}, fmt.Errorf("settings.requests_flow steps must be a non-empty array")
	}

	steps := make([]requestsFlowStep, 0, len(rawSteps))
	for i, raw := range rawSteps {
		entry, ok := raw.(map[string]any)
		if !ok {
			return requestsFlowConfig{}, fmt.Errorf("settings.requests_flow steps[%d] is invalid", i)
		}

		role := strings.TrimSpace(anyToString(entry["role"]))
		if role == "" {
			return requestsFlowConfig{}, fmt.Errorf("settings.requests_flow steps[%d] missing role", i)
		}
		if _, exists := requestAllowedRoles[role]; !exists {
			return requestsFlowConfig{}, fmt.Errorf("settings.requests_flow steps[%d] invalid role: %s", i, role)
		}

		action := strings.TrimSpace(anyToString(entry["action"]))
		if action == "" {
			return requestsFlowConfig{}, fmt.Errorf("settings.requests_flow steps[%d] missing action", i)
		}
		if _, exists := requestAllowedActions[action]; !exists {
			return requestsFlowConfig{}, fmt.Errorf("settings.requests_flow steps[%d] invalid action: %s", i, action)
		}

		label := strings.TrimSpace(anyToString(entry["label"]))
		steps = append(steps, requestsFlowStep{Role: role, Action: action, Label: label})
	}

	return requestsFlowConfig{
		Version: version,
		Steps:   steps,
	}, nil
}

func parseFlowVersion(raw any) int {
	switch value := raw.(type) {
	case float64:
		return int(value)
	case float32:
		return int(value)
	case int:
		return value
	case int64:
		return int(value)
	case int32:
		return int(value)
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return 0
		}
		return parsed
	default:
		return 0
	}
}

func requestProgressFromData(data map[string]any) (flowVersion int, stepIndex int) {
	flowVersion = parseFlowVersion(data[requestFlowVersionDataKey])
	rawStep := data[requestStepIndexDataKey]
	stepIndex = parseFlowVersion(rawStep)
	if flowVersion < 1 {
		flowVersion = 1
	}
	if stepIndex < 0 {
		stepIndex = 0
	}
	return flowVersion, stepIndex
}

func requiredFieldForAction(action string) string {
	switch action {
	case requestFlowActionAssignGroup:
		return "group"
	case requestFlowActionAssignGuardian:
		return "guardian"
	default:
		return ""
	}
}

func statusForStepIndex(stepIndex int, steps []requestsFlowStep) string {
	if stepIndex <= 0 {
		return requestStatusSubmitted
	}
	if len(steps) == 0 {
		return requestStatusSubmitted
	}
	if stepIndex > len(steps) {
		stepIndex = len(steps)
	}
	action := steps[stepIndex-1].Action
	if status, ok := requestActionToStatus[action]; ok {
		return status
	}
	return requestStatusSubmitted
}

func normalizeStatusValue(status string) string {
	trimmed := strings.TrimSpace(status)
	if trimmed == "" {
		return ""
	}
	clean := strings.TrimSpace(trimmed)
	if dash := strings.Index(clean, "-"); dash > 0 {
		numberPart := clean[:dash]
		if _, err := strconv.Atoi(numberPart); err == nil {
			return clean[dash+1:]
		}
	}
	return clean
}

func stepIndexFromStatus(status string, steps []requestsFlowStep) int {
	normalized := normalizeStatusValue(status)
	if normalized == "" || normalized == "submitted" {
		return 0
	}
	for i := range steps {
		targetStatus := normalizeStatusValue(statusForStepIndex(i+1, steps))
		if targetStatus == normalized {
			return i + 1
		}
	}
	return 0
}
