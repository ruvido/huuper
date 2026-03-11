package api

import (
	"net/http"

	backendinternal "members/internal"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

var publicSettingsNames = map[string]bool{
	"title":          true,
	"password_reset": true,
	"signup":         true,
}

var memberSettingsNames = map[string]bool{
	"onboarding":       true,
	"profile_schema":   true,
	"telegram":         true,
	"telegram_connect": true,
}

func unwrapSettingData(raw any) map[string]any {
	return backendinternal.UnwrapSettingData(raw)
}

func requireSettingsVisibility(e *core.RequestEvent, name string) error {
	if publicSettingsNames[name] {
		return nil
	}

	actor, err := backendinternal.RequireAuthenticatedActor(e)
	if err != nil {
		return err
	}

	if memberSettingsNames[name] {
		return nil
	}

	if !actor.GetBool("admin") {
		return apis.NewForbiddenError("Forbidden", nil)
	}

	return nil
}

// GetSettingsHandler returns settings by name
func GetSettingsHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		name := e.Request.PathValue("name")
		if name == "" {
			return apis.NewNotFoundError("Setting not found", nil)
		}

		if err := requireSettingsVisibility(e, name); err != nil {
			return err
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
