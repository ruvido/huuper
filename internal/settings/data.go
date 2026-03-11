package settings

import (
	"net/http"
	"strings"

	backendinternal "members/internal"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

type Scope string

const (
	ScopePublic Scope = "public"
	ScopeMember Scope = "member"
	ScopeAdmin  Scope = "admin"
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

func GetVisibleForScope(app *pocketbase.PocketBase, name string, scope Scope) (*core.Record, map[string]any, error) {
	rawName := strings.TrimSpace(name)
	if !isVisibleInScope(rawName, scope) {
		return nil, nil, apis.NewNotFoundError("Setting not found", nil)
	}
	return GetVisible(app, rawName)
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

func isVisibleInScope(name string, scope Scope) bool {
	switch scope {
	case ScopePublic:
		return PublicNames[name]
	case ScopeMember:
		return MemberNames[name]
	case ScopeAdmin:
		return true
	default:
		return false
	}
}
