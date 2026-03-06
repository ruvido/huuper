package api

import (
	"net/http"

	"members/bot"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// AdminSyncGroupMembershipsHandler triggers a full Telegram membership sync for all connected users.
func AdminSyncGroupMembershipsHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		_, err := requireAdmin(e)
		if err != nil {
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
