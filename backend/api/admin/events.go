package admin

import (
	"net/http"
	"strings"

	backendinternal "members/backend/internal"
	eventinternal "members/backend/internal/events"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

type registrationNotePayload struct {
	Note string `json:"note"`
}

func EventDetailsHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		eventID := e.Request.PathValue("id")
		if eventID == "" {
			return apis.NewBadRequestError("invalid_event", nil)
		}

		event, err := app.FindRecordById("events", eventID)
		if err != nil || event == nil {
			return apis.NewNotFoundError("invalid_event", err)
		}

		attendees, err := eventinternal.ActiveAttendeesForEvent(app, eventID)
		if err != nil {
			return apis.NewBadRequestError("failed_registrations", err)
		}
		pendingAttendees, err := eventinternal.PendingAttendeesForEvent(app, eventID)
		if err != nil {
			return apis.NewBadRequestError("failed_registrations", err)
		}
		cancelledAttendees, err := eventinternal.CancelledAttendeesForEvent(app, eventID)
		if err != nil {
			return apis.NewBadRequestError("failed_registrations", err)
		}

		eventData := backendinternal.ParseJSONMap(event.Get("data"))

		return e.JSON(http.StatusOK, map[string]any{
			"event": map[string]any{
				"id":         event.Id,
				"type":       event.GetString("type"),
				"slug":       event.GetString("slug"),
				"title":      event.GetString("title"),
				"event_date": event.GetString("event_date"),
				"url":        event.GetString("url"),
				"group":      event.GetString("group"),
				"series":     event.GetString("series"),
				"data":       eventData,
			},
			"registrations":           attendees,
			"pending_registrations":   pendingAttendees,
			"cancelled_registrations": cancelledAttendees,
		})
	}
}

func ApproveRegistrationHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
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

func CreateEventHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor, err := backendinternal.RequireAuthenticatedActor(e)
		if err != nil {
			return err
		}
		var input eventinternal.CreateInput
		if err := e.BindBody(&input); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}
		if !eventinternal.IsValidType(input.Type) {
			return apis.NewBadRequestError("invalid_event_type", nil)
		}
		records, err := eventinternal.Create(app, actor, input)
		if err != nil {
			return apis.NewBadRequestError("failed_event_create", err)
		}
		out := make([]eventinternal.Item, 0, len(records))
		for _, record := range records {
			out = append(out, eventinternal.MapItem(record))
		}
		return e.JSON(http.StatusCreated, map[string]any{"items": out})
	}
}

func UpdateEventHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		record, err := loadEventByID(app, e.Request.PathValue("id"))
		if err != nil {
			return err
		}
		var input eventinternal.UpdateInput
		if err := e.BindBody(&input); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}
		if err := eventinternal.Update(app, record, input); err != nil {
			return apis.NewBadRequestError("failed_event_update", err)
		}
		return e.JSON(http.StatusOK, eventinternal.MapItem(record))
	}
}

func RescheduleEventHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		record, err := loadEventByID(app, e.Request.PathValue("id"))
		if err != nil {
			return err
		}
		var payload struct {
			EventDate string `json:"event_date"`
		}
		if err := e.BindBody(&payload); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}
		if err := eventinternal.Reschedule(app, record, payload.EventDate); err != nil {
			return apis.NewBadRequestError("failed_event_reschedule", err)
		}
		return e.JSON(http.StatusOK, eventinternal.MapItem(record))
	}
}

func CancelEventHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		record, err := loadEventByID(app, e.Request.PathValue("id"))
		if err != nil {
			return err
		}
		var payload struct {
			Scope string `json:"scope"`
		}
		if err := e.BindBody(&payload); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}
		scope := strings.TrimSpace(payload.Scope)
		if scope == "" {
			scope = eventinternal.CancelScopeThis
		}
		count, err := eventinternal.Cancel(app, record, scope)
		if err != nil {
			return apis.NewBadRequestError("failed_event_cancel", err)
		}
		return e.JSON(http.StatusOK, map[string]any{"deleted": count})
	}
}

func MarkAttendanceHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		regID := strings.TrimSpace(e.Request.PathValue("id"))
		if regID == "" {
			return apis.NewBadRequestError("invalid_registration", nil)
		}
		registration, err := app.FindRecordById("event_registrations", regID)
		if err != nil || registration == nil {
			return apis.NewNotFoundError("registration_not_found", err)
		}
		var payload struct {
			Attended *bool `json:"attended"`
		}
		if err := e.BindBody(&payload); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}
		if payload.Attended == nil {
			if err := eventinternal.ClearAttendance(app, registration); err != nil {
				return apis.NewBadRequestError("failed_attendance_clear", err)
			}
		} else {
			if err := eventinternal.MarkAttendance(app, registration, *payload.Attended); err != nil {
				return apis.NewBadRequestError("failed_attendance", err)
			}
		}
		return e.JSON(http.StatusOK, map[string]any{"ok": true})
	}
}

func loadEventByID(app *pocketbase.PocketBase, id string) (*core.Record, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, apis.NewBadRequestError("invalid_event", nil)
	}
	record, err := app.FindRecordById("events", id)
	if err != nil || record == nil {
		return nil, apis.NewNotFoundError("event_not_found", err)
	}
	return record, nil
}
