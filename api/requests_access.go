package api

import (
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func hasRoleForRequest(app *pocketbase.PocketBase, actor *core.Record, record *core.Record, role string) (bool, error) {
	if actor != nil && actor.GetBool("admin") {
		return true, nil
	}

	switch role {
	case requestRoleAdmin:
		return actor != nil && actor.GetBool("admin"), nil
	case requestRoleGuardian:
		if actor == nil {
			return false, nil
		}
		return strings.TrimSpace(record.GetString("guardian")) == actor.Id, nil
	case requestRoleAssistant:
		if actor == nil {
			return false, nil
		}
		groupID := strings.TrimSpace(record.GetString("group"))
		if groupID == "" {
			return false, nil
		}
		group, err := app.FindRecordById("groups", groupID)
		if err != nil || group == nil {
			return false, err
		}
		return strings.TrimSpace(group.GetString("assistant")) == actor.Id, nil
	default:
		return false, nil
	}
}

func assistantGroupIDsForUser(app *pocketbase.PocketBase, user *core.Record) (map[string]struct{}, error) {
	out := map[string]struct{}{}
	if user == nil || user.GetBool("admin") {
		return out, nil
	}

	groups, err := app.FindRecordsByFilter(
		"groups",
		"assistant = {:assistant}",
		"",
		500,
		0,
		map[string]any{"assistant": user.Id},
	)
	if err != nil {
		return nil, err
	}

	for _, group := range groups {
		out[group.Id] = struct{}{}
	}
	return out, nil
}

func canViewRequest(actor *core.Record, request *core.Record, assistantGroups map[string]struct{}) bool {
	if actor == nil || request == nil {
		return false
	}
	if actor.GetBool("admin") {
		return true
	}

	if strings.TrimSpace(request.GetString("guardian")) == actor.Id {
		return true
	}

	groupID := strings.TrimSpace(request.GetString("group"))
	if groupID == "" {
		return false
	}
	_, ok := assistantGroups[groupID]
	return ok
}
