package internal

import (
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func RequireAuthenticatedActor(e *core.RequestEvent) (*core.Record, error) {
	authRecord := e.Auth
	if authRecord == nil {
		return nil, apis.NewUnauthorizedError("Unauthorized", nil)
	}
	return authRecord, nil
}

func RequireAdmin(e *core.RequestEvent) (*core.Record, error) {
	authRecord, err := RequireAuthenticatedActor(e)
	if err != nil {
		return nil, err
	}
	if !authRecord.GetBool("admin") {
		return nil, apis.NewForbiddenError("Forbidden", nil)
	}
	return authRecord, nil
}

func HasRoleForRequest(app *pocketbase.PocketBase, actor *core.Record, record *core.Record, role string, requestRoleAdmin string, requestRoleGuardian string, requestRoleAssistant string) (bool, error) {
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

func AssistantGroupIDsForUser(app *pocketbase.PocketBase, user *core.Record) (map[string]struct{}, error) {
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

func CanViewRequest(actor *core.Record, request *core.Record, assistantGroups map[string]struct{}) bool {
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

func VisibleRequestForActor(app *pocketbase.PocketBase, actor *core.Record, requestID string) (*core.Record, error) {
	if actor == nil {
		return nil, apis.NewUnauthorizedError("Unauthorized", nil)
	}

	id := strings.TrimSpace(requestID)
	if id == "" {
		return nil, apis.NewBadRequestError("invalid_request", nil)
	}

	record, err := app.FindRecordById("requests", id)
	if err != nil || record == nil {
		return nil, apis.NewNotFoundError("request_not_found", err)
	}

	if record.GetBool("rejected") && !actor.GetBool("admin") {
		return nil, apis.NewForbiddenError("forbidden_request", nil)
	}

	assistantGroups, err := AssistantGroupIDsForUser(app, actor)
	if err != nil {
		return nil, apis.NewBadRequestError("failed_groups_lookup", err)
	}
	if !CanViewRequest(actor, record, assistantGroups) {
		return nil, apis.NewForbiddenError("forbidden_request", nil)
	}

	return record, nil
}

func RequireGroupVisibility(app *pocketbase.PocketBase, actor *core.Record, group *core.Record) error {
	if actor == nil {
		return apis.NewUnauthorizedError("Unauthorized", nil)
	}
	if group == nil {
		return apis.NewBadRequestError("invalid_group", nil)
	}
	if actor.GetBool("admin") {
		return nil
	}
	if IsAssistantForGroup(actor, group) {
		return nil
	}
	ok, err := IsMemberOfGroup(app, actor.Id, group.Id)
	if err != nil {
		return apis.NewBadRequestError("failed_group_access_check", err)
	}
	if !ok {
		return apis.NewForbiddenError("forbidden_group", nil)
	}
	return nil
}

func IsAssistantForGroup(actor *core.Record, group *core.Record) bool {
	if actor == nil || group == nil {
		return false
	}
	return strings.TrimSpace(group.GetString("assistant")) == actor.Id
}

func IsMemberOfGroup(app *pocketbase.PocketBase, userID string, groupID string) (bool, error) {
	records, err := app.FindRecordsByFilter(
		"user_groups",
		"user = {:user} && group = {:group}",
		"",
		1,
		0,
		map[string]any{
			"user":  userID,
			"group": groupID,
		},
	)
	if err != nil {
		return false, err
	}
	return len(records) > 0, nil
}
