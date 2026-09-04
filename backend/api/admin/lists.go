package admin

import (
	"encoding/json"
	"net/http"
	"sort"
	"time"

	backendinternal "members/backend/internal"
	eventinternal "members/backend/internal/events"
	retreatsinternal "members/backend/internal/retreats"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func GroupsListHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		records, err := app.FindRecordsByFilter("groups", "", "name", 500, 0)
		if err != nil {
			return apis.NewBadRequestError("failed_groups", err)
		}

		items := make([]map[string]any, 0, len(records))
		for _, record := range records {
			items = append(items, map[string]any{
				"id":          record.Id,
				"name":        record.GetString("name"),
				"type":        record.GetString("type"),
				"description": record.GetString("description"),
				"assistant":   record.GetString("assistant"),
			})
		}

		return e.JSON(http.StatusOK, map[string]any{"items": items})
	}
}

// EventsListHandler returns a single merged, chronologically-ordered feed of
// both `events` (call/meetup) and `retreats` — the two backends are separate
// collections/modules (see docs/CLAUDE.md), but admins browse them as one
// unified "Events" list, exactly as they did before retreats were scorporated
// out of the `events` collection.
func EventsListHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor, err := backendinternal.RequireAuthenticatedActor(e)
		if err != nil {
			return err
		}
		window := e.Request.URL.Query().Get("window")
		if window == "" {
			window = eventinternal.WindowFuture
		}

		eventItems, err := eventinternal.ListForUser(app, actor, window)
		if err != nil {
			return apis.NewBadRequestError("failed_events", err)
		}
		retreatItems, err := retreatsinternal.ListItems(app, window)
		if err != nil {
			return apis.NewBadRequestError("failed_retreats", err)
		}

		merged := make([]map[string]any, 0, len(eventItems)+len(retreatItems))
		for _, item := range eventItems {
			merged = append(merged, toMergedMap(item))
		}
		for _, item := range retreatItems {
			row := toMergedMap(item)
			row["type"] = "retreat"
			row["type_label"] = "Retreat"
			merged = append(merged, row)
		}

		sortMergedByWindow(merged, window)

		return e.JSON(http.StatusOK, map[string]any{"items": merged})
	}
}

// toMergedMap round-trips a typed Item (events.Item or retreats.Item) through
// JSON into a plain map — the two Item types share most field names but
// aren't the same Go type, so a generic map is the simplest common shape for
// the merged feed.
func toMergedMap(item any) map[string]any {
	raw, err := json.Marshal(item)
	if err != nil {
		return map[string]any{}
	}
	var row map[string]any
	if err := json.Unmarshal(raw, &row); err != nil {
		return map[string]any{}
	}
	return row
}

func sortMergedByWindow(rows []map[string]any, window string) {
	sortKey := func(row map[string]any) time.Time {
		raw, _ := row["start_date"].(string)
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			t, err = time.Parse("2006-01-02 15:04:05.000Z", raw)
		}
		if err != nil {
			return time.Time{}
		}
		return t
	}
	sort.SliceStable(rows, func(i, j int) bool {
		ti, tj := sortKey(rows[i]), sortKey(rows[j])
		if window == eventinternal.WindowPast {
			return ti.After(tj)
		}
		return ti.Before(tj)
	})
}
