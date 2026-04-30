package events

import (
	"fmt"
	"strings"
	"time"

	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

const (
	CancelScopeThis   = "this"
	CancelScopeFuture = "future"
	CancelScopeAll    = "all"
)

type CreateInput struct {
	Type  string         `json:"type"`
	Group string         `json:"group,omitempty"`
	Title string         `json:"title,omitempty"`
	URL   string         `json:"url,omitempty"`
	Dates []string       `json:"dates"`
	Data  map[string]any `json:"data,omitempty"`
}

type UpdateInput struct {
	Title string         `json:"title"`
	URL   string         `json:"url"`
	Data  map[string]any `json:"data"`
}

func Create(app *pocketbase.PocketBase, creator *core.Record, in CreateInput) ([]*core.Record, error) {
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
	if err := validateCreate(in, typeDef, cfg); err != nil {
		return nil, err
	}

	dates, err := parseDates(in.Dates)
	if err != nil {
		return nil, err
	}

	collection, err := app.FindCollectionByNameOrId("events")
	if err != nil {
		return nil, err
	}

	series := ""
	if len(dates) > 1 {
		series = backendinternal.RandomToken()
	}

	title := strings.TrimSpace(in.Title)
	url := strings.TrimSpace(in.URL)
	groupID := strings.TrimSpace(in.Group)

	out := make([]*core.Record, 0, len(dates))
	for _, date := range dates {
		record := core.NewRecord(collection)
		record.Set("type", in.Type)
		record.Set("event_date", date)
		record.Set("title", title)
		record.Set("slug", generateSlug(in.Type, date))
		record.Set("active", true)
		record.Set("created_by", creator.Id)
		if url != "" {
			record.Set("url", url)
		}
		if groupID != "" {
			record.Set("group", groupID)
		}
		if series != "" {
			record.Set("series", series)
		}
		if in.Data != nil {
			record.Set("data", in.Data)
		}
		if err := app.Save(record); err != nil {
			return nil, err
		}
		out = append(out, record)
	}
	return out, nil
}

func Update(app *pocketbase.PocketBase, record *core.Record, in UpdateInput) error {
	if record == nil {
		return fmt.Errorf("missing event")
	}
	if isPast(record) {
		return fmt.Errorf("past events cannot be edited")
	}
	if strings.TrimSpace(in.Title) != "" {
		record.Set("title", strings.TrimSpace(in.Title))
	}
	record.Set("url", strings.TrimSpace(in.URL))
	if in.Data != nil {
		record.Set("data", in.Data)
	}
	return app.Save(record)
}

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

func Cancel(app *pocketbase.PocketBase, record *core.Record, scope string) (int, error) {
	if record == nil {
		return 0, fmt.Errorf("missing event")
	}
	series := strings.TrimSpace(record.GetString("series"))
	if series == "" || scope == CancelScopeThis {
		if err := deleteEventWithRegistrations(app, record); err != nil {
			return 0, err
		}
		return 1, nil
	}

	switch scope {
	case CancelScopeFuture:
		anchor := record.GetDateTime("event_date").Time()
		records, err := app.FindRecordsByFilter(
			"events",
			"series = {:series} && event_date >= {:anchor}",
			"event_date",
			0, 0,
			map[string]any{"series": series, "anchor": anchor},
		)
		if err != nil {
			return 0, err
		}
		return deleteEvents(app, records)
	case CancelScopeAll:
		records, err := app.FindRecordsByFilter(
			"events",
			"series = {:series}",
			"event_date",
			0, 0,
			map[string]any{"series": series},
		)
		if err != nil {
			return 0, err
		}
		return deleteEvents(app, records)
	}
	return 0, fmt.Errorf("invalid cancel scope %q", scope)
}

func deleteEvents(app *pocketbase.PocketBase, records []*core.Record) (int, error) {
	count := 0
	for _, r := range records {
		if err := deleteEventWithRegistrations(app, r); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
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

func validateCreate(in CreateInput, typeDef TypeDef, cfg *Config) error {
	if len(in.Dates) == 0 {
		return fmt.Errorf("at least one date is required")
	}
	if max := cfg.MaxOccurrences(); len(in.Dates) > max {
		return fmt.Errorf("too many occurrences (max %d)", max)
	}
	if typeDef.RequiresTitle && strings.TrimSpace(in.Title) == "" {
		return fmt.Errorf("title is required for type %q", in.Type)
	}
	if typeDef.RequiresGroup && strings.TrimSpace(in.Group) == "" {
		return fmt.Errorf("group is required for type %q", in.Type)
	}
	return nil
}

func parseDates(raw []string) ([]time.Time, error) {
	out := make([]time.Time, 0, len(raw))
	seen := make(map[string]struct{}, len(raw))
	for _, value := range raw {
		parsed, err := parseDate(value)
		if err != nil {
			return nil, err
		}
		key := parsed.Format(time.RFC3339)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, parsed)
	}
	return out, nil
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

func generateSlug(eventType string, date time.Time) string {
	return fmt.Sprintf("%s-%s-%s", eventType, date.Format("2006-01-02"), backendinternal.RandomToken()[:8])
}

func isPast(record *core.Record) bool {
	if record == nil {
		return false
	}
	eventDate := record.GetDateTime("event_date").Time()
	return !eventDate.IsZero() && eventDate.Before(time.Now())
}
