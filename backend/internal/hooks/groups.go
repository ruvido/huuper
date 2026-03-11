package hooks

import (
	"fmt"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterGroupsValidation enforces groups-level integrity rules.
func RegisterGroupsValidation(app *pocketbase.PocketBase) {
	app.OnRecordValidate("groups").BindFunc(func(e *core.RecordEvent) error {
		groupType := strings.TrimSpace(e.Record.GetString("type"))
		if groupType != "default" {
			return e.Next()
		}

		currentID := strings.TrimSpace(e.Record.Id)
		filter := "type = 'default'"
		params := map[string]any{}
		if currentID != "" {
			filter = "type = 'default' && id != {:id}"
			params["id"] = currentID
		}

		existing, err := e.App.FindFirstRecordByFilter("groups", filter, params)
		if err == nil && existing != nil {
			return apis.NewBadRequestError("invalid_group_type", fmt.Errorf("only one default group is allowed"))
		}

		return e.Next()
	})
}
