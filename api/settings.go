package api

import (
	"net/http"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// GetSettingsHandler returns settings by name
func GetSettingsHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		name := e.Request.PathValue("name")

		record, err := app.FindFirstRecordByFilter(
			"settings",
			"name = {:name}",
			map[string]any{
				"name": name,
			},
		)

		if err != nil {
			return apis.NewNotFoundError("Setting not found", err)
		}

		return e.JSON(http.StatusOK, map[string]interface{}{
			"name": record.GetString("name"),
			"data": record.Get("data"),
		})
	}
}
