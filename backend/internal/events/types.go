package events

import (
	"strconv"

	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase/core"
)

// eventCount reads the count field as int. It's stored as a SelectField string
// (one of "1"/"3"/"6"/"12") since PB has no enum-of-numbers type. Returns 1
// when the value is missing or unparseable so render code can rely on count>=1.
func eventCount(record *core.Record) int {
	if record == nil {
		return 1
	}
	raw := record.GetString("count")
	if n, err := strconv.Atoi(raw); err == nil && n >= 1 {
		return n
	}
	return 1
}

type Item struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	TypeLabel    string         `json:"type_label,omitempty"`
	Title        string         `json:"title"`
	Slug         string         `json:"slug,omitempty"`
	StartDate    string         `json:"start_date"`
	EndDate      string         `json:"end_date,omitempty"`
	Cadence      string         `json:"cadence"`
	Count        int            `json:"count"`
	Location     string         `json:"location,omitempty"`
	URL          string         `json:"url,omitempty"`
	Group        string         `json:"group,omitempty"`
	GroupName    string         `json:"group_name,omitempty"`
	Registration bool           `json:"registration"`
	CreatedBy    string         `json:"created_by,omitempty"`
	Active       bool           `json:"active"`
	Data         map[string]any `json:"data"`
	Created      string         `json:"created"`
	Updated      string         `json:"updated"`
}

func ApplyTypeConfig(item *Item, cfg *Config) {
	if item == nil || cfg == nil {
		return
	}
	if typeDef, ok := cfg.TypeDef(item.Type); ok {
		item.TypeLabel = typeDef.Label
		item.Registration = typeDef.Registration.Enabled
	}
}

func MapItem(record *core.Record) Item {
	endDate := ""
	if !record.GetDateTime("end_date").IsZero() {
		endDate = record.GetDateTime("end_date").String()
	}
	count := eventCount(record)
	cadence := record.GetString("cadence")
	if cadence == "" {
		cadence = CadenceOnce
	}
	return Item{
		ID:        record.Id,
		Type:      record.GetString("type"),
		Title:     record.GetString("title"),
		Slug:      record.GetString("slug"),
		StartDate: record.GetDateTime("event_date").String(),
		EndDate:   endDate,
		Cadence:   cadence,
		Count:     count,
		Location:  record.GetString("location"),
		URL:       record.GetString("url"),
		Group:     record.GetString("group"),
		CreatedBy: record.GetString("created_by"),
		Active:    record.GetBool("active"),
		Data:      backendinternal.ParseJSONMap(record.Get("data")),
		Created:   record.GetDateTime("created").String(),
		Updated:   record.GetDateTime("updated").String(),
	}
}
