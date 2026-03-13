package admin

import (
	"net/http"

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
			})
		}

		return e.JSON(http.StatusOK, map[string]any{"items": items})
	}
}

func EventsListHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		records, err := app.FindRecordsByFilter("events", "", "event_date", 500, 0)
		if err != nil {
			return apis.NewBadRequestError("failed_events", err)
		}

		items := make([]map[string]any, 0, len(records))
		for _, record := range records {
			items = append(items, map[string]any{
				"id":         record.Id,
				"title":      record.GetString("title"),
				"slug":       record.GetString("slug"),
				"event_date": record.GetString("event_date"),
			})
		}

		return e.JSON(http.StatusOK, map[string]any{"items": items})
	}
}
