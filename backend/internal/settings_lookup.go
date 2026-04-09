package internal

import (
	"fmt"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

func FindSettingData(app core.App, name string) (map[string]any, error) {
	record, err := app.FindFirstRecordByFilter(
		"settings",
		"name = {:name}",
		map[string]any{"name": strings.TrimSpace(name)},
	)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, fmt.Errorf("%s settings not found", strings.TrimSpace(name))
	}
	return UnwrapSettingData(record.Get("data")), nil
}
