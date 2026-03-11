package settings

import (
	"net/http"
	"strings"

	backendinternal "members/internal"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

var PublicNames = map[string]bool{
	"title":          true,
	"password_reset": true,
	"signup":         true,
}

var MemberNames = map[string]bool{
	"onboarding":       true,
	"profile_schema":   true,
	"telegram":         true,
	"telegram_connect": true,
}

func Unwrap(raw any) map[string]any {
	return backendinternal.UnwrapSettingData(raw)
}

func GetVisible(app *pocketbase.PocketBase, name string) (*core.Record, map[string]any, error) {
	rawName := strings.TrimSpace(name)
	if rawName == "" {
		return nil, nil, apis.NewNotFoundError("Setting not found", nil)
	}

	record, err := app.FindFirstRecordByFilter(
		"settings",
		"name = {:name}",
		map[string]any{"name": rawName},
	)
	if err != nil {
		return nil, nil, apis.NewNotFoundError("Setting not found", err)
	}

	return record, sanitize(rawName, Unwrap(record.Get("data"))), nil
}

func WriteJSON(e *core.RequestEvent, name string, data map[string]any) error {
	return e.JSON(http.StatusOK, map[string]any{
		"name": name,
		"data": data,
	})
}

func sanitize(name string, data map[string]any) map[string]any {
	if name != "telegram" {
		return data
	}
	safe := map[string]any{}
	if value, ok := data["name"].(string); ok {
		safe["name"] = value
	}
	return safe
}
