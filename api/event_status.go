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

		event, err := findEventBySlug(app, slug)
		if err != nil || event == nil {
			return apis.NewNotFoundError("invalid_event", err)
		}

		if authRecord != nil {
			registration, err := findEventRegistrationByUser(app, event.Id, authRecord.Id, true)
			if err == nil && registration != nil {
				return e.JSON(http.StatusOK, map[string]any{"registered": true})
			}
		}

		email, err := normalizeEmail(authRecord.GetString("email"))
		if err != nil {
			return apis.NewBadRequestError("invalid_email", nil)
		}

		registration, err := findEventRegistrationByEmail(app, event.Id, email, true)
		if err != nil || registration == nil {
			return e.JSON(http.StatusOK, map[string]any{"registered": false})
		}

		return e.JSON(http.StatusOK, map[string]any{"registered": true})
	}
}
