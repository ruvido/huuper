package admin

import (
	"net/http"

	"members/bot"
	backendinternal "members/internal"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func SyncGroupMembershipsHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if _, err := backendinternal.RequireAdmin(e); err != nil {
			return err
		}

		if bot.GetBot() == nil {
			return apis.NewBadRequestError("telegram_bot_unavailable", nil)
		}

		bot.SyncAllUsersMemberships()

		return e.JSON(http.StatusOK, map[string]any{
			"ok": true,
		})
	}
}
