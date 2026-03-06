package api

import (
	"encoding/json"
	"strconv"
	"strings"
)

func anyToString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case json.Number:
		return typed.String()
	case int:
		return strconv.Itoa(typed)
	case int32:
		return strconv.FormatInt(int64(typed), 10)
	case int64:
		return strconv.FormatInt(typed, 10)
	case float64:
		return strconv.FormatInt(int64(typed), 10)
	default:
		return ""
	}
}

func anyToInt64(value any) (int64, bool) {
	switch typed := value.(type) {
	case int:
		return int64(typed), true
	case int32:
		return int64(typed), true
	case int64:
		return typed, true
	case float64:
		return int64(typed), true
	case json.Number:
		n, err := typed.Int64()
		return n, err == nil
	case string:
		raw := strings.TrimSpace(typed)
		if raw == "" {
			return 0, false
		}
		n, err := strconv.ParseInt(raw, 10, 64)
		return n, err == nil
	default:
		return 0, false
	}
}
