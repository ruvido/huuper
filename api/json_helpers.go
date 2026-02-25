package api

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/tools/types"
)

func parseJSONMap(raw any) map[string]any {
	data := map[string]any{}
	if raw == nil {
		return data
	}

	switch typed := raw.(type) {
	case map[string]any:
		data = typed
	case types.JSONRaw:
		_ = json.Unmarshal(typed, &data)
	case string:
		_ = json.Unmarshal([]byte(typed), &data)
	case []byte:
		_ = json.Unmarshal(typed, &data)
	default:
		if payload, err := json.Marshal(typed); err == nil {
			_ = json.Unmarshal(payload, &data)
		}
	}

	return data
}
