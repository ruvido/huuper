package events

import (
	"sort"
	"strconv"
	"strings"
	"time"

	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

const (
	WindowFuture = "future"
	WindowPast   = "past"
	WindowAll    = "all"
)

func CanView(app *pocketbase.PocketBase, actor *core.Record, event *core.Record) (bool, error) {
	if actor == nil || event == nil {
		return false, nil
	}
	if actor.GetBool("admin") {
		return true, nil
	}
	groupID := strings.TrimSpace(event.GetString("group"))
	if groupID == "" {
		return true, nil
	}
	return isGroupMember(app, actor.Id, groupID)
}

func CanCreate(app *pocketbase.PocketBase, actor *core.Record, group *core.Record, eventType string) bool {
	if app == nil || actor == nil {
		return false
	}
	cfg, err := LoadConfig(app)
	if err != nil {
		return false
	}
	typeDef, ok := cfg.TypeDef(eventType)
	if !ok {
		return false
	}
	if actor.GetBool("admin") && typeDef.AllowsCreator("admin") {
		return true
	}
	if !typeDef.AllowsCreator("assistant") {
		return false
	}
	if group != nil {
		return strings.TrimSpace(group.GetString("assistant")) == actor.Id
	}
	groups, err := app.FindRecordsByFilter(
		"groups",
		"assistant = {:assistant}",
		"",
		1,
		0,
		map[string]any{"assistant": actor.Id},
	)
	return err == nil && len(groups) > 0
}

func CanEdit(app *pocketbase.PocketBase, actor *core.Record, event *core.Record) (bool, error) {
	if actor == nil || event == nil {
		return false, nil
	}
	if actor.GetBool("admin") {
		return true, nil
	}
	if strings.TrimSpace(event.GetString("created_by")) == actor.Id {
		return true, nil
	}
	groupID := strings.TrimSpace(event.GetString("group"))
	if groupID == "" {
		return false, nil
	}
	group, err := app.FindRecordById("groups", groupID)
	if err != nil || group == nil {
		return false, nil
	}
	return strings.TrimSpace(group.GetString("assistant")) == actor.Id, nil
}

func ListForUser(app *pocketbase.PocketBase, actor *core.Record, window string) ([]Item, error) {
	if actor == nil {
		return nil, nil
	}
	groupIDs, err := groupIDsForUser(app, actor.Id)
	if err != nil {
		return nil, err
	}

	// Visibility filter is SQL-friendly (admin or group membership). Window
	// filter (past/future) is applied in Go after expanding cadence — a
	// recurring event with start in past + future occurrences must show in
	// the future tab, which a simple event_date comparison can't capture.
	filter, params := visibilityFilter(actor, groupIDs)
	records, err := app.FindRecordsByFilter("events", filter, "event_date", 0, 0, params)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	filtered := make([]*core.Record, 0, len(records))
	for _, record := range records {
		next, hasFuture := nextNonCancelledOccurrence(record, now)
		switch window {
		case WindowPast:
			if !hasFuture {
				filtered = append(filtered, record)
			}
		case WindowAll:
			filtered = append(filtered, record)
		default: // future
			if hasFuture {
				filtered = append(filtered, record)
			}
		}
		_ = next
	}

	sort.SliceStable(filtered, func(i, j int) bool {
		ni, _ := nextNonCancelledOccurrence(filtered[i], now)
		nj, _ := nextNonCancelledOccurrence(filtered[j], now)
		if window == WindowPast {
			return ni.After(nj)
		}
		return ni.Before(nj)
	})

	items := make([]Item, 0, len(filtered))
	groupNames, err := groupNamesByID(app, collectGroupIDs(filtered))
	if err != nil {
		return nil, err
	}
	cfg, err := LoadConfig(app)
	if err != nil {
		return nil, err
	}
	for _, record := range filtered {
		item := MapItem(record)
		if item.Group != "" {
			item.GroupName = groupNames[item.Group]
		}
		ApplyTypeConfig(&item, cfg)
		items = append(items, item)
	}
	return items, nil
}

// nextNonCancelledOccurrence returns the soonest upcoming occurrence (>= now)
// whose date isn't in data.cancelled_dates. Falls back to the latest past
// non-cancelled occurrence when nothing future remains — used by the window
// classifier (hasFuture=false → record is "past").
func nextNonCancelledOccurrence(record *core.Record, now time.Time) (time.Time, bool) {
	start := record.GetDateTime("event_date").Time()
	if start.IsZero() {
		return time.Time{}, false
	}
	cadence := record.GetString("cadence")
	if cadence == "" {
		cadence = CadenceOnce
	}
	count := eventCount(record)
	cancelled := cancelledDates(record)
	next, hasFuture, err := NextOccurrence(start, cadence, count, now, cancelled)
	if err != nil {
		return start, !start.Before(now)
	}
	return next, hasFuture
}

func cancelledDates(record *core.Record) []string {
	data := backendinternal.ParseJSONMap(record.Get("data"))
	raw, _ := data["cancelled_dates"].([]any)
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		if s, ok := v.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

// Occurrence is the render-friendly view of a single computed event date.
type Occurrence struct {
	Date      string `json:"date"`
	Cancelled bool   `json:"cancelled"`
	Past      bool   `json:"past"`
}

// OccurrencesFor returns the canonical schedule for an event record, with
// cancelled and past flags so the frontend can render a per-occurrence list
// without re-implementing the cadence math.
func OccurrencesFor(record *core.Record) ([]Occurrence, error) {
	if record == nil {
		return nil, nil
	}
	start := record.GetDateTime("event_date").Time()
	if start.IsZero() {
		return nil, nil
	}
	cadence := record.GetString("cadence")
	if cadence == "" {
		cadence = CadenceOnce
	}
	count := eventCount(record)
	dates, err := ComputeOccurrences(start, cadence, count)
	if err != nil {
		return nil, err
	}
	cancelled := cancelledDates(record)
	skip := map[string]struct{}{}
	for _, c := range cancelled {
		if t, err := time.Parse(time.RFC3339, c); err == nil {
			skip[t.Format("2006-01-02")] = struct{}{}
		} else if t, err := time.Parse("2006-01-02", c); err == nil {
			skip[t.Format("2006-01-02")] = struct{}{}
		}
	}
	now := time.Now()
	out := make([]Occurrence, 0, len(dates))
	for _, d := range dates {
		_, isCancelled := skip[d.Format("2006-01-02")]
		out = append(out, Occurrence{
			Date:      d.Format(time.RFC3339),
			Cancelled: isCancelled,
			Past:      d.Before(now),
		})
	}
	return out, nil
}

func visibilityFilter(actor *core.Record, groupIDs []string) (string, map[string]any) {
	params := map[string]any{}
	if actor.GetBool("admin") {
		return "1=1", params
	}
	groupClauses := []string{"group = ''"}
	for index, groupID := range groupIDs {
		key := "g" + strconv.Itoa(index)
		groupClauses = append(groupClauses, "group = {:"+key+"}")
		params[key] = groupID
	}
	return "(" + strings.Join(groupClauses, " || ") + ")", params
}

func groupIDsForUser(app *pocketbase.PocketBase, userID string) ([]string, error) {
	rels, err := app.FindRecordsByFilter(
		"user_groups",
		"user = {:user}",
		"",
		0, 0,
		map[string]any{"user": userID},
	)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rels))
	seen := make(map[string]struct{}, len(rels))
	for _, rel := range rels {
		groupID := strings.TrimSpace(rel.GetString("group"))
		if groupID == "" {
			continue
		}
		if _, ok := seen[groupID]; ok {
			continue
		}
		seen[groupID] = struct{}{}
		out = append(out, groupID)
	}
	return out, nil
}

func isGroupMember(app *pocketbase.PocketBase, userID string, groupID string) (bool, error) {
	rel, err := app.FindFirstRecordByFilter(
		"user_groups",
		"user = {:user} && group = {:group}",
		map[string]any{"user": userID, "group": groupID},
	)
	if err != nil {
		return false, nil
	}
	return rel != nil, nil
}

func collectGroupIDs(records []*core.Record) []string {
	seen := map[string]struct{}{}
	out := []string{}
	for _, record := range records {
		groupID := strings.TrimSpace(record.GetString("group"))
		if groupID == "" {
			continue
		}
		if _, ok := seen[groupID]; ok {
			continue
		}
		seen[groupID] = struct{}{}
		out = append(out, groupID)
	}
	return out
}

func groupNamesByID(app *pocketbase.PocketBase, groupIDs []string) (map[string]string, error) {
	out := map[string]string{}
	if len(groupIDs) == 0 {
		return out, nil
	}
	records, err := app.FindRecordsByIds("groups", groupIDs)
	if err != nil {
		return nil, err
	}
	for _, record := range records {
		out[record.Id] = strings.TrimSpace(record.GetString("name"))
	}
	return out, nil
}
