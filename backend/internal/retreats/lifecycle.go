package retreats

import (
	"fmt"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
)

// CreateInput describes a new retreat. Recurrence does not exist here on
// purpose — retreats are single, rare events (see plan rationale).
type CreateInput struct {
	Title         string         `json:"title"`
	Tagline       string         `json:"tagline,omitempty"`
	Slug          string         `json:"slug"`
	Location      string         `json:"location,omitempty"`
	StartDate     string         `json:"start_date"`
	EndDate       string         `json:"end_date,omitempty"`
	Active        bool           `json:"active"`
	TelegramGroup string         `json:"telegram_group,omitempty"`
	Data          map[string]any `json:"data,omitempty"`
}

// UpdateInput is a true partial update: every field is a pointer so that
// "absent from the JSON body" (nil → leave untouched) is distinguishable from
// "explicitly set to empty" (pointer to "" → clear it). Using plain strings
// here silently wipes any field the caller didn't send.
type UpdateInput struct {
	Title         *string        `json:"title"`
	Tagline       *string        `json:"tagline"`
	Slug          *string        `json:"slug"`
	Location      *string        `json:"location"`
	StartDate     *string        `json:"start_date"`
	EndDate       *string        `json:"end_date"`
	Active        *bool          `json:"active"`
	Capacity      *int           `json:"capacity"`
	TelegramGroup *string        `json:"telegram_group"`
	Data          map[string]any `json:"data"`
}

func Create(app *pocketbase.PocketBase, in CreateInput) (*core.Record, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil, fmt.Errorf("missing required field title")
	}
	slug := strings.TrimSpace(in.Slug)
	if slug == "" {
		return nil, fmt.Errorf("missing required field slug")
	}
	start, err := parseDate(in.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date: %w", err)
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

	if err := validateTelegramGroup(app, in.TelegramGroup); err != nil {
		return nil, err
	}

	collection, err := app.FindCollectionByNameOrId("retreats")
	if err != nil {
		return nil, err
	}

	record := core.NewRecord(collection)
	record.Set("title", title)
	record.Set("tagline", strings.TrimSpace(in.Tagline))
	record.Set("slug", slug)
	record.Set("location", strings.TrimSpace(in.Location))
	record.Set("start_date", start)
	if endPtr != nil {
		record.Set("end_date", *endPtr)
	}
	record.Set("active", in.Active)
	if strings.TrimSpace(in.TelegramGroup) != "" {
		record.Set("telegram_group", strings.TrimSpace(in.TelegramGroup))
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
		return fmt.Errorf("missing retreat")
	}
	// title / slug / start_date are required: they may be changed, never cleared.
	if in.Title != nil {
		title := strings.TrimSpace(*in.Title)
		if title == "" {
			return fmt.Errorf("missing required field title")
		}
		record.Set("title", title)
	}
	if in.Slug != nil {
		slug := strings.TrimSpace(*in.Slug)
		if slug == "" {
			return fmt.Errorf("missing required field slug")
		}
		record.Set("slug", slug)
	}
	// tagline / location are optional: sending "" clears them on purpose.
	if in.Tagline != nil {
		record.Set("tagline", strings.TrimSpace(*in.Tagline))
	}
	if in.Location != nil {
		record.Set("location", strings.TrimSpace(*in.Location))
	}
	if in.StartDate != nil {
		raw := strings.TrimSpace(*in.StartDate)
		if raw == "" {
			return fmt.Errorf("missing required field start_date")
		}
		start, err := parseDate(raw)
		if err != nil {
			return fmt.Errorf("invalid start_date: %w", err)
		}
		record.Set("start_date", start)
	}
	if in.EndDate != nil {
		raw := strings.TrimSpace(*in.EndDate)
		if raw == "" {
			record.Set("end_date", nil)
		} else {
			end, err := parseDate(raw)
			if err != nil {
				return fmt.Errorf("invalid end_date: %w", err)
			}
			record.Set("end_date", end)
		}
	}
	if in.Active != nil {
		record.Set("active", *in.Active)
	}
	if in.Capacity != nil {
		record.Set("capacity", *in.Capacity)
	}
	if in.TelegramGroup != nil {
		group := strings.TrimSpace(*in.TelegramGroup)
		if group != "" {
			if err := validateTelegramGroup(app, group); err != nil {
				return err
			}
		}
		record.Set("telegram_group", group)
	}
	if in.Data != nil {
		record.Set("data", in.Data)
	}

	start := record.GetDateTime("start_date").Time()
	end := record.GetDateTime("end_date").Time()
	if !start.IsZero() && !end.IsZero() && end.Before(start) {
		return fmt.Errorf("end_date precedes start_date")
	}

	return app.Save(record)
}

// AppendGalleryFiles uploads and appends new images to the retreat's
// gallery file field without discarding the existing ones.
func AppendGalleryFiles(app *pocketbase.PocketBase, record *core.Record, files []*filesystem.File) error {
	if record == nil {
		return fmt.Errorf("missing retreat")
	}
	if len(files) == 0 {
		return nil
	}
	existing := record.GetStringSlice("gallery")
	merged := make([]any, 0, len(existing)+len(files))
	for _, name := range existing {
		merged = append(merged, name)
	}
	for _, file := range files {
		merged = append(merged, file)
	}
	record.Set("gallery", merged)
	return app.Save(record)
}

// RemoveGalleryFile drops a single filename from the retreat's gallery.
func RemoveGalleryFile(app *pocketbase.PocketBase, record *core.Record, filename string) error {
	if record == nil {
		return fmt.Errorf("missing retreat")
	}
	filename = strings.TrimSpace(filename)
	if filename == "" {
		return fmt.Errorf("missing filename")
	}
	existing := record.GetStringSlice("gallery")
	kept := make([]string, 0, len(existing))
	for _, name := range existing {
		if name != filename {
			kept = append(kept, name)
		}
	}
	record.Set("gallery", kept)
	return app.Save(record)
}

// Cancel deletes the retreat record and all of its registrations.
func Cancel(app *pocketbase.PocketBase, record *core.Record) error {
	if record == nil {
		return fmt.Errorf("missing retreat")
	}
	registrations, err := app.FindRecordsByFilter(
		"retreat_registrations",
		"retreat = {:retreat}",
		"",
		0, 0,
		map[string]any{"retreat": record.Id},
	)
	if err != nil {
		return err
	}
	for _, registration := range registrations {
		if err := app.Delete(registration); err != nil {
			return err
		}
	}
	return app.Delete(record)
}

func validateTelegramGroup(app *pocketbase.PocketBase, groupID string) error {
	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		return nil
	}
	if _, err := app.FindRecordById("groups", groupID); err != nil {
		return fmt.Errorf("invalid telegram_group: %w", err)
	}
	return nil
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
