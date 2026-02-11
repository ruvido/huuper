package api

import (
	"net/http"
	"strings"

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

func GroupMembersHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor := e.Auth
		if actor == nil {
			return apis.NewUnauthorizedError("Unauthorized", nil)
		}

		group, err := findGroupByPathID(app, e)
		if err != nil {
			return err
		}
		if err := requireGroupVisibility(app, actor, group); err != nil {
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
		actor := e.Auth
		if actor == nil {
			return apis.NewUnauthorizedError("Unauthorized", nil)
		}

		group, err := findGroupByPathID(app, e)
		if err != nil {
			return err
		}
		if err := requireGroupVisibility(app, actor, group); err != nil {
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

func requireGroupVisibility(app *pocketbase.PocketBase, actor *core.Record, group *core.Record) error {
	if actor == nil {
		return apis.NewUnauthorizedError("Unauthorized", nil)
	}
	if group == nil {
		return apis.NewBadRequestError("invalid_group", nil)
	}
	if actor.GetBool("admin") {
		return nil
	}
	if isAssistantForGroup(actor, group) {
		return nil
	}
	ok, err := isMemberOfGroup(app, actor.Id, group.Id)
	if err != nil {
		return apis.NewBadRequestError("failed_group_access_check", err)
	}
	if !ok {
		return apis.NewForbiddenError("forbidden_group", nil)
	}
	return nil
}

func isAssistantForGroup(actor *core.Record, group *core.Record) bool {
	if actor == nil || group == nil {
		return false
	}
	return strings.TrimSpace(group.GetString("assistant")) == actor.Id
}

func isMemberOfGroup(app *pocketbase.PocketBase, userID string, groupID string) (bool, error) {
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

func userDisplayName(user *core.Record) string {
	if user == nil {
		return ""
	}
	data := parseJSONMap(user.Get("data"))
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
