package requests

import "testing"

func TestValidateAndBuildDataNormalizesEmailAndFiltersFields(t *testing.T) {
	signup := SignupSettingsConfig{
		Steps: []SignupFieldConfig{
			{Field: "email"},
			{Field: "full_name"},
			{Field: "mobile"},
			{Field: "motivation"},
		},
	}
	profile := ProfileSchemaConfig{
		Fields: []ProfileFieldConfig{
			{Key: "email", Type: "email", Required: true},
			{Key: "full_name", Type: "text", Required: true},
			{Key: "mobile", Type: "text", Required: true},
			{Key: "motivation", Type: "textarea", Required: true},
			{Key: "region", Type: "select"},
		},
	}

	data, email, err := ValidateAndBuildData(map[string]any{
		"email":      " Candidate@Example.com ",
		"full_name":  "Mario Rossi",
		"mobile":     "+39 5645567",
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
	if got := data["mobile"]; got != "+39 5645567" {
		t.Fatalf("unexpected mobile: %#v", got)
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

func TestValidateAndBuildDataRejectsInvalidEmail(t *testing.T) {
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
		"email": "not-an-email",
	}, signup, profile)
	if err == nil {
		t.Fatalf("expected validation error for invalid email")
	}
	if got := err.Error(); got != "invalid field email: Enter a valid email address." {
		t.Fatalf("unexpected error message: %q", got)
	}
}

func TestValidateAndBuildDataRejectsInvalidPhone(t *testing.T) {
	signup := SignupSettingsConfig{
		Steps: []SignupFieldConfig{
			{Field: "email"},
			{Field: "mobile"},
		},
	}
	profile := ProfileSchemaConfig{
		Fields: []ProfileFieldConfig{
			{Key: "email", Type: "email", Required: true},
			{Key: "mobile", Type: "text", Required: true},
		},
	}

	_, _, err := ValidateAndBuildData(map[string]any{
		"email":  "candidate@example.com",
		"mobile": "+39 1234",
	}, signup, profile)
	if err == nil {
		t.Fatalf("expected validation error for invalid phone")
	}
	if got := err.Error(); got != "invalid field mobile: Enter a valid phone number." {
		t.Fatalf("unexpected error message: %q", got)
	}
}

func TestValidateAndBuildDataRejectsSingleWordFullName(t *testing.T) {
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
		"full_name": "Mario",
	}, signup, profile)
	if err == nil {
		t.Fatalf("expected validation error for single-word full name")
	}
	if got := err.Error(); got != "Both First and Last name are required" {
		t.Fatalf("unexpected error message: %q", got)
	}
}

func TestValidateAndBuildDataNormalizesLowercaseSurname(t *testing.T) {
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

	data, _, err := ValidateAndBuildData(map[string]any{
		"email":     "candidate@example.com",
		"full_name": "Mario rossi",
	}, signup, profile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := data["full_name"]; got != "Mario Rossi" {
		t.Fatalf("unexpected normalized full_name: %#v", got)
	}
}

func TestValidateAndBuildDataAcceptsCapitalizedFullName(t *testing.T) {
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

	data, _, err := ValidateAndBuildData(map[string]any{
		"email":     "candidate@example.com",
		"full_name": "Mario Rossi",
	}, signup, profile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := data["full_name"]; got != "Mario Rossi" {
		t.Fatalf("unexpected full_name: %#v", got)
	}
}
