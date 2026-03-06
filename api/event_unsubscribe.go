package api

import (
	"net/http"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// EventUnsubscribeHandler removes a registration for the authenticated user.
func EventUnsubscribeHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
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

		var record *core.Record
		if authRecord != nil {
			record, err = findEventRegistrationByUser(app, event.Id, authRecord.Id, false)
			if err == nil && record != nil {
				if err := app.Delete(record); err != nil {
					return apis.NewBadRequestError("error_generic", err)
				}
				return e.JSON(http.StatusOK, map[string]any{"unsubscribed": true})
			}
		}

		email, err := normalizeEmail(authRecord.GetString("email"))
		if err != nil {
			return apis.NewBadRequestError("invalid_email", nil)
		}

		record, err = findEventRegistrationByEmail(app, event.Id, email, false)
		if err != nil || record == nil {
			return e.JSON(http.StatusOK, map[string]any{"unsubscribed": false})
		}

		if err := app.Delete(record); err != nil {
			return apis.NewBadRequestError("error_generic", err)
		}

		return e.JSON(http.StatusOK, map[string]any{"unsubscribed": true})
	}
}
