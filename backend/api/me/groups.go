package me

import (
	"net/http"
	"strings"

	backendinternal "members/backend/internal"
	groupinternal "members/backend/internal/groups"
	backendrequests "members/backend/internal/requests"

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
		if actor.GetBool("admin") {
			requestsVisible = true
		}
		requestsCount := 0
		pendingRequests := []groupinternal.PendingRequestItem{}
		if requestsVisible {
			requests, err := app.FindRecordsByFilter("requests", "group = {:group} && rejected = false", "", 500, 0, map[string]any{"group": group.Id})
			if err != nil {
				return apis.NewBadRequestError("failed_group_requests", err)
			}
			requestsCount = len(requests)
			flow, err := backendrequests.LoadFlowSettings(app)
			if err != nil {
				return apis.NewBadRequestError("invalid_requests_flow_settings", err)
			}
			pendingRequests = make([]groupinternal.PendingRequestItem, 0, len(requests))
			for _, record := range requests {
				item, err := backendrequests.MapItemWithWorkflow(app, actor, record, flow)
				if err != nil {
					return apis.NewBadRequestError("failed_group_request_workflow", err)
				}
				pendingRequests = append(pendingRequests, groupinternal.PendingRequestItem{
					ID:          item.ID,
					FullName:    strings.TrimSpace(backendrequests.DisplayName(item.Data, strings.TrimSpace(item.Email), item.ID)),
					Email:       strings.TrimSpace(item.Email),
					Status:      strings.TrimSpace(item.Status),
					StatusLabel: requestStatusLabel(item.Status, item.StepIndex, flow),
					Created:     strings.TrimSpace(item.Created),
					AssignedAt:  requestAssignedAt(item.Data),
					Data:        item.Data,
					Workflow:    item.Workflow,
				})
			}
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
			"pending_requests": pendingRequests,
			"assistant":        group.GetString("assistant"),
			"members":          membersResponse.Items,
			"guardians":        guardiansResponse.Items,
		})
	}
}
