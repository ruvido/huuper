package retreats

import (
	"fmt"
	"time"

	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase"
)

// GenerateAcceptToken mirrors events.GenerateAcceptToken: a random,
// collection-unique token stored alongside a registration for potential
// future secure-link flows (kept for schema parity with the events module).
func GenerateAcceptToken(app *pocketbase.PocketBase) (string, error) {
	const attempts = 5
	for i := 0; i < attempts; i++ {
		token := backendinternal.RandomToken()
		if token == "" {
			continue
		}
		unique, err := isTokenUnique(app, token)
		if err != nil {
			return "", err
		}
		if unique {
			return token, nil
		}
	}
	return "", fmt.Errorf("unable to generate unique accept token")
}

// AcceptTokenExpiry gives the token a generous, fixed validity window.
// Retreats don't recur and registration windows are open-ended admin
// decisions, so unlike events there's no single "occurrence date" to expire
// against.
func AcceptTokenExpiry() time.Time {
	return time.Now().UTC().AddDate(0, 0, 30)
}

func isTokenUnique(app *pocketbase.PocketBase, token string) (bool, error) {
	records, err := app.FindRecordsByFilter(
		"retreat_registrations",
		"accept_token = {:token}",
		"",
		1,
		0,
		map[string]any{"token": token},
	)
	if err != nil {
		return false, err
	}
	return len(records) == 0, nil
}
