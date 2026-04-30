package admin

import (
	backendinternal "members/backend/internal"
	settinginternal "members/backend/internal/settings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func SettingsHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		record, data, err := settinginternal.GetVisibleForScope(app, e.Request.PathValue("name"), settinginternal.ScopeAdmin)
		if err != nil {
			return err
		}
		if title, ok := data["title"].(string); ok {
			if html, rendered := backendinternal.RenderInlineMarkdown(title); rendered {
				data["title"] = html
			}
		}
		return settinginternal.WriteJSON(e, record.GetString("name"), data)
	}
}
