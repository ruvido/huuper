package events

import (
	"encoding/json"

	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase"
)

type Config struct {
	Types      []TypeDef     `json:"types"`
	Recurrence RecurrenceDef `json:"recurrence"`
	List       ListDef       `json:"list"`
}

type TypeDef struct {
	Value                string `json:"value"`
	Creator              string `json:"creator"`
	RequiresGroup        bool   `json:"requires_group"`
	RegistrationApproval bool   `json:"registration_approval"`
	RegistrationScope    string `json:"registration_scope"`
	RequiresTitle        bool   `json:"requires_title"`
}

type RecurrenceDef struct {
	MaxOccurrences int             `json:"max_occurrences"`
	Rules          []RecurrenceRule `json:"rules"`
}

type RecurrenceRule struct {
	Type string `json:"type"`
}

type ListDef struct {
	DefaultWindow  string `json:"default_window"`
	CollapseSeries bool   `json:"collapse_series"`
}

func LoadConfig(app *pocketbase.PocketBase) (*Config, error) {
	raw, err := backendinternal.FindSettingData(app, "eventflow")
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
	return &cfg, nil
}

func (c *Config) TypeDef(value string) (TypeDef, bool) {
	for _, t := range c.Types {
		if t.Value == value {
			return t, true
		}
	}
	return TypeDef{}, false
}

func (c *Config) IsValidRecurrenceRule(value string) bool {
	for _, r := range c.Recurrence.Rules {
		if r.Type == value {
			return true
		}
	}
	return false
}

func (c *Config) MaxOccurrences() int {
	if c.Recurrence.MaxOccurrences <= 0 {
		return 52
	}
	return c.Recurrence.MaxOccurrences
}
