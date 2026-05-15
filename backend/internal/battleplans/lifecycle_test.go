package battleplans

import (
	"errors"
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func TestCanTransition(t *testing.T) {
	cases := []struct {
		from string
		to   string
		want bool
	}{
		// identity is always allowed (idempotent re-application)
		{StatusActive, StatusActive, true},
		{StatusDraft, StatusDraft, true},
		{StatusCompleted, StatusCompleted, true},
		{StatusArchived, StatusArchived, true},

		// active → ...
		{StatusActive, StatusCompleted, true},
		{StatusActive, StatusArchived, true},
		{StatusActive, StatusDraft, true},

		// draft → ...
		{StatusDraft, StatusActive, true},
		{StatusDraft, StatusArchived, true},
		{StatusDraft, StatusCompleted, false},

		// completed → ... (terminal except archive)
		{StatusCompleted, StatusActive, false},
		{StatusCompleted, StatusDraft, false},
		{StatusCompleted, StatusArchived, true},

		// archived → ... (fully terminal)
		{StatusArchived, StatusActive, false},
		{StatusArchived, StatusDraft, false},
		{StatusArchived, StatusCompleted, false},

		// unknown source statuses are never transitional
		{"", StatusActive, false},
		{"unknown", StatusActive, false},
		{"unknown", "unknown", false},
	}
	for _, tc := range cases {
		got := CanTransition(tc.from, tc.to)
		if got != tc.want {
			t.Errorf("CanTransition(%q, %q) = %v, want %v", tc.from, tc.to, got, tc.want)
		}
	}
}

func TestIsEditable(t *testing.T) {
	cases := []struct {
		status string
		want   bool
	}{
		{StatusActive, true},
		{StatusDraft, true},
		{StatusCompleted, false},
		{StatusArchived, false},
		{"", false},
		{"unknown", false},
	}
	for _, tc := range cases {
		if got := IsEditable(tc.status); got != tc.want {
			t.Errorf("IsEditable(%q) = %v, want %v", tc.status, got, tc.want)
		}
	}
}

func TestCreateRequiresUser(t *testing.T) {
	// Create checks userID before touching the app, so a nil app is acceptable
	// for this branch — we never reach LoadConfig.
	_, err := Create(nil, "", Input{})
	if err == nil {
		t.Fatal("Create with empty userID should return error")
	}
	if !strings.Contains(err.Error(), "missing user") {
		t.Errorf("Create error = %q, want error containing %q", err.Error(), "missing user")
	}

	_, err = Create(nil, "   ", Input{})
	if err == nil {
		t.Fatal("Create with whitespace-only userID should return error")
	}
	if !strings.Contains(err.Error(), "missing user") {
		t.Errorf("Create error = %q, want error containing %q", err.Error(), "missing user")
	}
}

func TestDeleteNilRecord(t *testing.T) {
	// Delete checks the record before touching the app, so nil app is fine.
	if err := Delete(nil, nil); err == nil {
		t.Fatal("Delete with nil record should return error")
	} else if !strings.Contains(err.Error(), "missing battleplan") {
		t.Errorf("Delete error = %q, want error containing %q", err.Error(), "missing battleplan")
	}
	// Delete with a valid record requires a real DB (app.Delete) so we skip
	// that path here — covered by integration tests at the API layer.
}

func TestIsValidStatus(t *testing.T) {
	cases := []struct {
		status string
		want   bool
	}{
		{StatusActive, true},
		{StatusDraft, true},
		{StatusCompleted, true},
		{StatusArchived, true},
		{"", false},
		{"invalid", false},
	}
	for _, tc := range cases {
		if got := IsValidStatus(tc.status); got != tc.want {
			t.Errorf("IsValidStatus(%q) = %v, want %v", tc.status, got, tc.want)
		}
	}
}

func TestIsValidInitialStatus(t *testing.T) {
	cases := []struct {
		status string
		want   bool
	}{
		{StatusActive, true},
		{StatusDraft, true},
		{StatusCompleted, false},
		{StatusArchived, false},
		{"", false},
		{"invalid", false},
	}
	for _, tc := range cases {
		if got := IsValidInitialStatus(tc.status); got != tc.want {
			t.Errorf("IsValidInitialStatus(%q) = %v, want %v", tc.status, got, tc.want)
		}
	}
}

func TestStampNewRoutinesPreservesExistingMetadata(t *testing.T) {
	existing := Data{Pillars: map[string]Pillar{
		"body": {
			Routines: []Routine{{
				ID:      "routine-1",
				Title:   "Old title",
				Trigger: "Old trigger",
				Cadence: Cadence{Type: CadenceDaily},
				Created: "2026-01-01 00:00:00.000Z",
			}},
		},
	}}
	incoming := Data{
		Priority: Priority{Title: "Priority"},
		Pillars: map[string]Pillar{
			"body": {
				Routines: []Routine{{
					ID:      "routine-1",
					Title:   "New title",
					Trigger: "New trigger",
					Cadence: Cadence{Type: CadenceTimesPerWeek, Times: 3},
				}},
			},
		},
	}

	got := stampNewRoutines(existing, incoming, "2026-02-01 00:00:00.000Z")
	routines := got.Pillars["body"].Routines
	if len(routines) != 1 {
		t.Fatalf("len(routines) = %d, want 1", len(routines))
	}
	if routines[0].ID != "routine-1" {
		t.Fatalf("routine ID = %q, want %q", routines[0].ID, "routine-1")
	}
	if routines[0].Created != "2026-01-01 00:00:00.000Z" {
		t.Fatalf("routine Created = %q, want existing timestamp", routines[0].Created)
	}
	if routines[0].Title != "New title" || routines[0].Trigger != "New trigger" {
		t.Fatalf("routine content was not updated: %#v", routines[0])
	}
}

func TestStampNewRoutinesAssignsMetadataToNewRoutine(t *testing.T) {
	existing := Data{Pillars: map[string]Pillar{
		"body": {
			Routines: []Routine{{
				ID:      "routine-1",
				Created: "2026-01-01 00:00:00.000Z",
			}},
		},
	}}
	incoming := Data{Pillars: map[string]Pillar{
		"body": {
			Routines: []Routine{{
				ID:      "unknown-routine",
				Title:   "New",
				Trigger: "After waking",
				Cadence: Cadence{Type: CadenceDaily},
			}},
		},
	}}

	got := stampNewRoutines(existing, incoming, "2026-02-01 00:00:00.000Z")
	routine := got.Pillars["body"].Routines[0]
	if routine.ID == "" || routine.ID == "unknown-routine" {
		t.Fatalf("routine ID = %q, want generated ID", routine.ID)
	}
	if routine.Created != "2026-02-01 00:00:00.000Z" {
		t.Fatalf("routine Created = %q, want new timestamp", routine.Created)
	}
}

func TestCreateRejectsTerminalInitialStatus(t *testing.T) {
	app, userID := testBattleplanApp(t)
	input := validBattleplanInput()
	input.Status = StatusCompleted

	_, err := Create(app, userID, input)
	if err == nil {
		t.Fatal("Create with completed initial status should fail")
	}
	if !strings.Contains(err.Error(), "invalid initial status") {
		t.Fatalf("Create error = %q, want invalid initial status", err.Error())
	}
}

func TestCreateDetectsActiveAndDraftCollisions(t *testing.T) {
	app, userID := testBattleplanApp(t)

	if _, err := Create(app, userID, validBattleplanInput()); err != nil {
		t.Fatalf("create active: %v", err)
	}
	_, err := Create(app, userID, validBattleplanInput())
	assertStatusCollision(t, err, StatusActive)

	draft := validBattleplanInput()
	draft.Status = StatusDraft
	if _, err := Create(app, userID, draft); err != nil {
		t.Fatalf("create draft: %v", err)
	}
	_, err = Create(app, userID, draft)
	assertStatusCollision(t, err, StatusDraft)
}

func TestActivateReplacesExistingActive(t *testing.T) {
	app, userID := testBattleplanApp(t)

	active, err := Create(app, userID, validBattleplanInput())
	if err != nil {
		t.Fatalf("create active: %v", err)
	}
	draftInput := validBattleplanInput()
	draftInput.Status = StatusDraft
	draft, err := Create(app, userID, draftInput)
	if err != nil {
		t.Fatalf("create draft: %v", err)
	}

	if err := Activate(app, draft); err != nil {
		t.Fatalf("Activate returned error: %v", err)
	}

	refreshedActive, err := app.FindRecordById("battleplans", active.Id)
	if err != nil {
		t.Fatalf("reload active: %v", err)
	}
	refreshedDraft, err := app.FindRecordById("battleplans", draft.Id)
	if err != nil {
		t.Fatalf("reload draft: %v", err)
	}
	if got := refreshedActive.GetString("status"); got != StatusArchived {
		t.Fatalf("old active status = %q, want %q", got, StatusArchived)
	}
	if got := refreshedDraft.GetString("status"); got != StatusActive {
		t.Fatalf("draft status = %q, want %q", got, StatusActive)
	}
}

func TestActivateRejectsArchivedBattleplan(t *testing.T) {
	app, userID := testBattleplanApp(t)
	input := validBattleplanInput()
	input.Status = StatusDraft
	record, err := Create(app, userID, input)
	if err != nil {
		t.Fatalf("create draft: %v", err)
	}
	if err := SetStatus(app, record, StatusArchived); err != nil {
		t.Fatalf("archive draft: %v", err)
	}

	if err := Activate(app, record); err == nil {
		t.Fatal("Activate archived battleplan should fail")
	}
}

func assertStatusCollision(t *testing.T, err error, status string) {
	t.Helper()
	var collision StatusCollisionError
	if !errors.As(err, &collision) {
		t.Fatalf("error = %#v, want StatusCollisionError", err)
	}
	if collision.Status != status {
		t.Fatalf("collision status = %q, want %q", collision.Status, status)
	}
	if collision.ExistingID == "" {
		t.Fatal("collision ExistingID should be set")
	}
}

func validBattleplanInput() Input {
	return Input{
		StartDate:    "2026-05-01",
		DurationDays: 30,
		Visibility:   "group",
		Data: Data{
			Priority: Priority{Title: "Priority", Why: "Because"},
			Pillars: map[string]Pillar{
				"body": {
					Objective: "Train",
					Routines: []Routine{{
						Title:   "Walk",
						Trigger: "After lunch",
						Cadence: Cadence{Type: CadenceDaily},
					}},
				},
			},
		},
	}
}

func testBattleplanApp(t *testing.T) (*pocketbase.PocketBase, string) {
	t.Helper()
	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir:       t.TempDir(),
		DefaultEncryptionEnv: "pb_test_env",
	})
	if err := app.Bootstrap(); err != nil {
		t.Fatalf("bootstrap test app: %v", err)
	}

	users, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		t.Fatalf("find users collection: %v", err)
	}
	user := core.NewRecord(users)
	user.Set("email", "member@example.com")
	user.SetPassword("test-password-123")
	if err := app.Save(user); err != nil {
		t.Fatalf("save user: %v", err)
	}

	settings := core.NewBaseCollection("settings")
	settings.Fields.Add(
		&core.TextField{Name: "name", Required: true},
		&core.JSONField{Name: "data", Required: true},
	)
	if err := app.Save(settings); err != nil {
		t.Fatalf("save settings collection: %v", err)
	}
	battleplanSettings := core.NewRecord(settings)
	battleplanSettings.Set("name", "battleplan")
	battleplanSettings.Set("data", validBattleplanConfigData())
	if err := app.Save(battleplanSettings); err != nil {
		t.Fatalf("save battleplan settings: %v", err)
	}

	battleplans := core.NewBaseCollection("battleplans")
	battleplans.Fields.Add(
		&core.AutodateField{Name: "created", OnCreate: true},
		&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
		&core.RelationField{
			Name:          "user",
			Required:      true,
			CollectionId:  users.Id,
			MaxSelect:     1,
			CascadeDelete: true,
		},
		&core.DateField{Name: "start_date", Required: true},
		&core.DateField{Name: "end_date", Required: true},
		&core.SelectField{
			Name:      "status",
			Required:  true,
			MaxSelect: 1,
			Values:    []string{StatusActive, StatusCompleted, StatusArchived, StatusDraft},
		},
		&core.SelectField{
			Name:      "visibility",
			Required:  true,
			MaxSelect: 1,
			Values:    []string{"group", "public"},
		},
		&core.JSONField{Name: "data", Required: true},
	)
	battleplans.AddIndex("idx_battleplans_user_active", true, "user", "status = 'active'")
	battleplans.AddIndex("idx_battleplans_user_draft", true, "user", "status = 'draft'")
	if err := app.Save(battleplans); err != nil {
		t.Fatalf("save battleplans collection: %v", err)
	}

	return app, user.Id
}

func validBattleplanConfigData() map[string]any {
	return map[string]any{
		"priority": map[string]any{
			"new":  map[string]any{"title": "New", "text": "New text"},
			"edit": map[string]any{"title": "Edit", "text": "Edit text"},
		},
		"durations": []map[string]any{
			{"value": 30, "default": true},
			{"value": 60},
		},
		"pillars": []map[string]any{
			{"key": "body", "label": "Body", "description": ""},
		},
		"cadences": []map[string]any{
			{"type": CadenceDaily, "label": "Daily", "default": true},
			{"type": CadenceSpecificDays, "label": "Days"},
			{"type": CadenceTimesPerWeek, "label": "Times"},
			{"type": CadencePaused, "label": "Paused"},
		},
		"visibility": []map[string]any{
			{"value": "group", "label": "Group", "default": true},
			{"value": "public", "label": "Public"},
		},
		"wizard": map[string]any{},
	}
}
