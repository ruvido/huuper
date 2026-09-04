package events

import (
	"encoding/json"
	"log"
	"strings"

	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase"
)

type Config struct {
	Types []TypeDef `json:"types"`
}

type TypeDef struct {
	Value        string          `json:"value"`
	Label        string          `json:"label"`
	Description  string          `json:"description"`
	Creators     []string        `json:"creators"`
	Required     map[string]bool `json:"required"`
	Registration RegistrationDef `json:"registration"`
	// Legacy fields are read only to tolerate old settings during migration.
	Creator       string `json:"creator,omitempty"`
	RequiresGroup bool   `json:"requires_group,omitempty"`
	RequiresTitle bool   `json:"requires_title,omitempty"`
}

type RegistrationDef struct {
	Enabled      bool `json:"enabled"`
	Approval     bool `json:"approval"`
	DepositCents int  `json:"deposit_cents"`
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
	value = strings.TrimSpace(value)
	for _, t := range c.Types {
		if t.Value == value {
			t.normalize()
			return t, true
		}
	}
	return TypeDef{}, false
}

func (t *TypeDef) normalize() {
	if t.Required == nil {
		t.Required = map[string]bool{}
	}
	if t.Creator != "" && len(t.Creators) == 0 {
		switch strings.TrimSpace(t.Creator) {
		case "admin_or_assistant":
			t.Creators = []string{"admin", "assistant"}
		case "admin", "assistant":
			t.Creators = []string{strings.TrimSpace(t.Creator)}
		}
	}
	if t.RequiresGroup {
		t.Required["group"] = true
	}
	if t.RequiresTitle {
		t.Required["title"] = true
	}
}

func (t TypeDef) AllowsCreator(role string) bool {
	role = strings.TrimSpace(role)
	for _, creator := range t.Creators {
		if strings.TrimSpace(creator) == role {
			return true
		}
	}
	return false
}

func (t TypeDef) Requires(field string) bool {
	field = strings.TrimSpace(field)
	for configured, required := range t.Required {
		configured = strings.TrimSpace(configured)
		if !isKnownRequiredField(configured) {
			log.Printf("eventflow: unknown required field %q ignored", configured)
			continue
		}
		if configured == field {
			return required
		}
	}
	return false
}

func isKnownRequiredField(field string) bool {
	switch field {
	case "title", "url", "location", "group", "end_date", "time", "description":
		return true
	default:
		return false
	}
}
