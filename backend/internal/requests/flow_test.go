package requests

import (
	"testing"

	"github.com/pocketbase/pocketbase"
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
		&core.BoolField{Name: "archived"},
		&core.JSONField{Name: "data"},
	)
	record := core.NewRecord(collection)
	record.Set("email", "candidate@example.com")
	record.Set("data", map[string]any{})
	record.Set("archived", false)
	return record
}

func testUserRecord(id string, admin bool) *core.Record {
	collection := core.NewBaseCollection("users")
	collection.Fields.Add(
		&core.TextField{Name: "email"},
		&core.BoolField{Name: "admin"},
		&core.JSONField{Name: "data"},
	)
	record := core.NewRecord(collection)
	record.Set("id", id)
	record.Set("admin", admin)
	return record
}

func testPocketBase(t *testing.T) *pocketbase.PocketBase {
	t.Helper()

	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir:       t.TempDir(),
		DefaultEncryptionEnv: "pb_test_env",
	})
	if err := app.Bootstrap(); err != nil {
		t.Fatalf("failed to bootstrap test app: %v", err)
	}
	return app
}

func testGroupsCollection(t *testing.T, app *pocketbase.PocketBase) *core.Collection {
	t.Helper()

	groups := core.NewBaseCollection("groups")
	groups.Fields.Add(&core.TextField{Name: "assistant"})
	if err := app.Save(groups); err != nil {
		t.Fatalf("failed to save groups collection: %v", err)
	}
	return groups
}

func testGroupWithAssistant(t *testing.T, app *pocketbase.PocketBase, groups *core.Collection, id string, assistantID string) *core.Record {
	t.Helper()

	group := core.NewRecord(groups)
	group.Set("id", id)
	group.Set("assistant", assistantID)
	if err := app.Save(group); err != nil {
		t.Fatalf("failed to save group: %v", err)
	}
	return group
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
	data["mentoring"] = map[string]any{
		"notes":   []any{map[string]any{"text": "done", "at": "2026-04-09T10:00:00Z"}},
		"done_at": "2026-04-09T10:00:00Z",
	}
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
		"assign_group": map[string]any{
			"assigned_at": "2026-04-09T09:30:00Z",
			"assigned_by": "Admin",
		},
		"guardian": map[string]any{
			"name":        "Guardian",
			"assigned_at": "2026-04-09T10:00:00Z",
		},
		"mentoring": map[string]any{
			"notes":   []any{map[string]any{"text": "done", "at": "2026-04-09T10:30:00Z"}},
			"done_at": "2026-04-09T10:30:00Z",
		},
		"group_approved_at": "2026-04-09T11:00:00Z",
	}

	ResetStepsAfterAction(testFlow(), record, data, FlowActionAssignGroup)

	if got := record.GetString("guardian"); got != "" {
		t.Fatalf("expected guardian to be reset, got %q", got)
	}
	assignGroup, ok := data["assign_group"].(map[string]any)
	if !ok {
		t.Fatalf("expected assign_group payload to remain")
	}
	if assignedAt, _ := assignGroup["assigned_at"].(string); assignedAt == "" {
		t.Fatalf("expected assign_group assigned_at to remain")
	}
	if _, ok := data["guardian"]; ok {
		t.Fatalf("expected guardian payload to be removed")
	}
	if _, ok := data["mentoring"]; ok {
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
		CanArchive:         true,
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

func TestBuildWorkflowStateUsesCtaLabelFallback(t *testing.T) {
	record := testRequestRecord()
	record.Set("data", map[string]any{})

	flow := FlowConfig{
		Version: 1,
		Steps: []FlowStep{
			{Role: RoleAdmin, Action: FlowActionAdminApproved, Label: "In verifica", Cta: "Accetta"},
		},
	}

	collection := core.NewBaseCollection("users")
	collection.Fields.Add(
		&core.TextField{Name: "email"},
		&core.BoolField{Name: "admin"},
		&core.JSONField{Name: "data"},
	)
	admin := core.NewRecord(collection)
	admin.Set("id", "admin-1")
	admin.Set("admin", true)

	state, err := BuildWorkflowState(nil, admin, record, map[string]any{}, false, flow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state.CurrentActionLabel != "Accetta" {
		t.Fatalf("expected cta label to win, got %q", state.CurrentActionLabel)
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
		"mentoring": map[string]any{
			"notes":   []any{map[string]any{"text": "done", "at": "2026-04-09T10:30:00Z"}},
			"done_at": "2026-04-09T10:30:00Z",
		},
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

func TestBuildWorkflowStateAllowsAssignedGuardianOnly(t *testing.T) {
	record := testRequestRecord()
	record.Set("group", "group-1")
	record.Set("guardian", "guardian-1")
	data := map[string]any{
		"guardian": map[string]any{
			"name":        "Guardian",
			"assigned_at": "2026-04-09T10:00:00Z",
		},
	}
	record.Set("data", SetRequestFlowSnapshot(data, testFlow()))

	collection := core.NewBaseCollection("users")
	collection.Fields.Add(
		&core.TextField{Name: "email"},
		&core.BoolField{Name: "admin"},
		&core.JSONField{Name: "data"},
	)
	guardian := core.NewRecord(collection)
	guardian.Set("id", "guardian-1")
	other := core.NewRecord(collection)
	other.Set("id", "guardian-2")

	state, err := BuildWorkflowState(nil, guardian, record, data, false, testFlow())
	if err != nil {
		t.Fatalf("unexpected error for assigned guardian: %v", err)
	}
	if !state.CanTakeAction {
		t.Fatalf("expected assigned guardian to be allowed to take action")
	}
	if !state.ActorIsAssigned {
		t.Fatalf("expected assigned guardian to be flagged as ActorIsAssigned")
	}
	if state.RequiredField != "mentoring_notes" {
		t.Fatalf("expected mentoring_notes required field, got %q", state.RequiredField)
	}
	if state.CurrentAction != ActionSetMentoring {
		t.Fatalf("expected current action %q, got %q", ActionSetMentoring, state.CurrentAction)
	}

	state, err = BuildWorkflowState(nil, other, record, data, false, testFlow())
	if err != nil {
		t.Fatalf("unexpected error for unrelated guardian: %v", err)
	}
	if state.CanTakeAction {
		t.Fatalf("expected unrelated guardian to be blocked")
	}
	if state.ActorIsAssigned {
		t.Fatalf("expected unrelated guardian to not be flagged as ActorIsAssigned")
	}
}

func TestBuildWorkflowStateAllowsGroupAssistantToCompleteMentoring(t *testing.T) {
	app := testPocketBase(t)
	groups := testGroupsCollection(t, app)

	assistant := testUserRecord("assistant000001", false)
	group := testGroupWithAssistant(t, app, groups, "group0000000001", assistant.Id)

	record := testRequestRecord()
	record.Set("group", group.Id)
	record.Set("guardian", "guardian0000001")
	data := map[string]any{
		"guardian": map[string]any{
			"name":        "Guardian",
			"assigned_at": "2026-04-09T10:00:00Z",
		},
	}
	record.Set("data", SetRequestFlowSnapshot(data, testFlow()))

	state, err := BuildWorkflowState(app, assistant, record, data, false, testFlow())
	if err != nil {
		t.Fatalf("unexpected error for group assistant on mentoring step: %v", err)
	}
	if !state.CanTakeAction {
		t.Fatalf("expected group assistant to complete mentoring in own group")
	}
	if state.ActorIsAssigned {
		t.Fatalf("expected group assistant not to be primary actor for guardian mentoring step")
	}

	if err := ApplyAddMentoringNoteAction(app, assistant, record, data, ActionPayload{MentoringNotes: "Assistant note"}); err != nil {
		t.Fatalf("expected group assistant to add mentoring note: %v", err)
	}
	if err := ApplySetMentoringAction(app, assistant, record, data, ActionPayload{}); err != nil {
		t.Fatalf("expected group assistant to finalize mentoring: %v", err)
	}
}

func TestBuildWorkflowStateAdminNotAssignedToGuardianStep(t *testing.T) {
	record := testRequestRecord()
	record.Set("group", "group-1")
	record.Set("guardian", "guardian-1")
	data := map[string]any{
		"guardian": map[string]any{
			"name":        "Guardian",
			"assigned_at": "2026-04-09T10:00:00Z",
		},
	}
	record.Set("data", SetRequestFlowSnapshot(data, testFlow()))

	collection := core.NewBaseCollection("users")
	collection.Fields.Add(
		&core.TextField{Name: "email"},
		&core.BoolField{Name: "admin"},
		&core.JSONField{Name: "data"},
	)
	admin := core.NewRecord(collection)
	admin.Set("id", "admin-1")
	admin.Set("admin", true)

	state, err := BuildWorkflowState(nil, admin, record, data, false, testFlow())
	if err != nil {
		t.Fatalf("unexpected error for admin on guardian step: %v", err)
	}
	if !state.CanTakeAction {
		t.Fatalf("expected admin override to allow action")
	}
	if state.ActorIsAssigned {
		t.Fatalf("expected admin NOT personally assigned as guardian to have ActorIsAssigned=false")
	}
}

func TestBuildWorkflowStateAdminAlsoAssignedAsGuardian(t *testing.T) {
	record := testRequestRecord()
	record.Set("group", "group-1")
	record.Set("guardian", "admin-1")
	data := map[string]any{
		"guardian": map[string]any{
			"name":        "Admin",
			"assigned_at": "2026-04-09T10:00:00Z",
		},
	}
	record.Set("data", SetRequestFlowSnapshot(data, testFlow()))

	collection := core.NewBaseCollection("users")
	collection.Fields.Add(
		&core.TextField{Name: "email"},
		&core.BoolField{Name: "admin"},
		&core.JSONField{Name: "data"},
	)
	admin := core.NewRecord(collection)
	admin.Set("id", "admin-1")
	admin.Set("admin", true)

	state, err := BuildWorkflowState(nil, admin, record, data, false, testFlow())
	if err != nil {
		t.Fatalf("unexpected error for admin+guardian: %v", err)
	}
	if !state.CanTakeAction {
		t.Fatalf("expected admin+guardian to be allowed to take action")
	}
	if !state.ActorIsAssigned {
		t.Fatalf("expected admin also assigned as guardian to have ActorIsAssigned=true")
	}
}

func TestBuildWorkflowStateAllowsAdminOverride(t *testing.T) {
	record := testRequestRecord()
	data := map[string]any{}
	record.Set("data", SetRequestFlowSnapshot(data, testFlow()))

	admin := testUserRecord("admin0000000001", true)

	state, err := BuildWorkflowState(nil, admin, record, data, false, testFlow())
	if err != nil {
		t.Fatalf("unexpected error for admin: %v", err)
	}
	if !state.CanTakeAction {
		t.Fatalf("expected admin override to allow current step action")
	}
	if !state.CanArchive {
		t.Fatalf("expected admin to be allowed to archive")
	}
	if state.CurrentAction != ActionSetGroup {
		t.Fatalf("expected current action %q, got %q", ActionSetGroup, state.CurrentAction)
	}
	if !state.ActorIsAssigned {
		t.Fatalf("expected admin on admin-role step to have ActorIsAssigned=true")
	}

	if err := ApplyArchiveAction(nil, admin, record, data, "No fit"); err != nil {
		t.Fatalf("expected admin archive to remain allowed without flow lookup: %v", err)
	}
}

func TestAssistantCanArchiveOwnGroupRequest(t *testing.T) {
	app := testPocketBase(t)
	groups := testGroupsCollection(t, app)

	assistant := testUserRecord("assistant000001", false)
	other := testUserRecord("assistant000002", false)
	group := testGroupWithAssistant(t, app, groups, "group0000000001", assistant.Id)

	record := testRequestRecord()
	record.Set("group", group.Id)
	record.Set("guardian", "guardian0000001")
	data := map[string]any{
		"guardian": map[string]any{
			"name":        "Guardian",
			"assigned_at": "2026-04-09T10:00:00Z",
		},
		"mentoring": map[string]any{
			"notes":   []any{map[string]any{"text": "done", "at": "2026-04-09T10:30:00Z"}},
			"done_at": "2026-04-09T10:30:00Z",
		},
	}
	record.Set("data", SetRequestFlowSnapshot(data, testFlow()))

	state, err := BuildWorkflowState(app, assistant, record, data, false, testFlow())
	if err != nil {
		t.Fatalf("unexpected error for assigned assistant: %v", err)
	}
	if !state.CanArchive {
		t.Fatalf("expected assigned assistant to archive while group approval is pending")
	}
	if err := ApplyArchiveAction(app, assistant, record, data, "No fit"); err != nil {
		t.Fatalf("expected assigned assistant archive to succeed: %v", err)
	}

	record = testRequestRecord()
	record.Set("group", group.Id)
	record.Set("guardian", "guardian0000001")
	data = map[string]any{
		"guardian": map[string]any{
			"name":        "Guardian",
			"assigned_at": "2026-04-09T10:00:00Z",
		},
	}
	record.Set("data", SetRequestFlowSnapshot(data, testFlow()))

	state, err = BuildWorkflowState(app, assistant, record, data, false, testFlow())
	if err != nil {
		t.Fatalf("unexpected error before group approval step: %v", err)
	}
	if !state.CanArchive {
		t.Fatalf("expected assigned assistant archive to be allowed before group approval is pending")
	}
	if err := ApplyArchiveAction(app, assistant, record, data, "No fit"); err != nil {
		t.Fatalf("expected assigned assistant archive to succeed before group approval is pending: %v", err)
	}

	record = testRequestRecord()
	record.Set("group", group.Id)
	record.Set("guardian", "guardian0000001")
	data = map[string]any{
		"guardian": map[string]any{
			"name":        "Guardian",
			"assigned_at": "2026-04-09T10:00:00Z",
		},
		"mentoring": map[string]any{
			"notes":   []any{map[string]any{"text": "done", "at": "2026-04-09T10:30:00Z"}},
			"done_at": "2026-04-09T10:30:00Z",
		},
	}
	record.Set("data", SetRequestFlowSnapshot(data, testFlow()))

	state, err = BuildWorkflowState(app, other, record, data, false, testFlow())
	if err != nil {
		t.Fatalf("unexpected error for other assistant: %v", err)
	}
	if state.CanArchive {
		t.Fatalf("expected non-assigned assistant archive to be blocked")
	}
	if err := ApplyArchiveAction(app, other, record, data, "No fit"); err == nil {
		t.Fatalf("expected non-assigned assistant archive to fail")
	}
}

func TestGuardianCanArchiveAssignedRequest(t *testing.T) {
	record := testRequestRecord()
	record.Set("group", "group-1")
	record.Set("guardian", "guardian-1")
	data := map[string]any{
		"guardian": map[string]any{
			"name":        "Guardian",
			"assigned_at": "2026-04-09T10:00:00Z",
		},
	}
	record.Set("data", SetRequestFlowSnapshot(data, testFlow()))

	guardian := testUserRecord("guardian-1", false)

	state, err := BuildWorkflowState(nil, guardian, record, data, false, testFlow())
	if err != nil {
		t.Fatalf("unexpected error for assigned guardian: %v", err)
	}
	if !state.CanArchive {
		t.Fatalf("expected assigned guardian to be allowed to archive")
	}
	if err := ApplyArchiveAction(nil, guardian, record, data, "No fit"); err != nil {
		t.Fatalf("expected assigned guardian archive to succeed: %v", err)
	}

	unarchiveState, err := BuildWorkflowState(nil, guardian, record, data, true, testFlow())
	if err != nil {
		t.Fatalf("unexpected error for archived request: %v", err)
	}
	if !unarchiveState.CanUnarchive {
		t.Fatalf("expected assigned guardian to be allowed to unarchive")
	}
	if err := ApplyUnarchiveAction(nil, guardian, record, data); err != nil {
		t.Fatalf("expected assigned guardian unarchive to succeed: %v", err)
	}
}

func TestAssistantCannotTakeFinalAdminApproval(t *testing.T) {
	app := testPocketBase(t)
	groups := testGroupsCollection(t, app)

	assistant := testUserRecord("assistant000001", false)
	group := testGroupWithAssistant(t, app, groups, "group0000000001", assistant.Id)

	record := testRequestRecord()
	record.Set("group", group.Id)
	record.Set("guardian", "guardian0000001")
	data := map[string]any{
		"guardian": map[string]any{
			"name":        "Guardian",
			"assigned_at": "2026-04-09T10:00:00Z",
		},
		"mentoring": map[string]any{
			"notes":   []any{map[string]any{"text": "done", "at": "2026-04-09T10:30:00Z"}},
			"done_at": "2026-04-09T10:30:00Z",
		},
		"group_approved_at": "2026-04-09T11:00:00Z",
	}
	record.Set("data", SetRequestFlowSnapshot(data, testFlow()))

	state, err := BuildWorkflowState(app, assistant, record, data, false, testFlow())
	if err != nil {
		t.Fatalf("unexpected error for assistant on final approval step: %v", err)
	}
	if state.CanTakeAction {
		t.Fatalf("expected assistant not to take final admin approval")
	}
	if err := ApplySetAdminApprovedAction(app, assistant, record, data, ActionPayload{}); err == nil {
		t.Fatalf("expected assistant final admin approval to fail")
	}
	if _, err := ApplyPromoteAction(app, assistant, record, data); err == nil {
		t.Fatalf("expected assistant promote to fail")
	}
}

func TestBuildWorkflowStateArchivedRequestDisablesActions(t *testing.T) {
	record := testRequestRecord()
	record.Set("archived", true)
	data := map[string]any{
		"archived": map[string]any{
			"reason": "No fit",
		},
	}
	record.Set("data", SetRequestFlowSnapshot(data, testFlow()))

	collection := core.NewBaseCollection("users")
	collection.Fields.Add(
		&core.TextField{Name: "email"},
		&core.BoolField{Name: "admin"},
		&core.JSONField{Name: "data"},
	)
	admin := core.NewRecord(collection)
	admin.Set("id", "admin-1")
	admin.Set("admin", true)

	state, err := BuildWorkflowState(nil, admin, record, data, true, testFlow())
	if err != nil {
		t.Fatalf("unexpected error for archived request: %v", err)
	}
	if state.Status != StatusArchived {
		t.Fatalf("expected archived status, got %q", state.Status)
	}
	if state.CanTakeAction {
		t.Fatalf("expected archived request to block step actions")
	}
	if state.CanArchive {
		t.Fatalf("expected archived request to block re-archive action")
	}
	if !state.CanUnarchive {
		t.Fatalf("expected admin to be allowed to unarchive")
	}
	if state.ActorIsAssigned {
		t.Fatalf("expected archived request to have ActorIsAssigned=false")
	}
}

func TestApplyUnarchiveActionRestoresPriorFlowStatus(t *testing.T) {
	app := testPocketBase(t)
	groups := testGroupsCollection(t, app)

	assistant := testUserRecord("assistant000001", false)
	group := testGroupWithAssistant(t, app, groups, "group0000000001", assistant.Id)

	record := testRequestRecord()
	record.Set("group", group.Id)
	record.Set("guardian", "guardian0000001")
	data := map[string]any{
		"guardian": map[string]any{
			"name":        "Guardian",
			"assigned_at": "2026-04-09T10:00:00Z",
		},
		"mentoring": map[string]any{
			"notes":   []any{map[string]any{"text": "in progress", "at": "2026-04-09T10:30:00Z"}},
			"done_at": "2026-04-09T11:00:00Z",
		},
	}
	record.Set("data", SetRequestFlowSnapshot(data, testFlow()))

	statusBefore, err := StatusForRecord(app, record)
	if err != nil {
		t.Fatalf("unexpected error computing status before archive: %v", err)
	}
	if statusBefore != StatusMentoring {
		t.Fatalf("expected status before archive to be mentoring, got %q", statusBefore)
	}

	if err := ApplyArchiveAction(app, assistant, record, data, "Interrupted"); err != nil {
		t.Fatalf("expected assistant archive to succeed: %v", err)
	}

	statusArchived, err := StatusForRecord(app, record)
	if err != nil {
		t.Fatalf("unexpected error computing status while archived: %v", err)
	}
	if statusArchived != StatusArchived {
		t.Fatalf("expected archived status while archived, got %q", statusArchived)
	}

	if err := ApplyUnarchiveAction(app, assistant, record, data); err != nil {
		t.Fatalf("expected assistant unarchive to succeed: %v", err)
	}

	statusAfter, err := StatusForRecord(app, record)
	if err != nil {
		t.Fatalf("unexpected error computing status after unarchive: %v", err)
	}
	if statusAfter != StatusMentoring {
		t.Fatalf("expected status to be restored to mentoring after unarchive, got %q", statusAfter)
	}

	notes := mentoringNotes(data)
	if len(notes) != 1 {
		t.Fatalf("expected mentoring notes to survive archive/unarchive, got %d notes", len(notes))
	}
}
