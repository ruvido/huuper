package retreats

import (
	"fmt"
	"strings"
	"time"

	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
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

// AcceptRequestView is what the organiser sees before deciding: the request
// as the guest filled it in, plus which retreat it is for. Built here, next to
// the token that unlocks it, so the public handler carries no knowledge of
// where these fields live.
func AcceptRequestView(app *pocketbase.PocketBase, registration *core.Record) map[string]any {
	data := backendinternal.ParseJSONMap(registration.Get("data"))
	field := func(key string) string {
		return strings.TrimSpace(backendinternal.AnyToString(data[key]))
	}

	retreatView := map[string]any{}
	if retreat, err := app.FindRecordById("retreats", registration.GetString("retreat")); err == nil && retreat != nil {
		retreatView = map[string]any{
			"title": strings.TrimSpace(retreat.GetString("title")),
			"slug":  strings.TrimSpace(retreat.GetString("slug")),
			"dates": formatDateRange(retreat.GetDateTime("start_date").Time(), retreat.GetDateTime("end_date").Time()),
		}
	}

	previous := ""
	if at, note := PreviousRequest(registration); !at.IsZero() {
		previous = formatDay(at)
		if note != "" {
			previous += " · " + note
		}
	}

	// For a request put aside, the note the organiser left: the page shows it
	// so a tap on an old email explains itself instead of just refusing.
	status := registration.GetString("status")
	note := ""
	if IsClosedStatus(status) {
		note = field(status)
	}

	return map[string]any{
		"status":   status,
		"note":     note,
		"retreat":  retreatView,
		"previous": previous,
		"registration": map[string]any{
			"email":          strings.TrimSpace(registration.GetString("email")),
			"full_name":      field("full_name"),
			"mobile":         field("mobile"),
			"birth_year":     field("birth_year"),
			"marital_status": field("marital_status"),
			"provenance":     field("provenance"),
			"received_on":    formatDay(registration.GetDateTime("created").Time()),
		},
	}
}

// FindByAcceptToken resolves a still-valid accept token to its registration.
// An expired or unknown token yields no record rather than an error, so the
// caller can answer with one indistinguishable "invalid link" either way.
func FindByAcceptToken(app *pocketbase.PocketBase, token string) *core.Record {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil
	}
	records, err := app.FindRecordsByFilter(
		"retreat_registrations",
		"accept_token = {:token}",
		"",
		1,
		0,
		map[string]any{"token": token},
	)
	if err != nil || len(records) == 0 {
		return nil
	}
	record := records[0]
	expiry := record.GetDateTime("accept_expires_at").Time()
	if !expiry.IsZero() && time.Now().UTC().After(expiry) {
		return nil
	}
	return record
}
