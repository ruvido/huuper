package requests

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
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
	EmailToAdmin     = "admin"
	EmailToAssistant = "assistant"
	EmailToGuardian  = "guardian"
	EmailToCandidate = "candidate"
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
	Role            string `json:"role"`
	Action          string `json:"action"`
	Label           string `json:"label"`
	Cta             string `json:"cta,omitempty"`
	Notes           string `json:"notes,omitempty"`
	Filter          string `json:"filter,omitempty"`
	EmailTo         string `json:"email_to,omitempty"`
	TelegramMessage bool   `json:"telegram_message,omitempty"`
}

type FlowConfig struct {
	Version int        `json:"version"`
	Steps   []FlowStep `json:"steps"`
}

const (
	FlowVersionDataKey = "__flow_version"
	RequestFlowDataKey = "__request_flow"
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

var allowedEmailTargets = map[string]struct{}{
	EmailToAdmin:     {},
	EmailToAssistant: {},
	EmailToGuardian:  {},
	EmailToCandidate: {},
}

var actionToStatus = map[string]string{
	FlowActionAssignGroup:    StatusGroupAssigned,
	FlowActionAssignGuardian: StatusGuardianAssigned,
	FlowActionMentoring:      StatusMentoring,
	FlowActionGroupApproved:  StatusGroupApproved,
	FlowActionAdminApproved:  StatusAdminApproved,
}

type flowActionSpec struct {
	RequiredField string
	IsDone        func(record *core.Record, data map[string]any) bool
	Apply         func(app core.App, actor *core.Record, record *core.Record, data map[string]any, payload ActionPayload, step FlowStep) error
	Reset         func(record *core.Record, data map[string]any)
}

var flowActionSpecs = map[string]flowActionSpec{
	FlowActionAssignGroup: {
		RequiredField: "group",
		IsDone: func(record *core.Record, _ map[string]any) bool {
			return strings.TrimSpace(record.GetString("group")) != ""
		},
		Apply: func(app core.App, actor *core.Record, record *core.Record, data map[string]any, payload ActionPayload, step FlowStep) error {
			pb, ok := app.(*pocketbase.PocketBase)
			if !ok {
				return fmt.Errorf("invalid app")
			}
			return applyGroupAssignment(pb, record, data, actor, strings.TrimSpace(payload.GroupID), step.Filter)
		},
		Reset: func(record *core.Record, data map[string]any) {
			record.Set("group", "")
			delete(data, "assign_group")
		},
	},
	FlowActionAssignGuardian: {
		RequiredField: "guardian",
		IsDone: func(record *core.Record, _ map[string]any) bool {
			return strings.TrimSpace(record.GetString("guardian")) != ""
		},
		Apply: func(app core.App, actor *core.Record, record *core.Record, data map[string]any, payload ActionPayload, step FlowStep) error {
			pb, ok := app.(*pocketbase.PocketBase)
			if !ok {
				return fmt.Errorf("invalid app")
			}
			return applyGuardianAssignment(pb, record, data, actor, strings.TrimSpace(payload.GuardianID), step.Filter)
		},
		Reset: func(record *core.Record, data map[string]any) {
			record.Set("guardian", "")
			delete(data, "guardian")
		},
	},
	FlowActionMentoring: {
		RequiredField: "mentoring_notes",
		IsDone: func(_ *core.Record, data map[string]any) bool {
			return mentoringDoneAt(data) != ""
		},
		Apply: func(_ core.App, actor *core.Record, _ *core.Record, data map[string]any, payload ActionPayload, _ FlowStep) error {
			note := strings.TrimSpace(payload.MentoringNotes)
			if note != "" {
				appendMentoringNote(data, note, actor)
			}
			if len(mentoringNotes(data)) == 0 {
				return apis.NewBadRequestError("missing_mentoring_notes", nil)
			}
			markMentoringDone(data, actor)
			return nil
		},
		Reset: func(_ *core.Record, data map[string]any) {
			delete(data, mentoringDataKey)
		},
	},
	FlowActionGroupApproved: {
		IsDone: func(_ *core.Record, data map[string]any) bool {
			value, _ := data["group_approved_at"].(string)
			return strings.TrimSpace(value) != ""
		},
		Apply: func(_ core.App, actor *core.Record, _ *core.Record, data map[string]any, _ ActionPayload, _ FlowStep) error {
			data["group_approved_at"] = time.Now().UTC().Format(time.RFC3339)
			if actor != nil {
				data["group_approved_by"] = actorDisplayName(actor)
			}
			return nil
		},
		Reset: func(_ *core.Record, data map[string]any) {
			delete(data, "group_approved_at")
			delete(data, "group_approved_by")
		},
	},
	FlowActionAdminApproved: {
		IsDone: func(_ *core.Record, data map[string]any) bool {
			value, _ := data["admin_approved_at"].(string)
			return strings.TrimSpace(value) != ""
		},
		Apply: func(_ core.App, actor *core.Record, _ *core.Record, data map[string]any, _ ActionPayload, _ FlowStep) error {
			data["admin_approved_at"] = time.Now().UTC().Format(time.RFC3339)
			if actor != nil {
				data["admin_approved_by"] = actorDisplayName(actor)
			}
			return nil
		},
		Reset: func(_ *core.Record, data map[string]any) {
			delete(data, "admin_approved_at")
			delete(data, "admin_approved_by")
		},
	},
}

const (
	FilterLocal        = "local"
	FilterGroupMembers = "group_members"
)

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
		if err := validateRoleForAction(action, role); err != nil {
			return FlowConfig{}, fmt.Errorf("settings.requests_flow steps[%d] %w", i, err)
		}

		label := strings.TrimSpace(backendinternal.AnyToString(entry["label"]))
		cta := strings.TrimSpace(backendinternal.AnyToString(entry["cta"]))
		notes := strings.TrimSpace(backendinternal.AnyToString(entry["notes"]))
		filter := strings.TrimSpace(backendinternal.AnyToString(entry["filter"]))
		emailTo := strings.TrimSpace(backendinternal.AnyToString(entry["email_to"]))
		if emailTo != "" {
			if _, exists := allowedEmailTargets[emailTo]; !exists {
				return FlowConfig{}, fmt.Errorf("settings.requests_flow steps[%d] invalid email_to: %s", i, emailTo)
			}
		}
		telegramMessage := false
		switch value := entry["telegram_message"].(type) {
		case bool:
			telegramMessage = value
		default:
			parsed, err := ParseBoolQuery(backendinternal.AnyToString(value), false)
			if err != nil {
				return FlowConfig{}, fmt.Errorf("settings.requests_flow steps[%d] invalid telegram_message", i)
			}
			telegramMessage = parsed
		}
		if err := validateFilterForAction(action, filter); err != nil {
			return FlowConfig{}, fmt.Errorf("settings.requests_flow steps[%d] %w", i, err)
		}
		steps = append(steps, FlowStep{
			Role:            role,
			Action:          action,
			Label:           label,
			Cta:             cta,
			Notes:           notes,
			Filter:          filter,
			EmailTo:         emailTo,
			TelegramMessage: telegramMessage,
		})
	}

	return FlowConfig{Version: version, Steps: steps}, nil
}

func validateRoleForAction(action, role string) error {
	switch action {
	case FlowActionAssignGroup:
		if role != RoleAdmin {
			return fmt.Errorf("invalid role for %s: %s", action, role)
		}
	}
	return nil
}

func validateFilterForAction(action, filter string) error {
	if filter == "" {
		return nil
	}

	switch action {
	case FlowActionAssignGroup:
		if filter != FilterLocal {
			return fmt.Errorf("invalid filter for %s: %s", action, filter)
		}
		return nil
	case FlowActionAssignGuardian:
		if filter != FilterGroupMembers {
			return fmt.Errorf("invalid filter for %s: %s", action, filter)
		}
		return nil
	default:
		return fmt.Errorf("filter not allowed for %s", action)
	}
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

func FlowVersionFromData(data map[string]any) int {
	flowVersion := ParseVersion(data[FlowVersionDataKey])
	if flowVersion < 1 {
		if snapshot, ok := RequestFlowSnapshotFromData(data); ok {
			return snapshot.Version
		}
	}
	if flowVersion < 1 {
		return 1
	}
	return flowVersion
}

func RequiredFieldForAction(action string) string {
	spec, ok := flowActionSpecs[action]
	if !ok {
		return ""
	}
	return spec.RequiredField
}

func FlowStepForAction(flow FlowConfig, action string) (FlowStep, bool) {
	for _, step := range flow.Steps {
		if step.Action == action {
			return step, true
		}
	}
	return FlowStep{}, false
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
		target := NormalizeStatus(StatusForStepIndex(i+1, steps))
		if target == normalized {
			return i + 1
		}
	}
	return 0
}

func StepSatisfied(record *core.Record, data map[string]any, action string) bool {
	spec, ok := flowActionSpecs[action]
	if !ok || spec.IsDone == nil {
		return false
	}
	return spec.IsDone(record, data)
}

func ResetStep(record *core.Record, data map[string]any, action string) {
	spec, ok := flowActionSpecs[action]
	if !ok || spec.Reset == nil {
		return
	}
	spec.Reset(record, data)
}

func ResetStepsAfterAction(flow FlowConfig, record *core.Record, data map[string]any, action string) {
	found := false
	for _, step := range flow.Steps {
		if found {
			ResetStep(record, data, step.Action)
		}
		if step.Action == action {
			found = true
		}
	}
}

func ApplyStepAction(app core.App, actor *core.Record, record *core.Record, data map[string]any, payload ActionPayload, step FlowStep) error {
	spec, ok := flowActionSpecs[step.Action]
	if !ok || spec.Apply == nil {
		return nil
	}
	return spec.Apply(app, actor, record, data, payload, step)
}

func ActionForFlowAction(action string) string {
	switch strings.TrimSpace(action) {
	case FlowActionAssignGroup:
		return ActionSetGroup
	case FlowActionAssignGuardian:
		return ActionSetGuardian
	case FlowActionMentoring:
		return ActionSetMentoring
	case FlowActionGroupApproved:
		return ActionSetGroupApprove
	case FlowActionAdminApproved:
		return ActionSetAdminApprove
	default:
		return ""
	}
}

func FlowActionForAction(action string) string {
	switch strings.TrimSpace(action) {
	case ActionSetGroup:
		return FlowActionAssignGroup
	case ActionSetGuardian:
		return FlowActionAssignGuardian
	case ActionSetMentoring:
		return FlowActionMentoring
	case ActionSetGroupApprove:
		return FlowActionGroupApproved
	case ActionSetAdminApprove:
		return FlowActionAdminApproved
	default:
		return ""
	}
}
