package me

import (
	"net/http"

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

		response, err := groupinternal.MembersResponseForGroup(app, group.Id)
		if err != nil {
			return apis.NewBadRequestError("failed_guardians", err)
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
