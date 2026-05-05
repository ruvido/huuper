package events

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// CreateInput describes a single event record. Recurrence is expressed by the
// (StartDate + Cadence + Count) triple — never by repeating dates client-side.
type CreateInput struct {
	Type      string         `json:"type"`
	Group     string         `json:"group,omitempty"`
	Title     string         `json:"title,omitempty"`
	URL       string         `json:"url,omitempty"`
	StartDate string         `json:"start_date"`
	EndDate   string         `json:"end_date,omitempty"`
	Cadence   string         `json:"cadence"`
	Count     int            `json:"count"`
	Location  string         `json:"location,omitempty"`
	Data      map[string]any `json:"data,omitempty"`
}

type UpdateInput struct {
	Title     string         `json:"title"`
	URL       string         `json:"url"`
	Location  string         `json:"location"`
	StartDate string         `json:"start_date,omitempty"`
	EndDate   string         `json:"end_date,omitempty"`
	Cadence   string         `json:"cadence,omitempty"`
	Count     int            `json:"count,omitempty"`
	Data      map[string]any `json:"data"`
}

func Create(app *pocketbase.PocketBase, creator *core.Record, in CreateInput) (*core.Record, error) {
	if creator == nil {
		return nil, fmt.Errorf("missing creator")
	}
	cfg, err := LoadConfig(app)
	if err != nil {
		return nil, err
	}
	typeDef, ok := cfg.TypeDef(in.Type)
	if !ok {
		return nil, fmt.Errorf("invalid type %q", in.Type)
	}
	cadence := strings.TrimSpace(in.Cadence)
	if cadence == "" {
		cadence = CadenceOnce
	}
	count := in.Count
	if cadence == CadenceOnce {
		count = 1
	}
	if count < 1 {
		count = 1
	}
	start, err := parseDate(in.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date: %w", err)
	}
	if _, err := ComputeOccurrences(start, cadence, count); err != nil {
		return nil, err
	}
	var endPtr *time.Time
	if strings.TrimSpace(in.EndDate) != "" {
		end, err := parseDate(in.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date: %w", err)
		}
		if end.Before(start) {
			return nil, fmt.Errorf("end_date precedes start_date")
		}
		endPtr = &end
	}
	if err := validateInputRequirements(app, typeDef, in, start, endPtr); err != nil {
		return nil, err
	}

	collection, err := app.FindCollectionByNameOrId("events")
	if err != nil {
		return nil, err
	}

	title := strings.TrimSpace(in.Title)
	url := strings.TrimSpace(in.URL)
	location := strings.TrimSpace(in.Location)
	groupID := strings.TrimSpace(in.Group)

	record := core.NewRecord(collection)
	record.Set("type", in.Type)
	record.Set("event_date", start)
	record.Set("title", title)
	record.Set("slug", generateSlug(in.Type, start))
	record.Set("active", true)
	record.Set("created_by", creator.Id)
	record.Set("cadence", cadence)
	record.Set("count", strconv.Itoa(count))
	if endPtr != nil {
		record.Set("end_date", *endPtr)
	}
	if url != "" {
		record.Set("url", url)
	}
	if location != "" {
		record.Set("location", location)
	}
	if groupID != "" {
		record.Set("group", groupID)
	}
	if in.Data != nil {
		record.Set("data", in.Data)
	}
	if err := app.Save(record); err != nil {
		return nil, err
	}
	return record, nil
}

func Update(app *pocketbase.PocketBase, record *core.Record, in UpdateInput) error {
	if record == nil {
		return fmt.Errorf("missing event")
	}
	if isPast(record) {
		return fmt.Errorf("past events cannot be edited")
	}
	cfg, err := LoadConfig(app)
	if err != nil {
		return err
	}
	typeDef, ok := cfg.TypeDef(record.GetString("type"))
	if !ok {
		return fmt.Errorf("invalid type %q", record.GetString("type"))
	}
	if strings.TrimSpace(in.Title) != "" {
		record.Set("title", strings.TrimSpace(in.Title))
	}
	record.Set("url", strings.TrimSpace(in.URL))
	record.Set("location", strings.TrimSpace(in.Location))
	if strings.TrimSpace(in.Cadence) != "" {
		record.Set("cadence", strings.TrimSpace(in.Cadence))
	}
	if in.Count > 0 {
		record.Set("count", strconv.Itoa(in.Count))
	}
	if strings.TrimSpace(in.StartDate) != "" {
		start, err := parseDate(in.StartDate)
		if err != nil {
			return fmt.Errorf("invalid start_date: %w", err)
		}
		record.Set("event_date", start)
	}
	if strings.TrimSpace(in.EndDate) != "" {
		end, err := parseDate(in.EndDate)
		if err != nil {
			return fmt.Errorf("invalid end_date: %w", err)
		}
		record.Set("end_date", end)
	}
	if in.Data != nil {
		record.Set("data", in.Data)
	}
	if err := validateRecordRequirements(app, typeDef, record); err != nil {
		return err
	}
	return app.Save(record)
}

// Reschedule shifts the start_date of the recurring series — all computed
// occurrences move by the same delta. Past occurrences are immutable
// historically but the field change is accepted (caller's responsibility).
func Reschedule(app *pocketbase.PocketBase, record *core.Record, newDate string) error {
	if record == nil {
		return fmt.Errorf("missing event")
	}
	if isPast(record) {
		return fmt.Errorf("past events cannot be rescheduled")
	}
	parsed, err := parseDate(newDate)
	if err != nil {
		return err
	}
	record.Set("event_date", parsed)
	return app.Save(record)
}

// Cancel deletes the entire event record (and all its registrations).
// To skip a single occurrence of a recurring event, use CancelOccurrence.
func Cancel(app *pocketbase.PocketBase, record *core.Record) error {
	if record == nil {
		return fmt.Errorf("missing event")
	}
	return deleteEventWithRegistrations(app, record)
}

// CancelOccurrence appends a single date (YYYY-MM-DD) to data.cancelled_dates
// on a recurring event. Render-time logic skips dates present in this list.
// Idempotent — adding the same date twice is a no-op.
func CancelOccurrence(app *pocketbase.PocketBase, record *core.Record, dateStr string) error {
	if record == nil {
		return fmt.Errorf("missing event")
	}
	parsed, err := parseDate(dateStr)
	if err != nil {
		return fmt.Errorf("invalid date: %w", err)
	}
	key := parsed.Format("2006-01-02")
	occurrences, err := OccurrencesFor(record)
	if err != nil {
		return err
	}
	found := false
	for _, occurrence := range occurrences {
		occurrenceDate, err := time.Parse(time.RFC3339, occurrence.Date)
		if err != nil {
			continue
		}
		if occurrenceDate.Format("2006-01-02") != key {
			continue
		}
		found = true
		if occurrence.Past {
			return fmt.Errorf("past occurrences cannot be cancelled")
		}
		break
	}
	if !found {
		return fmt.Errorf("date is not an occurrence")
	}
	data := backendinternal.ParseJSONMap(record.Get("data"))
	existing, _ := data["cancelled_dates"].([]any)
	for _, v := range existing {
		if s, ok := v.(string); ok && s == key {
			return nil
		}
	}
	existing = append(existing, key)
	data["cancelled_dates"] = existing
	record.Set("data", data)
	return app.Save(record)
}

func deleteEventWithRegistrations(app *pocketbase.PocketBase, event *core.Record) error {
	if event == nil {
		return nil
	}
	registrations, err := app.FindRecordsByFilter(
		"event_registrations",
		"event = {:event}",
		"",
		0, 0,
		map[string]any{"event": event.Id},
	)
	if err != nil {
		return err
	}
	for _, registration := range registrations {
		if err := app.Delete(registration); err != nil {
			return err
		}
	}
	return app.Delete(event)
}

func parseDate(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05.000Z", "2006-01-02 15:04:05Z", "2006-01-02T15:04", "2006-01-02"} {
		if t, err := time.Parse(layout, raw); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized date format: %s", raw)
}

func validateInputRequirements(app *pocketbase.PocketBase, typeDef TypeDef, in CreateInput, start time.Time, end *time.Time) error {
	if typeDef.Requires("title") && strings.TrimSpace(in.Title) == "" {
		return fmt.Errorf("missing required field title")
	}
	if typeDef.Requires("url") {
		if err := validateURL(in.URL); err != nil {
			return fmt.Errorf("invalid required field url: %w", err)
		}
	} else if strings.TrimSpace(in.URL) != "" {
		if err := validateURL(in.URL); err != nil {
			return fmt.Errorf("invalid url: %w", err)
		}
	}
	if typeDef.Requires("location") && strings.TrimSpace(in.Location) == "" {
		return fmt.Errorf("missing required field location")
	}
	if typeDef.Requires("group") && strings.TrimSpace(in.Group) == "" {
		return fmt.Errorf("missing required field group")
	}
	if strings.TrimSpace(in.Group) != "" {
		if _, err := app.FindRecordById("groups", strings.TrimSpace(in.Group)); err != nil {
			return fmt.Errorf("invalid group: %w", err)
		}
	}
	if typeDef.Requires("end_date") && end == nil {
		return fmt.Errorf("missing required field end_date")
	}
	if end != nil && end.Before(start) {
		return fmt.Errorf("end_date precedes start_date")
	}
	return nil
}

func validateRecordRequirements(app *pocketbase.PocketBase, typeDef TypeDef, record *core.Record) error {
	if typeDef.Requires("title") && strings.TrimSpace(record.GetString("title")) == "" {
		return fmt.Errorf("missing required field title")
	}
	if typeDef.Requires("url") {
		if err := validateURL(record.GetString("url")); err != nil {
			return fmt.Errorf("invalid required field url: %w", err)
		}
	} else if strings.TrimSpace(record.GetString("url")) != "" {
		if err := validateURL(record.GetString("url")); err != nil {
			return fmt.Errorf("invalid url: %w", err)
		}
	}
	if typeDef.Requires("location") && strings.TrimSpace(record.GetString("location")) == "" {
		return fmt.Errorf("missing required field location")
	}
	if typeDef.Requires("group") && strings.TrimSpace(record.GetString("group")) == "" {
		return fmt.Errorf("missing required field group")
	}
	if strings.TrimSpace(record.GetString("group")) != "" {
		if _, err := app.FindRecordById("groups", strings.TrimSpace(record.GetString("group"))); err != nil {
			return fmt.Errorf("invalid group: %w", err)
		}
	}
	if typeDef.Requires("end_date") && record.GetDateTime("end_date").IsZero() {
		return fmt.Errorf("missing required field end_date")
	}
	start := record.GetDateTime("event_date").Time()
	end := record.GetDateTime("end_date").Time()
	if !start.IsZero() && !end.IsZero() && end.Before(start) {
		return fmt.Errorf("end_date precedes start_date")
	}
	return nil
}

func validateURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fmt.Errorf("empty")
	}
	parsed, err := url.ParseRequestURI(raw)
	if err != nil {
		return err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("missing scheme or host")
	}
	return nil
}

func generateSlug(eventType string, date time.Time) string {
	return fmt.Sprintf("%s-%s-%s", eventType, date.Format("2006-01-02"), backendinternal.RandomToken()[:8])
}

func isPast(record *core.Record) bool {
	if record == nil {
		return false
	}
	start := record.GetDateTime("event_date").Time()
	if start.IsZero() {
		return false
	}
	cadence := record.GetString("cadence")
	if cadence == "" {
		cadence = CadenceOnce
	}
	count := eventCount(record)
	last, err := LastOccurrence(start, cadence, count)
	if err != nil || last.IsZero() {
		last = start
	}
	return last.Before(time.Now())
}
