package requests

import "testing"

func TestValidateAndBuildDataNormalizesEmailAndFiltersFields(t *testing.T) {
	signup := SignupSettingsConfig{
		Steps: []SignupFieldConfig{
			{Field: "email"},
			{Field: "full_name"},
			{Field: "motivation"},
		},
	}
	profile := ProfileSchemaConfig{
		Fields: []ProfileFieldConfig{
			{Key: "email"},
			{Key: "full_name"},
			{Key: "motivation"},
			{Key: "region"},
		},
	}

	data, email, err := ValidateAndBuildData(map[string]any{
		"email":      " Candidate@Example.com ",
		"full_name":  "Mario Rossi",
		"motivation": "Test",
	}, signup, profile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if email != "candidate@example.com" {
		t.Fatalf("expected normalized email, got %q", email)
	}
	if _, ok := data["email"]; ok {
		t.Fatalf("email should not remain in request data")
	}
	if got := data["full_name"]; got != "Mario Rossi" {
		t.Fatalf("unexpected full_name: %#v", got)
	}
}

func TestValidateAndBuildDataRejectsUnknownFields(t *testing.T) {
	signup := SignupSettingsConfig{
		Steps: []SignupFieldConfig{
			{Field: "email"},
		},
	}
	profile := ProfileSchemaConfig{
		Fields: []ProfileFieldConfig{
			{Key: "email"},
		},
	}

	_, _, err := ValidateAndBuildData(map[string]any{
		"email": "candidate@example.com",
		"extra": "nope",
	}, signup, profile)
	if err == nil {
		t.Fatalf("expected validation error for unknown field")
	}
}
