package events

import (
	"sort"
	"strconv"
	"strings"
	"time"

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

func CanCreate(actor *core.Record, group *core.Record, eventType string) bool {
	if actor == nil {
		return false
	}
	if actor.GetBool("admin") {
		return true
	}
	if !IsAssistantCreatableType(eventType) {
		return false
	}
	if group == nil {
		return false
	}
	return strings.TrimSpace(group.GetString("assistant")) == actor.Id
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

	filter, params := visibilityFilter(actor, groupIDs, window)
	order := "event_date"
	if window == WindowPast {
		order = "-event_date"
	}

	records, err := app.FindRecordsByFilter("events", filter, order, 0, 0, params)
	if err != nil {
		return nil, err
	}

	collapsed := collapseSeries(records, window)

	items := make([]Item, 0, len(collapsed))
	groupNames, err := groupNamesByID(app, collectGroupIDs(collapsed))
	if err != nil {
		return nil, err
	}
	for _, record := range collapsed {
		item := MapItem(record)
		if item.Group != "" {
			item.GroupName = groupNames[item.Group]
		}
		items = append(items, item)
	}
	return items, nil
}

func visibilityFilter(actor *core.Record, groupIDs []string, window string) (string, map[string]any) {
	params := map[string]any{}
	parts := []string{}

	if actor.GetBool("admin") {
		parts = append(parts, "1=1")
	} else {
		groupClauses := []string{"group = ''"}
		for index, groupID := range groupIDs {
			key := "g" + strconv.Itoa(index)
			groupClauses = append(groupClauses, "group = {:"+key+"}")
			params[key] = groupID
		}
		parts = append(parts, "("+strings.Join(groupClauses, " || ")+")")
	}

	switch window {
	case WindowPast:
		parts = append(parts, "event_date < {:now}")
		params["now"] = time.Now()
	case WindowAll:
		// no date filter
	default:
		parts = append(parts, "event_date >= {:now}")
		params["now"] = time.Now()
	}

	return strings.Join(parts, " && "), params
}

func collapseSeries(records []*core.Record, window string) []*core.Record {
	type bucket struct {
		record *core.Record
	}
	out := make([]*core.Record, 0, len(records))
	bestBySeries := map[string]bucket{}
	for _, record := range records {
		series := strings.TrimSpace(record.GetString("series"))
		if series == "" {
			out = append(out, record)
			continue
		}
		current, ok := bestBySeries[series]
		if !ok {
			bestBySeries[series] = bucket{record: record}
			continue
		}
		if window == WindowPast {
			// keep most recent past occurrence
			if record.GetDateTime("event_date").Time().After(current.record.GetDateTime("event_date").Time()) {
				bestBySeries[series] = bucket{record: record}
			}
		} else {
			// keep next upcoming occurrence
			if record.GetDateTime("event_date").Time().Before(current.record.GetDateTime("event_date").Time()) {
				bestBySeries[series] = bucket{record: record}
			}
		}
	}
	for _, b := range bestBySeries {
		out = append(out, b.record)
	}
	sort.SliceStable(out, func(i, j int) bool {
		ti := out[i].GetDateTime("event_date").Time()
		tj := out[j].GetDateTime("event_date").Time()
		if window == WindowPast {
			return ti.After(tj)
		}
		return ti.Before(tj)
	})
	return out
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

