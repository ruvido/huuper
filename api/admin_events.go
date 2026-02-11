package api

import (
	"net/http"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

type adminRegistrationItem struct {
	ID      string         `json:"id"`
	Email   string         `json:"email"`
	Status  string         `json:"status"`
	Created string         `json:"created"`
	HasUser bool           `json:"hasUser"`
	Data    map[string]any `json:"data"`
}

// AdminEventDetailsHandler returns the event and its registrations.
func AdminEventDetailsHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if _, err := requireAdmin(e); err != nil {
			return err
		}

		eventId := e.Request.PathValue("id")
		if eventId == "" {
			return apis.NewBadRequestError("invalid_event", nil)
		}

		event, err := app.FindRecordById("events", eventId)
		if err != nil || event == nil {
			return apis.NewNotFoundError("invalid_event", err)
		}

		registrations, err := app.FindRecordsByFilter(
			"event_registrations",
			"event = {:event}",
			"created",
			0,
			0,
			map[string]any{"event": eventId},
		)
		if err != nil {
			return apis.NewBadRequestError("failed_registrations", err)
		}

		items := make([]adminRegistrationItem, 0, len(registrations))
		for _, record := range registrations {
			items = append(items, mapRegistration(record))
		}

		eventData := parseJSONMap(event.Get("data"))

		return e.JSON(http.StatusOK, map[string]any{
			"event": map[string]any{
				"id":         event.Id,
				"title":      event.GetString("title"),
				"event_date": event.GetString("event_date"),
				"data":       eventData,
			},
			"registrations": items,
		})
	}
}

// AdminApproveRegistrationHandler marks a registration as accepted.
func AdminApproveRegistrationHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if _, err := requireAdmin(e); err != nil {
			return err
		}

		regId := e.Request.PathValue("id")
		if regId == "" {
			return apis.NewBadRequestError("invalid_registration", nil)
		}

		record, err := app.FindRecordById("event_registrations", regId)
		if err != nil || record == nil {
			return apis.NewNotFoundError("invalid_registration", err)
		}

		if record.GetString("status") == "active" {
			return e.JSON(http.StatusOK, map[string]any{"status": "already_accepted"})
		}

		record.Set("status", "active")
		if err := app.Save(record); err != nil {
			return apis.NewBadRequestError("failed_update", err)
		}

		return e.JSON(http.StatusOK, map[string]any{"status": "accepted"})
	}
}

// AdminCancelRegistrationHandler marks a registration as cancelled.
func AdminCancelRegistrationHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if _, err := requireAdmin(e); err != nil {
			return err
		}

		regId := e.Request.PathValue("id")
		if regId == "" {
			return apis.NewBadRequestError("invalid_registration", nil)
		}

		record, err := app.FindRecordById("event_registrations", regId)
		if err != nil || record == nil {
			return apis.NewNotFoundError("invalid_registration", err)
		}

		if record.GetString("status") == "cancelled" {
			return e.JSON(http.StatusOK, map[string]any{"status": "already_cancelled"})
		}

		record.Set("status", "cancelled")
		if err := app.Save(record); err != nil {
			return apis.NewBadRequestError("failed_update", err)
		}

		return e.JSON(http.StatusOK, map[string]any{"status": "cancelled"})
	}
}

func mapRegistration(record *core.Record) adminRegistrationItem {
	data := parseJSONMap(record.Get("data"))

	userId := record.GetString("user")

	return adminRegistrationItem{
		ID:      record.Id,
		Email:   record.GetString("email"),
		Status:  record.GetString("status"),
		Created: record.GetString("created"),
		HasUser: userId != "",
		Data:    data,
	}
}
