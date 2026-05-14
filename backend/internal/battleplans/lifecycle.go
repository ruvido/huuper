package battleplans

import (
	"fmt"
	"strings"
	"time"

	backendinternal "members/backend/internal"
	groupinternal "members/backend/internal/groups"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

type Input struct {
	StartDate    string `json:"start_date"`
	DurationDays int    `json:"duration_days"`
	EndDate      string `json:"end_date,omitempty"`
	Visibility   string `json:"visibility"`
	Status       string `json:"status,omitempty"`
	Data         Data   `json:"data"`
}

func validateInput(in Input, cfg *Config) error {
	start, err := parseDate(in.StartDate)
	if err != nil {
		return fmt.Errorf("invalid start_date")
	}
	if strings.TrimSpace(in.EndDate) != "" {
		end, err := parseDate(in.EndDate)
		if err != nil {
			return fmt.Errorf("invalid end_date")
		}
		if !end.After(start) {
			return fmt.Errorf("end_date must be after start_date")
		}
	} else if !cfg.IsValidDuration(in.DurationDays) {
		return fmt.Errorf("duration_days must be one of the configured values")
	}
	if !cfg.IsValidVisibility(in.Visibility) {
		return fmt.Errorf("invalid visibility")
	}
	return validateData(in.Data, cfg)
}

func validateData(data Data, cfg *Config) error {
	if strings.TrimSpace(data.Priority.Title) == "" {
		return fmt.Errorf("priority.title is required")
	}
	for key, pillar := range data.Pillars {
		if !cfg.IsValidPillarKey(key) {
			return fmt.Errorf("unknown pillar %q", key)
		}
		for i, r := range pillar.Routines {
			if strings.TrimSpace(r.Title) == "" {
				return fmt.Errorf("routine %d in %q missing title", i, key)
			}
			if strings.TrimSpace(r.Trigger) == "" {
				return fmt.Errorf("routine %d in %q missing trigger", i, key)
			}
			if !IsValidCadence(r.Cadence) {
				return fmt.Errorf("routine %d in %q has invalid cadence", i, key)
			}
		}
	}
	return nil
}

func parseDate(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}
	for _, layout := range []string{"2006-01-02", time.RFC3339, "2006-01-02 15:04:05.000Z", "2006-01-02 15:04:05Z"} {
		if t, err := time.Parse(layout, raw); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized date format")
}

// stampNewRoutines assigns id+created to routines without an id (genuinely new ones).
// Existing routines (matched by id within the same pillar) keep their original created.
func stampNewRoutines(existing Data, incoming Data, now string) Data {
	prevByKey := map[string]Routine{}
	for pkey, pillar := range existing.Pillars {
		for _, r := range pillar.Routines {
			if r.ID != "" {
				prevByKey[pkey+"|"+r.ID] = r
			}
		}
	}

	out := Data{Priority: incoming.Priority, Pillars: map[string]Pillar{}, Notes: existing.Notes}
	for pkey, pillar := range incoming.Pillars {
		stamped := make([]Routine, len(pillar.Routines))
		for i, r := range pillar.Routines {
			if r.ID != "" {
				if prev, ok := prevByKey[pkey+"|"+r.ID]; ok {
					r.Created = prev.Created
					stamped[i] = r
					continue
				}
			}
			r.ID = backendinternal.RandomToken()
			r.Created = now
			stamped[i] = r
		}
		pillar.Routines = stamped
		out.Pillars[pkey] = pillar
	}
	return out
}

func Create(app *pocketbase.PocketBase, userID string, in Input) (*core.Record, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("missing user")
	}
	cfg, err := LoadConfig(app)
	if err != nil {
		return nil, err
	}
	if err := validateInput(in, cfg); err != nil {
		return nil, err
	}

	start, _ := parseDate(in.StartDate)
	var end time.Time
	if strings.TrimSpace(in.EndDate) != "" {
		end, _ = parseDate(in.EndDate)
	} else {
		end = start.AddDate(0, 0, in.DurationDays)
	}

	collection, err := app.FindCollectionByNameOrId("battleplans")
	if err != nil {
		return nil, err
	}

	status := strings.TrimSpace(in.Status)
	if status == "" {
		status = StatusActive
	}
	if !IsValidStatus(status) {
		return nil, fmt.Errorf("invalid status %q", status)
	}

	record := core.NewRecord(collection)
	record.Set("user", userID)
	record.Set("start_date", start)
	record.Set("end_date", end)
	record.Set("status", status)
	record.Set("visibility", in.Visibility)
	record.Set("data", stampNewRoutines(Data{Pillars: map[string]Pillar{}}, in.Data, types.NowDateTime().String()))

	if err := app.Save(record); err != nil {
		return nil, err
	}
	return record, nil
}

func Update(app *pocketbase.PocketBase, record *core.Record, in Input) error {
	if !IsEditable(record.GetString("status")) {
		return fmt.Errorf("only active or draft battleplans can be edited")
	}
	cfg, err := LoadConfig(app)
	if err != nil {
		return err
	}
	if err := validateData(in.Data, cfg); err != nil {
		return err
	}
	if strings.TrimSpace(in.Visibility) != "" {
		if !cfg.IsValidVisibility(in.Visibility) {
			return fmt.Errorf("invalid visibility")
		}
		record.Set("visibility", in.Visibility)
	}

	existing := ParseData(record.Get("data"))
	merged := stampNewRoutines(existing, in.Data, types.NowDateTime().String())
	record.Set("data", merged)
	return app.Save(record)
}

func AddNote(app *pocketbase.PocketBase, actor *core.Record, record *core.Record, note string) error {
	if record == nil {
		return fmt.Errorf("missing battleplan")
	}
	if record.GetString("status") != StatusActive {
		return apis.NewBadRequestError("battleplan_not_active", nil)
	}
	note = strings.TrimSpace(note)
	if note == "" {
		return apis.NewBadRequestError("missing_note", nil)
	}
	data := ParseData(record.Get("data"))
	data.Notes = append(data.Notes, Note{
		Text: note,
		At:   time.Now().UTC().Format(time.RFC3339),
		By:   groupinternal.UserDisplayName(actor),
	})
	record.Set("data", data)
	return app.Save(record)
}

func SetStatus(app *pocketbase.PocketBase, record *core.Record, status string) error {
	if !IsValidStatus(status) {
		return fmt.Errorf("invalid status")
	}
	record.Set("status", status)
	return app.Save(record)
}

func Delete(app *pocketbase.PocketBase, record *core.Record) error {
	if record == nil {
		return fmt.Errorf("missing battleplan")
	}
	return app.Delete(record)
}

func ListForUser(app *pocketbase.PocketBase, userID string, perPage, offset int) ([]ListItem, error) {
	records, err := app.FindRecordsByFilter(
		"battleplans",
		"user = {:user}",
		"-start_date",
		perPage,
		offset,
		map[string]any{"user": userID},
	)
	if err != nil {
		return nil, err
	}
	items := make([]ListItem, 0, len(records))
	for _, record := range records {
		items = append(items, MapItem(record))
	}
	return items, nil
}

func CanView(app *pocketbase.PocketBase, actor *core.Record, record *core.Record) (bool, error) {
	if actor == nil || record == nil {
		return false, nil
	}
	if actor.GetBool("admin") {
		return true, nil
	}
	if record.GetString("user") == actor.Id {
		return true, nil
	}
	if record.GetString("visibility") == "public" {
		return true, nil
	}
	return shareGroup(app, actor.Id, record.GetString("user"))
}

func shareGroup(app *pocketbase.PocketBase, actorID, ownerID string) (bool, error) {
	mine, err := app.FindRecordsByFilter(
		"user_groups",
		"user = {:user}",
		"",
		500, 0,
		map[string]any{"user": actorID},
	)
	if err != nil {
		return false, err
	}
	if len(mine) == 0 {
		return false, nil
	}
	groupIDs := make(map[string]struct{}, len(mine))
	for _, m := range mine {
		groupIDs[m.GetString("group")] = struct{}{}
	}
	theirs, err := app.FindRecordsByFilter(
		"user_groups",
		"user = {:user}",
		"",
		500, 0,
		map[string]any{"user": ownerID},
	)
	if err != nil {
		return false, err
	}
	for _, t := range theirs {
		if _, ok := groupIDs[t.GetString("group")]; ok {
			return true, nil
		}
	}
	return false, nil
}
