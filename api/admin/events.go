package admin

import (
	"net/http"
	"strings"

	backendinternal "members/internal"
	eventinternal "members/internal/events"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

type registrationItem struct {
	ID      string         `json:"id"`
	Email   string         `json:"email"`
	Status  string         `json:"status"`
	Created string         `json:"created"`
	HasUser bool           `json:"hasUser"`
	Data    map[string]any `json:"data"`
}

type registrationNotePayload struct {
	Note string `json:"note"`
}

func EventDetailsHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if _, err := backendinternal.RequireAdmin(e); err != nil {
			return err
		}

		eventID := e.Request.PathValue("id")
		if eventID == "" {
			return apis.NewBadRequestError("invalid_event", nil)
		}

		event, err := app.FindRecordById("events", eventID)
		if err != nil || event == nil {
			return apis.NewNotFoundError("invalid_event", err)
		}

		registrations, err := app.FindRecordsByFilter(
			"event_registrations",
			"event = {:event}",
			"created",
			0,
			0,
			map[string]any{"event": eventID},
		)
		if err != nil {
			return apis.NewBadRequestError("failed_registrations", err)
		}

		items := make([]registrationItem, 0, len(registrations))
		for _, record := range registrations {
			items = append(items, mapRegistration(app, record))
		}

		eventData := backendinternal.ParseJSONMap(event.Get("data"))

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

func ApproveRegistrationHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if _, err := backendinternal.RequireAdmin(e); err != nil {
			return err
		}

		regID := e.Request.PathValue("id")
		if regID == "" {
			return apis.NewBadRequestError("invalid_registration", nil)
		}

		record, err := app.FindRecordById("event_registrations", regID)
		if err != nil || record == nil {
			return apis.NewNotFoundError("invalid_registration", err)
		}

		if record.GetString("status") == "active" {
			return e.JSON(http.StatusOK, map[string]any{"status": "already_accepted"})
		}

		if err := eventinternal.ActivateRegistration(app, record, "events.user.registration_accepted"); err != nil {
			return apis.NewBadRequestError("failed_update", err)
		}

		return e.JSON(http.StatusOK, map[string]any{"status": "accepted"})
	}
}

func CancelRegistrationHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if _, err := backendinternal.RequireAdmin(e); err != nil {
			return err
		}

		regID := e.Request.PathValue("id")
		if regID == "" {
			return apis.NewBadRequestError("invalid_registration", nil)
		}

		var payload registrationNotePayload
		if err := e.BindBody(&payload); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}
		note := strings.TrimSpace(payload.Note)
		if note == "" {
			return apis.NewBadRequestError("invalid_cancelled_note", nil)
		}

		record, err := app.FindRecordById("event_registrations", regID)
		if err != nil || record == nil {
			return apis.NewNotFoundError("invalid_registration", err)
		}

		if record.GetString("status") == "cancelled" {
			return e.JSON(http.StatusOK, map[string]any{"status": "already_cancelled"})
		}

		data := backendinternal.ParseJSONMap(record.Get("data"))
		data["cancelled"] = note
		record.Set("data", data)
		record.Set("status", "cancelled")
		if err := app.Save(record); err != nil {
			return apis.NewBadRequestError("failed_update", err)
		}

		return e.JSON(http.StatusOK, map[string]any{"status": "cancelled"})
	}
}

func RejectRegistrationHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if _, err := backendinternal.RequireAdmin(e); err != nil {
			return err
		}

		regID := e.Request.PathValue("id")
		if regID == "" {
			return apis.NewBadRequestError("invalid_registration", nil)
		}

		var payload registrationNotePayload
		if err := e.BindBody(&payload); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}

		note := strings.TrimSpace(payload.Note)
		if note == "" {
			return apis.NewBadRequestError("invalid_rejected_note", nil)
		}

		record, err := app.FindRecordById("event_registrations", regID)
		if err != nil || record == nil {
			return apis.NewNotFoundError("invalid_registration", err)
		}

		data := backendinternal.ParseJSONMap(record.Get("data"))
		data["rejected"] = note
		record.Set("data", data)
		record.Set("status", "rejected")
		if err := app.Save(record); err != nil {
			return apis.NewBadRequestError("failed_update", err)
		}

		return e.JSON(http.StatusOK, map[string]any{"status": "rejected"})
	}
}

func mapRegistration(app *pocketbase.PocketBase, record *core.Record) registrationItem {
	data := backendinternal.ParseJSONMap(record.Get("data"))
	userID := strings.TrimSpace(record.GetString("user"))
	if userID != "" {
		if user, err := app.FindRecordById("users", userID); err == nil && user != nil {
			data = backendinternal.ParseJSONMap(user.Get("data"))
		}
	}

	return registrationItem{
		ID:      record.Id,
		Email:   record.GetString("email"),
		Status:  record.GetString("status"),
		Created: record.GetString("created"),
		HasUser: userID != "",
		Data:    data,
	}
}
