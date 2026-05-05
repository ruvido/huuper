package me

import (
	"net/http"
	"strings"

	backendinternal "members/backend/internal"
	eventinternal "members/backend/internal/events"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func EventStatusHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		authRecord, err := backendinternal.RequireAuthenticatedActor(e)
		if err != nil {
			return err
		}

		slug := e.Request.PathValue("slug")
		if slug == "" {
			return apis.NewBadRequestError("invalid_event", nil)
		}

		event, err := eventinternal.FindBySlug(app, slug)
		if err != nil || event == nil {
			return apis.NewNotFoundError("invalid_event", err)
		}

		if authRecord != nil {
			registration, err := eventinternal.FindRegistrationByUser(app, event.Id, authRecord.Id, true)
			if err == nil && registration != nil {
				return e.JSON(http.StatusOK, map[string]any{"registered": true})
			}
		}

		email, err := backendinternal.NormalizeEmail(authRecord.GetString("email"))
		if err != nil {
			return apis.NewBadRequestError("invalid_email", nil)
		}

		registration, err := eventinternal.FindRegistrationByEmail(app, event.Id, email, true)
		if err != nil || registration == nil {
			return e.JSON(http.StatusOK, map[string]any{"registered": false})
		}

		return e.JSON(http.StatusOK, map[string]any{"registered": true})
	}
}

func EventGetHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		authRecord, err := backendinternal.RequireAuthenticatedActor(e)
		if err != nil {
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
		canView, err := eventinternal.CanView(app, authRecord, event)
		if err != nil {
			return apis.NewBadRequestError("failed_event_visibility", err)
		}
		if !canView {
			return apis.NewForbiddenError("forbidden_event", nil)
		}

		registered := false
		attendees, err := eventinternal.ActiveAttendeesForEvent(app, event.Id)
		if err != nil {
			return apis.NewBadRequestError("failed_registrations", err)
		}
		if authRecord != nil {
			registration, err := eventinternal.FindRegistrationByUser(app, event.Id, authRecord.Id, true)
			if err == nil && registration != nil {
				registered = true
			} else {
				email, normalizeErr := backendinternal.NormalizeEmail(authRecord.GetString("email"))
				if normalizeErr == nil {
					registration, err = eventinternal.FindRegistrationByEmail(app, event.Id, email, true)
					registered = err == nil && registration != nil
				}
			}
		}

		item := eventinternal.MapItem(event)
		if cfg, err := eventinternal.LoadConfig(app); err == nil {
			eventinternal.ApplyTypeConfig(&item, cfg)
		}
		occurrences, _ := eventinternal.OccurrencesFor(event)
		return e.JSON(http.StatusOK, map[string]any{
			"event":                   item,
			"occurrences":             occurrences,
			"registered":              registered,
			"registrations":           attendees,
			"pending_registrations":   []any{},
			"cancelled_registrations": []any{},
		})
	}
}

func EventUnsubscribeHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		authRecord, err := backendinternal.RequireAuthenticatedActor(e)
		if err != nil {
			return err
		}

		slug := e.Request.PathValue("slug")
		if slug == "" {
			return apis.NewBadRequestError("invalid_event", nil)
		}

		event, err := eventinternal.FindBySlug(app, slug)
		if err != nil || event == nil {
			return apis.NewNotFoundError("invalid_event", err)
		}

		var record *core.Record
		if authRecord != nil {
			record, err = eventinternal.FindRegistrationByUser(app, event.Id, authRecord.Id, false)
			if err == nil && record != nil {
				if err := app.Delete(record); err != nil {
					return apis.NewBadRequestError("error_generic", err)
				}
				return e.JSON(http.StatusOK, map[string]any{"unsubscribed": true})
			}
		}

		email, err := backendinternal.NormalizeEmail(authRecord.GetString("email"))
		if err != nil {
			return apis.NewBadRequestError("invalid_email", nil)
		}

		record, err = eventinternal.FindRegistrationByEmail(app, event.Id, email, false)
		if err != nil || record == nil {
			return e.JSON(http.StatusOK, map[string]any{"unsubscribed": false})
		}

		if err := app.Delete(record); err != nil {
			return apis.NewBadRequestError("error_generic", err)
		}

		return e.JSON(http.StatusOK, map[string]any{"unsubscribed": true})
	}
}

func ListEventsHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor, err := backendinternal.RequireAuthenticatedActor(e)
		if err != nil {
			return err
		}
		window := strings.TrimSpace(e.Request.URL.Query().Get("window"))
		if window == "" {
			window = eventinternal.WindowFuture
		}
		items, err := eventinternal.ListForUser(app, actor, window)
		if err != nil {
			return apis.NewBadRequestError("failed_events", err)
		}
		return e.JSON(http.StatusOK, map[string]any{"items": items})
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
		var group *core.Record
		if groupID := strings.TrimSpace(input.Group); groupID != "" {
			group, err = app.FindRecordById("groups", groupID)
			if err != nil || group == nil {
				return apis.NewBadRequestError("invalid_group", err)
			}
		}
		if !eventinternal.CanCreate(app, actor, group, input.Type) {
			return apis.NewForbiddenError("forbidden_event_create", nil)
		}
		record, err := eventinternal.Create(app, actor, input)
		if err != nil {
			return apis.NewBadRequestError("failed_event_create", err)
		}
		item := eventinternal.MapItem(record)
		if cfg, err := eventinternal.LoadConfig(app); err == nil {
			eventinternal.ApplyTypeConfig(&item, cfg)
		}
		return e.JSON(http.StatusCreated, item)
	}
}

func UpdateEventHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor, err := backendinternal.RequireAuthenticatedActor(e)
		if err != nil {
			return err
		}
		record, err := loadEvent(app, e.Request.PathValue("id"))
		if err != nil {
			return err
		}
		ok, err := eventinternal.CanEdit(app, actor, record)
		if err != nil {
			return apis.NewBadRequestError("failed_event_edit_check", err)
		}
		if !ok {
			return apis.NewForbiddenError("forbidden_event_edit", nil)
		}
		var input eventinternal.UpdateInput
		if err := e.BindBody(&input); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}
		if err := eventinternal.Update(app, record, input); err != nil {
			return apis.NewBadRequestError("failed_event_update", err)
		}
		item := eventinternal.MapItem(record)
		if cfg, err := eventinternal.LoadConfig(app); err == nil {
			eventinternal.ApplyTypeConfig(&item, cfg)
		}
		return e.JSON(http.StatusOK, item)
	}
}

func RescheduleEventHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor, err := backendinternal.RequireAuthenticatedActor(e)
		if err != nil {
			return err
		}
		record, err := loadEvent(app, e.Request.PathValue("id"))
		if err != nil {
			return err
		}
		ok, err := eventinternal.CanEdit(app, actor, record)
		if err != nil {
			return apis.NewBadRequestError("failed_event_edit_check", err)
		}
		if !ok {
			return apis.NewForbiddenError("forbidden_event_edit", nil)
		}
		var payload struct {
			EventDate string `json:"event_date"`
			StartDate string `json:"start_date"`
		}
		if err := e.BindBody(&payload); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}
		newDate := strings.TrimSpace(payload.StartDate)
		if newDate == "" {
			newDate = payload.EventDate
		}
		if err := eventinternal.Reschedule(app, record, newDate); err != nil {
			return apis.NewBadRequestError("failed_event_reschedule", err)
		}
		item := eventinternal.MapItem(record)
		if cfg, err := eventinternal.LoadConfig(app); err == nil {
			eventinternal.ApplyTypeConfig(&item, cfg)
		}
		return e.JSON(http.StatusOK, item)
	}
}

func CancelEventHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor, err := backendinternal.RequireAuthenticatedActor(e)
		if err != nil {
			return err
		}
		record, err := loadEvent(app, e.Request.PathValue("id"))
		if err != nil {
			return err
		}
		ok, err := eventinternal.CanEdit(app, actor, record)
		if err != nil {
			return apis.NewBadRequestError("failed_event_edit_check", err)
		}
		if !ok {
			return apis.NewForbiddenError("forbidden_event_cancel", nil)
		}
		if err := eventinternal.Cancel(app, record); err != nil {
			return apis.NewBadRequestError("failed_event_cancel", err)
		}
		return e.JSON(http.StatusOK, map[string]any{"deleted": 1})
	}
}

func CancelOccurrenceHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor, err := backendinternal.RequireAuthenticatedActor(e)
		if err != nil {
			return err
		}
		record, err := loadEvent(app, e.Request.PathValue("id"))
		if err != nil {
			return err
		}
		ok, err := eventinternal.CanEdit(app, actor, record)
		if err != nil {
			return apis.NewBadRequestError("failed_event_edit_check", err)
		}
		if !ok {
			return apis.NewForbiddenError("forbidden_event_cancel", nil)
		}
		var payload struct {
			Date string `json:"date"`
		}
		if err := e.BindBody(&payload); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}
		if err := eventinternal.CancelOccurrence(app, record, payload.Date); err != nil {
			return apis.NewBadRequestError("failed_occurrence_cancel", err)
		}
		occurrences, _ := eventinternal.OccurrencesFor(record)
		return e.JSON(http.StatusOK, map[string]any{"occurrences": occurrences})
	}
}

func RegisterEventHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor, err := backendinternal.RequireAuthenticatedActor(e)
		if err != nil {
			return err
		}
		event, err := loadEvent(app, e.Request.PathValue("id"))
		if err != nil {
			return err
		}
		cfg, err := eventinternal.LoadConfig(app)
		if err != nil {
			return apis.NewBadRequestError("failed_event_config", err)
		}
		typeDef, ok := cfg.TypeDef(event.GetString("type"))
		if !ok || !typeDef.Registration.Enabled {
			return apis.NewBadRequestError("registration_not_allowed", nil)
		}
		canView, err := eventinternal.CanView(app, actor, event)
		if err != nil {
			return apis.NewBadRequestError("failed_event_visibility", err)
		}
		if !canView {
			return apis.NewForbiddenError("forbidden_event", nil)
		}

		registered, err := registerForEvents(app, actor, []*core.Record{event})
		if err != nil {
			return apis.NewBadRequestError("failed_event_register", err)
		}
		return e.JSON(http.StatusOK, map[string]any{"registered": registered})
	}
}

func UnregisterEventHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor, err := backendinternal.RequireAuthenticatedActor(e)
		if err != nil {
			return err
		}
		event, err := loadEvent(app, e.Request.PathValue("id"))
		if err != nil {
			return err
		}
		removed := 0
		registration, _ := eventinternal.FindRegistrationByUser(app, event.Id, actor.Id, false)
		if registration != nil {
			if err := app.Delete(registration); err != nil {
				return apis.NewBadRequestError("failed_event_unregister", err)
			}
			removed = 1
		}
		return e.JSON(http.StatusOK, map[string]any{"unregistered": removed})
	}
}

func MarkAttendanceHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor, err := backendinternal.RequireAuthenticatedActor(e)
		if err != nil {
			return err
		}
		regID := strings.TrimSpace(e.Request.PathValue("id"))
		if regID == "" {
			return apis.NewBadRequestError("invalid_registration", nil)
		}
		registration, err := app.FindRecordById("event_registrations", regID)
		if err != nil || registration == nil {
			return apis.NewNotFoundError("registration_not_found", err)
		}
		event, err := app.FindRecordById("events", registration.GetString("event"))
		if err != nil || event == nil {
			return apis.NewNotFoundError("event_not_found", err)
		}
		ok, err := eventinternal.CanEdit(app, actor, event)
		if err != nil {
			return apis.NewBadRequestError("failed_event_edit_check", err)
		}
		if !ok {
			return apis.NewForbiddenError("forbidden_attendance", nil)
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

func loadEvent(app *pocketbase.PocketBase, id string) (*core.Record, error) {
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

func registerForEvents(app *pocketbase.PocketBase, actor *core.Record, events []*core.Record) (int, error) {
	collection, err := app.FindCollectionByNameOrId("event_registrations")
	if err != nil {
		return 0, err
	}
	registered := 0
	for _, event := range events {
		existing, _ := eventinternal.FindRegistrationByUser(app, event.Id, actor.Id, false)
		if existing != nil {
			if existing.GetString("status") != "active" {
				existing.Set("status", "active")
				if err := app.Save(existing); err != nil {
					return registered, err
				}
			}
			registered++
			continue
		}
		record := core.NewRecord(collection)
		record.Set("event", event.Id)
		record.Set("user", actor.Id)
		record.Set("email", strings.TrimSpace(actor.GetString("email")))
		record.Set("status", "active")
		record.Set("data", map[string]any{})
		if err := app.Save(record); err != nil {
			return registered, err
		}
		registered++
	}
	return registered, nil
}
