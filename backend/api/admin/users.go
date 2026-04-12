package admin

import (
	"net/http"
	"strings"

	backendinternal "members/backend/internal"
	groupinternal "members/backend/internal/groups"
	requestinternal "members/backend/internal/requests"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func DeleteUserHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
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

func UserGetHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		userID := strings.TrimSpace(e.Request.PathValue("id"))
		if userID == "" {
			return apis.NewBadRequestError("invalid_user", nil)
		}

		user, err := app.FindRecordById("users", userID)
		if err != nil || user == nil {
			return apis.NewNotFoundError("user_not_found", err)
		}

		memberships, err := app.FindRecordsByFilter("user_groups", "user = {:user}", "", 500, 0, map[string]any{"user": userID})
		if err != nil {
			return apis.NewBadRequestError("failed_user_groups", err)
		}

		groupIDs := make([]string, 0, len(memberships))
		for _, rel := range memberships {
			groupID := strings.TrimSpace(rel.GetString("group"))
			if groupID != "" {
				groupIDs = append(groupIDs, groupID)
			}
		}

		groups, err := app.FindRecordsByIds("groups", groupIDs)
		if err != nil {
			return apis.NewBadRequestError("failed_groups", err)
		}

		groupItems := make([]map[string]any, 0, len(groups))
		for _, group := range groups {
			if group == nil {
				continue
			}
			groupItems = append(groupItems, map[string]any{
				"id":        group.Id,
				"name":      group.GetString("name"),
				"type":      group.GetString("type"),
				"assistant": group.GetString("assistant"),
			})
		}

		telegram := backendinternal.ParseJSONMap(user.Get("telegram"))
		data := backendinternal.ParseJSONMap(user.Get("data"))
		guardianRequests, err := requestinternal.GuardianRequestsForUser(app, userID, nil)
		if err != nil {
			return apis.NewBadRequestError("failed_guardian_requests", err)
		}
		return e.JSON(http.StatusOK, map[string]any{
			"id":                user.Id,
			"email":             user.GetString("email"),
			"full_name":         groupinternal.UserDisplayName(user),
			"avatar":            strings.TrimSpace(user.GetString("avatar")),
			"data":              data,
			"status":            user.GetString("status"),
			"admin":             user.GetBool("admin"),
			"telegram":          telegram,
			"groups":            groupItems,
			"guardian_requests": guardianRequests,
		})
	}
}

func deleteUserGroupMemberships(app *pocketbase.PocketBase, userID string) (int, error) {
	relations, err := app.FindRecordsByFilter("user_groups", "user = {:user}", "", 0, 0, map[string]any{"user": userID})
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
	groups, err := app.FindRecordsByFilter("groups", "assistant = {:user}", "", 0, 0, map[string]any{"user": userID})
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
	requests, err := app.FindRecordsByFilter("requests", "guardian = {:guardian}", "", 0, 0, map[string]any{"guardian": userID})
	if err != nil {
		return 0, err
	}

	cleared := 0
	for _, record := range requests {
		if record == nil {
			continue
		}
		record.Set("guardian", "")
		data := backendinternal.ParseJSONMap(record.Get("data"))
		delete(data, "guardian")
		record.Set("data", data)
		if err := app.Save(record); err != nil {
			return cleared, err
		}
		cleared++
	}
	return cleared, nil
}
