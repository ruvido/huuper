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
			"group = {:group} && archived = false",
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

		requestsVisible := true
		requestsCount := 0
		pendingRequests := []groupinternal.PendingRequestItem{}
		if requestsVisible {
			requests, err := app.FindRecordsByFilter("requests", "group = {:group} && archived = false", "", 500, 0, map[string]any{"group": group.Id})
			if err != nil {
				return apis.NewBadRequestError("failed_group_requests", err)
			}
			requestsCount = len(requests)
			pendingRequests = make([]groupinternal.PendingRequestItem, 0, len(requests))
			for _, record := range requests {
				item, err := backendrequests.MapItemWithWorkflow(app, actor, record)
				if err != nil {
					return apis.NewBadRequestError("failed_group_request_workflow", err)
				}
				statusLabel := strings.TrimSpace(backendinternal.AnyToString(item.Workflow["pending_action_label"]))
				if statusLabel == "" {
					flow, err := backendrequests.LoadFlowForRequest(app, item.Data)
					if err != nil {
						return apis.NewBadRequestError("invalid_requests_flow_settings", err)
					}
					stepIndex := backendrequests.EffectiveStepIndex(record, item.Data, flow)
					statusLabel = requestStatusLabel(item.Status, stepIndex, flow)
				}
				pendingRequests = append(pendingRequests, groupinternal.PendingRequestItem{
					ID:          item.ID,
					FullName:    strings.TrimSpace(backendrequests.DisplayName(item.Data, strings.TrimSpace(item.Email), item.ID)),
					Email:       strings.TrimSpace(item.Email),
					Status:      strings.TrimSpace(item.Status),
					StatusLabel: statusLabel,
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
			"members_count":    len(membersResponse.Items),
			"requests_visible": requestsVisible,
			"requests_count":   requestsCount,
			"pending_requests": pendingRequests,
			"assistant":        group.GetString("assistant"),
			"members":          membersResponse.Items,
			"guardians":        guardiansResponse.Items,
		})
	}
}

func GroupAssistantHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
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
		if !actor.GetBool("admin") && !backendinternal.IsAssistantForGroup(actor, group) {
			return apis.NewForbiddenError("Forbidden", nil)
		}

		var payload struct {
			Assistant string `json:"assistant"`
		}
		if err := e.BindBody(&payload); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}

		if err := groupinternal.UpdateAssistant(app, group.Id, payload.Assistant); err != nil {
			return err
		}

		return e.JSON(http.StatusOK, map[string]any{
			"ok":        true,
			"group_id":  group.Id,
			"assistant": strings.TrimSpace(payload.Assistant),
		})
	}
}
