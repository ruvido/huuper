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
			{Key: "email", Type: "email", Required: true},
			{Key: "full_name", Type: "text", Required: true},
			{Key: "motivation", Type: "textarea", Required: true},
			{Key: "region", Type: "select"},
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
			{Key: "email", Type: "email", Required: true},
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

func TestValidateAndBuildDataRejectsSelectOptionOutsideSchema(t *testing.T) {
	signup := SignupSettingsConfig{
		Steps: []SignupFieldConfig{
			{Field: "email"},
			{Field: "region"},
		},
	}
	profile := ProfileSchemaConfig{
		Fields: []ProfileFieldConfig{
			{Key: "email", Type: "email", Required: true},
			{Key: "region", Type: "select", Required: true, Options: []string{"Lazio", "Lombardia"}},
		},
	}

	_, _, err := ValidateAndBuildData(map[string]any{
		"email":  "candidate@example.com",
		"region": "Mars",
	}, signup, profile)
	if err == nil {
		t.Fatalf("expected validation error for invalid select option")
	}
}

func TestValidateAndBuildDataAcceptsCustomSelectOptionWhenSchemaAllowsInput(t *testing.T) {
	signup := SignupSettingsConfig{
		Steps: []SignupFieldConfig{
			{Field: "email"},
			{Field: "region"},
		},
	}
	profile := ProfileSchemaConfig{
		Fields: []ProfileFieldConfig{
			{Key: "email", Type: "email", Required: true},
			{Key: "region", Type: "select", Required: true, Options: []string{"Lazio", "Estero:input"}},
		},
	}

	data, _, err := ValidateAndBuildData(map[string]any{
		"email":  "candidate@example.com",
		"region": "Svizzera",
	}, signup, profile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := data["region"]; got != "Svizzera" {
		t.Fatalf("unexpected region value: %#v", got)
	}
}

func TestValidateAndBuildDataRejectsNonStringTextField(t *testing.T) {
	signup := SignupSettingsConfig{
		Steps: []SignupFieldConfig{
			{Field: "email"},
			{Field: "full_name"},
		},
	}
	profile := ProfileSchemaConfig{
		Fields: []ProfileFieldConfig{
			{Key: "email", Type: "email", Required: true},
			{Key: "full_name", Type: "text", Required: true},
		},
	}

	_, _, err := ValidateAndBuildData(map[string]any{
		"email":     "candidate@example.com",
		"full_name": 42,
	}, signup, profile)
	if err == nil {
		t.Fatalf("expected validation error for non-string text field")
	}
}
