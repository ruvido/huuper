package battleplans

import (
	"encoding/json"
	"slices"

	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase/core"
)

const (
	StatusActive    = "active"
	StatusCompleted = "completed"
	StatusArchived  = "archived"
	StatusDraft     = "draft"

	CadenceDaily        = "daily"
	CadencePaused       = "paused"
	CadenceSpecificDays = "specific_days"
	CadenceTimesPerWeek = "times_per_week"
)

var weekDays = []string{"mon", "tue", "wed", "thu", "fri", "sat", "sun"}

type Routine struct {
	ID      string  `json:"id"`
	Title   string  `json:"title"`
	Trigger string  `json:"trigger"`
	Cadence Cadence `json:"cadence"`
	Created string  `json:"created"`
}

type Cadence struct {
	Type  string   `json:"type"`
	Days  []string `json:"days,omitempty"`
	Times int      `json:"times,omitempty"`
}

type Pillar struct {
	Objective string    `json:"objective"`
	Routines  []Routine `json:"routines"`
}

type Priority struct {
	Title string `json:"title"`
	Why   string `json:"why,omitempty"`
}

type Note struct {
	Text string `json:"text"`
	At   string `json:"at"`
	By   string `json:"by,omitempty"`
}

type Data struct {
	Priority Priority          `json:"priority"`
	Pillars  map[string]Pillar `json:"pillars"`
	Notes    []Note            `json:"notes,omitempty"`
}

type ListItem struct {
	ID         string `json:"id"`
	User       string `json:"user"`
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
	Status     string `json:"status"`
	Visibility string `json:"visibility"`
	Data       Data   `json:"data"`
	Created    string `json:"created"`
	Updated    string `json:"updated"`
}

func IsValidStatus(s string) bool {
	return s == StatusActive || s == StatusCompleted || s == StatusArchived || s == StatusDraft
}

func IsValidCadence(c Cadence) bool {
	switch c.Type {
	case CadencePaused:
		return true
	case CadenceDaily:
		return true
	case CadenceSpecificDays:
		if len(c.Days) == 0 {
			return false
		}
		for _, d := range c.Days {
			if !slices.Contains(weekDays, d) {
				return false
			}
		}
		return true
	case CadenceTimesPerWeek:
		return c.Times >= 1 && c.Times <= 7
	}
	return false
}

func ParseData(raw any) Data {
	data := Data{Pillars: map[string]Pillar{}}
	bytes, err := json.Marshal(backendinternal.ParseJSONMap(raw))
	if err != nil {
		return data
	}
	_ = json.Unmarshal(bytes, &data)
	if data.Pillars == nil {
		data.Pillars = map[string]Pillar{}
	}
	return data
}

func MapItem(record *core.Record) ListItem {
	return ListItem{
		ID:         record.Id,
		User:       record.GetString("user"),
		StartDate:  record.GetDateTime("start_date").String(),
		EndDate:    record.GetDateTime("end_date").String(),
		Status:     record.GetString("status"),
		Visibility: record.GetString("visibility"),
		Data:       ParseData(record.Get("data")),
		Created:    record.GetDateTime("created").String(),
		Updated:    record.GetDateTime("updated").String(),
	}
}
