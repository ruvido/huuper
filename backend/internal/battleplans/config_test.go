package battleplans

import "testing"

func TestConfigValidateRejectsMissingRequiredSections(t *testing.T) {
	cases := []struct {
		name string
		cfg  Config
	}{
		{name: "durations", cfg: Config{}},
		{name: "visibility", cfg: Config{Durations: []DurationDef{{Value: 30}}}},
		{name: "pillars", cfg: Config{
			Durations:  []DurationDef{{Value: 30}},
			Visibility: []VisibilityDef{{Value: "group"}},
		}},
		{name: "cadences", cfg: Config{
			Durations:  []DurationDef{{Value: 30}},
			Visibility: []VisibilityDef{{Value: "group"}},
			Pillars:    []PillarDef{{Key: "body"}},
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.cfg.Validate(); err == nil {
				t.Fatal("Validate should fail")
			}
		})
	}
}

func TestConfigValidateAcceptsMinimalConfig(t *testing.T) {
	cfg := Config{
		Durations:  []DurationDef{{Value: 30}},
		Visibility: []VisibilityDef{{Value: "group"}},
		Pillars:    []PillarDef{{Key: "body"}},
		Cadences:   []CadenceDef{{Type: CadenceDaily}},
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
}
