package admin

import (
	"net/http"

	backendinternal "members/backend/internal"
	eventinternal "members/backend/internal/events"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func GroupsListHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		records, err := app.FindRecordsByFilter("groups", "", "name", 500, 0)
		if err != nil {
			return apis.NewBadRequestError("failed_groups", err)
		}

		items := make([]map[string]any, 0, len(records))
		for _, record := range records {
			items = append(items, map[string]any{
				"id":          record.Id,
				"name":        record.GetString("name"),
				"type":        record.GetString("type"),
				"description": record.GetString("description"),
				"assistant":   record.GetString("assistant"),
			})
		}

		return e.JSON(http.StatusOK, map[string]any{"items": items})
	}
}

func EventsListHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor, err := backendinternal.RequireAuthenticatedActor(e)
		if err != nil {
			return err
		}
		window := e.Request.URL.Query().Get("window")
		if window == "" {
			window = eventinternal.WindowFuture
		}
		items, err := eventinternal.ListForUser(app, actor, window)
		if err != nil {
			return apis.NewBadRequestError("failed_events", err)
		}
		return e.JSON(http.StatusOK, map[string]any{"items": items})
	}
}
