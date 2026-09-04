package retreats

import (
	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase/core"
)

// Item is the public-safe JSON projection of a retreats record. Unlike
// events.Item it is never gated on `active` before being returned to a
// caller — the detail endpoint always renders it, active or not.
type Item struct {
	ID            string         `json:"id"`
	Title         string         `json:"title"`
	Tagline       string         `json:"tagline,omitempty"`
	Slug          string         `json:"slug"`
	Location      string         `json:"location,omitempty"`
	StartDate     string         `json:"start_date"`
	EndDate       string         `json:"end_date,omitempty"`
	Active        bool           `json:"active"`
	TelegramGroup string         `json:"telegram_group,omitempty"`
	Gallery       []string       `json:"gallery"`
	Data          map[string]any `json:"data"`
	Created       string         `json:"created"`
	Updated       string         `json:"updated"`
}

func MapItem(record *core.Record) Item {
	if record == nil {
		return Item{}
	}
	endDate := ""
	if !record.GetDateTime("end_date").IsZero() {
		endDate = record.GetDateTime("end_date").String()
	}
	return Item{
		ID:            record.Id,
		Title:         record.GetString("title"),
		Tagline:       record.GetString("tagline"),
		Slug:          record.GetString("slug"),
		Location:      record.GetString("location"),
		StartDate:     record.GetDateTime("start_date").String(),
		EndDate:       endDate,
		Active:        record.GetBool("active"),
		TelegramGroup: record.GetString("telegram_group"),
		Gallery:       record.GetStringSlice("gallery"),
		Data:          backendinternal.ParseJSONMap(record.Get("data")),
		Created:       record.GetDateTime("created").String(),
		Updated:       record.GetDateTime("updated").String(),
	}
}

// RegistrationItem is the admin-safe JSON projection of a retreat
// registration.
type RegistrationItem struct {
	ID      string         `json:"id"`
	Retreat string         `json:"retreat"`
	Email   string         `json:"email"`
	User    string         `json:"user,omitempty"`
	IsUser  bool           `json:"is_user"`
	Status  string         `json:"status"`
	Data    map[string]any `json:"data"`
	Created string         `json:"created"`
	Updated string         `json:"updated"`
}

func MapRegistrationItem(record *core.Record) RegistrationItem {
	if record == nil {
		return RegistrationItem{}
	}
	userID := record.GetString("user")
	return RegistrationItem{
		ID:      record.Id,
		Retreat: record.GetString("retreat"),
		Email:   record.GetString("email"),
		User:    userID,
		IsUser:  userID != "",
		Status:  record.GetString("status"),
		Data:    backendinternal.ParseJSONMap(record.Get("data")),
		Created: record.GetDateTime("created").String(),
		Updated: record.GetDateTime("updated").String(),
	}
}

func parseData(record *core.Record) map[string]any {
	if record == nil {
		return map[string]any{}
	}
	return backendinternal.ParseJSONMap(record.Get("data"))
}

// DataInt reads an integer stored in the retreat's `data` JSON (price_cents,
// deposit_cents, capacity, ...). Returns 0 when missing or of the wrong type.
func DataInt(data map[string]any, key string) int {
	if data == nil {
		return 0
	}
	n, _ := backendinternal.AnyToInt64(data[key])
	return int(n)
}

// DataString reads a string stored in the retreat's `data` JSON.
func DataString(data map[string]any, key string) string {
	if data == nil {
		return ""
	}
	if s, ok := data[key].(string); ok {
		return s
	}
	return ""
}
