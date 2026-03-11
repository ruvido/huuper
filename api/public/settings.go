package public

import (
	settinginternal "members/internal/settings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func SettingsHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		name := e.Request.PathValue("name")
		if !settinginternal.PublicNames[name] {
			return apis.NewNotFoundError("Setting not found", nil)
		}
		record, data, err := settinginternal.GetVisible(app, name)
		if err != nil {
			return err
		}
		return settinginternal.WriteJSON(e, record.GetString("name"), data)
	}
}
