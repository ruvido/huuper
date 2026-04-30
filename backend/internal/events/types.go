package events

import (
	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase/core"
)

const (
	TypeRally  = "rally"
	TypeCall   = "call"
	TypeMeetup = "meetup"
)

func IsValidType(value string) bool {
	switch value {
	case TypeRally, TypeCall, TypeMeetup:
		return true
	}
	return false
}

func IsAssistantCreatableType(value string) bool {
	switch value {
	case TypeCall, TypeMeetup:
		return true
	}
	return false
}

func RequiresGroup(value string) bool {
	return value == TypeMeetup
}

type Item struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	Title     string         `json:"title"`
	Slug      string         `json:"slug,omitempty"`
	EventDate string         `json:"event_date"`
	URL       string         `json:"url,omitempty"`
	Group     string         `json:"group,omitempty"`
	GroupName string         `json:"group_name,omitempty"`
	Series    string         `json:"series,omitempty"`
	CreatedBy string         `json:"created_by,omitempty"`
	Active    bool           `json:"active"`
	Data      map[string]any `json:"data"`
	Created   string         `json:"created"`
	Updated   string         `json:"updated"`
}

func MapItem(record *core.Record) Item {
	return Item{
		ID:        record.Id,
		Type:      record.GetString("type"),
		Title:     record.GetString("title"),
		Slug:      record.GetString("slug"),
		EventDate: record.GetDateTime("event_date").String(),
		URL:       record.GetString("url"),
		Group:     record.GetString("group"),
		Series:    record.GetString("series"),
		CreatedBy: record.GetString("created_by"),
		Active:    record.GetBool("active"),
		Data:      backendinternal.ParseJSONMap(record.Get("data")),
		Created:   record.GetDateTime("created").String(),
		Updated:   record.GetDateTime("updated").String(),
	}
}
