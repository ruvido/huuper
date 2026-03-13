package me

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

func UserGetHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor, err := backendinternal.RequireAuthenticatedActor(e)
		if err != nil {
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

		if actor.Id != userID {
			shared, err := sharedGroupCount(app, actor.Id, userID)
			if err != nil {
				return apis.NewBadRequestError("failed_user_groups", err)
			}
			if shared == 0 && !actor.GetBool("admin") {
				return apis.NewForbiddenError("forbidden_user", nil)
			}
		}

		visibleGroupIDs, err := sharedGroupIDs(app, actor.Id, userID)
		if err != nil {
			return apis.NewBadRequestError("failed_user_groups", err)
		}
		guardianRequests, err := requestinternal.GuardianRequestsForUser(app, userID, visibleGroupIDs)
		if err != nil {
			return apis.NewBadRequestError("failed_guardian_requests", err)
		}

		telegram := backendinternal.ParseJSONMap(user.Get("telegram"))
		return e.JSON(http.StatusOK, map[string]any{
			"id":                user.Id,
			"email":             user.GetString("email"),
			"full_name":         groupinternal.UserDisplayName(user),
			"telegram":          telegram,
			"guardian_requests": guardianRequests,
		})
	}
}

func sharedGroupCount(app *pocketbase.PocketBase, actorID string, targetID string) (int, error) {
	groupIDs, err := sharedGroupIDs(app, actorID, targetID)
	if err != nil {
		return 0, err
	}
	return len(groupIDs), nil
}

func sharedGroupIDs(app *pocketbase.PocketBase, actorID string, targetID string) (map[string]struct{}, error) {
	actorMemberships, err := app.FindRecordsByFilter("user_groups", "user = {:user}", "", 500, 0, map[string]any{"user": actorID})
	if err != nil {
		return nil, err
	}

	actorGroups := make(map[string]struct{}, len(actorMemberships))
	for _, rel := range actorMemberships {
		groupID := strings.TrimSpace(rel.GetString("group"))
		if groupID != "" {
			actorGroups[groupID] = struct{}{}
		}
	}

	targetMemberships, err := app.FindRecordsByFilter("user_groups", "user = {:user}", "", 500, 0, map[string]any{"user": targetID})
	if err != nil {
		return nil, err
	}

	shared := map[string]struct{}{}
	for _, rel := range targetMemberships {
		groupID := strings.TrimSpace(rel.GetString("group"))
		if groupID == "" {
			continue
		}
		if _, ok := actorGroups[groupID]; ok {
			shared[groupID] = struct{}{}
		}
	}
	return shared, nil
}
