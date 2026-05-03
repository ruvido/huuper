package battleplans

import (
	"strings"
	"testing"
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
		{"unknown", "unknown", true}, // identity short-circuits
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
