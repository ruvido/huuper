package api

import (
	"net/http"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func unwrapSettingData(raw any) map[string]any {
	data := parseJSONMap(raw)
	if nested, ok := data["data"].(map[string]any); ok && nested != nil {
		return nested
	}
	return data
}

// GetSettingsHandler returns settings by name
func GetSettingsHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		name := e.Request.PathValue("name")
		if name == "" {
			return apis.NewNotFoundError("Setting not found", nil)
		}

		publicNames := map[string]bool{
			"title":          true,
			"password_reset": true,
			"signup":         true,
		}

		if !publicNames[name] && e.Auth == nil {
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

		settingData := unwrapSettingData(record.Get("data"))

		if name == "telegram" {
			var telegramData struct {
				Name string `json:"name"`
			}
			if value, ok := settingData["name"].(string); ok {
				telegramData.Name = value
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
			"data": settingData,
		})
	}
}
