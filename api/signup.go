package api

import (
	"net/http"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

type signupEmailPayload struct {
	Email string `json:"email"`
}

// CheckSignupEmailHandler verifies if a signup request already exists for the email.
func CheckSignupEmailHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		var payload signupEmailPayload
		if err := e.BindBody(&payload); err != nil {
			return apis.NewBadRequestError("Invalid request", err)
		}

		email := strings.TrimSpace(strings.ToLower(payload.Email))
		if email == "" {
			return apis.NewBadRequestError("Missing email", nil)
		}

		records, err := app.FindRecordsByFilter(
			"requests",
			"email = {:email}",
			"",
			1,
			0,
			map[string]any{"email": email},
		)
		if err != nil {
			return apis.NewBadRequestError("Failed to check email", err)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"unique": len(records) == 0,
		})
	}
}
