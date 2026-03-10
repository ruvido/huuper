package api

import (
	"net/http"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func AdminDeleteUserHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if _, err := requireAdmin(e); err != nil {
			return err
		}

		userID := strings.TrimSpace(e.Request.PathValue("id"))
		if userID == "" {
			return apis.NewBadRequestError("invalid_user", nil)
		}

		user, err := app.FindRecordById("users", userID)
		if err != nil || user == nil {
			return apis.NewNotFoundError("user_not_found", err)
		}

		removedMemberships, err := deleteUserGroupMemberships(app, userID)
		if err != nil {
			return apis.NewBadRequestError("failed_cleanup_user_groups", err)
		}

		clearedGroups, err := clearUserAsGroupAssistant(app, userID)
		if err != nil {
			return apis.NewBadRequestError("failed_cleanup_groups", err)
		}

		clearedGuardians, err := clearUserAsRequestGuardian(app, userID)
		if err != nil {
			return apis.NewBadRequestError("failed_cleanup_requests", err)
		}

		if err := app.Delete(user); err != nil {
			return apis.NewBadRequestError("failed_delete_user", err)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"ok":                  true,
			"deleted_user":        userID,
			"removed_memberships": removedMemberships,
			"cleared_groups":      clearedGroups,
			"cleared_guardians":   clearedGuardians,
		})
	}
}

func deleteUserGroupMemberships(app *pocketbase.PocketBase, userID string) (int, error) {
	relations, err := app.FindRecordsByFilter(
		"user_groups",
		"user = {:user}",
		"",
		0,
		0,
		map[string]any{"user": userID},
	)
	if err != nil {
		return 0, err
	}

	removed := 0
	for _, rel := range relations {
		if rel == nil {
			continue
		}
		if err := app.Delete(rel); err != nil {
			return removed, err
		}
		removed++
	}
	return removed, nil
}

func clearUserAsGroupAssistant(app *pocketbase.PocketBase, userID string) (int, error) {
	groups, err := app.FindRecordsByFilter(
		"groups",
		"assistant = {:user}",
		"",
		0,
		0,
		map[string]any{"user": userID},
	)
	if err != nil {
		return 0, err
	}

	cleared := 0
	for _, group := range groups {
		if group == nil {
			continue
		}
		group.Set("assistant", "")
		if err := app.Save(group); err != nil {
			return cleared, err
		}
		cleared++
	}
	return cleared, nil
}

func clearUserAsRequestGuardian(app *pocketbase.PocketBase, userID string) (int, error) {
	requests, err := app.FindRecordsByFilter(
		"requests",
		"guardian = {:guardian}",
		"",
		0,
		0,
		map[string]any{"guardian": userID},
	)
	if err != nil {
		return 0, err
	}

	cleared := 0
	for _, record := range requests {
		if record == nil {
			continue
		}
		record.Set("guardian", "")
		data := parseJSONMap(record.Get("data"))
		delete(data, "guardian")
		record.Set("data", data)
		if err := app.Save(record); err != nil {
			return cleared, err
		}
		cleared++
	}
	return cleared, nil
}
