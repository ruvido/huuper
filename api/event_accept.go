package api

import (
	"net/http"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// AcceptEventHandler marks a registration as accepted by token.
func AcceptEventHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		token := e.Request.URL.Query().Get("token")
		if token == "" {
			return apis.NewBadRequestError("Token mancante", nil)
		}

		record, err := app.FindFirstRecordByFilter(
			"event_registrations",
			"accept_token = {:token}",
			map[string]any{"token": token},
		)
	if err != nil || record == nil {
		return apis.NewNotFoundError("Registrazione non trovata", err)
	}

	expiresAt := record.GetDateTime("accept_expires_at")
	if !expiresAt.IsZero() && time.Now().After(expiresAt.Time()) {
		return apis.NewBadRequestError("token_expired", nil)
	}

	if record.GetBool("accepted") {
		return e.JSON(http.StatusOK, map[string]any{"status": "already_accepted"})
	}

		record.Set("accepted", true)
		if err := app.Save(record); err != nil {
			return apis.NewBadRequestError("Aggiornamento fallito", err)
		}

		return e.JSON(http.StatusOK, map[string]any{"status": "accepted"})
	}
}
