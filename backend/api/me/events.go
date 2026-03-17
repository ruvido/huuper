package me

import (
	"net/http"

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

		return e.JSON(http.StatusOK, map[string]any{
			"event": map[string]any{
				"id":         event.Id,
				"slug":       event.GetString("slug"),
				"title":      event.GetString("title"),
				"event_date": event.GetString("event_date"),
				"data":       backendinternal.ParseJSONMap(event.Get("data")),
			},
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
