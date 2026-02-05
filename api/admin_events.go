package api

import (
	"net/http"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

type adminRegistrationItem struct {
	ID       string         `json:"id"`
	Email    string         `json:"email"`
	Accepted bool           `json:"accepted"`
	Created  string         `json:"created"`
	HasUser  bool           `json:"hasUser"`
	Data     map[string]any `json:"data"`
}

// AdminEventDetailsHandler returns the event and its registrations.
func AdminEventDetailsHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		authRecord := e.Auth
		if authRecord == nil {
			return apis.NewUnauthorizedError("Unauthorized", nil)
		}
		if !authRecord.GetBool("admin") {
			return apis.NewForbiddenError("Forbidden", nil)
		}

		eventId := e.Request.PathValue("id")
		if eventId == "" {
			return apis.NewBadRequestError("invalid_event", nil)
		}

		event, err := app.FindRecordById("events", eventId)
		if err != nil || event == nil {
			return apis.NewNotFoundError("invalid_event", err)
		}

		pending, err := app.FindRecordsByFilter(
			"event_registrations",
			"event = {:event} && accepted = false",
			"created",
			0,
			0,
			map[string]any{"event": eventId},
		)
		if err != nil {
			return apis.NewBadRequestError("failed_registrations", err)
		}

		approved, err := app.FindRecordsByFilter(
			"event_registrations",
			"event = {:event} && accepted = true",
			"created",
			0,
			0,
			map[string]any{"event": eventId},
		)
		if err != nil {
			return apis.NewBadRequestError("failed_registrations", err)
		}

		items := make([]adminRegistrationItem, 0, len(pending)+len(approved))
		for _, record := range pending {
			items = append(items, mapRegistration(record))
		}
		for _, record := range approved {
			items = append(items, mapRegistration(record))
		}

		return e.JSON(http.StatusOK, map[string]any{
			"event": map[string]any{
				"id":         event.Id,
				"title":      event.GetString("title"),
				"event_date": event.GetString("event_date"),
			},
			"registrations": items,
		})
	}
}

// AdminApproveRegistrationHandler marks a registration as accepted.
func AdminApproveRegistrationHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		authRecord := e.Auth
		if authRecord == nil {
			return apis.NewUnauthorizedError("Unauthorized", nil)
		}
		if !authRecord.GetBool("admin") {
			return apis.NewForbiddenError("Forbidden", nil)
		}

		regId := e.Request.PathValue("id")
		if regId == "" {
			return apis.NewBadRequestError("invalid_registration", nil)
		}

		record, err := app.FindRecordById("event_registrations", regId)
		if err != nil || record == nil {
			return apis.NewNotFoundError("invalid_registration", err)
		}

		if record.GetBool("accepted") {
			return e.JSON(http.StatusOK, map[string]any{"status": "already_accepted"})
		}

		record.Set("accepted", true)
		if err := app.Save(record); err != nil {
			return apis.NewBadRequestError("failed_update", err)
		}

		return e.JSON(http.StatusOK, map[string]any{"status": "accepted"})
	}
}

func mapRegistration(record *core.Record) adminRegistrationItem {
	data := map[string]any{}
	if raw := record.Get("data"); raw != nil {
		if typed, ok := raw.(map[string]any); ok {
			data = typed
		}
	}

	userId := record.GetString("user")

	return adminRegistrationItem{
		ID:       record.Id,
		Email:    record.GetString("email"),
		Accepted: record.GetBool("accepted"),
		Created:  record.GetString("created"),
		HasUser:  userId != "",
		Data:     data,
	}
}
