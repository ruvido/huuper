package api

import (
	"net/http"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

type approvalRequest struct {
	User  string `json:"user"`
	Group string `json:"group"`
}

// LeaderApproveHandler marks a user as leader-approved for a group.
func LeaderApproveHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		authRecord := e.Auth
		if authRecord == nil {
			return apis.NewUnauthorizedError("Unauthorized", nil)
		}

		var payload approvalRequest
		if err := e.BindBody(&payload); err != nil {
			return apis.NewBadRequestError("Invalid request", err)
		}

		if payload.User == "" || payload.Group == "" {
			return apis.NewBadRequestError("Missing user or group", nil)
		}

		group, err := app.FindRecordById("groups", payload.Group)
		if err != nil {
			return apis.NewNotFoundError("Group not found", err)
		}

		if group.GetString("leader") != authRecord.Id {
			return apis.NewForbiddenError("Forbidden", nil)
		}

		if _, err := app.FindRecordById("users", payload.User); err != nil {
			return apis.NewNotFoundError("User not found", err)
		}

		approvals, err := app.FindCollectionByNameOrId("approvals")
		if err != nil {
			return apis.NewNotFoundError("Approvals collection not found", err)
		}

		record, err := app.FindFirstRecordByFilter(
			"approvals",
			"user = {:user}",
			map[string]any{
				"user": payload.User,
			},
		)

		if err != nil || record == nil {
			record = core.NewRecord(approvals)
			record.Set("user", payload.User)
			record.Set("group", payload.Group)
		} else if record.GetString("group") != payload.Group {
			return apis.NewBadRequestError("User assigned to a different group", nil)
		}

		if record.GetDateTime("leader_approved_at").IsZero() {
			record.Set("leader_approved_at", types.NowDateTime())
		}

		if err := app.Save(record); err != nil {
			return apis.NewBadRequestError("Failed to save approval", err)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"id":                 record.Id,
			"user":               record.GetString("user"),
			"group":              record.GetString("group"),
			"leader_approved_at": record.GetString("leader_approved_at"),
			"admin_confirmed_at": record.GetString("admin_confirmed_at"),
		})
	}
}

// AdminApproveHandler marks a user as admin-confirmed.
func AdminApproveHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		authRecord := e.Auth
		if authRecord == nil {
			return apis.NewUnauthorizedError("Unauthorized", nil)
		}

		if !authRecord.GetBool("admin") {
			return apis.NewForbiddenError("Forbidden", nil)
		}

		var payload approvalRequest
		if err := e.BindBody(&payload); err != nil {
			return apis.NewBadRequestError("Invalid request", err)
		}

		if payload.User == "" {
			return apis.NewBadRequestError("Missing user", nil)
		}

		record, err := app.FindFirstRecordByFilter(
			"approvals",
			"user = {:user}",
			map[string]any{
				"user": payload.User,
			},
		)
		if err != nil || record == nil {
			return apis.NewNotFoundError("Approval record not found", err)
		}

		if record.GetDateTime("leader_approved_at").IsZero() {
			return apis.NewBadRequestError("Leader approval required", nil)
		}

		if record.GetDateTime("admin_confirmed_at").IsZero() {
			record.Set("admin_confirmed_at", types.NowDateTime())
		}

		if err := app.Save(record); err != nil {
			return apis.NewBadRequestError("Failed to save approval", err)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"id":                 record.Id,
			"user":               record.GetString("user"),
			"group":              record.GetString("group"),
			"leader_approved_at": record.GetString("leader_approved_at"),
			"admin_confirmed_at": record.GetString("admin_confirmed_at"),
		})
	}
}
