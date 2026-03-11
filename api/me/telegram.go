package me

import (
	"net/http"

	backendinternal "members/internal"
	tginternal "members/internal/telegram"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func GenerateTelegramTokenHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		authRecord, err := backendinternal.RequireAuthenticatedActor(e)
		if err != nil {
			return err
		}
		token, err := tginternal.GenerateUserToken(app, authRecord)
		if err != nil {
			return err
		}
		return e.JSON(http.StatusOK, map[string]any{"token": token})
	}
}

func DefaultGroupInviteHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor, err := backendinternal.RequireAuthenticatedActor(e)
		if err != nil {
			return err
		}
		groupID, link, err := tginternal.DefaultGroupInvite(app, actor)
		if err != nil {
			return err
		}
		return e.JSON(http.StatusOK, map[string]any{
			"group_id":    groupID,
			"invite_link": link,
		})
	}
}
