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

		publicNames := map[string]bool{
			"title": true,
		}
		authNames := map[string]bool{
			"onboarding":       true,
			"telegram":         true,
			"telegram_connect": true,
			"welcome":          true,
		}

		if !publicNames[name] && !authNames[name] {
			return apis.NewNotFoundError("Setting not found", nil)
		}

		if authNames[name] && e.Auth == nil {
			return apis.NewUnauthorizedError("Unauthorized", nil)
		}

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

		if name == "telegram" {
			var telegramData struct {
				Name string `json:"name"`
			}
			if err := record.UnmarshalJSONField("data", &telegramData); err != nil {
				return apis.NewBadRequestError("Invalid setting data", err)
			}

			return e.JSON(http.StatusOK, map[string]interface{}{
				"name": record.GetString("name"),
				"data": map[string]interface{}{
					"name": telegramData.Name,
				},
			})
		}

		return e.JSON(http.StatusOK, map[string]interface{}{
			"name": record.GetString("name"),
			"data": record.Get("data"),
		})
	}
}
