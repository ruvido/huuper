package me

import (
	"net/http"
	"strings"

	backendinternal "members/internal"
	groupinternal "members/internal/groups"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func GroupRequestsCountHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor, err := backendinternal.RequireAuthenticatedActor(e)
		if err != nil {
			return err
		}

		group, err := groupinternal.FindByPathID(app, e)
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

		group, err := groupinternal.FindByPathID(app, e)
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
		guardianCounts, err := groupinternal.GuardianCounts(app, group.Id)
		if err != nil {
			return apis.NewBadRequestError("failed_guardians", err)
		}

		items := make([]groupinternal.MemberItem, 0, len(relations))
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
			items = append(items, groupinternal.MemberItem{
				ID:         user.Id,
				Email:      user.GetString("email"),
				FullName:   groupinternal.UserDisplayName(user),
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

		group, err := groupinternal.FindByPathID(app, e)
		if err != nil {
			return err
		}
		if err := backendinternal.RequireGroupVisibility(app, actor, group); err != nil {
			return err
		}

		guardianCounts, err := groupinternal.GuardianCounts(app, group.Id)
		if err != nil {
			return apis.NewBadRequestError("failed_guardians", err)
		}

		items := make([]groupinternal.GuardianItem, 0, len(guardianCounts))
		for userID, protegesCount := range guardianCounts {
			user, err := app.FindRecordById("users", userID)
			if err != nil || user == nil {
				continue
			}

			items = append(items, groupinternal.GuardianItem{
				ID:            user.Id,
				Email:         user.GetString("email"),
				FullName:      groupinternal.UserDisplayName(user),
				ProtegesCount: protegesCount,
			})
		}

		return e.JSON(http.StatusOK, map[string]any{
			"group_id": group.Id,
			"items":    items,
		})
	}
}
