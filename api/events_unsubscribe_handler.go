package api

import (
	"net/http"

	backendinternal "members/internal"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// EventUnsubscribeHandler removes a registration for the authenticated user.
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

		event, err := backendinternal.FindEventBySlug(app, slug)
		if err != nil || event == nil {
			return apis.NewNotFoundError("invalid_event", err)
		}

		var record *core.Record
		if authRecord != nil {
			record, err = backendinternal.FindEventRegistrationByUser(app, event.Id, authRecord.Id, false)
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

		record, err = backendinternal.FindEventRegistrationByEmail(app, event.Id, email, false)
		if err != nil || record == nil {
			return e.JSON(http.StatusOK, map[string]any{"unsubscribed": false})
		}

		if err := app.Delete(record); err != nil {
			return apis.NewBadRequestError("error_generic", err)
		}

		return e.JSON(http.StatusOK, map[string]any{"unsubscribed": true})
	}
}
