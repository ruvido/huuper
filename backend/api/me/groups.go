package me

import (
	"net/http"

	backendinternal "members/backend/internal"
	groupinternal "members/backend/internal/groups"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func GroupsListHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor, err := backendinternal.RequireAuthenticatedActor(e)
		if err != nil {
			return err
		}

		items, err := groupinternal.ListForUser(app, actor.Id)
		if err != nil {
			return apis.NewBadRequestError("failed_groups", err)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"items": items,
		})
	}
}

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

		response, err := groupinternal.MembersResponseForGroup(app, group.Id)
		if err != nil {
			return apis.NewBadRequestError("failed_members", err)
		}
		return e.JSON(http.StatusOK, response)
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

		response, err := groupinternal.GuardiansResponseForGroup(app, group.Id)
		if err != nil {
			return apis.NewBadRequestError("failed_guardians", err)
		}
		return e.JSON(http.StatusOK, response)
	}
}

func GroupGetHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
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

		members, err := app.FindRecordsByFilter("user_groups", "group = {:group}", "", 500, 0, map[string]any{"group": group.Id})
		if err != nil {
			return apis.NewBadRequestError("failed_group_memberships", err)
		}

		requestsVisible := backendinternal.IsAssistantForGroup(actor, group)
		requestsCount := 0
		if requestsVisible {
			requests, err := app.FindRecordsByFilter("requests", "group = {:group} && rejected = false", "", 500, 0, map[string]any{"group": group.Id})
			if err != nil {
				return apis.NewBadRequestError("failed_group_requests", err)
			}
			requestsCount = len(requests)
		}

		membersResponse, err := groupinternal.MembersResponseForGroup(app, group.Id)
		if err != nil {
			return apis.NewBadRequestError("failed_group_members", err)
		}

		guardiansResponse, err := groupinternal.GuardiansResponseForGroup(app, group.Id)
		if err != nil {
			return apis.NewBadRequestError("failed_group_guardians", err)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"id":               group.Id,
			"name":             group.GetString("name"),
			"type":             group.GetString("type"),
			"description":      group.GetString("description"),
			"members_count":    len(members),
			"requests_visible": requestsVisible,
			"requests_count":   requestsCount,
			"assistant":        group.GetString("assistant"),
			"members":          membersResponse.Items,
			"guardians":        guardiansResponse.Items,
		})
	}
}
