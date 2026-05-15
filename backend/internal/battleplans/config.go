package battleplans

import (
	"encoding/json"
	"fmt"
	"strings"

	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase/core"
)

type Config struct {
	Priority   PriorityDef     `json:"priority"`
	Durations  []DurationDef   `json:"durations"`
	Pillars    []PillarDef     `json:"pillars"`
	Cadences   []CadenceDef    `json:"cadences"`
	Visibility []VisibilityDef `json:"visibility"`
	Wizard     map[string]any  `json:"wizard"`
}

type PriorityDef struct {
	New  PriorityCopy `json:"new"`
	Edit PriorityCopy `json:"edit"`
}

type PriorityCopy struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type DurationDef struct {
	Value   int  `json:"value"`
	Default bool `json:"default,omitempty"`
}

type PillarDef struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type CadenceDef struct {
	Type    string `json:"type"`
	Label   string `json:"label"`
	Default bool   `json:"default,omitempty"`
}

type VisibilityDef struct {
	Value   string `json:"value"`
	Label   string `json:"label"`
	Default bool   `json:"default,omitempty"`
}

func LoadConfig(app core.App) (*Config, error) {
	raw, err := backendinternal.FindSettingData(app, "battleplan")
	if err != nil {
		return nil, err
	}
	bytes, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(bytes, &cfg); err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("battleplan config missing")
	}
	if len(c.Durations) == 0 {
		return fmt.Errorf("battleplan config missing durations")
	}
	for _, item := range c.Durations {
		if item.Value <= 0 {
			return fmt.Errorf("battleplan config invalid duration")
		}
	}
	if len(c.Visibility) == 0 {
		return fmt.Errorf("battleplan config missing visibility")
	}
	seenVisibility := map[string]struct{}{}
	for _, item := range c.Visibility {
		value := strings.TrimSpace(item.Value)
		if value == "" {
			return fmt.Errorf("battleplan config invalid visibility")
		}
		if _, ok := seenVisibility[value]; ok {
			return fmt.Errorf("battleplan config duplicate visibility %q", value)
		}
		seenVisibility[value] = struct{}{}
	}
	if len(c.Pillars) == 0 {
		return fmt.Errorf("battleplan config missing pillars")
	}
	seenPillars := map[string]struct{}{}
	for _, item := range c.Pillars {
		key := strings.TrimSpace(item.Key)
		if key == "" {
			return fmt.Errorf("battleplan config invalid pillar")
		}
		if _, ok := seenPillars[key]; ok {
			return fmt.Errorf("battleplan config duplicate pillar %q", key)
		}
		seenPillars[key] = struct{}{}
	}
	if len(c.Cadences) == 0 {
		return fmt.Errorf("battleplan config missing cadences")
	}
	for _, item := range c.Cadences {
		if !IsValidCadence(Cadence{Type: item.Type, Days: []string{"mon"}, Times: 1}) {
			return fmt.Errorf("battleplan config invalid cadence %q", item.Type)
		}
	}
	return nil
}

func (c *Config) IsValidPillarKey(key string) bool {
	for _, p := range c.Pillars {
		if p.Key == key {
			return true
		}
	}
	return false
}

func (c *Config) IsValidDuration(d int) bool {
	for _, item := range c.Durations {
		if item.Value == d {
			return true
		}
	}
	return false
}

func (c *Config) IsValidVisibility(v string) bool {
	for _, item := range c.Visibility {
		if item.Value == v {
			return true
		}
	}
	return false
}
