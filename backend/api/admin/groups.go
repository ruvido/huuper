package admin

import (
	"net/http"
	"strings"

	meapi "members/backend/api/me"
	"members/backend/bot"
	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func SyncGroupMembershipsHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if bot.GetBot() == nil {
			return apis.NewBadRequestError("telegram_bot_unavailable", nil)
		}

		bot.SyncAllUsersMemberships()

		return e.JSON(http.StatusOK, map[string]any{
			"ok": true,
		})
	}
}

func GroupGetHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return meapi.GroupGetHandler(app)
}

func GroupAssistantHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if _, err := backendinternal.RequireAdmin(e); err != nil {
			return err
		}

		groupID := strings.TrimSpace(e.Request.PathValue("id"))
		if groupID == "" {
			return apis.NewBadRequestError("invalid_group", nil)
		}

		group, err := app.FindRecordById("groups", groupID)
		if err != nil || group == nil {
			return apis.NewNotFoundError("group_not_found", err)
		}

		var payload struct {
			Assistant string `json:"assistant"`
		}
		if err := e.BindBody(&payload); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}

		assistantID := strings.TrimSpace(payload.Assistant)
		if assistantID == "" {
			return apis.NewBadRequestError("invalid_assistant", nil)
		}

		assistant, err := app.FindRecordById("users", assistantID)
		if err != nil || assistant == nil {
			return apis.NewNotFoundError("assistant_not_found", err)
		}

		ok, err := backendinternal.IsMemberOfGroup(app, assistantID, groupID)
		if err != nil {
			return apis.NewBadRequestError("failed_group_membership_check", err)
		}
		if !ok {
			return apis.NewBadRequestError("assistant_must_be_group_member", nil)
		}

		group.Set("assistant", assistantID)
		if err := app.Save(group); err != nil {
			return apis.NewBadRequestError("failed_update_group", err)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"ok":        true,
			"group_id":  groupID,
			"assistant": assistantID,
		})
	}
}
