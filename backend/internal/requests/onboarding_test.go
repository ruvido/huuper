package requests

import (
	"reflect"
	"testing"

	"github.com/pocketbase/pocketbase/core"
)

func TestMissingOnboardingFieldsAcceptsPresentRecordField(t *testing.T) {
	collection := core.NewBaseCollection("users")
	collection.Fields.Add(&core.FileField{Name: "photo"})
	user := core.NewRecord(collection)
	user.Set("photo", "profile.png")
	settings := OnboardingSettingsConfig{
		Steps: []OnboardingStepConfig{
			{Field: "work"},
			{Field: "skills"},
			{Field: "photo"},
		},
	}
	data := map[string]any{
		"work":   "Falegname",
		"skills": []any{"legno"},
	}

	missing := MissingOnboardingFields(data, settings, user)

	if len(missing) != 0 {
		t.Fatalf("expected no missing fields, got %v", missing)
	}
}

func TestMissingOnboardingFieldsRequiresAbsentRecordField(t *testing.T) {
	collection := core.NewBaseCollection("users")
	collection.Fields.Add(&core.FileField{Name: "photo"})
	user := core.NewRecord(collection)
	settings := OnboardingSettingsConfig{
		Steps: []OnboardingStepConfig{
			{Field: "work"},
			{Field: "photo"},
		},
	}
	data := map[string]any{
		"work": "Falegname",
	}

	missing := MissingOnboardingFields(data, settings, user)

	if !reflect.DeepEqual(missing, []string{"photo"}) {
		t.Fatalf("missing mismatch: got %v", missing)
	}
}
