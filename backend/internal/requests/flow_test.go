package requests

import (
	"testing"

	"github.com/pocketbase/pocketbase/core"
)

func testFlow() FlowConfig {
	return FlowConfig{
		Version: 2,
		Steps: []FlowStep{
			{Role: RoleAdmin, Action: FlowActionAssignGroup, Label: "Assign group"},
			{Role: RoleAssistant, Action: FlowActionAssignGuardian, Label: "Assign guardian"},
			{Role: RoleGuardian, Action: FlowActionMentoring, Label: "Complete mentoring"},
			{Role: RoleAssistant, Action: FlowActionGroupApproved, Label: "Approve group"},
			{Role: RoleAdmin, Action: FlowActionAdminApproved, Label: "Approve request"},
		},
	}
}

func testRequestRecord() *core.Record {
	collection := core.NewBaseCollection("requests")
	collection.Fields.Add(
		&core.TextField{Name: "email"},
		&core.TextField{Name: "group"},
		&core.TextField{Name: "guardian"},
		&core.BoolField{Name: "rejected"},
		&core.JSONField{Name: "data"},
	)
	record := core.NewRecord(collection)
	record.Set("email", "candidate@example.com")
	record.Set("data", map[string]any{})
	record.Set("rejected", false)
	return record
}

func TestEffectiveStepIndexTracksSatisfiedSteps(t *testing.T) {
	record := testRequestRecord()
	data := map[string]any{}
	flow := testFlow()

	if got := EffectiveStepIndex(record, data, flow); got != 0 {
		t.Fatalf("expected step index 0, got %d", got)
	}

	record.Set("group", "group-1")
	if got := EffectiveStepIndex(record, data, flow); got != 1 {
		t.Fatalf("expected step index 1 after group assignment, got %d", got)
	}

	record.Set("guardian", "guardian-1")
	data["mentoring_done_at"] = "2026-04-09T10:00:00Z"
	data["group_approved_at"] = "2026-04-09T11:00:00Z"
	if got := EffectiveStepIndex(record, data, flow); got != 4 {
		t.Fatalf("expected step index 4 after four completed steps, got %d", got)
	}
}

func TestResetStepsAfterActionClearsDownstreamState(t *testing.T) {
	record := testRequestRecord()
	record.Set("group", "group-1")
	record.Set("guardian", "guardian-1")
	data := map[string]any{
		"guardian": map[string]any{
			"name":        "Guardian",
			"assigned_at": "2026-04-09T10:00:00Z",
		},
		"mentoring_notes":   "done",
		"mentoring_done_at": "2026-04-09T10:30:00Z",
		"group_approved_at": "2026-04-09T11:00:00Z",
	}

	ResetStepsAfterAction(testFlow(), record, data, FlowActionAssignGroup)

	if got := record.GetString("guardian"); got != "" {
		t.Fatalf("expected guardian to be reset, got %q", got)
	}
	if _, ok := data["guardian"]; ok {
		t.Fatalf("expected guardian payload to be removed")
	}
	if _, ok := data["mentoring_done_at"]; ok {
		t.Fatalf("expected mentoring state to be removed")
	}
	if _, ok := data["group_approved_at"]; ok {
		t.Fatalf("expected group approval state to be removed")
	}
}

func TestBuildWorkflowPayloadUsesExplicitPendingFields(t *testing.T) {
	state := WorkflowState{
		HasNext:            true,
		NextStep:           FlowStep{Role: RoleAssistant, Action: FlowActionAssignGuardian, Label: "Assign guardian", Notes: "Pick a guardian"},
		RequiredField:      "guardian",
		CanTakeAction:      true,
		CanReject:          true,
		CurrentAction:      ActionSetGuardian,
		CurrentActionLabel: "Assign guardian",
	}

	payload := BuildWorkflowPayload(state, testFlow())

	if payload["pending_action"] != ActionSetGuardian {
		t.Fatalf("expected pending_action %q, got %#v", ActionSetGuardian, payload["pending_action"])
	}
	if payload["pending_role"] != RoleAssistant {
		t.Fatalf("expected pending_role %q, got %#v", RoleAssistant, payload["pending_role"])
	}
	if payload["pending_action_label"] != "Assign guardian" {
		t.Fatalf("unexpected pending_action_label: %#v", payload["pending_action_label"])
	}
	if payload["can_take_pending_action"] != true {
		t.Fatalf("expected can_take_pending_action true, got %#v", payload["can_take_pending_action"])
	}
	if _, ok := payload["current_action"]; ok {
		t.Fatalf("payload should not expose ambiguous current_action")
	}
	if _, ok := payload["next_action"]; ok {
		t.Fatalf("payload should not expose ambiguous next_action")
	}
}

func TestRecordMatchesStatusUsesDerivedWorkflowStatus(t *testing.T) {
	record := testRequestRecord()
	record.Set("group", "group-1")
	data := map[string]any{
		"guardian": map[string]any{
			"name":        "Guardian",
			"assigned_at": "2026-04-09T10:00:00Z",
		},
		"mentoring_done_at": "2026-04-09T10:30:00Z",
	}
	record.Set("guardian", "guardian-1")
	record.Set("data", SetRequestFlowSnapshot(data, testFlow()))

	matches, err := RecordMatchesStatus(nil, record, "mentoring")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !matches {
		t.Fatalf("expected mentoring status to match after mentoring is completed")
	}

	matches, err = RecordMatchesStatus(nil, record, "group_approved")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if matches {
		t.Fatalf("group_approved should not match before group approval is completed")
	}
}

func TestParseFlowConfigRejectsAssistantAssignGroup(t *testing.T) {
	_, err := ParseFlowConfig(map[string]any{
		"version": 1,
		"steps": []any{
			map[string]any{
				"role":   RoleAssistant,
				"action": FlowActionAssignGroup,
				"label":  "Assign group",
				"filter": FilterLocal,
			},
		},
	})
	if err == nil {
		t.Fatalf("expected assign_group with assistant role to be rejected")
	}
}
