package requests

import (
	"fmt"
	"strconv"
	"strings"

	backendinternal "members/backend/internal"
)

const (
	RoleAdmin     = "admin"
	RoleAssistant = "assistant"
	RoleGuardian  = "guardian"
)

const (
	FlowActionAssignGroup    = "assign_group"
	FlowActionAssignGuardian = "assign_guardian"
	FlowActionMentoring      = "mentoring"
	FlowActionGroupApproved  = "group_approved"
	FlowActionAdminApproved  = "admin_approved"
)

const (
	StatusSubmitted        = "1-submitted"
	StatusGroupAssigned    = "2-group_assigned"
	StatusGuardianAssigned = "3-guardian_assigned"
	StatusMentoring        = "4-mentoring"
	StatusGroupApproved    = "5-group_approved"
	StatusAdminApproved    = "6-admin_approved"
	StatusRejected         = "rejected"
)

type FlowStep struct {
	Role   string `json:"role"`
	Action string `json:"action"`
	Label  string `json:"label"`
	Notes  string `json:"notes,omitempty"`
}

type FlowConfig struct {
	Version int        `json:"version"`
	Steps   []FlowStep `json:"steps"`
}

const (
	FlowVersionDataKey = "__flow_version"
	StepIndexDataKey   = "__step_index"
)

var allowedRoles = map[string]struct{}{
	RoleAdmin:     {},
	RoleAssistant: {},
	RoleGuardian:  {},
}

var allowedActions = map[string]struct{}{
	FlowActionAssignGroup:    {},
	FlowActionAssignGuardian: {},
	FlowActionMentoring:      {},
	FlowActionGroupApproved:  {},
	FlowActionAdminApproved:  {},
}

var actionToStatus = map[string]string{
	FlowActionAssignGroup:    StatusGroupAssigned,
	FlowActionAssignGuardian: StatusGuardianAssigned,
	FlowActionMentoring:      StatusMentoring,
	FlowActionGroupApproved:  StatusGroupApproved,
	FlowActionAdminApproved:  StatusAdminApproved,
}

func ParseFlowConfig(data map[string]any) (FlowConfig, error) {
	version := ParseVersion(data["version"])
	if version < 1 {
		return FlowConfig{}, fmt.Errorf("settings.requests_flow version must be >= 1")
	}

	rawSteps, ok := data["steps"].([]any)
	if !ok || len(rawSteps) == 0 {
		return FlowConfig{}, fmt.Errorf("settings.requests_flow steps must be a non-empty array")
	}

	steps := make([]FlowStep, 0, len(rawSteps))
	for i, raw := range rawSteps {
		entry, ok := raw.(map[string]any)
		if !ok {
			return FlowConfig{}, fmt.Errorf("settings.requests_flow steps[%d] is invalid", i)
		}

		role := strings.TrimSpace(backendinternal.AnyToString(entry["role"]))
		if role == "" {
			return FlowConfig{}, fmt.Errorf("settings.requests_flow steps[%d] missing role", i)
		}
		if _, exists := allowedRoles[role]; !exists {
			return FlowConfig{}, fmt.Errorf("settings.requests_flow steps[%d] invalid role: %s", i, role)
		}

		action := strings.TrimSpace(backendinternal.AnyToString(entry["action"]))
		if action == "" {
			return FlowConfig{}, fmt.Errorf("settings.requests_flow steps[%d] missing action", i)
		}
		if _, exists := allowedActions[action]; !exists {
			return FlowConfig{}, fmt.Errorf("settings.requests_flow steps[%d] invalid action: %s", i, action)
		}

		label := strings.TrimSpace(backendinternal.AnyToString(entry["label"]))
		notes := strings.TrimSpace(backendinternal.AnyToString(entry["notes"]))
		steps = append(steps, FlowStep{Role: role, Action: action, Label: label, Notes: notes})
	}

	return FlowConfig{Version: version, Steps: steps}, nil
}

func ParseVersion(raw any) int {
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

func ProgressFromData(data map[string]any) (flowVersion int, stepIndex int) {
	flowVersion = ParseVersion(data[FlowVersionDataKey])
	stepIndex = ParseVersion(data[StepIndexDataKey])
	if flowVersion < 1 {
		flowVersion = 1
	}
	if stepIndex < 0 {
		stepIndex = 0
	}
	return flowVersion, stepIndex
}

func RequiredFieldForAction(action string) string {
	switch action {
	case FlowActionAssignGroup:
		return "group"
	case FlowActionAssignGuardian:
		return "guardian"
	default:
		return ""
	}
}

func StatusForStepIndex(stepIndex int, steps []FlowStep) string {
	if stepIndex <= 0 || len(steps) == 0 {
		return StatusSubmitted
	}
	if stepIndex > len(steps) {
		stepIndex = len(steps)
	}
	action := steps[stepIndex-1].Action
	if status, ok := actionToStatus[action]; ok {
		return status
	}
	return StatusSubmitted
}

func StatusForItem(rejected bool, stepIndex int, steps []FlowStep) string {
	if rejected {
		return StatusRejected
	}
	return StatusForStepIndex(stepIndex, steps)
}

func NormalizeStatus(status string) string {
	trimmed := strings.TrimSpace(status)
	if trimmed == "" {
		return ""
	}
	if dash := strings.Index(trimmed, "-"); dash > 0 {
		numberPart := trimmed[:dash]
		if _, err := strconv.Atoi(numberPart); err == nil {
			return trimmed[dash+1:]
		}
	}
	return trimmed
}

func StepIndexFromStatus(status string, steps []FlowStep) int {
	normalized := NormalizeStatus(status)
	if normalized == "" || normalized == "submitted" {
		return 0
	}
	for i := range steps {
		target := NormalizeStatus(StatusForStepIndex(i + 1, steps))
		if target == normalized {
			return i + 1
		}
	}
	return 0
}
