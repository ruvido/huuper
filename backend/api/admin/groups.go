package admin

import (
	"net/http"
	"strings"

	meapi "members/backend/api/me"
	"members/backend/bot"
	backendinternal "members/backend/internal"
	groupinternal "members/backend/internal/groups"

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
		var payload struct {
			Assistant string `json:"assistant"`
		}
		if err := e.BindBody(&payload); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}

		if err := groupinternal.UpdateAssistant(app, groupID, payload.Assistant); err != nil {
			return err
		}

		return e.JSON(http.StatusOK, map[string]any{
			"ok":        true,
			"group_id":  groupID,
			"assistant": strings.TrimSpace(payload.Assistant),
		})
	}
}
