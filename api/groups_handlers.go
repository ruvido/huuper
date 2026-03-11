package api

import (
	"net/http"
	"strings"

	backendinternal "members/internal"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

type groupMemberItem struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	FullName   string `json:"full_name"`
	Role       string `json:"role"`
	IsGuardian bool   `json:"is_guardian"`
}

type groupGuardianItem struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	FullName      string `json:"full_name"`
	ProtegesCount int    `json:"proteges_count"`
}

func GroupRequestsCountHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor, err := backendinternal.RequireAuthenticatedActor(e)
		if err != nil {
			return err
		}

		group, err := findGroupByPathID(app, e)
		if err != nil {
			return err
		}
		if err := backendinternal.RequireGroupVisibility(app, actor, group); err != nil {
			return err
		}

		records, err := app.FindRecordsByFilter(
			"requests",
			"group = {:group} && rejected = false",
			"",
			0,
			0,
			map[string]any{"group": group.Id},
		)
		if err != nil {
			return apis.NewBadRequestError("failed_requests_count", err)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"group_id": group.Id,
			"count":    len(records),
		})
	}
}

func GroupMembersHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor, err := backendinternal.RequireAuthenticatedActor(e)
		if err != nil {
			return err
		}

		group, err := findGroupByPathID(app, e)
		if err != nil {
			return err
		}
		if err := backendinternal.RequireGroupVisibility(app, actor, group); err != nil {
			return err
		}

		relations, err := app.FindRecordsByFilter(
			"user_groups",
			"group = {:group}",
			"created",
			500,
			0,
			map[string]any{"group": group.Id},
		)
		if err != nil {
			return apis.NewBadRequestError("failed_members", err)
		}
		guardianCounts, err := guardianCountsForGroup(app, group.Id)
		if err != nil {
			return apis.NewBadRequestError("failed_guardians", err)
		}

		items := make([]groupMemberItem, 0, len(relations))
		for _, rel := range relations {
			userID := strings.TrimSpace(rel.GetString("user"))
			if userID == "" {
				continue
			}
			user, err := app.FindRecordById("users", userID)
			if err != nil || user == nil {
				continue
			}
			role := strings.TrimSpace(rel.GetString("role"))
			_, isGuardian := guardianCounts[userID]
			items = append(items, groupMemberItem{
				ID:         user.Id,
				Email:      user.GetString("email"),
				FullName:   userDisplayName(user),
				Role:       role,
				IsGuardian: isGuardian,
			})
		}

		return e.JSON(http.StatusOK, map[string]any{
			"group_id": group.Id,
			"items":    items,
		})
	}
}

func GroupGuardiansHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor, err := backendinternal.RequireAuthenticatedActor(e)
		if err != nil {
			return err
		}

		group, err := findGroupByPathID(app, e)
		if err != nil {
			return err
		}
		if err := backendinternal.RequireGroupVisibility(app, actor, group); err != nil {
			return err
		}

		guardianCounts, err := guardianCountsForGroup(app, group.Id)
		if err != nil {
			return apis.NewBadRequestError("failed_guardians", err)
		}

		items := make([]groupGuardianItem, 0, len(guardianCounts))
		for userID, protegesCount := range guardianCounts {
			user, err := app.FindRecordById("users", userID)
			if err != nil || user == nil {
				continue
			}

			items = append(items, groupGuardianItem{
				ID:            user.Id,
				Email:         user.GetString("email"),
				FullName:      userDisplayName(user),
				ProtegesCount: protegesCount,
			})
		}

		return e.JSON(http.StatusOK, map[string]any{
			"group_id": group.Id,
			"items":    items,
		})
	}
}

func findGroupByPathID(app *pocketbase.PocketBase, e *core.RequestEvent) (*core.Record, error) {
	id := strings.TrimSpace(e.Request.PathValue("id"))
	if id == "" {
		return nil, apis.NewBadRequestError("invalid_group", nil)
	}
	group, err := app.FindRecordById("groups", id)
	if err != nil || group == nil {
		return nil, apis.NewNotFoundError("group_not_found", err)
	}
	return group, nil
}

func userDisplayName(user *core.Record) string {
	if user == nil {
		return ""
	}
	data := backendinternal.ParseJSONMap(user.Get("data"))
	if fullName, ok := data["full_name"].(string); ok && strings.TrimSpace(fullName) != "" {
		return strings.TrimSpace(fullName)
	}
	return strings.TrimSpace(user.GetString("email"))
}

func guardianCountsForGroup(app *pocketbase.PocketBase, groupID string) (map[string]int, error) {
	records, err := app.FindRecordsByFilter(
		"requests",
		"group = {:group} && guardian != '' && rejected = false",
		"",
		500,
		0,
		map[string]any{"group": groupID},
	)
	if err != nil {
		return nil, err
	}

	out := map[string]int{}
	for _, record := range records {
		guardianID := strings.TrimSpace(record.GetString("guardian"))
		if guardianID == "" {
			continue
		}
		out[guardianID]++
	}
	return out, nil
}
