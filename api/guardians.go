package api

import (
	"net/http"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

type guardianApprovalRequest struct {
	Request string `json:"request"`
	Group   string `json:"group"`
}

// LeaderApproveGuardianHandler marks leader approval for a guardian record.
func LeaderApproveGuardianHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		authRecord := e.Auth
		if authRecord == nil {
			return apis.NewUnauthorizedError("Unauthorized", nil)
		}

		var payload guardianApprovalRequest
		if err := e.BindBody(&payload); err != nil {
			return apis.NewBadRequestError("Invalid request", err)
		}

		if payload.Request == "" {
			return apis.NewBadRequestError("Missing request", nil)
		}

		record, err := app.FindFirstRecordByFilter(
			"guardians",
			"request = {:request}",
			map[string]any{
				"request": payload.Request,
			},
		)
		if err != nil || record == nil {
			return apis.NewNotFoundError("Guardian record not found", err)
		}

		groupID := record.GetString("group")
		if payload.Group != "" && payload.Group != groupID {
			return apis.NewBadRequestError("User assigned to a different group", nil)
		}

		group, err := app.FindRecordById("groups", groupID)
		if err != nil {
			return apis.NewNotFoundError("Group not found", err)
		}

		if group.GetString("leader") != authRecord.Id {
			return apis.NewForbiddenError("Forbidden", nil)
		}

		if record.GetDateTime("leader_approved_at").IsZero() {
			record.Set("leader_approved_at", types.NowDateTime())
		}

		if err := app.Save(record); err != nil {
			return apis.NewBadRequestError("Failed to save approval", err)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"id":                 record.Id,
			"request":            record.GetString("request"),
			"group":              record.GetString("group"),
			"leader_approved_at": record.GetString("leader_approved_at"),
			"admin_confirmed_at": record.GetString("admin_confirmed_at"),
		})
	}
}

// AdminConfirmGuardianHandler marks admin confirmation for a guardian record.
func AdminConfirmGuardianHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		authRecord := e.Auth
		if authRecord == nil {
			return apis.NewUnauthorizedError("Unauthorized", nil)
		}

		if !authRecord.GetBool("admin") {
			return apis.NewForbiddenError("Forbidden", nil)
		}

		var payload guardianApprovalRequest
		if err := e.BindBody(&payload); err != nil {
			return apis.NewBadRequestError("Invalid request", err)
		}

		if payload.Request == "" {
			return apis.NewBadRequestError("Missing request", nil)
		}

		record, err := app.FindFirstRecordByFilter(
			"guardians",
			"request = {:request}",
			map[string]any{
				"request": payload.Request,
			},
		)
		if err != nil || record == nil {
			return apis.NewNotFoundError("Guardian record not found", err)
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
			"request":            record.GetString("request"),
			"group":              record.GetString("group"),
			"leader_approved_at": record.GetString("leader_approved_at"),
			"admin_confirmed_at": record.GetString("admin_confirmed_at"),
		})
	}
}
