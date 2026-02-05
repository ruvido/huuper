package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

type adminEventNext struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	EventDate     string `json:"event_date"`
	Registrations int    `json:"registrations"`
	Pending       int    `json:"pending"`
}

// AdminSummaryHandler returns compact stats for the admin dashboard.
func AdminSummaryHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if _, err := requireAdmin(e); err != nil {
			return err
		}

		users, err := app.FindRecordsByFilter("users", "", "", 0, 0)
		if err != nil {
			return apis.NewBadRequestError("failed_users", err)
		}
		telegramMissing := 0
		notActive := 0
		for _, user := range users {
			if user.GetString("status") != "active" {
				notActive++
			}
			rawTelegram := user.Get("telegram")
			if isTelegramMissing(rawTelegram) {
				telegramMissing++
			}
		}
		groups, err := app.FindRecordsByFilter("groups", "", "", 0, 0)
		if err != nil {
			return apis.NewBadRequestError("failed_groups", err)
		}
		eventCount, err := app.FindRecordsByFilter("events", "", "", 0, 0)
		if err != nil {
			return apis.NewBadRequestError("failed_events_count", err)
		}

		today := time.Now().Format("2006-01-02")
		eventFilter := fmt.Sprintf("event_date >= '%s'", today)
		events, err := app.FindRecordsByFilter("events", eventFilter, "event_date", 1, 0)
		if err != nil {
			return apis.NewBadRequestError("failed_events", err)
		}

		var next *adminEventNext
		if len(events) > 0 {
			event := events[0]
			eventID := event.Id
			registrations, err := app.FindRecordsByFilter(
				"event_registrations",
				fmt.Sprintf("event = '%s'", eventID),
				"",
				0,
				0,
			)
			if err != nil {
				return apis.NewBadRequestError("failed_registrations", err)
			}

			pending, err := app.FindRecordsByFilter(
				"event_registrations",
				fmt.Sprintf("event = '%s' && accepted = false", eventID),
				"",
				0,
				0,
			)
			if err != nil {
				return apis.NewBadRequestError("failed_pending", err)
			}

			next = &adminEventNext{
				ID:            eventID,
				Title:         event.GetString("title"),
				EventDate:     event.GetString("event_date"),
				Registrations: len(registrations),
				Pending:       len(pending),
			}
		}

		return e.JSON(http.StatusOK, map[string]any{
			"users": map[string]any{
				"total":      len(users),
				"noTelegram": telegramMissing,
				"notActive":  notActive,
			},
			"groups": map[string]any{
				"total": len(groups),
			},
			"events": map[string]any{
				"total": len(eventCount),
				"next":  next,
			},
		})
	}
}

func isTelegramMissing(raw any) bool {
	if raw == nil {
		return true
	}

	switch v := raw.(type) {
	case types.JSONRaw:
		return isNullJSON(string(bytes.TrimSpace(v)))
	case string:
		return isNullJSON(v)
	case []byte:
		return isNullJSON(string(bytes.TrimSpace(v)))
	default:
		return false
	}
}

func isNullJSON(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || trimmed == "null" {
		return true
	}
	var parsed any
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		return false
	}
	return parsed == nil
}
