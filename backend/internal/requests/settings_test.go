package requests

import "testing"

func TestMergeUserDataKeepsExistingAndAppliesUpdates(t *testing.T) {
	base := map[string]any{
		"full_name": "Mario Rossi",
		"city":      "Roma",
		"tags":      []any{"a"},
	}
	updates := map[string]any{
		"city": "Milano",
		"bio":  "Ciao",
	}

	merged := MergeUserData(base, updates)

	if got := merged["full_name"]; got != "Mario Rossi" {
		t.Fatalf("full_name mismatch: got %v", got)
	}
	if got := merged["city"]; got != "Milano" {
		t.Fatalf("city mismatch: got %v", got)
	}
	if got := merged["bio"]; got != "Ciao" {
		t.Fatalf("bio mismatch: got %v", got)
	}
}

func TestMergeUserDataFiltersInternalKeys(t *testing.T) {
	base := map[string]any{
		"__private": "x",
		"name":      "ok",
	}
	updates := map[string]any{
		"__tmp": "y",
	}

	merged := MergeUserData(base, updates)

	if _, ok := merged["__private"]; ok {
		t.Fatalf("expected __private removed")
	}
	if _, ok := merged["__tmp"]; ok {
		t.Fatalf("expected __tmp removed")
	}
	if got := merged["name"]; got != "ok" {
		t.Fatalf("name mismatch: got %v", got)
	}
}
