package api

import (
	"net/http"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// EventStatusHandler returns whether the authenticated user is registered for the event.
func EventStatusHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		authRecord := e.Auth
		if authRecord == nil {
			return apis.NewUnauthorizedError("Unauthorized", nil)
		}

		slug := e.Request.PathValue("slug")
		if slug == "" {
			return apis.NewBadRequestError("invalid_event", nil)
		}

		event, err := app.FindFirstRecordByFilter(
			"events",
			"slug = {:slug}",
			map[string]any{"slug": slug},
		)
		if err != nil || event == nil {
			return apis.NewNotFoundError("invalid_event", err)
		}

		if authRecord != nil {
			registration, err := app.FindFirstRecordByFilter(
				"event_registrations",
				"event = {:event} && user = {:user}",
				map[string]any{
					"event": event.Id,
					"user":  authRecord.Id,
				},
			)
			if err == nil && registration != nil {
				return e.JSON(http.StatusOK, map[string]any{"registered": true})
			}
		}

		email, err := normalizeEmail(authRecord.GetString("email"))
		if err != nil {
			return apis.NewBadRequestError("invalid_email", nil)
		}

		registration, err := app.FindFirstRecordByFilter(
			"event_registrations",
			"event = {:event} && email = {:email}",
			map[string]any{
				"event": event.Id,
				"email": email,
			},
		)
		if err != nil || registration == nil {
			return e.JSON(http.StatusOK, map[string]any{"registered": false})
		}

		return e.JSON(http.StatusOK, map[string]any{"registered": true})
	}
}
