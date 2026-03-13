package admin

import (
	"net/http"

	meapi "members/backend/api/me"
	"members/backend/bot"

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
